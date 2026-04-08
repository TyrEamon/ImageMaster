package meta

import "strings"

type VersionInfo struct {
	Version    string `json:"version"`
	Display    string `json:"display"`
	Commit     string `json:"commit"`
	BuildTime  string `json:"buildTime"`
	IsDevBuild bool   `json:"isDevBuild"`
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
