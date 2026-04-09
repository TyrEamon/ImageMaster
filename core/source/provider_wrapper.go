package source

import (
	"context"
	"fmt"
)

type wrappedProvider struct {
	provider Provider
	summary  Summary
}

func newWrappedProvider(provider Provider, summary Summary) Provider {
	return &wrappedProvider{
		provider: provider,
		summary:  summary,
	}
}

func (w *wrappedProvider) Summary() Summary {
	return w.summary
}

func (w *wrappedProvider) Search(query string, page int) (SearchResult, error) {
	result, err := w.provider.Search(query, page)
	if err != nil {
		return SearchResult{}, err
	}
	result.Source = w.summary
	return result, nil
}

func (w *wrappedProvider) Detail(itemID string) (DetailResult, error) {
	provider, ok := w.provider.(DetailProvider)
	if !ok {
		return DetailResult{}, fmt.Errorf("source %s does not support detail yet", w.summary.ID)
	}

	result, err := provider.Detail(itemID)
	if err != nil {
		return DetailResult{}, err
	}
	result.Source = w.summary
	return result, nil
}

func (w *wrappedProvider) Images(chapterID string) (ImageResult, error) {
	provider, ok := w.provider.(ImageProvider)
	if !ok {
		return ImageResult{}, fmt.Errorf("source %s does not support chapter reading yet", w.summary.ID)
	}

	result, err := provider.Images(chapterID)
	if err != nil {
		return ImageResult{}, err
	}
	result.Source = w.summary
	return result, nil
}

func (w *wrappedProvider) Ranking(kind string, page int) (RankingResult, error) {
	provider, ok := w.provider.(RankingProvider)
	if !ok {
		return RankingResult{}, fmt.Errorf("source %s does not support ranking yet", w.summary.ID)
	}

	result, err := provider.Ranking(kind, page)
	if err != nil {
		return RankingResult{}, err
	}
	result.Source = w.summary
	return result, nil
}

func (w *wrappedProvider) SetContext(ctx context.Context) {
	if withContext, ok := w.provider.(interface{ SetContext(context.Context) }); ok {
		withContext.SetContext(ctx)
	}
}
