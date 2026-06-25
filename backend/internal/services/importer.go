package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"nsh-guild-analytics/backend/internal/domain"
)

var requiredCSVColumns = []string{
	"帮会名", "玩家", "等级", "职业", "所在团长", "击败", "助攻", "战备资源",
	"对玩家伤害", "对建筑伤害", "治疗值", "承受伤害", "重伤", "青灯焚骨", "化羽", "控制",
}

var csvAliases = map[string]string{
	"击杀":   "击败",
	"团长":   "所在团长",
	"玩家名":  "玩家",
	"帮会":   "帮会名",
	"拆塔伤害": "对建筑伤害",
	"玩家伤害": "对玩家伤害",
	"死亡":   "重伤",
}

var battleTimePattern = regexp.MustCompile(`(?P<year>\d{4})_(?P<month>\d{2})_(?P<day>\d{2})_(?P<hour>\d{2})_(?P<minute>\d{2})_(?P<second>\d{2})`)

type ImportService struct{}

func NewImportService() *ImportService {
	return &ImportService{}
}

func (s *ImportService) ParsePreview(ctx context.Context, filename string, data []byte) (domain.ImportPreview, error) {
	_ = ctx
	sha := sha256.Sum256(data)
	text, encodingName, err := decodeCSVBytes(data)
	if err != nil {
		return domain.ImportPreview{}, err
	}
	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return domain.ImportPreview{}, fmt.Errorf("CSV 解析失败：%w", err)
	}
	preview := domain.ImportPreview{
		SourceFilename:  filepath.Base(filename),
		SourceSHA256:    hex.EncodeToString(sha[:]),
		OriginalRowCount: maxInt(len(records)-1, 0),
		Warnings: []domain.ImportMessage{{
			Level:   "info",
			Code:    "encoding_detected",
			Message: "文件编码识别为 " + encodingName,
		}},
	}
	if inferred, ok := inferBattleTime(filename); ok {
		preview.InferredBattleAt = &inferred
	}
	if len(records) == 0 {
		preview.Errors = append(preview.Errors, domain.ImportMessage{Level: "error", Code: "empty_file", Message: "CSV 文件为空"})
		return preview, nil
	}
	headers := normalizeHeaders(records[0])
	index := map[string]int{}
	for i, name := range headers {
		if _, exists := index[name]; !exists {
			index[name] = i
		}
	}
	for _, required := range requiredCSVColumns {
		if _, ok := index[required]; !ok {
			preview.Errors = append(preview.Errors, domain.ImportMessage{Level: "error", Code: "missing_column", Message: "缺少必填字段：" + required})
		}
	}
	if len(preview.Errors) > 0 {
		return preview, nil
	}

	profiles := domain.CareerProfiles()
	guilds := map[string]*domain.GuildPreview{}
	unknownCareers := map[string]bool{}
	for lineNo, record := range records[1:] {
		rowNumber := lineNo + 2
		if isEmptyRecord(record) {
			continue
		}
		if isRepeatedHeader(record, headers) {
			preview.Warnings = append(preview.Warnings, domain.ImportMessage{Level: "warning", Code: "repeated_header", RowNumber: rowNumber, Message: "已忽略重复表头"})
			continue
		}
		stat, rowErrors := parseStatRow(record, index, rowNumber)
		if len(rowErrors) > 0 {
			preview.Errors = append(preview.Errors, rowErrors...)
			continue
		}
		if _, ok := profiles[stat.Career]; !ok {
			unknownCareers[stat.Career] = true
		}
		preview.Rows = append(preview.Rows, stat)
		gp := guilds[stat.GuildName]
		if gp == nil {
			gp = &domain.GuildPreview{Name: stat.GuildName, Careers: map[string]int{}, Teams: map[string]int{}}
			guilds[stat.GuildName] = gp
		}
		gp.MemberCount++
		gp.Careers[stat.Career]++
		gp.Teams[stat.TeamLeader]++
	}

	preview.ValidRowCount = len(preview.Rows)
	for _, gp := range guilds {
		preview.Guilds = append(preview.Guilds, *gp)
	}
	sort.Slice(preview.Guilds, func(i, j int) bool { return preview.Guilds[i].Name < preview.Guilds[j].Name })
	for career := range unknownCareers {
		preview.UnknownCareers = append(preview.UnknownCareers, career)
	}
	sort.Strings(preview.UnknownCareers)
	if len(preview.UnknownCareers) > 0 {
		preview.Errors = append(preview.Errors, domain.ImportMessage{Level: "error", Code: "unknown_career", Message: "存在未配置职业：" + strings.Join(preview.UnknownCareers, "、")})
	}
	if len(preview.Guilds) != 2 {
		preview.Errors = append(preview.Errors, domain.ImportMessage{Level: "error", Code: "guild_count_invalid", Message: fmt.Sprintf("同一文件必须包含两个帮会，当前识别到 %d 个。", len(preview.Guilds))})
	}
	if len(preview.Rows) > 20 {
		preview.PreviewRows = append([]domain.RawStat(nil), preview.Rows[:20]...)
	} else {
		preview.PreviewRows = append([]domain.RawStat(nil), preview.Rows...)
	}
	preview.RangeSuggestions = SuggestRanges(preview.Rows)
	return preview, nil
}

