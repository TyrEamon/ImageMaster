package source

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type rankingCachePayload struct {
	CachedAt time.Time     `json:"cachedAt"`
	Result   RankingResult `json:"result"`
}

var rankingCacheFilePattern = regexp.MustCompile(`[^a-z0-9_-]+`)

func (s *JmSource) loadRankingCache(kind string, page int) (RankingResult, bool, bool) {
	cachePath, err := s.rankingCachePath(kind, page)
	if err != nil {
		return RankingResult{}, false, false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return RankingResult{}, false, false
	}

	var payload rankingCachePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return RankingResult{}, false, false
	}

	if payload.CachedAt.IsZero() {
		payload.CachedAt = time.Unix(0, 0)
	}
	if payload.Result.Source.ID == "" {
		payload.Result.Source = s.Summary()
	}

	fresh := time.Since(payload.CachedAt) <= s.rankingCacheTTL(kind)
	return payload.Result, fresh, true
}

func (s *JmSource) saveRankingCache(kind string, page int, result RankingResult) error {
	cachePath, err := s.rankingCachePath(kind, page)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	payload := rankingCachePayload{
		CachedAt: time.Now(),
		Result:   result,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		return err
	}

	s.cleanupOldRankingCacheFiles(filepath.Dir(cachePath))
	return nil
}

func (s *JmSource) rankingCachePath(kind string, page int) (string, error) {
	rootDir := s.jmCacheBaseDir()
	cacheDir := filepath.Join(rootDir, "meta", "rankings")
	fileName := sanitizeRankingCacheSegment(kind)
	if page > 1 {
		fileName += "-" + strconvInt(page)
	}
	return filepath.Join(cacheDir, fileName+".json"), nil
}

func (s *JmSource) rankingCacheTTL(kind string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "day":
		return 3 * time.Hour
	case "week":
		return 12 * time.Hour
	case "month":
		return 24 * time.Hour
	case "year":
		return 48 * time.Hour
	default:
		return 12 * time.Hour
	}
}

func (s *JmSource) cleanupOldRankingCacheFiles(cacheDir string) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}

	expireBefore := time.Now().Add(-7 * 24 * time.Hour)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(expireBefore) {
			_ = os.Remove(path)
		}
	}
}

func (s *JmSource) jmCacheBaseDir() string {
	configured := ""
	if s.configManager != nil {
		configured = strings.TrimSpace(s.configManager.GetJmCacheDir())
	}
	if configured != "" {
		return configured
	}
	return filepath.Join(os.TempDir(), "imagemaster-jm-cache")
}

func sanitizeRankingCacheSegment(value string) string {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "" {
		return "week"
	}
	text = rankingCacheFilePattern.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-_.")
	if text == "" {
		return "week"
	}
	return text
}

func strconvInt(value int) string {
	return strconv.Itoa(value)
}
