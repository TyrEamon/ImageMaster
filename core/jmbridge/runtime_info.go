package jmbridge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RuntimeInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Engine       string `json:"engine"`
	Upstream     string `json:"upstream"`
	BuildTime    string `json:"buildTime"`
	ManifestPath string `json:"manifestPath"`
	HelperPath   string `json:"helperPath"`
	Available    bool   `json:"available"`
	Source       string `json:"source"`
}

func GetRuntimeInfo() RuntimeInfo {
	info := loadManifest()
	if info.Name == "" {
		info.Name = "JM Runtime"
	}
	if info.Engine == "" {
		info.Engine = "jmcomic"
	}
	if info.Upstream == "" {
		info.Upstream = "hect0x7/JMComic-Crawler-Python"
	}

	cmd, err := resolveBridgeCommand()
	if err == nil && cmd != nil {
		info.Available = true
		info.HelperPath = cmd.Executable
		info.Source = cmd.Source
	}

	return info
}

func loadManifest() RuntimeInfo {
	for _, baseDir := range candidateBaseDirs() {
		manifestPath := filepath.Join(baseDir, "runtime", "runtime-manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var info RuntimeInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}

		info.ManifestPath = manifestPath
		info.Name = strings.TrimSpace(info.Name)
		info.Version = strings.TrimSpace(info.Version)
		info.Engine = strings.TrimSpace(info.Engine)
		info.Upstream = strings.TrimSpace(info.Upstream)
		info.BuildTime = normalizeBuildTime(info.BuildTime)
		return info
	}

	return RuntimeInfo{}
}

func normalizeBuildTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format(time.RFC3339)
	}

	return value
}
