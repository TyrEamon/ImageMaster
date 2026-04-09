package source

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ImageMaster/core/download"
	"ImageMaster/core/jmbridge"
	"ImageMaster/core/types"
	"ImageMaster/core/types/dto"
	"ImageMaster/core/utils"
)

var invalidSegmentChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

type API struct {
	registry      *Registry
	configManager types.ConfigManager
	historyStore  types.HistoryStore
	ctx           context.Context
}

func NewAPI(configManager types.ConfigManager, historyStore types.HistoryStore) *API {
	return &API{
		registry:      NewRegistry(configManager),
		configManager: configManager,
		historyStore:  historyStore,
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
	if sourceID == "jmcomic" {
		return a.getJMReadableImages(chapterID)
	}
	return a.registry.Images(sourceID, chapterID)
}

func (a *API) GetSourceRanking(sourceID string, kind string, page int) (RankingResult, error) {
	return a.registry.Ranking(sourceID, kind, page)
}

func (a *API) GetSourceChapterDownloadStatus(sourceID string, chapterID string) (ChapterDownloadStatusResult, error) {
	rootDir := a.resolveDownloadRootDir()
	if rootDir == "" {
		return ChapterDownloadStatusResult{}, nil
	}

	imageResult, err := a.registry.Images(sourceID, chapterID)
	if err != nil {
		return ChapterDownloadStatusResult{}, err
	}

	saveDir := buildChapterSaveDir(rootDir, imageResult, sourceID)
	fileCount, err := countFilesInDir(saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			fileCount = 0
		} else {
			return ChapterDownloadStatusResult{}, err
		}
	}

	return ChapterDownloadStatusResult{
		Source:       imageResult.Source,
		ComicTitle:   imageResult.ComicTitle,
		ChapterTitle: imageResult.ChapterTitle,
		SaveDir:      saveDir,
		FileCount:    fileCount,
		Downloaded:   fileCount > 0,
	}, nil
}

func (a *API) DownloadSourceChapter(sourceID string, chapterID string) (DownloadChapterResult, error) {
	startTime := time.Now()
	if a.configManager == nil {
		err := fmt.Errorf("config manager is not available")
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, DownloadChapterResult{}, err)
		return DownloadChapterResult{}, err
	}

	rootDir := a.resolveDownloadRootDir()
	if rootDir == "" {
		err := fmt.Errorf("no active library or output directory is configured")
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, DownloadChapterResult{}, err)
		return DownloadChapterResult{}, err
	}

	if sourceID == "jmcomic" {
		result, err := a.downloadJMChapter(rootDir, chapterID)
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, result, err)
		return result, err
	}

	imageResult, err := a.registry.Images(sourceID, chapterID)
	if err != nil {
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, DownloadChapterResult{}, err)
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
		err := fmt.Errorf("no downloadable images returned by source")
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, DownloadChapterResult{}, err)
		return DownloadChapterResult{}, err
	}

	saveDir := buildChapterSaveDir(rootDir, imageResult, sourceID)

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
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, DownloadChapterResult{}, err)
		return DownloadChapterResult{}, err
	}
	if successCount != len(jobs) {
		err := fmt.Errorf("chapter download incomplete: %d/%d files saved", successCount, len(jobs))
		a.recordSourceDownloadHistory(startTime, sourceID, chapterID, DownloadChapterResult{}, err)
		return DownloadChapterResult{}, err
	}

	result := DownloadChapterResult{
		Source:       imageResult.Source,
		ComicTitle:   imageResult.ComicTitle,
		ChapterTitle: imageResult.ChapterTitle,
		SaveDir:      saveDir,
		FileCount:    len(jobs),
	}
	a.recordSourceDownloadHistory(startTime, sourceID, chapterID, result, nil)
	return result, nil
}

func (a *API) downloadJMChapter(rootDir string, chapterID string) (DownloadChapterResult, error) {
	imageResult, err := a.registry.Images("jmcomic", chapterID)
	if err != nil {
		return DownloadChapterResult{}, err
	}

	proxy := ""
	if a.configManager != nil {
		proxy = strings.TrimSpace(a.configManager.GetProxy())
	}

	saveDir, err := jmbridge.Download(a.ctx, nil, chapterID, rootDir, proxy)
	if err != nil {
		return DownloadChapterResult{}, err
	}

	fileCount, err := countFilesInDir(saveDir)
	if err != nil {
		return DownloadChapterResult{}, err
	}

	return DownloadChapterResult{
		Source:       imageResult.Source,
		ComicTitle:   imageResult.ComicTitle,
		ChapterTitle: imageResult.ChapterTitle,
		SaveDir:      saveDir,
		FileCount:    fileCount,
	}, nil
}

