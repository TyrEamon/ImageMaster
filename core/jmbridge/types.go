package jmbridge

type SearchItem struct {
	ID             string
	Title          string
	Cover          string
	Summary        string
	PrimaryLabel   string
	SecondaryLabel string
	DetailURL      string
}

type SearchResult struct {
	Query   string
	Page    int
	HasMore bool
	Total   int
	Items   []SearchItem
}

type RankingResult struct {
	Kind  string
	Page  int
	Total int
	Items []SearchItem
}

type ChapterItem struct {
	ID           string
	Name         string
	URL          string
	Index        int
	UpdatedLabel string
}

type DetailItem struct {
	ID        string
	Title     string
	Cover     string
	Summary   string
	Author    string
	Status    string
	Tags      []string
	DetailURL string
	Chapters  []ChapterItem
}

type DetailResult struct {
	Item DetailItem
}

type ImageEntry struct {
	URL     string            `json:"url"`
	Referer string            `json:"referer,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type ImageResult struct {
	ComicTitle   string
	ChapterTitle string
	ChapterURL   string
	Images       []string
	Entries      []ImageEntry
	HasNext      bool
	NextURL      string
}
