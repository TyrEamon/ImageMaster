package source

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type manifestIndex struct {
	Version string   `json:"version"`
	Sources []string `json:"sources"`
}

type SourceManifest struct {
	ID           string   `json:"id"`
	Adapter      string   `json:"adapter"`
	Provider     string   `json:"provider,omitempty"`
	Engine       string   `json:"engine,omitempty"`
	Script       string   `json:"script,omitempty"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Language     string   `json:"language"`
	Website      string   `json:"website"`
	Version      string   `json:"version"`
	Package      string   `json:"package,omitempty"`
	Author       string   `json:"author,omitempty"`
	License      string   `json:"license,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	NSFW         *bool    `json:"nsfw,omitempty"`
	Enabled      *bool    `json:"enabled"`
	BuiltIn      *bool    `json:"builtIn"`
	Capabilities []string `json:"capabilities"`
	RankingKinds []string `json:"rankingKinds"`
	Description  string   `json:"description"`
	Upstream     string   `json:"upstream,omitempty"`
	UpstreamURL  string   `json:"upstreamUrl,omitempty"`
	ScriptURL    string   `json:"scriptUrl,omitempty"`
	ImportedAt   string   `json:"importedAt,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	ManifestPath string   `json:"-"`
}

func loadSourceManifests() []SourceManifest {
	sourceDirs := resolveSourcesDirs()
	if len(sourceDirs) == 0 {
		return nil
	}

	manifestByID := map[string]SourceManifest{}
	for _, sourceDir := range sourceDirs {
		indexPath := filepath.Join(sourceDir, "index.json")
		manifestPaths := make([]string, 0, 8)

		if isFile(indexPath) {
			manifestPaths = append(manifestPaths, readManifestPathsFromIndex(sourceDir, indexPath)...)
		} else {
			manifestPaths = append(manifestPaths, scanManifestPaths(sourceDir)...)
		}

		for _, manifestPath := range manifestPaths {
			clean := filepath.Clean(strings.TrimSpace(manifestPath))
			if clean == "" {
				continue
			}

			manifest, ok := readManifest(clean)
			if !ok {
				continue
			}
			manifestByID[manifest.ID] = manifest
		}
	}

	manifests := make([]SourceManifest, 0, len(manifestByID))
	for _, manifest := range manifestByID {
		manifests = append(manifests, manifest)
	}
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ID < manifests[j].ID
	})

	return manifests
}

func resolveSourcesDirs() []string {
	candidates := []string{}
	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "sources"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "sources"))
	}
	candidates = append(candidates, ResolveUserSourcesDir())

	dirs := make([]string, 0, len(candidates))
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		clean := filepath.Clean(strings.TrimSpace(candidate))
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}

		if info, err := os.Stat(clean); err == nil && info.IsDir() {
			dirs = append(dirs, clean)
		}
	}

	return dirs
}

func ResolveUserSourcesDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			return filepath.Join(cwd, "user-sources")
		}
		return ""
	}

	return filepath.Join(configDir, "imagemaster-sources")
}

func readManifestPathsFromIndex(sourceDir string, indexPath string) []string {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil
	}
	data = normalizeJSONBytes(data)

	var index manifestIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil
	}

	paths := make([]string, 0, len(index.Sources))
	for _, item := range index.Sources {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		paths = append(paths, filepath.Join(sourceDir, item))
	}

	return paths
}

func scanManifestPaths(sourceDir string) []string {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") || strings.EqualFold(name, "index.json") {
			continue
		}
		paths = append(paths, filepath.Join(sourceDir, name))
	}

	sort.Strings(paths)
	return paths
}

func readManifest(path string) (SourceManifest, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceManifest{}, false
	}
	data = normalizeJSONBytes(data)

	var manifest SourceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return SourceManifest{}, false
	}

	manifest.ID = strings.TrimSpace(manifest.ID)
	manifest.Adapter = strings.TrimSpace(manifest.Adapter)
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Type = strings.TrimSpace(manifest.Type)
	manifest.Language = strings.TrimSpace(manifest.Language)
	manifest.Website = strings.TrimSpace(manifest.Website)
	manifest.Version = strings.TrimSpace(manifest.Version)
	manifest.Provider = strings.TrimSpace(manifest.Provider)
	manifest.Engine = strings.TrimSpace(manifest.Engine)
	manifest.Script = strings.TrimSpace(manifest.Script)
	manifest.Package = strings.TrimSpace(manifest.Package)
	manifest.Author = strings.TrimSpace(manifest.Author)
	manifest.License = strings.TrimSpace(manifest.License)
	manifest.Icon = strings.TrimSpace(manifest.Icon)
	manifest.Description = strings.TrimSpace(manifest.Description)
	manifest.Upstream = strings.TrimSpace(manifest.Upstream)
	manifest.UpstreamURL = strings.TrimSpace(manifest.UpstreamURL)
	manifest.ScriptURL = strings.TrimSpace(manifest.ScriptURL)
	manifest.ImportedAt = strings.TrimSpace(manifest.ImportedAt)
	manifest.Notes = strings.TrimSpace(manifest.Notes)
	manifest.ManifestPath = path

	if manifest.ID == "" {
		return SourceManifest{}, false
	}

	return manifest, true
}

func normalizeJSONBytes(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func mergeSummary(base Summary, manifest SourceManifest) Summary {
	summary := base

	if manifest.ID != "" {
		summary.ID = manifest.ID
	}
	if manifest.Name != "" {
		summary.Name = manifest.Name
	}
	if manifest.Type != "" {
		summary.Type = manifest.Type
	}
	if manifest.Language != "" {
		summary.Language = manifest.Language
	}
	if manifest.Website != "" {
		summary.Website = manifest.Website
	}
	if manifest.Version != "" {
		summary.Version = manifest.Version
	}
	if manifest.BuiltIn != nil {
		summary.BuiltIn = *manifest.BuiltIn
	}
	if len(manifest.Capabilities) > 0 {
		summary.Capabilities = append([]string{}, manifest.Capabilities...)
	}
	if len(manifest.RankingKinds) > 0 {
		summary.RankingKinds = append([]string{}, manifest.RankingKinds...)
	}
	if manifest.Description != "" {
		summary.Description = manifest.Description
	}

	return summary
}
