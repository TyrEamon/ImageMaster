package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type RepositorySyncResult struct {
	RepoURL              string `json:"repoUrl"`
	IndexURL             string `json:"indexUrl"`
	LocalDir             string `json:"localDir"`
	ManifestCount        int    `json:"manifestCount"`
	ScriptCount          int    `json:"scriptCount"`
	EnabledManifestCount int    `json:"enabledManifestCount"`
	UpdatedAt            string `json:"updatedAt"`
}

type SourceStorageInfo struct {
	BundledDir string `json:"bundledDir"`
	UserDir    string `json:"userDir"`
}

func GetSourceStorageInfo() SourceStorageInfo {
	info := SourceStorageInfo{
		UserDir: ResolveUserSourcesDir(),
	}

	for _, dir := range resolveSourcesDirs() {
		if strings.EqualFold(filepath.Clean(dir), filepath.Clean(info.UserDir)) {
			continue
		}
		info.BundledDir = dir
		break
	}

	return info
}

func SyncSourceRepositoryFromRemote(repoURL string) (RepositorySyncResult, error) {
	indexURL, err := normalizeSourceRepoIndexURL(repoURL)
	if err != nil {
		return RepositorySyncResult{}, err
	}

	userDir := ResolveUserSourcesDir()
	if userDir == "" {
		return RepositorySyncResult{}, fmt.Errorf("failed to resolve local user sources directory")
	}
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return RepositorySyncResult{}, fmt.Errorf("create user sources directory: %w", err)
	}

	indexData, err := fetchRemoteFile(indexURL)
	if err != nil {
		return RepositorySyncResult{}, err
	}

	var index manifestIndex
	if err := json.Unmarshal(indexData, &index); err != nil {
		return RepositorySyncResult{}, fmt.Errorf("parse remote source index: %w", err)
	}

	baseURL, err := url.Parse(indexURL)
	if err != nil {
		return RepositorySyncResult{}, fmt.Errorf("parse repo index url: %w", err)
	}

	scriptPaths := map[string]struct{}{}
	manifestCount := 0
	enabledCount := 0

	for _, sourcePath := range index.Sources {
		trimmedPath := filepath.ToSlash(strings.TrimSpace(sourcePath))
		if trimmedPath == "" {
			continue
		}

		manifestURL := resolveRemotePath(baseURL, trimmedPath)
		manifestData, err := fetchRemoteFile(manifestURL)
		if err != nil {
			return RepositorySyncResult{}, fmt.Errorf("download manifest %s: %w", trimmedPath, err)
		}

		localManifestPath := filepath.Join(userDir, filepath.FromSlash(trimmedPath))
		if err := os.MkdirAll(filepath.Dir(localManifestPath), 0o755); err != nil {
			return RepositorySyncResult{}, fmt.Errorf("prepare manifest directory %s: %w", trimmedPath, err)
		}
		if err := os.WriteFile(localManifestPath, manifestData, 0o644); err != nil {
			return RepositorySyncResult{}, fmt.Errorf("write manifest %s: %w", trimmedPath, err)
		}
		manifestCount++

		var manifest SourceManifest
		if err := json.Unmarshal(manifestData, &manifest); err == nil {
			if manifest.Enabled == nil || *manifest.Enabled {
				enabledCount++
			}

			scriptRel := strings.TrimSpace(manifest.Script)
			if scriptRel != "" {
				remoteScriptURL := resolveRemotePathFromManifestURL(manifestURL, scriptRel)
				localScriptPath := filepath.Clean(filepath.Join(filepath.Dir(localManifestPath), filepath.FromSlash(scriptRel)))
				if _, exists := scriptPaths[localScriptPath]; !exists {
					scriptData, scriptErr := fetchRemoteFile(remoteScriptURL)
					if scriptErr != nil {
						return RepositorySyncResult{}, fmt.Errorf("download script %s: %w", scriptRel, scriptErr)
					}
					if err := os.MkdirAll(filepath.Dir(localScriptPath), 0o755); err != nil {
						return RepositorySyncResult{}, fmt.Errorf("prepare script directory %s: %w", scriptRel, err)
					}
					if err := os.WriteFile(localScriptPath, scriptData, 0o644); err != nil {
						return RepositorySyncResult{}, fmt.Errorf("write script %s: %w", scriptRel, err)
					}
					scriptPaths[localScriptPath] = struct{}{}
				}
			}
		}
	}

	localIndexPath := filepath.Join(userDir, "index.json")
	if err := os.WriteFile(localIndexPath, indexData, 0o644); err != nil {
		return RepositorySyncResult{}, fmt.Errorf("write local source index: %w", err)
	}

	return RepositorySyncResult{
		RepoURL:              strings.TrimSpace(repoURL),
		IndexURL:             indexURL,
		LocalDir:             userDir,
		ManifestCount:        manifestCount,
		ScriptCount:          len(scriptPaths),
		EnabledManifestCount: enabledCount,
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}, nil
}

func normalizeSourceRepoIndexURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("source repository url is empty")
	}

	if strings.Contains(trimmed, "raw.githubusercontent.com") && strings.HasSuffix(strings.ToLower(trimmed), "index.json") {
		return trimmed, nil
	}

	if strings.HasPrefix(trimmed, "https://github.com/") || strings.HasPrefix(trimmed, "http://github.com/") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("parse github repository url: %w", err)
		}

		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 2 {
			owner := parts[0]
			repo := parts[1]
			branch := "main"
			return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/sources/index.json", owner, repo, branch), nil
		}
	}

	if strings.HasSuffix(strings.ToLower(trimmed), "index.json") {
		return trimmed, nil
	}

	return strings.TrimRight(trimmed, "/") + "/index.json", nil
}

func fetchRemoteFile(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "ImageMaster/0.2")
	req.Header.Set("Accept", "application/json,text/plain,*/*")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("request failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return io.ReadAll(resp.Body)
}

func resolveRemotePath(indexURL *url.URL, relativePath string) string {
	next := *indexURL
	next.Path = path.Join(path.Dir(next.Path), filepath.ToSlash(strings.TrimSpace(relativePath)))
	next.RawQuery = ""
	next.Fragment = ""
	return next.String()
}

func resolveRemotePathFromManifestURL(manifestURL string, relativePath string) string {
	baseURL, err := url.Parse(strings.TrimSpace(manifestURL))
	if err != nil {
		return strings.TrimSpace(relativePath)
	}
	targetURL, err := url.Parse(strings.TrimSpace(relativePath))
	if err != nil {
		return strings.TrimSpace(relativePath)
	}
	return baseURL.ResolveReference(targetURL).String()
}
