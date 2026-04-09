package source

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"ImageMaster/core/download"
	"ImageMaster/core/types"
	"ImageMaster/core/utils"
)

var invalidSegmentChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

type API struct {
	registry      *Registry
	configManager types.ConfigManager
	ctx           context.Context
}

func NewAPI(configManager types.ConfigManager) *API {
	return &API{
		registry:      NewRegistry(configManager),
		configManager: configManager,
	}
}

func (a *API) SetContext(ctx context.Context) {
	a.ctx = ctx

	for _, provider := range a.registry.providers {
		if withContext, ok := provider.(interface{ SetContext(context.Context) }); ok {
			withContext.SetContext(ctx)
		}
	}
}

func (a *API) ListSources() []Summary {
	return a.registry.List()
}

func (a *API) SearchSources(sourceID string, query string, page int) (SearchResult, error) {
	return a.registry.Search(sourceID, query, page)
}

func (a *API) GetSourceDetail(sourceID string, itemID string) (DetailResult, error) {
	return a.registry.Detail(sourceID, itemID)
}

func (a *API) GetSourceImages(sourceID string, chapterID string) (ImageResult, error) {
	return a.registry.Images(sourceID, chapterID)
}

func (a *API) GetSourceRanking(sourceID string, kind string, page int) (RankingResult, error) {
	return a.registry.Ranking(sourceID, kind, page)
}

func (a *API) DownloadSourceChapter(sourceID string, chapterID string) (DownloadChapterResult, error) {
	if a.configManager == nil {
		return DownloadChapterResult{}, fmt.Errorf("config manager is not available")
	}

	rootDir := strings.TrimSpace(a.configManager.GetActiveLibrary())
	if rootDir == "" {
		rootDir = strings.TrimSpace(a.configManager.GetOutputDir())
	}
	if rootDir == "" {
		return DownloadChapterResult{}, fmt.Errorf("no active library or output directory is configured")
	}

	imageResult, err := a.registry.Images(sourceID, chapterID)
	if err != nil {
		return DownloadChapterResult{}, err
	}

	entries := make([]ImageEntry, 0, len(imageResult.Entries)+len(imageResult.Images))
	if len(imageResult.Entries) > 0 {
		entries = append(entries, imageResult.Entries...)
	} else {
		for _, rawURL := range imageResult.Images {
			entries = append(entries, ImageEntry{
				URL:     rawURL,
				Referer: imageResult.ChapterURL,
			})
		}
	}

	if len(entries) == 0 {
		return DownloadChapterResult{}, fmt.Errorf("no downloadable images returned by source")
	}

	comicDir := sanitizeSegment(fallbackString(imageResult.ComicTitle, sourceID))
	chapterDir := sanitizeSegment(fallbackString(imageResult.ChapterTitle, "chapter"))
	saveDir := utils.NormalizePath(filepath.Join(rootDir, comicDir, chapterDir))

	jobs := make([]download.DownloadJob, 0, len(entries))
	for index, entry := range entries {
		extension := guessImageExtension(entry.URL)
		headers := cloneHeaders(entry.Headers)
		if entry.Referer != "" {
			if headers == nil {
				headers = map[string]string{}
			}
			if _, exists := headers["Referer"]; !exists {
				headers["Referer"] = entry.Referer
			}
		}

		jobs = append(jobs, download.DownloadJob{
			URL:      entry.URL,
			FilePath: filepath.Join(saveDir, fmt.Sprintf("%03d%s", index+1, extension)),
			Headers:  headers,
		})
	}

	downloader := download.NewDownloader(download.Config{
		RetryCount:  3,
		RetryDelay:  2,
		ShowProcess: true,
	})
	downloader.SetConfigManager(a.configManager)
	if a.ctx != nil {
		downloader.SetContext(a.ctx)
	}

	successCount, err := downloader.BatchDownloadJobs(jobs)
	if err != nil {
		return DownloadChapterResult{}, err
	}
	if successCount != len(jobs) {
		return DownloadChapterResult{}, fmt.Errorf("chapter download incomplete: %d/%d files saved", successCount, len(jobs))
	}

	return DownloadChapterResult{
		Source:       imageResult.Source,
		ComicTitle:   imageResult.ComicTitle,
		ChapterTitle: imageResult.ChapterTitle,
		SaveDir:      saveDir,
		FileCount:    len(jobs),
	}, nil
}

func sanitizeSegment(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "untitled"
	}

	cleaned := invalidSegmentChars.ReplaceAllString(trimmed, "_")
	cleaned = strings.Trim(cleaned, " .")
	if cleaned == "" {
		return "untitled"
	}

	return cleaned
}

func cloneHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(headers))
	for key, value := range headers {
		cloned[key] = value
	}
	return cloned
}

func guessImageExtension(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ".jpg"
	}

	extension := strings.ToLower(path.Ext(parsed.Path))
	switch extension {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".avif":
		return extension
	default:
		return ".jpg"
	}
}
