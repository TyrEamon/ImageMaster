package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	mangaDexBaseURL    = "https://api.mangadex.org"
	mangaDexWebsiteURL = "https://mangadex.org"
	mangaDexPageSize   = 18
)

type MangaDexSource struct {
	client *http.Client
}

func NewMangaDexSource() *MangaDexSource {
	return &MangaDexSource{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *MangaDexSource) Summary() Summary {
	return Summary{
		ID:       "mangadex",
		Name:     "MangaDex",
		Type:     "manga",
		Language: "all",
		Website:  mangaDexWebsiteURL,
		Version:  "0.1.0",
		BuiltIn:  true,
		Capabilities: []string{
			CapabilitySearch,
		},
		Description: "Built-in sample source for validating the ImageMaster online source flow.",
	}
}

func (s *MangaDexSource) Search(query string, page int) (SearchResult, error) {
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

	searchURL := fmt.Sprintf(
		"%s/manga?title=%s&limit=%d&offset=%d&includes[]=cover_art&includes[]=author&contentRating[]=safe&contentRating[]=suggestive&contentRating[]=erotica&contentRating[]=pornographic",
		mangaDexBaseURL,
		url.QueryEscape(trimmedQuery),
		mangaDexPageSize,
		(page-1)*mangaDexPageSize,
	)

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return SearchResult{}, err
	}

	req.Header.Set("User-Agent", "ImageMaster/0.2")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return SearchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return SearchResult{}, fmt.Errorf("mangadex search failed: %s %s", resp.Status, string(body))
	}

	var payload mangaDexSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SearchResult{}, err
	}

	items := make([]SearchItem, 0, len(payload.Data))
	for _, entry := range payload.Data {
		title := chooseMangaDexTitle(entry.Attributes.Title, entry.Attributes.AltTitles)
		coverFileName := ""
		authorName := ""

		for _, relationship := range entry.Relationships {
			switch relationship.Type {
			case "cover_art":
				coverFileName = relationship.Attributes.FileName
			case "author":
				if authorName == "" {
					authorName = relationship.Attributes.Name
				}
			}
		}

		coverURL := ""
		if coverFileName != "" {
			coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.256.jpg", entry.ID, coverFileName)
		}

		secondaryParts := make([]string, 0, 2)
		if entry.Attributes.Status != "" {
			secondaryParts = append(secondaryParts, entry.Attributes.Status)
		}
		if entry.Attributes.Year != 0 {
			secondaryParts = append(secondaryParts, fmt.Sprintf("%d", entry.Attributes.Year))
		}

		items = append(items, SearchItem{
			ID:             entry.ID,
			Title:          title,
			Cover:          coverURL,
			Summary:        chooseMangaDexDescription(entry.Attributes.Description),
			PrimaryLabel:   fallbackString(authorName, entry.Attributes.OriginalLanguage),
			SecondaryLabel: strings.Join(secondaryParts, " / "),
			DetailURL:      fmt.Sprintf("%s/title/%s", mangaDexWebsiteURL, entry.ID),
		})
	}

	return SearchResult{
		Source:  s.Summary(),
		Query:   trimmedQuery,
		Page:    page,
		HasMore: page*mangaDexPageSize < payload.Total,
		Total:   payload.Total,
		Items:   items,
	}, nil
}

func (s *MangaDexSource) Detail(itemID string) (DetailResult, error) {
	return DetailResult{}, fmt.Errorf("MangaDex detail is not wired yet; this source currently supports search only")
}

func chooseMangaDexTitle(title map[string]string, altTitles []map[string]string) string {
	for _, key := range []string{"zh-hans", "zh-cn", "zh", "en", "ja-ro", "ja"} {
		if value := strings.TrimSpace(title[key]); value != "" {
			return value
		}
	}

	for _, candidate := range title {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}

	for _, altTitle := range altTitles {
		for _, key := range []string{"zh-hans", "zh-cn", "zh", "en", "ja-ro", "ja"} {
			if value := strings.TrimSpace(altTitle[key]); value != "" {
				return value
			}
		}
	}

	for _, altTitle := range altTitles {
		for _, candidate := range altTitle {
			if strings.TrimSpace(candidate) != "" {
				return candidate
			}
		}
	}

	return "Untitled"
}

func chooseMangaDexDescription(description map[string]string) string {
	for _, key := range []string{"zh-hans", "zh-cn", "zh", "en", "ja"} {
		if value := strings.TrimSpace(description[key]); value != "" {
			return value
		}
	}

	for _, candidate := range description {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}

	return "No description available."
}

func fallbackString(primary string, fallback string) string {
	primary = strings.TrimSpace(primary)
	if primary != "" {
		return primary
	}

	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}

	return "Unknown"
}

type mangaDexSearchResponse struct {
	Total int `json:"total"`
	Data  []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title            map[string]string   `json:"title"`
			AltTitles        []map[string]string `json:"altTitles"`
			Description      map[string]string   `json:"description"`
			OriginalLanguage string              `json:"originalLanguage"`
			Status           string              `json:"status"`
			Year             int                 `json:"year"`
		} `json:"attributes"`
		Relationships []struct {
			Type       string `json:"type"`
			Attributes struct {
				Name     string `json:"name"`
				FileName string `json:"fileName"`
			} `json:"attributes"`
		} `json:"relationships"`
	} `json:"data"`
}
