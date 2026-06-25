package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"nsh-guild-analytics/backend/internal/config"
	"nsh-guild-analytics/backend/internal/domain"
	"nsh-guild-analytics/backend/internal/services"
)

const (
	sessionCookie = "nsh_session"
	csrfCookie    = "nsh_csrf"
	sessionKey    = "session"
)

type Server struct {
	cfg      config.Config
	pool     *pgxpool.Pool
	redis    *redis.Client
	auth     *services.AuthService
	importer *services.ImportService
	store    *services.Store
	router   *gin.Engine
}

func NewServer(cfg config.Config, pool *pgxpool.Pool, redisClient *redis.Client) *Server {
	gin.SetMode(cfg.AppEnv)
	server := &Server{
		cfg:      cfg,
		pool:     pool,
		redis:    redisClient,
		auth:     services.NewAuthService(cfg, pool, redisClient),
		importer: services.NewImportService(),
		store:    services.NewStore(cfg, pool),
	}
	_ = server.store.EnsureDefaultScoring(context.Background())
	server.router = server.buildRouter()
	return server
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) buildRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.MaxMultipartMemory = 64 << 20

	api := router.Group("/api")
	api.GET("/health", s.health)
	api.POST("/auth/login", s.login)
	api.Use(s.authMiddleware())
	api.POST("/auth/logout", s.requireCSRF(), s.logout)
	api.GET("/auth/me", s.me)
	api.POST("/auth/change-password", s.requireCSRF(), s.changePassword)
	api.GET("/system/version", s.versionInfo)
	api.POST("/system/update", s.requireCSRF(), s.applyUpdate)
	api.POST("/battles/import/preview", s.requireCSRF(), s.importPreview)
	api.POST("/battles/import/confirm", s.requireCSRF(), s.importConfirm)
	api.GET("/battles", s.listBattles)
	api.GET("/battles/:battle_id", s.getBattle)
	api.DELETE("/battles/:battle_id", s.requireCSRF(), s.deleteBattle)
	api.POST("/battles/:battle_id/reanalyze", s.requireCSRF(), s.reanalyzeBattle)
	api.GET("/battles/:battle_id/overview", s.overview)
	api.GET("/battles/:battle_id/rankings", s.rankings)
	api.GET("/battles/:battle_id/players/:stat_id", s.playerDetail)
	api.GET("/battles/:battle_id/team-top3", s.teamTop3)
	api.GET("/battles/:battle_id/guild-comparison", s.guildComparison)
	api.GET("/battles/:battle_id/squad-comparison", s.squadComparison)
	api.GET("/rankings/history", s.historyRankings)
	api.GET("/rankings/players/aggregate", s.historyRankings)
	api.GET("/settings", s.settings)
	api.PUT("/settings", s.requireCSRF(), s.putSettings)
	api.GET("/scoring-rules", s.scoringRules)
	api.POST("/scoring-rules", s.requireCSRF(), s.createScoringRule)
	api.POST("/scoring-rules/validate", s.requireCSRF(), s.validateScoringRules)
	api.POST("/scoring-rules/range-suggestions", s.requireCSRF(), s.rangeSuggestions)
	api.POST("/scoring-rules/:version/publish", s.requireCSRF(), s.publishScoringRule)
	api.GET("/scoring-ranges", s.scoringRanges)
	api.POST("/scoring-ranges", s.requireCSRF(), s.publishScoringRange)
	api.PUT("/players/:player_id/avatar", s.requireCSRF(), s.uploadPlayerAvatar)
	api.DELETE("/players/:player_id/avatar", s.requireCSRF(), s.deletePlayerAvatar)
	api.PUT("/careers/:career/avatar", s.requireCSRF(), s.uploadCareerAvatar)
	api.DELETE("/careers/:career/avatar", s.requireCSRF(), s.deleteCareerAvatar)
	api.POST("/backups", s.requireCSRF(), s.backup)

	router.Static("/assets", filepath.Join(s.cfg.StaticDir, "assets"))
	router.Static("/uploads", s.cfg.UploadDir)
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		index := filepath.Join(s.cfg.StaticDir, "index.html")
		if _, err := os.Stat(index); err == nil {
			c.File(index)
			return
		}
		c.JSON(http.StatusOK, gin.H{"app": s.cfg.AppName, "message": "frontend build not found"})
	})
	return router
}

