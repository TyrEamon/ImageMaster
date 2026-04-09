package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	dmzjBaseURL     = "https://www.dmzj.com"
	dmzjPageSize    = 20
	dmzjRankingSize = 24
)

type DmzjSource struct {
	client *http.Client
}

func NewDmzjSource() *DmzjSource {
	return &DmzjSource{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *DmzjSource) Summary() Summary {
	return Summary{
		ID:       "dmzj",
		Name:     "动漫之家",
		Type:     "manga",
		Language: "zh-cn",
		Website:  dmzjBaseURL,
		Version:  "0.1.0",
		BuiltIn:  true,
		Capabilities: []string{
			CapabilitySearch,
			CapabilityDetail,
			CapabilityRead,
			CapabilityRanking,
		},
		RankingKinds: []string{"latest"},
		Description:  "Built-in DMZJ source adapted from the Miru source flow.",
	}
}

func (s *DmzjSource) Search(query string, page int) (SearchResult, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return SearchResult{
			Source: s.Summary(),
			Query:  "",
			Page:   1,
			Items:  []SearchItem{},
		}, nil
	}

	if page < 1 {
		page = 1
	}

	endpoint := fmt.Sprintf(
		"/api/v1/comic1/search?keyword=%s&page=%d",
		url.QueryEscape(trimmedQuery),
		page,
	)

	var payload dmzjSearchResponse
	if err := s.fetchJSON(endpoint, &payload); err != nil {
		return SearchResult{}, err
	}

	items := make([]SearchItem, 0, len(payload.Data.ComicList))
	for _, comic := range payload.Data.ComicList {
		if strings.TrimSpace(comic.ComicPy) == "" || strings.TrimSpace(comic.Name) == "" {
			continue
		}

		items = append(items, SearchItem{
			ID:             strings.TrimSpace(comic.ComicPy),
			Title:          strings.TrimSpace(comic.Name),
			Cover:          strings.TrimSpace(comic.Cover),
			Summary:        fallbackString(strings.TrimSpace(comic.LastUpdateChapterName), "动漫之家漫画"),
			PrimaryLabel:   fallbackString(strings.TrimSpace(comic.Authors), "动漫之家"),
			SecondaryLabel: strings.TrimSpace(comic.LastUpdateChapterName),
			DetailURL:      s.detailURL(comic.ComicPy),
		})
	}

	return SearchResult{
		Source:  s.Summary(),
		Query:   trimmedQuery,
		Page:    page,
		HasMore: len(items) >= dmzjPageSize,
		Total:   len(items),
		Items:   items,
	}, nil
}

func (s *DmzjSource) Ranking(kind string, page int) (RankingResult, error) {
	normalizedKind := strings.ToLower(strings.TrimSpace(kind))
	if normalizedKind == "" {
		normalizedKind = "latest"
	}
	if page < 1 {
		page = 1
	}

	endpoint := fmt.Sprintf(
		"/api/v1/comic1/update_list?channel=pc&app_name=dmzj&version=1.0.0&page=%d&size=%d",
		page,
		dmzjRankingSize,
	)

	var payload dmzjLatestResponse
	if err := s.fetchJSON(endpoint, &payload); err != nil {
		return RankingResult{}, err
	}

	items := make([]SearchItem, 0, len(payload.Data.List))
	for _, comic := range payload.Data.List {
		if strings.TrimSpace(comic.ComicPy) == "" || strings.TrimSpace(comic.Title) == "" {
			continue
		}

		items = append(items, SearchItem{
			ID:             strings.TrimSpace(comic.ComicPy),
			Title:          strings.TrimSpace(comic.Title),
			Cover:          strings.TrimSpace(comic.Cover),
			Summary:        fallbackString(strings.TrimSpace(comic.LastUpdateChapterName), "最新更新"),
			PrimaryLabel:   "动漫之家",
			SecondaryLabel: strings.TrimSpace(comic.LastUpdateChapterName),
			DetailURL:      s.detailURL(comic.ComicPy),
		})
	}

	return RankingResult{
		Source: s.Summary(),
		Kind:   normalizedKind,
		Page:   page,
		Total:  len(items),
		Items:  items,
	}, nil
}

