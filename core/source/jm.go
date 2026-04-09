package source

import (
	"context"
	"strings"

	"ImageMaster/core/jmbridge"
	"ImageMaster/core/types"
)

type JmSource struct {
	configManager types.ConfigManager
	ctx           context.Context
}

func NewJmSource(configManager types.ConfigManager) *JmSource {
	return &JmSource{configManager: configManager}
}

func (s *JmSource) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *JmSource) Summary() Summary {
	return Summary{
		ID:       "jmcomic",
		Name:     "JM漫画",
		Type:     "manga",
		Language: "zh",
		Website:  "https://18comic.vip",
		Version:  "0.1.0",
		BuiltIn:  true,
		Capabilities: []string{
			CapabilitySearch,
			CapabilityDetail,
			CapabilityRead,
			CapabilityRanking,
		},
		RankingKinds: []string{"week", "month", "day"},
		Description:  "内置 JM 漫画源，底层由打包的 JM runtime 提供搜索、榜单、详情与阅读能力。",
	}
}

func (s *JmSource) Search(query string, page int) (SearchResult, error) {
	result, err := jmbridge.Search(s.ctx, query, s.proxy(), page)
	if err != nil {
		return SearchResult{}, err
	}

	return SearchResult{
		Source:  s.Summary(),
		Query:   result.Query,
		Page:    result.Page,
		HasMore: result.HasMore,
		Total:   result.Total,
		Items:   s.mapItems(result.Items),
	}, nil
}

func (s *JmSource) Ranking(kind string, page int) (RankingResult, error) {
	normalizedKind := strings.ToLower(strings.TrimSpace(kind))
	if normalizedKind == "" {
		normalizedKind = "week"
	}
	if page < 1 {
		page = 1
	}

	if cached, fresh, ok := s.loadRankingCache(normalizedKind, page); ok && fresh {
		return cached, nil
	}

	result, err := jmbridge.Ranking(s.ctx, normalizedKind, s.proxy(), page)
	if err != nil {
		if cached, _, ok := s.loadRankingCache(normalizedKind, page); ok {
			return cached, nil
		}
		return RankingResult{}, err
	}

	mapped := RankingResult{
		Source: s.Summary(),
		Kind:   result.Kind,
		Page:   result.Page,
		Total:  result.Total,
		Items:  s.mapItems(result.Items),
	}
	_ = s.saveRankingCache(normalizedKind, page, mapped)
	return mapped, nil
}

func (s *JmSource) Detail(itemID string) (DetailResult, error) {
	result, err := jmbridge.Detail(s.ctx, itemID, s.proxy())
	if err != nil {
		return DetailResult{}, err
	}

	return DetailResult{
		Source: s.Summary(),
		Item: DetailItem{
			ID:        strings.TrimSpace(result.Item.ID),
			Title:     result.Item.Title,
			Cover:     result.Item.Cover,
			Summary:   result.Item.Summary,
			Author:    result.Item.Author,
			Status:    result.Item.Status,
			Tags:      result.Item.Tags,
			DetailURL: result.Item.DetailURL,
			Chapters:  s.mapChapters(result.Item.Chapters),
		},
	}, nil
}

func (s *JmSource) Images(chapterID string) (ImageResult, error) {
	result, err := jmbridge.Images(s.ctx, chapterID, s.proxy())
	if err != nil {
		return ImageResult{}, err
	}

	return ImageResult{
		Source:       s.Summary(),
		ComicTitle:   result.ComicTitle,
		ChapterTitle: result.ChapterTitle,
		ChapterURL:   result.ChapterURL,
		Images:       result.Images,
		Entries:      s.mapEntries(result.Entries),
		HasNext:      result.HasNext,
		NextURL:      result.NextURL,
	}, nil
}

func (s *JmSource) proxy() string {
	if s.configManager == nil {
		return ""
	}
	return strings.TrimSpace(s.configManager.GetProxy())
}

func (s *JmSource) mapItems(items []jmbridge.SearchItem) []SearchItem {
	mapped := make([]SearchItem, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, SearchItem{
			ID:             strings.TrimSpace(item.ID),
			Title:          strings.TrimSpace(item.Title),
			Cover:          strings.TrimSpace(item.Cover),
			Summary:        strings.TrimSpace(item.Summary),
			PrimaryLabel:   strings.TrimSpace(item.PrimaryLabel),
			SecondaryLabel: strings.TrimSpace(item.SecondaryLabel),
			DetailURL:      strings.TrimSpace(item.DetailURL),
		})
	}
	return mapped
}

func (s *JmSource) mapChapters(items []jmbridge.ChapterItem) []ChapterItem {
	mapped := make([]ChapterItem, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, ChapterItem{
			ID:           strings.TrimSpace(item.ID),
			Name:         strings.TrimSpace(item.Name),
			URL:          strings.TrimSpace(item.URL),
			Index:        item.Index,
			UpdatedLabel: strings.TrimSpace(item.UpdatedLabel),
		})
	}
	return mapped
}

func (s *JmSource) mapEntries(items []jmbridge.ImageEntry) []ImageEntry {
	mapped := make([]ImageEntry, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, ImageEntry{
			URL:     strings.TrimSpace(item.URL),
			Referer: strings.TrimSpace(item.Referer),
			Headers: cloneHeaders(item.Headers),
		})
	}
	return mapped
}
