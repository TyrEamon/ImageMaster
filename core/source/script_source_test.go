package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScriptSourceBasicFlow(t *testing.T) {
	tempDir := t.TempDir()
	scriptDir := filepath.Join(tempDir, "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}

	scriptPath := filepath.Join(scriptDir, "demo.js")
	scriptBody := `
var source = {
  search: function(query, page, ctx) {
    return {
      query: query,
      page: page,
      hasMore: false,
      total: 1,
      items: [{
        id: "item-1",
        title: "Demo " + query,
        cover: "",
        summary: "summary",
        primaryLabel: "demo",
        secondaryLabel: "page-" + page,
        detailUrl: ctx.resolveURL("https://example.com", "/detail/item-1")
      }]
    };
  },
  detail: function(itemID, ctx) {
    return {
      item: {
        id: itemID,
        title: "Detail " + itemID,
        cover: "",
        summary: "detail summary",
        author: "tester",
        status: "ongoing",
        tags: ["demo"],
        detailUrl: "https://example.com/detail/" + itemID,
        chapters: [{
          id: "chapter-1",
          name: "Chapter 1",
          url: "chapter-1",
          index: 0,
          updatedLabel: "latest"
        }]
      }
    };
  },
  images: function(chapterID, ctx) {
    return {
      comicTitle: "Comic",
      chapterTitle: chapterID,
      chapterUrl: "https://example.com/read/" + chapterID,
      images: ["https://example.com/1.jpg"],
      entries: [{ url: "https://example.com/1.jpg" }],
      hasNext: false,
      nextUrl: ""
    };
  },
  ranking: function(kind, page, ctx) {
    return {
      kind: kind,
      page: page,
      total: 1,
      items: [{
        id: "rank-1",
        title: "Ranked",
        cover: "",
        summary: "rank summary",
        primaryLabel: "demo",
        secondaryLabel: kind,
        detailUrl: "https://example.com/rank-1"
      }]
    };
  }
};`
	if err := os.WriteFile(scriptPath, []byte(scriptBody), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	manifestPath := filepath.Join(tempDir, "demo.json")
	manifest := SourceManifest{
		ID:           "demo-source",
		Engine:       "script",
		Script:       "scripts/demo.js",
		Name:         "Demo Source",
		Type:         "manga",
		Language:     "all",
		Website:      "https://example.com",
		Version:      "0.1.0",
		Capabilities: []string{CapabilitySearch, CapabilityDetail, CapabilityRead, CapabilityRanking},
		RankingKinds: []string{"popular"},
		Description:  "Test source",
		ManifestPath: manifestPath,
	}

	provider, err := NewScriptSource(manifest)
	if err != nil {
		t.Fatalf("new script source: %v", err)
	}

	searchResult, err := provider.Search("demo", 2)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if searchResult.Total != 1 || len(searchResult.Items) != 1 {
		t.Fatalf("unexpected search result: %+v", searchResult)
	}

	detailResult, err := provider.Detail("item-1")
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detailResult.Item.Title != "Detail item-1" {
		t.Fatalf("unexpected detail title: %+v", detailResult.Item)
	}

	imageResult, err := provider.Images("chapter-1")
	if err != nil {
		t.Fatalf("images: %v", err)
	}
	if len(imageResult.Entries) != 1 || imageResult.ChapterTitle != "chapter-1" {
		t.Fatalf("unexpected image result: %+v", imageResult)
	}

	rankingResult, err := provider.Ranking("popular", 1)
	if err != nil {
		t.Fatalf("ranking: %v", err)
	}
	if rankingResult.Kind != "popular" || len(rankingResult.Items) != 1 {
		t.Fatalf("unexpected ranking result: %+v", rankingResult)
	}
}