func (s *DmzjSource) Detail(itemID string) (DetailResult, error) {
	trimmedID := strings.TrimSpace(itemID)
	if trimmedID == "" {
		return DetailResult{}, fmt.Errorf("missing dmzj item id")
	}

	endpoint := fmt.Sprintf(
		"/api/v1/comic1/comic/detail?channel=pc&app_name=dmzj&version=1.0.0&comic_py=%s",
		url.QueryEscape(trimmedID),
	)

	var payload dmzjDetailResponse
	if err := s.fetchJSON(endpoint, &payload); err != nil {
		return DetailResult{}, err
	}

	comic := payload.Data.ComicInfo
	if strings.TrimSpace(comic.Title) == "" {
		return DetailResult{}, fmt.Errorf("dmzj detail returned empty comic info")
	}

	chapters := make([]ChapterItem, 0, 64)
	for _, group := range comic.ChapterList {
		groupChapters := make([]ChapterItem, 0, len(group.Data))
		for _, chapter := range group.Data {
			chapterID := strings.TrimSpace(strconv.FormatInt(chapter.ChapterID, 10))
			if chapterID == "" {
				continue
			}

			groupChapters = append(groupChapters, ChapterItem{
				ID:           fmt.Sprintf("%d|%s", comic.ID, chapterID),
				Name:         strings.TrimSpace(chapter.ChapterTitle),
				URL:          fmt.Sprintf("%d|%s", comic.ID, chapterID),
				Index:        len(chapters) + len(groupChapters),
				UpdatedLabel: strings.TrimSpace(group.Title),
			})
		}

		for left, right := 0, len(groupChapters)-1; left < right; left, right = left+1, right-1 {
			groupChapters[left], groupChapters[right] = groupChapters[right], groupChapters[left]
		}
		chapters = append(chapters, groupChapters...)
	}

	tags := make([]string, 0, 4)
	if strings.TrimSpace(comic.Zone) != "" {
		tags = append(tags, strings.TrimSpace(comic.Zone))
	}
	if strings.TrimSpace(comic.Status) != "" {
		tags = append(tags, strings.TrimSpace(comic.Status))
	}

	return DetailResult{
		Source: s.Summary(),
		Item: DetailItem{
			ID:        trimmedID,
			Title:     strings.TrimSpace(comic.Title),
			Cover:     strings.TrimSpace(comic.Cover),
			Summary:   fallbackString(strings.TrimSpace(comic.Description), "No summary available."),
			Author:    fallbackString(strings.TrimSpace(comic.Authors), "Unknown author"),
			Status:    fallbackString(strings.TrimSpace(comic.Status), "Unknown"),
			Tags:      tags,
			DetailURL: s.detailURL(trimmedID),
			Chapters:  chapters,
		},
	}, nil
}

func (s *DmzjSource) Images(chapterID string) (ImageResult, error) {
	parts := strings.Split(strings.TrimSpace(chapterID), "|")
	if len(parts) != 2 {
		return ImageResult{}, fmt.Errorf("invalid dmzj chapter id: %s", chapterID)
	}

	comicID := strings.TrimSpace(parts[0])
	rawChapterID := strings.TrimSpace(parts[1])
	if comicID == "" || rawChapterID == "" {
		return ImageResult{}, fmt.Errorf("invalid dmzj chapter id: %s", chapterID)
	}

	endpoint := fmt.Sprintf(
		"/api/v1/comic1/chapter/detail?channel=pc&app_name=dmzj&version=1.0.0&comic_id=%s&chapter_id=%s",
		url.QueryEscape(comicID),
		url.QueryEscape(rawChapterID),
	)

	var payload dmzjChapterResponse
	if err := s.fetchJSON(endpoint, &payload); err != nil {
		return ImageResult{}, err
	}

	urls := parseDMZJPageURLs(payload.Data.ChapterInfo.PageURLHD)
	if len(urls) == 0 {
		urls = parseDMZJPageURLs(payload.Data.ChapterInfo.PageURL)
	}
	if len(urls) == 0 {
		return ImageResult{}, fmt.Errorf("dmzj chapter returned no readable page urls")
	}

	entries := make([]ImageEntry, 0, len(urls))
	for _, imageURL := range urls {
		trimmed := strings.TrimSpace(imageURL)
		if trimmed == "" {
			continue
		}
		entries = append(entries, ImageEntry{URL: trimmed})
	}

	chapterTitle := strings.TrimSpace(payload.Data.ChapterInfo.ChapterTitle)
	if chapterTitle == "" {
		chapterTitle = fmt.Sprintf("Chapter %s", rawChapterID)
	}

	comicTitle := strings.TrimSpace(payload.Data.ChapterInfo.ComicTitle)
	if comicTitle == "" {
		comicTitle = "动漫之家"
	}

	return ImageResult{
		Source:       s.Summary(),
		ComicTitle:   comicTitle,
		ChapterTitle: chapterTitle,
		ChapterURL:   s.chapterURL(comicID, rawChapterID),
		Images:       urls,
		Entries:      entries,
		HasNext:      false,
		NextURL:      "",
	}, nil
}

