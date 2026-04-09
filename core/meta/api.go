package meta

import (
	"strings"

	"ImageMaster/core/jmbridge"
)

type VersionInfo struct {
	Version    string `json:"version"`
	Display    string `json:"display"`
	Commit     string `json:"commit"`
	BuildTime  string `json:"buildTime"`
	IsDevBuild bool   `json:"isDevBuild"`
}

type JmRuntimeInfo struct {
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

type API struct {
	version   string
	commit    string
	buildTime string
}

func NewAPI(version string, commit string, buildTime string) *API {
	return &API{
		version:   strings.TrimSpace(version),
		commit:    strings.TrimSpace(commit),
		buildTime: strings.TrimSpace(buildTime),
	}
}

func (a *API) GetVersionInfo() VersionInfo {
	version := a.version
	if version == "" {
		version = "0.0.0-dev"
	}

	display := version
	if !strings.HasPrefix(strings.ToLower(display), "v") {
		display = "v" + display
	}

	commit := a.commit
	if commit == "" {
		commit = "local"
	}

	return VersionInfo{
		Version:    version,
		Display:    display,
		Commit:     commit,
		BuildTime:  a.buildTime,
		IsDevBuild: strings.Contains(strings.ToLower(version), "dev"),
	}
}

func (a *API) GetJmRuntimeInfo() JmRuntimeInfo {
	info := jmbridge.GetRuntimeInfo()

	version := strings.TrimSpace(info.Version)
	if version == "" {
		version = "unversioned"
	}

	return JmRuntimeInfo{
		Name:         info.Name,
		Version:      version,
		Engine:       info.Engine,
		Upstream:     info.Upstream,
		BuildTime:    info.BuildTime,
		ManifestPath: info.ManifestPath,
		HelperPath:   info.HelperPath,
		Available:    info.Available,
		Source:       info.Source,
	}
}