func (a *API) getJMReadableImages(chapterID string) (ImageResult, error) {
	proxy := ""
	cacheDir := ""
	retentionHours := 24
	sizeLimitMB := 2048
	if a.configManager != nil {
		proxy = strings.TrimSpace(a.configManager.GetProxy())
		cacheDir = strings.TrimSpace(a.configManager.GetJmCacheDir())
		retentionHours = a.configManager.GetJmCacheRetentionHours()
		sizeLimitMB = a.configManager.GetJmCacheSizeLimitMB()
	}

	result, err := jmbridge.ReadableImages(a.ctx, chapterID, proxy, cacheDir, retentionHours, sizeLimitMB)
	if err != nil {
		return ImageResult{}, err
	}

	entries := make([]ImageEntry, 0, len(result.Entries))
	for _, entry := range result.Entries {
		entries = append(entries, ImageEntry{
			URL:     strings.TrimSpace(entry.URL),
			Referer: strings.TrimSpace(entry.Referer),
			Headers: cloneHeaders(entry.Headers),
		})
	}

	summary, summaryErr := a.registry.GetSummary("jmcomic")
	if summaryErr == nil {
		return ImageResult{
			Source:       summary,
			ComicTitle:   result.ComicTitle,
			ChapterTitle: result.ChapterTitle,
			ChapterURL:   result.ChapterURL,
			Images:       result.Images,
			Entries:      entries,
			HasNext:      result.HasNext,
			NextURL:      result.NextURL,
		}, nil
	}

	return ImageResult{
		ComicTitle:   result.ComicTitle,
		ChapterTitle: result.ChapterTitle,
		ChapterURL:   result.ChapterURL,
		Images:       result.Images,
		Entries:      entries,
		HasNext:      result.HasNext,
		NextURL:      result.NextURL,
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

func (a *API) resolveDownloadRootDir() string {
	if a.configManager == nil {
		return ""
	}

	rootDir := strings.TrimSpace(a.configManager.GetActiveLibrary())
	if rootDir == "" {
		rootDir = strings.TrimSpace(a.configManager.GetOutputDir())
	}

	return utils.NormalizePath(rootDir)
}

func buildChapterSaveDir(rootDir string, imageResult ImageResult, sourceID string) string {
	comicDir := sanitizeSegment(fallbackString(imageResult.ComicTitle, sourceID))
	chapterDir := sanitizeSegment(fallbackString(imageResult.ChapterTitle, "chapter"))
	return utils.NormalizePath(filepath.Join(rootDir, comicDir, chapterDir))
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

func countFilesInDir(targetDir string) (int, error) {
	info, err := os.Stat(targetDir)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect download directory: %w", err)
	}
	if !info.IsDir() {
		return 0, nil
	}

	count := 0
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() {
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count downloaded files: %w", err)
	}

	return count, nil
}

func (a *API) recordSourceDownloadHistory(startTime time.Time, sourceID string, chapterID string, result DownloadChapterResult, err error) {
	if a.historyStore == nil {
		return
	}

	now := time.Now()
	record := &dto.DownloadTaskDTO{
		ID:           fmt.Sprintf("source:%s:%s:%d", strings.TrimSpace(sourceID), strings.TrimSpace(chapterID), startTime.UnixNano()),
		URL:          strings.TrimSpace(chapterID),
		SavePath:     strings.TrimSpace(result.SaveDir),
		StartTime:    startTime,
		CompleteTime: now,
		UpdatedAt:    now,
		Name:         buildSourceDownloadHistoryName(sourceID, result),
	}

	if err != nil {
		record.Status = string(types.StatusFailed)
		record.Error = err.Error()
	} else {
		record.Status = string(types.StatusCompleted)
		record.Progress.Current = result.FileCount
		record.Progress.Total = result.FileCount
	}

	a.historyStore.AddDownloadRecord(record)
}

func buildSourceDownloadHistoryName(sourceID string, result DownloadChapterResult) string {
	comic := strings.TrimSpace(result.ComicTitle)
	chapter := strings.TrimSpace(result.ChapterTitle)
	switch {
	case comic != "" && chapter != "":
		return fmt.Sprintf("%s / %s", comic, chapter)
	case chapter != "":
		return chapter
	case comic != "":
		return comic
	default:
		return fmt.Sprintf("%s online chapter", strings.TrimSpace(sourceID))
	}
}
