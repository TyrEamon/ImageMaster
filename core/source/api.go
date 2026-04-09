package source

type API struct {
	registry *Registry
}

func NewAPI() *API {
	return &API{
		registry: NewRegistry(),
	}
}

func (a *API) ListSources() []Summary {
	return a.registry.List()
}

func (a *API) SearchSources(sourceID string, query string, page int) (SearchResult, error) {
	return a.registry.Search(sourceID, query, page)
}
