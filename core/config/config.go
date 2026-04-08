package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"ImageMaster/core/logger"
	"ImageMaster/core/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var _ types.ConfigProvider = (*Manager)(nil)
var _ types.ConfigManager = (*Manager)(nil)

var defaultConfig = Config{
	Libraries:     []string{},
	OutputDir:     "",
	ProxyURL:      "",
	ActiveLibrary: "",
}

type Config struct {
	Libraries     []string `json:"libraries"`
	OutputDir     string   `json:"output_dir"`
	ProxyURL      string   `json:"proxy_url"`
	ActiveLibrary string   `json:"active_library"`
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

	logger.Debug("Config loaded successfully")
	return true
}

func (m *Manager) SaveConfig() bool {
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