func (s *Server) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	dbOK := s.pool.Ping(ctx) == nil
	redisOK := s.redis.Ping(ctx).Err() == nil
	status := http.StatusOK
	if !dbOK || !redisOK {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{
		"app":      s.cfg.AppName,
		"database": dbOK,
		"redis":    redisOK,
		"version":  s.cfg.AppVersion,
	})
}

func (s *Server) versionInfo(c *gin.Context) {
	c.JSON(http.StatusOK, services.CheckUpdate(c.Request.Context(), s.cfg))
}

func (s *Server) applyUpdate(c *gin.Context) {
	result := services.ApplyUpdate(c.Request.Context(), s.cfg)
	if result.Error != "" {
		c.JSON(http.StatusBadRequest, result)
		return
	}
	c.JSON(http.StatusAccepted, result)
}

func (s *Server) login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	session, err := s.auth.Authenticate(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		if errors.Is(err, services.ErrLoginLimited) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too_many_attempts", "message": "登录失败次数过多，请稍后再试。"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "账号或密码错误。"})
		return
	}
	s.setSessionCookies(c, session.TokenHash, session.CSRFToken, session.ExpiresAt)
	c.JSON(http.StatusOK, gin.H{"user": session.User, "csrf_token": session.CSRFToken})
}

func (s *Server) logout(c *gin.Context) {
	token, _ := c.Cookie(sessionCookie)
	_ = s.auth.Logout(c.Request.Context(), token)
	s.clearSessionCookies(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) me(c *gin.Context) {
	session := currentSession(c)
	c.JSON(http.StatusOK, gin.H{"user": session.User, "csrf_token": session.CSRFToken})
}

func (s *Server) changePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	session := currentSession(c)
	if err := s.auth.ChangePassword(c.Request.Context(), session.User.ID, session.ID, req.CurrentPassword, req.NewPassword); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": "change_password_failed", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) importPreview(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_file"})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "open_file_failed"})
		return
	}
	defer opened.Close()
	data, err := readAllLimited(opened, 32<<20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large_or_invalid", "message": err.Error()})
		return
	}
	preview, err := s.importer.ParsePreview(c.Request.Context(), file.Filename, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parse_failed", "message": err.Error()})
		return
	}
	if exists, err := s.store.SourceExists(c.Request.Context(), preview.SourceSHA256); err == nil && exists {
		preview.Errors = append(preview.Errors, domain.ImportMessage{Level: "error", Code: "duplicate_file", Message: "该 CSV 文件已经导入过，已按 SHA-256 阻止重复入库。"})
	}
	token, err := services.RandomToken(24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_failed"})
		return
	}
	preview.Token = token
	raw, _ := json.Marshal(preview)
	if err = s.redis.Set(c.Request.Context(), "import_preview:"+token, raw, s.cfg.PreviewTTL).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache_failed"})
		return
	}
	response := preview
	response.Rows = nil
	c.JSON(http.StatusOK, response)
}

