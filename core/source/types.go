package source

type Summary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Language    string `json:"language"`
	Website     string `json:"website"`
	Description string `json:"description"`
}

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

type Provider interface {
	Summary() Summary
	Search(query string, page int) (SearchResult, error)
}
