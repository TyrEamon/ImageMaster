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

type detailCachePayload struct {
	CachedAt time.Time    `json:"cachedAt"`
	Result   DetailResult `json:"result"`
}

var cacheFilePattern = regexp.MustCompile(`[^a-z0-9_-]+`)

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

func (s *JmSource) loadDetailCache(itemID string) (DetailResult, bool, bool) {
	cachePath, err := s.detailCachePath(itemID)
	if err != nil {
		return DetailResult{}, false, false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return DetailResult{}, false, false
	}

	var payload detailCachePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return DetailResult{}, false, false
	}

	if payload.CachedAt.IsZero() {
		payload.CachedAt = time.Unix(0, 0)
	}
	if payload.Result.Source.ID == "" {
		payload.Result.Source = s.Summary()
	}

	fresh := time.Since(payload.CachedAt) <= s.detailCacheTTL()
	return payload.Result, fresh, true
}

func (s *JmSource) saveDetailCache(itemID string, result DetailResult) error {
	cachePath, err := s.detailCachePath(itemID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	payload := detailCachePayload{
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

	s.cleanupOldCacheFiles(filepath.Dir(cachePath), 7*24*time.Hour)
	return nil
}

func (s *JmSource) rankingCachePath(kind string, page int) (string, error) {
	rootDir := s.jmCacheBaseDir()
	cacheDir := filepath.Join(rootDir, "meta", "rankings")
	fileName := sanitizeCacheSegment(kind, "week")
	if page > 1 {
		fileName += "-" + strconvInt(page)
	}
	return filepath.Join(cacheDir, fileName+".json"), nil
}

func (s *JmSource) detailCachePath(itemID string) (string, error) {
	rootDir := s.jmCacheBaseDir()
	cacheDir := filepath.Join(rootDir, "meta", "details")
	fileName := sanitizeCacheSegment(itemID, "detail")
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

func (s *JmSource) detailCacheTTL() time.Duration {
	return 12 * time.Hour
}

func (s *JmSource) shouldRefreshRankingCache(result RankingResult) bool {
	checkCount := min(len(result.Items), 12)
	if checkCount == 0 {
		return false
	}

	for i := 0; i < checkCount; i++ {
		if strings.TrimSpace(result.Items[i].Cover) == "" {
			return true
		}
	}

	return false
}

func (s *JmSource) enrichRankingItemsFromDetailCache(items []SearchItem) ([]SearchItem, bool) {
	if len(items) == 0 {
		return items, false
	}

	changed := false
	enriched := make([]SearchItem, len(items))
	copy(enriched, items)

	for index := range enriched {
		if strings.TrimSpace(enriched[index].Cover) != "" {
			continue
		}

		detail, _, ok := s.loadDetailCache(enriched[index].ID)
		if !ok {
			continue
		}

		cover := strings.TrimSpace(detail.Item.Cover)
		if cover == "" {
			continue
		}

		enriched[index].Cover = cover
		changed = true
	}

	return enriched, changed
}

func (s *JmSource) cleanupOldCacheFiles(cacheDir string, maxAge time.Duration) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}

	expireBefore := time.Now().Add(-maxAge)
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

func (s *JmSource) cleanupOldRankingCacheFiles(cacheDir string) {
	s.cleanupOldCacheFiles(cacheDir, 7*24*time.Hour)
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

func sanitizeCacheSegment(value string, fallback string) string {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "" {
		return fallback
	}
	text = cacheFilePattern.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-_.")
	if text == "" {
		return fallback
	}
	return text
}

func strconvInt(value int) string {
	return strconv.Itoa(value)
}