func (s *Server) importConfirm(c *gin.Context) {
	var req struct {
		Token     string `json:"token"`
		HomeGuild string `json:"home_guild"`
		BattleAt  string `json:"battle_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	raw, err := s.redis.Get(c.Request.Context(), "import_preview:"+req.Token).Bytes()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preview_expired", "message": "导入预览已过期，请重新上传。"})
		return
	}
	var preview domain.ImportPreview
	if err = json.Unmarshal(raw, &preview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preview_invalid"})
		return
	}
	battleAt := time.Time{}
	if strings.TrimSpace(req.BattleAt) != "" {
		battleAt, _ = time.Parse(time.RFC3339, req.BattleAt)
	}
	session := currentSession(c)
	id, err := s.store.ConfirmImport(c.Request.Context(), preview, services.ConfirmImportRequest{HomeGuild: req.HomeGuild, BattleAt: battleAt, UserID: session.User.ID})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confirm_failed", "message": err.Error()})
		return
	}
	_ = s.redis.Del(c.Request.Context(), "import_preview:"+req.Token).Err()
	c.JSON(http.StatusOK, gin.H{"battle_id": id})
}

func (s *Server) listBattles(c *gin.Context) {
	items, err := s.store.ListBattles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) getBattle(c *gin.Context) {
	id, ok := pathID(c, "battle_id")
	if !ok {
		return
	}
	detail, err := s.store.BattleDetail(c.Request.Context(), id)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (s *Server) deleteBattle(c *gin.Context) {
	id, ok := pathID(c, "battle_id")
	if !ok {
		return
	}
	if err := s.store.DeleteBattle(c.Request.Context(), id); err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) reanalyzeBattle(c *gin.Context) {
	id, ok := pathID(c, "battle_id")
	if !ok {
		return
	}
	if err := s.store.ReanalyzeBattle(c.Request.Context(), id); err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) overview(c *gin.Context) {
	id, ok := s.battleIDOrLatest(c)
	if !ok {
		return
	}
	overview, err := s.store.Overview(c.Request.Context(), id)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (s *Server) rankings(c *gin.Context) {
	id, ok := s.battleIDOrLatest(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	resp, err := s.store.Rankings(c.Request.Context(), id, c.DefaultQuery("side", "home"), c.Query("career"), c.Query("team"), c.Query("search"), page, size)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) playerDetail(c *gin.Context) {
	battleID, ok := s.battleIDOrLatest(c)
	if !ok {
		return
	}
	statID, ok := pathID(c, "stat_id")
	if !ok {
		return
	}
	detail, err := s.store.PlayerDetail(c.Request.Context(), battleID, statID)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (s *Server) teamTop3(c *gin.Context) {
	id, ok := s.battleIDOrLatest(c)
	if !ok {
		return
	}
	resp, err := s.store.TeamTop3(c.Request.Context(), id, c.Query("side"))
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func (s *Server) guildComparison(c *gin.Context) {
	id, ok := s.battleIDOrLatest(c)
	if !ok {
		return
	}
	resp, err := s.store.GuildComparison(c.Request.Context(), id)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) squadComparison(c *gin.Context) {
	id, ok := s.battleIDOrLatest(c)
	if !ok {
		return
	}
	resp, err := s.store.SquadComparison(c.Request.Context(), id, c.Query("side"))
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) historyRankings(c *gin.Context) {
	minMatches, _ := strconv.Atoi(c.DefaultQuery("min_matches", "3"))
	resp, err := s.store.HistoryRankings(c.Request.Context(), c.Query("guild"), c.Query("career"), c.Query("search"), minMatches)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func (s *Server) settings(c *gin.Context) {
	resp, err := s.store.Settings(c.Request.Context())
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) putSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if err := s.store.PutSettings(c.Request.Context(), req); err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) scoringRules(c *gin.Context) {
	resp, err := s.store.ScoringRules(c.Request.Context())
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func (s *Server) createScoringRule(c *gin.Context) {
	var req struct {
		Name   string      `json:"name"`
		Config interface{} `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	version, err := s.store.CreateScoringRule(c.Request.Context(), req.Name, req.Config)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": version})
}

func (s *Server) validateScoringRules(c *gin.Context) {
	var profiles map[string]domain.CareerProfile
	if err := c.ShouldBindJSON(&profiles); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	var messages []domain.ImportMessage
	for career, profile := range profiles {
		if len(profile.Dimensions) != 6 {
			messages = append(messages, domain.ImportMessage{Level: "error", Code: "dimension_count", Message: career + " 必须配置 6 个维度。"})
		}
		sum := 0.0
		for _, dim := range profile.Dimensions {
			if dim.Enabled && dim.RankingWeight > 0 {
				sum += dim.RankingWeight
			}
		}
		if sum < 0.999 || sum > 1.001 {
			messages = append(messages, domain.ImportMessage{Level: "error", Code: "weight_sum", Message: career + " 参与排名权重合计必须为 100%。"})
		}
	}
	c.JSON(http.StatusOK, gin.H{"valid": len(messages) == 0, "messages": messages})
}

func (s *Server) rangeSuggestions(c *gin.Context) {
	var req struct {
		BattleID int64 `json:"battle_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.BattleID == 0 {
		id, err := s.store.LatestBattleID(c.Request.Context())
		if err != nil {
			respondStoreError(c, err)
			return
		}
		req.BattleID = id
	}
	rawRows, err := s.storeRowsForSuggestions(c.Request.Context(), req.BattleID)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": services.SuggestRanges(rawRows)})
}

func (s *Server) publishScoringRule(c *gin.Context) {
	session := currentSession(c)
	if err := s.store.PublishScoringRule(c.Request.Context(), c.Param("version"), session.User.ID); err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) scoringRanges(c *gin.Context) {
	resp, err := s.store.ScoringRanges(c.Request.Context())
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func (s *Server) publishScoringRange(c *gin.Context) {
	var req struct {
		Name   string      `json:"name"`
		Config interface{} `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	session := currentSession(c)
	version, err := s.store.PublishScoringRange(c.Request.Context(), req.Name, req.Config, session.User.ID)
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": version})
}

func (s *Server) uploadPlayerAvatar(c *gin.Context) {
	s.uploadAvatar(c, "players", c.Param("player_id"))
}

func (s *Server) deletePlayerAvatar(c *gin.Context) {
	if err := s.store.DeleteAvatar(c.Request.Context(), "players", c.Param("player_id")); err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) uploadCareerAvatar(c *gin.Context) {
	s.uploadAvatar(c, "careers", c.Param("career"))
}

func (s *Server) deleteCareerAvatar(c *gin.Context) {
	if err := s.store.DeleteAvatar(c.Request.Context(), "careers", c.Param("career")); err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) backup(c *gin.Context) {
	path, err := s.store.Backup(c.Request.Context())
	if err != nil {
		respondStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": path})
}

func (s *Server) uploadAvatar(c *gin.Context, kind, id string) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_file"})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "open_file_failed"})
		return
	}
	defer opened.Close()
	data, err := readAllLimited(opened, 5<<20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large"})
		return
	}
	session := currentSession(c)
	path, err := s.store.SaveAvatar(c.Request.Context(), kind, id, file.Filename, data, session.User.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar_failed", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"asset_path": path})
}

func (s *Server) setSessionCookies(c *gin.Context, token, csrf string, expires time.Time) {
	maxAge := int(time.Until(expires).Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, token, maxAge, "/", "", s.cfg.CookieSecure, true)
	c.SetCookie(csrfCookie, csrf, maxAge, "/", "", s.cfg.CookieSecure, false)
}

func (s *Server) clearSessionCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, "", -1, "/", "", s.cfg.CookieSecure, true)
	c.SetCookie(csrfCookie, "", -1, "/", "", s.cfg.CookieSecure, false)
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookie)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		session, err := s.auth.SessionFromToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if csrf, err := c.Cookie(csrfCookie); err == nil {
			session.CSRFToken = csrf
		}
		c.Set(sessionKey, session)
		c.Next()
	}
}

