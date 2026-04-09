package source

import (
	"fmt"
	"sort"
	"strings"

	"ImageMaster/core/types"
)

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(configManager types.ConfigManager) *Registry {
	registry := &Registry{
		providers: map[string]Provider{},
	}

	builtIns := map[string]Provider{
		"baozi":    NewBaoziSource(),
		"dmzj":     NewDmzjSource(),
		"mangadex": NewMangaDexSource(),
		"jmcomic":  NewJmSource(configManager),
	}

	registered := map[string]struct{}{}
	for _, manifest := range loadSourceManifests() {
		enabled := manifest.Enabled == nil || *manifest.Enabled
		if !enabled {
			continue
		}

		adapterID := manifest.Adapter
		if adapterID == "" {
			adapterID = manifest.ID
		}

		var provider Provider
		var ok bool

		switch strings.ToLower(strings.TrimSpace(manifest.Engine)) {
		case "script", "js":
			scriptProvider, err := NewScriptSource(manifest)
			if err != nil {
				continue
			}
			provider = scriptProvider
			ok = true
		default:
			provider, ok = builtIns[adapterID]
			if !ok {
				continue
			}
		}

		registry.Register(newWrappedProvider(provider, mergeSummary(provider.Summary(), manifest)))
		registered[adapterID] = struct{}{}
	}

	for id, provider := range builtIns {
		if _, ok := registered[id]; ok {
			continue
		}
		registry.Register(provider)
	}

	return registry
}

func (r *Registry) Register(provider Provider) {
	if provider == nil {
		return
	}

	r.providers[provider.Summary().ID] = provider
}

func (r *Registry) List() []Summary {
	summaries := make([]Summary, 0, len(r.providers))
	for _, provider := range r.providers {
		summaries = append(summaries, provider.Summary())
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries
}

func (r *Registry) GetSummary(sourceID string) (Summary, error) {
	provider, ok := r.providers[sourceID]
	if !ok {
		return Summary{}, fmt.Errorf("source not found: %s", sourceID)
	}

	return provider.Summary(), nil
}

func (r *Registry) Search(sourceID string, query string, page int) (SearchResult, error) {
	provider, ok := r.providers[sourceID]
	if !ok {
		return SearchResult{}, fmt.Errorf("source not found: %s", sourceID)
	}

	return provider.Search(query, page)
}

func (r *Registry) Detail(sourceID string, itemID string) (DetailResult, error) {
	provider, ok := r.providers[sourceID]
	if !ok {
		return DetailResult{}, fmt.Errorf("source not found: %s", sourceID)
	}

	detailProvider, ok := provider.(DetailProvider)
	if !ok {
		return DetailResult{}, fmt.Errorf("source %s does not support detail yet", sourceID)
	}

	return detailProvider.Detail(itemID)
}

func (r *Registry) Images(sourceID string, chapterID string) (ImageResult, error) {
	provider, ok := r.providers[sourceID]
	if !ok {
		return ImageResult{}, fmt.Errorf("source not found: %s", sourceID)
	}

	imageProvider, ok := provider.(ImageProvider)
	if !ok {
		return ImageResult{}, fmt.Errorf("source %s does not support chapter reading yet", sourceID)
	}

	return imageProvider.Images(chapterID)
}

func (r *Registry) Ranking(sourceID string, kind string, page int) (RankingResult, error) {
	provider, ok := r.providers[sourceID]
	if !ok {
		return RankingResult{}, fmt.Errorf("source not found: %s", sourceID)
	}

	rankingProvider, ok := provider.(RankingProvider)
	if !ok {
		return RankingResult{}, fmt.Errorf("source %s does not support ranking yet", sourceID)
	}

	return rankingProvider.Ranking(kind, page)
}
