package source

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baoziBaseURL = "https://www.baozimh.com"

var baoziImageRegexp = regexp.MustCompile(`"url"\s*:\s*"(https:[^"]+)"`)

type BaoziSource struct {
	client *http.Client
}

func NewBaoziSource() *BaoziSource {
	return &BaoziSource{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *BaoziSource) Summary() Summary {
	return Summary{
		ID:          "baozi",
		Name:        "包子漫画",
		Type:        "manga",
		Language:    "zh",
		Website:     baoziBaseURL,
		Description: "参考 Miru 的包子漫画源实现。当前版本先支持搜索、详情和在线阅读。",
	}
}

func (s *BaoziSource) Search(query string, page int) (SearchResult, error) {
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

	if page > 1 {
		return SearchResult{
			Source:  s.Summary(),
			Query:   trimmedQuery,
			Page:    page,
			HasMore: false,
			Total:   0,
			Items:   []SearchItem{},
		}, nil
	}

	searchURL := fmt.Sprintf("%s/search?q=%s", baoziBaseURL, url.QueryEscape(trimmedQuery))
	doc, err := s.fetchDocument(searchURL)
	if err != nil {
		return SearchResult{}, err
	}

	items := make([]SearchItem, 0, 18)
	doc.Find("div.comics-card").Each(func(_ int, selection *goquery.Selection) {
		posterLink := selection.Find("a.comics-card__poster").First()
		infoLink := selection.Find("a.comics-card__info").First()

		href, _ := posterLink.Attr("href")
		if strings.TrimSpace(href) == "" {
			href, _ = infoLink.Attr("href")
		}

		title := normalizeBaoziText(selection.Find(".comics-card__title h3").First().Text())
		if title == "" {
			title = normalizeBaoziText(posterLink.AttrOr("title", ""))
		}

		cover := strings.TrimSpace(html.UnescapeString(
			posterLink.Find("amp-img").First().AttrOr("src", ""),
		))
		author := normalizeBaoziText(selection.Find("small.tags").First().Text())
		tags := collectBaoziTags(selection)

		if title == "" || strings.TrimSpace(href) == "" {
			return
		}

		summary := "来自包子漫画搜索"
		if len(tags) > 0 {
			summary = strings.Join(tags, " / ")
		}

		items = append(items, SearchItem{
			ID:             strings.TrimPrefix(strings.TrimSpace(href), "/"),
			Title:          title,
			Cover:          cover,
			Summary:        summary,
			PrimaryLabel:   fallbackString(author, "未知作者"),
			SecondaryLabel: strings.Join(tags, " / "),
			DetailURL:      resolveBaoziURL(href),
		})
	})

	return SearchResult{
		Source:  s.Summary(),
		Query:   trimmedQuery,
		Page:    page,
		HasMore: false,
		Total:   len(items),
		Items:   items,
	}, nil
}

func (s *BaoziSource) Detail(itemID string) (DetailResult, error) {
	detailURL := resolveBaoziURL(itemID)
	doc, err := s.fetchDocument(detailURL)
	if err != nil {
		return DetailResult{}, err
	}

	title := fallbackString(
		normalizeBaoziText(doc.Find("meta[property='og:novel:book_name']").AttrOr("content", "")),
		normalizeBaoziText(doc.Find(".comics-detail__title").First().Text()),
	)

	cover := strings.TrimSpace(html.UnescapeString(
		doc.Find("meta[property='og:image']").AttrOr("content", ""),
	))
	if cover == "" {
		cover = strings.TrimSpace(html.UnescapeString(doc.Find("amp-img").First().AttrOr("src", "")))
	}

	summary := fallbackString(
		normalizeBaoziText(doc.Find(".comics-detail__desc").First().Text()),
		normalizeBaoziText(doc.Find("meta[name='description']").AttrOr("content", "")),
	)
	author := normalizeBaoziText(doc.Find("meta[property='og:novel:author']").AttrOr("content", ""))
	status := normalizeBaoziText(doc.Find("meta[property='og:novel:status']").AttrOr("content", ""))

	tags := make([]string, 0, 4)
	if category := normalizeBaoziText(doc.Find("meta[property='og:novel:category']").AttrOr("content", "")); category != "" {
		tags = append(tags, category)
	}

	chapters := make([]ChapterItem, 0, 32)
	seen := map[string]struct{}{}
	doc.Find(".comics-chapters__item").Each(func(_ int, selection *goquery.Selection) {
		href, _ := selection.Attr("href")
		resolvedURL := resolveBaoziURL(href)
		if resolvedURL == "" {
			return
		}
		if _, ok := seen[resolvedURL]; ok {
			return
		}
		seen[resolvedURL] = struct{}{}

		name := normalizeBaoziText(selection.Find("span").First().Text())
		if name == "" {
			name = normalizeBaoziText(selection.Text())
		}
		if name == "" {
			return
		}

		updatedLabel := normalizeBaoziText(selection.Parent().Find("em").First().Text())
		chapters = append(chapters, ChapterItem{
			ID:           strings.TrimPrefix(strings.TrimSpace(href), "/"),
			Name:         name,
			URL:          resolvedURL,
			UpdatedLabel: updatedLabel,
		})
	})

	return DetailResult{
		Source: s.Summary(),
		Item: DetailItem{
			ID:        strings.TrimPrefix(strings.TrimSpace(itemID), "/"),
			Title:     fallbackString(title, "未命名作品"),
			Cover:     cover,
			Summary:   fallbackString(summary, "暂无简介"),
			Author:    fallbackString(author, "未知作者"),
			Status:    fallbackString(status, "状态未知"),
			Tags:      tags,
			DetailURL: detailURL,
			Chapters:  chapters,
		},
	}, nil
}

func (s *BaoziSource) Images(chapterID string) (ImageResult, error) {
	chapterURL := resolveBaoziURL(chapterID)
	htmlText, finalURL, err := s.fetchHTML(chapterURL)
	if err != nil {
		return ImageResult{}, err
	}

	matches := baoziImageRegexp.FindAllStringSubmatch(htmlText, -1)
	seen := map[string]struct{}{}
	images := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		imageURL := strings.TrimSpace(match[1])
		if imageURL == "" {
			continue
		}
		if _, ok := seen[imageURL]; ok {
			continue
		}
		seen[imageURL] = struct{}{}
		images = append(images, imageURL)
	}

	if len(images) == 0 {
		return ImageResult{}, fmt.Errorf("no images found in chapter page")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		return ImageResult{}, err
	}

	comicTitle := fallbackString(
		normalizeBaoziText(doc.Find("meta[property='og:novel:book_name']").AttrOr("content", "")),
		normalizeBaoziText(doc.Find("a[href*='/comic/']").First().Text()),
	)

	chapterTitle := normalizeBaoziText(doc.Find("title").First().Text())
	if strings.Contains(chapterTitle, "-") {
		chapterTitle = normalizeBaoziText(strings.SplitN(chapterTitle, "-", 2)[0])
	}
	if chapterTitle == "" {
		chapterTitle = "在线章节"
	}

	nextURL := strings.TrimSpace(doc.Find("#next-chapter").AttrOr("href", ""))
	nextURL = resolveBaoziURL(nextURL)

	return ImageResult{
		Source:       s.Summary(),
		ComicTitle:   fallbackString(comicTitle, "包子漫画"),
		ChapterTitle: chapterTitle,
		ChapterURL:   finalURL,
		Images:       images,
		HasNext:      nextURL != "" && !strings.Contains(doc.Text(), "這是本作品最後一話了"),
		NextURL:      nextURL,
	}, nil
}

func (s *BaoziSource) fetchDocument(targetURL string) (*goquery.Document, error) {
	htmlText, _, err := s.fetchHTML(targetURL)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(strings.NewReader(htmlText))
}

func (s *BaoziSource) fetchHTML(targetURL string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", "ImageMaster/0.2")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("baozi request failed: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	finalURL := targetURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return string(bodyBytes), finalURL, nil
}

func collectBaoziTags(selection *goquery.Selection) []string {
	tags := make([]string, 0, 3)
	selection.Find(".tabs .tab").Each(func(_ int, tag *goquery.Selection) {
		value := normalizeBaoziText(tag.Text())
		if value != "" {
			tags = append(tags, value)
		}
	})
	return tags
}

func normalizeBaoziText(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	return strings.Join(fields, " ")
}

func resolveBaoziURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err == nil && parsed.IsAbs() {
		return parsed.String()
	}

	base, err := url.Parse(baoziBaseURL)
	if err != nil {
		return trimmed
	}

	relative, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	return base.ResolveReference(relative).String()
}