func decodeCSVBytes(data []byte) (string, string, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(data) {
		return string(data), "UTF-8", nil
	}
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GB18030.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", "", fmt.Errorf("无法按 UTF-8 或 GB18030 解码文件：%w", err)
	}
	if !utf8.Valid(decoded) {
		return "", "", fmt.Errorf("文件编码不是有效的 UTF-8/GB18030 CSV")
	}
	return string(decoded), "GB18030", nil
}

func normalizeHeaders(headers []string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		name := strings.TrimSpace(strings.TrimPrefix(h, "\ufeff"))
		if alias, ok := csvAliases[name]; ok {
			name = alias
		}
		out[i] = name
	}
	return out
}

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func isRepeatedHeader(record []string, headers []string) bool {
	if len(record) == 0 {
		return false
	}
	matches := 0
	for i, raw := range record {
		if i >= len(headers) {
			break
		}
		name := strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff"))
		if alias, ok := csvAliases[name]; ok {
			name = alias
		}
		if name == headers[i] {
			matches++
		}
	}
	return float64(matches)/math.Max(float64(len(requiredCSVColumns)), 1) >= 0.8
}

func parseStatRow(record []string, index map[string]int, rowNumber int) (domain.RawStat, []domain.ImportMessage) {
	var stat domain.RawStat
	stat.RowNumber = rowNumber
	value := func(column string) string {
		i := index[column]
		if i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}
	stat.GuildName = value("帮会名")
	stat.PlayerName = value("玩家")
	stat.Career = value("职业")
	stat.TeamLeader = value("所在团长")
	var messages []domain.ImportMessage
	for _, item := range []struct {
		column string
		value  string
	}{
		{"帮会名", stat.GuildName},
		{"玩家", stat.PlayerName},
		{"职业", stat.Career},
		{"所在团长", stat.TeamLeader},
	} {
		if item.value == "" {
			messages = append(messages, domain.ImportMessage{Level: "error", Code: "required_empty", RowNumber: rowNumber, Message: "必填字段为空：" + item.column})
		}
	}
	parseInt := func(column string) int {
		n, ok := parseNumber(value(column), column, rowNumber, &messages)
		if !ok {
			return 0
		}
		return int(n)
	}
	parseInt64 := func(column string) int64 {
		n, ok := parseNumber(value(column), column, rowNumber, &messages)
		if !ok {
			return 0
		}
		return n
	}
	stat.Level = parseInt("等级")
	stat.Kills = parseInt("击败")
	stat.Assists = parseInt("助攻")
	stat.Logistics = parseInt64("战备资源")
	stat.PlayerDamage = parseInt64("对玩家伤害")
	stat.BuildingDamage = parseInt64("对建筑伤害")
	stat.Healing = parseInt64("治疗值")
	stat.DamageTaken = parseInt64("承受伤害")
	stat.Deaths = parseInt("重伤")
	stat.Qingdeng = parseInt("青灯焚骨")
	stat.Revive = parseInt("化羽")
	stat.Control = parseInt("控制")
	return stat, messages
}

func parseNumber(raw, column string, rowNumber int, messages *[]domain.ImportMessage) (int64, bool) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	if cleaned == "" {
		cleaned = "0"
	}
	n, err := strconv.ParseInt(cleaned, 10, 64)
	if err != nil {
		*messages = append(*messages, domain.ImportMessage{Level: "error", Code: "invalid_number", RowNumber: rowNumber, Message: fmt.Sprintf("%s 不是合法数字：%q", column, raw)})
		return 0, false
	}
	if n < 0 {
		*messages = append(*messages, domain.ImportMessage{Level: "error", Code: "negative_number", RowNumber: rowNumber, Message: fmt.Sprintf("%s 不能为负数：%q", column, raw)})
		return 0, false
	}
	return n, true
}

func inferBattleTime(filename string) (time.Time, bool) {
	match := battleTimePattern.FindStringSubmatch(filename)
	if len(match) != 7 {
		return time.Time{}, false
	}
	parts := make([]int, 6)
	for i := range parts {
		n, err := strconv.Atoi(match[i+1])
		if err != nil {
			return time.Time{}, false
		}
		parts[i] = n
	}
	return time.Date(parts[0], time.Month(parts[1]), parts[2], parts[3], parts[4], parts[5], 0, time.Local), true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
