package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"ImageMaster/core/logger"
	"ImageMaster/core/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var _ types.ConfigProvider = (*Manager)(nil)
var _ types.ConfigManager = (*Manager)(nil)

var defaultConfig = Config{
	Libraries:             []string{},
	OutputDir:             "",
	ProxyURL:              "",
	EHentaiCookie:         "",
	BandizipPath:          "",
	SourceRepoURL:         "",
	JmCacheDir:            "",
	JmCacheRetentionHours: 24,
	JmCacheSizeLimitMB:    2048,
	ActiveLibrary:         "",
}

type Config struct {
	Libraries             []string `json:"libraries"`
	OutputDir             string   `json:"output_dir"`
	ProxyURL              string   `json:"proxy_url"`
	EHentaiCookie         string   `json:"ehentai_cookie"`
	BandizipPath          string   `json:"bandizip_path"`
	SourceRepoURL         string   `json:"source_repo_url"`
	JmCacheDir            string   `json:"jm_cache_dir"`
	JmCacheRetentionHours int      `json:"jm_cache_retention_hours"`
	JmCacheSizeLimitMB    int      `json:"jm_cache_size_limit_mb"`
	ActiveLibrary         string   `json:"active_library"`
}

type Manager struct {
	config     Config
	configPath string
	ctx        context.Context
}

func NewManager(configName string) *Manager {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir, _ = os.Getwd()
	}

	m := &Manager{
		config:     defaultConfig,
		configPath: filepath.Join(configDir, configName),
	}

	m.LoadConfig()
	return m
}

func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *Manager) LoadConfig() bool {
	data, err := os.ReadFile(m.configPath)
	logger.Debug("Loading config from: %s", m.configPath)
	if err != nil {
		logger.Warn("Failed to load config: %v, using default config", err)
		m.config = defaultConfig
		return false
	}

	if err := json.Unmarshal(data, &m.config); err != nil {
		logger.Error("Failed to parse config: %v, using default config", err)
		m.config = defaultConfig
		return false
	}

	m.normalizeConfig()

	logger.Debug("Config loaded successfully")
	return true
}

func (m *Manager) SaveConfig() bool {
	m.normalizeConfig()

	data, err := json.Marshal(m.config)
	if err != nil {
		logger.Error("Failed to marshal config: %v", err)
		return false
	}

	if err := os.WriteFile(m.configPath, data, 0o644); err != nil {
		logger.Error("Failed to save config: %v", err)
		return false
	}

	logger.Debug("Config saved successfully")
	return true
}

func (m *Manager) GetConfig() Config {
	return m.config
}

func (m *Manager) SetConfig(config Config) {
	m.config = config
}

func (m *Manager) GetLibraries() []string {
	return m.config.Libraries
}

func (m *Manager) SetActiveLibrary(library string) bool {
	m.config.ActiveLibrary = library
	logger.Info("Set active library: %s", library)
	return m.SaveConfig()
}

func (m *Manager) AddLibrary() bool {
	if m.ctx == nil {
		logger.Error("Cannot add library: missing Wails context")
		return false
	}

	dir, err := runtime.OpenDirectoryDialog(m.ctx, runtime.OpenDialogOptions{
		Title: "Select library directory",
	})
	if err != nil || dir == "" {
		return false
	}

	for _, lib := range m.config.Libraries {
		if lib == dir {
			logger.Warn("Library already exists: %s", dir)
			return false
		}
	}

	m.config.Libraries = append(m.config.Libraries, dir)
	logger.Info("Added library: %s", dir)
	return m.SaveConfig()
}

func (m *Manager) GetOutputDir() string {
	return m.config.OutputDir
}

func (m *Manager) SetOutputDir() bool {
	if m.ctx == nil {
		logger.Error("Cannot set output directory: missing Wails context")
		return false
	}

	dir, err := runtime.OpenDirectoryDialog(m.ctx, runtime.OpenDialogOptions{
		Title: "Select download directory",
	})
	if err != nil || dir == "" {
		return false
	}

	m.config.OutputDir = dir
	return m.SaveConfig()
}

func (m *Manager) GetActiveLibrary() string {
	return m.config.ActiveLibrary
}

func (m *Manager) SetProxy(proxyURL string) bool {
	m.config.ProxyURL = proxyURL
	logger.Debug("Set proxy: %s", proxyURL)
	return m.SaveConfig()
}

func (m *Manager) GetProxy() string {
	return m.config.ProxyURL
}

func (m *Manager) SetEHentaiCookie(cookie string) bool {
	m.config.EHentaiCookie = strings.TrimSpace(cookie)
	logger.Debug("Set EHentai cookie: %t", m.config.EHentaiCookie != "")
	return m.SaveConfig()
}

func (m *Manager) GetEHentaiCookie() string {
	return strings.TrimSpace(m.config.EHentaiCookie)
}

func (m *Manager) SetBandizipPath(path string) bool {
	m.config.BandizipPath = path
	logger.Debug("Set Bandizip path: %s", path)
	return m.SaveConfig()
}

func (m *Manager) GetBandizipPath() string {
	return m.config.BandizipPath
}

func (m *Manager) SetSourceRepoURL(rawURL string) bool {
	m.config.SourceRepoURL = rawURL
	logger.Debug("Set source repo url: %s", rawURL)
	return m.SaveConfig()
}

func (m *Manager) GetSourceRepoURL() string {
	return m.config.SourceRepoURL
}

func (m *Manager) SetJmCacheDir(path string) bool {
	m.config.JmCacheDir = path
	logger.Debug("Set JM cache dir: %s", path)
	return m.SaveConfig()
}

func (m *Manager) GetJmCacheDir() string {
	return m.config.JmCacheDir
}

func (m *Manager) SetJmCacheRetentionHours(hours int) bool {
	m.config.JmCacheRetentionHours = hours
	logger.Debug("Set JM cache retention hours: %d", hours)
	return m.SaveConfig()
}

func (m *Manager) GetJmCacheRetentionHours() int {
	if m.config.JmCacheRetentionHours <= 0 {
		return defaultConfig.JmCacheRetentionHours
	}
	return m.config.JmCacheRetentionHours
}

func (m *Manager) SetJmCacheSizeLimitMB(limit int) bool {
	m.config.JmCacheSizeLimitMB = limit
	logger.Debug("Set JM cache size limit MB: %d", limit)
	return m.SaveConfig()
}

func (m *Manager) GetJmCacheSizeLimitMB() int {
	if m.config.JmCacheSizeLimitMB <= 0 {
		return defaultConfig.JmCacheSizeLimitMB
	}
	return m.config.JmCacheSizeLimitMB
}

func (m *Manager) normalizeConfig() {
	if m.config.Libraries == nil {
		m.config.Libraries = []string{}
	}
	if m.config.JmCacheRetentionHours <= 0 {
		m.config.JmCacheRetentionHours = defaultConfig.JmCacheRetentionHours
	}
	if m.config.JmCacheSizeLimitMB <= 0 {
		m.config.JmCacheSizeLimitMB = defaultConfig.JmCacheSizeLimitMB
	}
}
