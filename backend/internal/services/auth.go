package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"nsh-guild-analytics/backend/internal/config"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLoginLimited       = errors.New("too many login attempts")
	ErrWeakPassword       = errors.New("password must be at least 10 characters")
)

type AuthService struct {
	cfg   config.Config
	pool  *pgxpool.Pool
	redis *redis.Client
}

type User struct {
	ID                  int64  `json:"id"`
	Username            string `json:"username"`
	ForcePasswordChange bool   `json:"force_password_change"`
}

type Session struct {
	ID           string
	User         User
	TokenHash    string
	CSRFHash     string
	CSRFToken    string
	ExpiresAt    time.Time
	NeedsRefresh bool
}

func NewAuthService(cfg config.Config, pool *pgxpool.Pool, redisClient *redis.Client) *AuthService {
	return &AuthService{cfg: cfg, pool: pool, redis: redisClient}
}

func (s *AuthService) BootstrapAdmin(ctx context.Context) error {
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM app_user`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	password, err := RandomPassword()
	if err != nil {
		return err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if _, err = s.pool.Exec(ctx, `
		INSERT INTO app_user(username, password_hash, is_admin, force_password_change)
		VALUES('admin', $1, TRUE, TRUE)
	`, hash); err != nil {
		return err
	}
	log.Printf("首次管理员账号已创建：username=admin initial_password=%s 该密码只在首次初始化日志中显示一次，登录后必须修改。", password)
	return nil
}

func (s *AuthService) Authenticate(ctx context.Context, username, password, ip string) (Session, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return Session{}, ErrInvalidCredentials
	}
	limited, err := s.loginLimited(ctx, username, ip)
	if err == nil && limited {
		return Session{}, ErrLoginLimited
	}

	var user User
	var passwordHash string
	err = s.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, force_password_change
		FROM app_user
		WHERE username=$1
	`, username).Scan(&user.ID, &user.Username, &passwordHash, &user.ForcePasswordChange)
	if err != nil {
		_ = s.recordLoginFailure(ctx, username, ip)
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrInvalidCredentials
		}
		return Session{}, err
	}
	ok, err := VerifyPassword(password, passwordHash)
	if err != nil || !ok {
		_ = s.recordLoginFailure(ctx, username, ip)
		return Session{}, ErrInvalidCredentials
	}
	_ = s.clearLoginFailures(ctx, username, ip)
	_, _ = s.pool.Exec(ctx, `UPDATE app_user SET last_login_at=now() WHERE id=$1`, user.ID)
	return s.createSession(ctx, user)
}

func (s *AuthService) createSession(ctx context.Context, user User) (Session, error) {
	token, err := RandomToken(32)
	if err != nil {
		return Session{}, err
	}
	csrf, err := RandomToken(32)
	if err != nil {
		return Session{}, err
	}
	sessionID := uuid.NewString()
	expires := time.Now().Add(s.cfg.SessionTTL)
	tokenHash := HashOpaque(token)
	csrfHash := HashOpaque(csrf)
	if _, err = s.pool.Exec(ctx, `
		INSERT INTO user_session(id, user_id, token_hash, csrf_token_hash, expires_at)
		VALUES($1, $2, $3, $4, $5)
	`, sessionID, user.ID, tokenHash, csrfHash, expires); err != nil {
		return Session{}, err
	}
	_ = s.redis.Set(ctx, "session:"+tokenHash, fmt.Sprintf("%d", user.ID), s.cfg.SessionTTL).Err()
	return Session{
		ID:        sessionID,
		User:      user,
		TokenHash: token,
		CSRFHash:  csrfHash,
		CSRFToken: csrf,
		ExpiresAt: expires,
	}, nil
}

func (s *AuthService) SessionFromToken(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, pgx.ErrNoRows
	}
	tokenHash := HashOpaque(token)
	var session Session
	err := s.pool.QueryRow(ctx, `
		SELECT us.id, us.token_hash, us.csrf_token_hash, us.expires_at,
		       u.id, u.username, u.force_password_change
		FROM user_session us
		JOIN app_user u ON u.id = us.user_id
		WHERE us.token_hash=$1 AND us.revoked_at IS NULL AND us.expires_at > now()
	`, tokenHash).Scan(
		&session.ID, &session.TokenHash, &session.CSRFHash, &session.ExpiresAt,
		&session.User.ID, &session.User.Username, &session.User.ForcePasswordChange,
	)
	if err != nil {
		return Session{}, err
	}
	_ = s.redis.Expire(ctx, "session:"+tokenHash, time.Until(session.ExpiresAt)).Err()
	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	tokenHash := HashOpaque(token)
	_, err := s.pool.Exec(ctx, `UPDATE user_session SET revoked_at=now() WHERE token_hash=$1 AND revoked_at IS NULL`, tokenHash)
	_ = s.redis.Del(ctx, "session:"+tokenHash).Err()
	return err
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, sessionID, currentPassword, newPassword string) error {
	if len([]rune(newPassword)) < 10 {
		return ErrWeakPassword
	}
	var currentHash string
	if err := s.pool.QueryRow(ctx, `SELECT password_hash FROM app_user WHERE id=$1`, userID).Scan(&currentHash); err != nil {
		return err
	}
	ok, err := VerifyPassword(currentPassword, currentHash)
	if err != nil || !ok {
		return ErrInvalidCredentials
	}
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `
		UPDATE app_user
		SET password_hash=$1, force_password_change=FALSE
		WHERE id=$2
	`, newHash, userID); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `
		UPDATE user_session
		SET revoked_at=now()
		WHERE user_id=$1 AND id<>$2 AND revoked_at IS NULL
		RETURNING token_hash
	`, userID, sessionID)
	if err != nil {
		return err
	}
	var revoked []string
	for rows.Next() {
		var tokenHash string
		if err = rows.Scan(&tokenHash); err != nil {
			rows.Close()
			return err
		}
		revoked = append(revoked, "session:"+tokenHash)
	}
	rows.Close()
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if len(revoked) > 0 {
		_ = s.redis.Del(ctx, revoked...).Err()
	}
	return nil
}

func (s *AuthService) loginLimited(ctx context.Context, username, ip string) (bool, error) {
	key := s.failureKey(username, ip)
	value, err := s.redis.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return value >= s.cfg.LoginFailureLimit, nil
}

func (s *AuthService) recordLoginFailure(ctx context.Context, username, ip string) error {
	key := s.failureKey(username, ip)
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		return s.redis.Expire(ctx, key, 15*time.Minute).Err()
	}
	return nil
}

func (s *AuthService) clearLoginFailures(ctx context.Context, username, ip string) error {
	return s.redis.Del(ctx, s.failureKey(username, ip)).Err()
}

func (s *AuthService) failureKey(username, ip string) string {
	return "login_fail:" + HashOpaque(strings.ToLower(username)+"|"+ip)
}