func (s *Server) requireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := currentSession(c)
		header := c.GetHeader(s.cfg.CSRFHeader)
		if header == "" {
			header = c.GetHeader("X-CSRF-Token")
		}
		if services.HashOpaque(header) != session.CSRFHash {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf_failed"})
			return
		}
		c.Next()
	}
}

func currentSession(c *gin.Context) services.Session {
	value, _ := c.Get(sessionKey)
	session, _ := value.(services.Session)
	return session
}

func (s *Server) battleIDOrLatest(c *gin.Context) (int64, bool) {
	raw := c.Param("battle_id")
	if raw == "" || raw == "latest" {
		id, err := s.store.LatestBattleID(c.Request.Context())
		if err != nil {
			respondStoreError(c, err)
			return 0, false
		}
		return id, true
	}
	return pathID(c, "battle_id")
}

func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return 0, false
	}
	return id, true
}

func respondStoreError(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, redis.Nil) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
}

func readAllLimited(reader interface{ Read([]byte) (int, error) }, limit int64) ([]byte, error) {
	limited := &io.LimitedReader{R: reader, N: limit + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("文件超过 %d 字节限制", limit)
	}
	return data, nil
}

func (s *Server) storeRowsForSuggestions(ctx context.Context, battleID int64) ([]domain.RawStat, error) {
	rows, err := s.store.BattleRows(ctx, battleID)
	if err != nil {
		return nil, err
	}
	rawRows := make([]domain.RawStat, 0, len(rows))
	for _, row := range rows {
		rawRows = append(rawRows, row.RawStat)
	}
	return rawRows, nil
}
