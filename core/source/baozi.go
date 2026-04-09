package source

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baoziBaseURL = "https://www.baozimh.com"

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
		Description: "参考 Miru 的包子漫画源实现。当前版本先支持搜索，用来验证 ImageMaster 在线源架构。",
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

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return SearchResult{}, err
	}

	req.Header.Set("User-Agent", "ImageMaster/0.2")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return SearchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SearchResult{}, fmt.Errorf("baozi search failed: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
			title, _ = posterLink.Attr("title")
			title = normalizeBaoziText(title)
		}

		cover, _ := posterLink.Find("amp-img").First().Attr("src")
		cover = strings.TrimSpace(html.UnescapeString(cover))

		author := normalizeBaoziText(selection.Find("small.tags").First().Text())
		tags := make([]string, 0, 3)
		selection.Find(".tabs .tab").Each(func(_ int, tag *goquery.Selection) {
			value := normalizeBaoziText(tag.Text())
			if value != "" {
				tags = append(tags, value)
			}
		})

		if title == "" || strings.TrimSpace(href) == "" {
			return
		}

		detailURL := resolveBaoziURL(href)
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
			DetailURL:      detailURL,
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