func (s *DmzjSource) fetchJSON(endpoint string, target any) error {
	fullURL := endpoint
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(endpoint)), "http") {
		fullURL = dmzjBaseURL + endpoint
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "ImageMaster/0.2")
	req.Header.Set("Accept", "application/json,text/plain,*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", dmzjBaseURL+"/")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("dmzj request failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (s *DmzjSource) detailURL(comicPy string) string {
	trimmed := strings.TrimSpace(comicPy)
	if trimmed == "" {
		return dmzjBaseURL
	}
	return fmt.Sprintf("%s/info/%s.html", dmzjBaseURL, trimmed)
}

func (s *DmzjSource) chapterURL(comicID string, chapterID string) string {
	return fmt.Sprintf("%s/view/%s/%s.html", dmzjBaseURL, strings.TrimSpace(comicID), strings.TrimSpace(chapterID))
}

func parseDMZJPageURLs(raw any) []string {
	result := make([]string, 0, 16)

	appendURL := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		result = append(result, trimmed)
	}

	switch value := raw.(type) {
	case []string:
		for _, item := range value {
			appendURL(item)
		}
	case []any:
		for _, item := range value {
			if text, ok := item.(string); ok {
				appendURL(text)
			}
		}
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return result
		}

		var stringList []string
		if json.Unmarshal([]byte(trimmed), &stringList) == nil {
			for _, item := range stringList {
				appendURL(item)
			}
			return result
		}

		var anyList []any
		if json.Unmarshal([]byte(trimmed), &anyList) == nil {
			for _, item := range anyList {
				if text, ok := item.(string); ok {
					appendURL(text)
				}
			}
			return result
		}

		appendURL(trimmed)
	}

	return result
}

type dmzjSearchResponse struct {
	Data struct {
		ComicList []struct {
			Name                  string `json:"name"`
			Cover                 string `json:"cover"`
			LastUpdateChapterName string `json:"last_update_chapter_name"`
			ComicPy               string `json:"comic_py"`
			Authors               string `json:"authors"`
		} `json:"comic_list"`
	} `json:"data"`
}

type dmzjLatestResponse struct {
	Data struct {
		List []struct {
			Title                 string `json:"title"`
			Cover                 string `json:"cover"`
			LastUpdateChapterName string `json:"lastUpdateChapterName"`
			ComicPy               string `json:"comic_py"`
		} `json:"list"`
	} `json:"data"`
}

type dmzjDetailResponse struct {
	Data struct {
		ComicInfo struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			Cover       string `json:"cover"`
			Description string `json:"description"`
			Authors     string `json:"authors"`
			Status      string `json:"status"`
			Zone        string `json:"zone"`
			ChapterList []struct {
				Title string `json:"title"`
				Data  []struct {
					ChapterTitle string `json:"chapter_title"`
					ChapterID    int64  `json:"chapter_id"`
				} `json:"data"`
			} `json:"chapterList"`
		} `json:"comicInfo"`
	} `json:"data"`
}

type dmzjChapterResponse struct {
	Data struct {
		ChapterInfo struct {
			ComicTitle   string `json:"comic_title"`
			ChapterTitle string `json:"chapter_title"`
			PageURL      any    `json:"page_url"`
			PageURLHD    any    `json:"page_url_hd"`
		} `json:"chapterInfo"`
	} `json:"data"`
}
