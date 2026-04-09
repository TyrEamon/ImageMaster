package source

type Summary struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Language     string   `json:"language"`
	Website      string   `json:"website"`
	Version      string   `json:"version"`
	BuiltIn      bool     `json:"builtIn"`
	Capabilities []string `json:"capabilities"`
	RankingKinds []string `json:"rankingKinds"`
	Description  string   `json:"description"`
}

const (
	CapabilitySearch  = "search"
	CapabilityDetail  = "detail"
	CapabilityRead    = "read"
	CapabilityRanking = "ranking"
)

type SearchItem struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Cover          string `json:"cover"`
	Summary        string `json:"summary"`
	PrimaryLabel   string `json:"primaryLabel"`
	SecondaryLabel string `json:"secondaryLabel"`
	DetailURL      string `json:"detailUrl"`
}

type SearchResult struct {
	Source  Summary      `json:"source"`
	Query   string       `json:"query"`
	Page    int          `json:"page"`
	HasMore bool         `json:"hasMore"`
	Total   int          `json:"total"`
	Items   []SearchItem `json:"items"`
}

type RankingResult struct {
	Source Summary      `json:"source"`
	Kind   string       `json:"kind"`
	Page   int          `json:"page"`
	Total  int          `json:"total"`
	Items  []SearchItem `json:"items"`
}

type ChapterItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Index        int    `json:"index"`
	UpdatedLabel string `json:"updatedLabel"`
}

type DetailItem struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Cover     string        `json:"cover"`
	Summary   string        `json:"summary"`
	Author    string        `json:"author"`
	Status    string        `json:"status"`
	Tags      []string      `json:"tags"`
	DetailURL string        `json:"detailUrl"`
	Chapters  []ChapterItem `json:"chapters"`
}

type DetailResult struct {
	Source Summary    `json:"source"`
	Item   DetailItem `json:"item"`
}

type ImageEntry struct {
	URL     string            `json:"url"`
	Referer string            `json:"referer,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type ImageResult struct {
	Source       Summary      `json:"source"`
	ComicTitle   string       `json:"comicTitle"`
	ChapterTitle string       `json:"chapterTitle"`
	ChapterURL   string       `json:"chapterUrl"`
	Images       []string     `json:"images"`
	Entries      []ImageEntry `json:"entries"`
	HasNext      bool         `json:"hasNext"`
	NextURL      string       `json:"nextUrl"`
}

type DownloadChapterResult struct {
	Source       Summary `json:"source"`
	ComicTitle   string  `json:"comicTitle"`
	ChapterTitle string  `json:"chapterTitle"`
	SaveDir      string  `json:"saveDir"`
	FileCount    int     `json:"fileCount"`
}

type Provider interface {
	Summary() Summary
	Search(query string, page int) (SearchResult, error)
}

type DetailProvider interface {
	Detail(itemID string) (DetailResult, error)
}

type ImageProvider interface {
	Images(chapterID string) (ImageResult, error)
}

type RankingProvider interface {
	Ranking(kind string, page int) (RankingResult, error)
}
