package main

import (
	"context"
	"embed"
	"log"

	archiveapi "ImageMaster/core/archive"
	"ImageMaster/core/config"
	crawlerapi "ImageMaster/core/crawler/api"
	"ImageMaster/core/history"
	"ImageMaster/core/library"
	appLogger "ImageMaster/core/logger"

	"github.com/wailsapp/wails/v2"
	wlogger "github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:front/dist
var assets embed.FS

const AppName = "imagemaster"

func main() {
	defer appLogger.Recover("main")

	_ = appLogger.Init(appLogger.FileConfig{
		Filename:    "",
		MaxSizeMB:   50,
		MaxBackups:  5,
		MaxAgeDays:  14,
		Compress:    true,
		WriteStdout: true,
	})

	// 创建历史记录API
	historyAPI := history.NewAPI(AppName)

	// 获取配置管理器
	configAPI := config.NewAPI(AppName)

	// 创建图书馆API
	libraryAPI := library.NewAPI(configAPI)

	// 创建解压管理API
	extractAPI := archiveapi.NewAPI(configAPI)

	// 创建爬虫API（构造注入历史存储）
	crawlerAPI := crawlerapi.NewCrawlerAPI(configAPI, historyAPI.GetStore())

	// 创建应用
	err := wails.Run(&options.App{
		Title:  "漫画查看器",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			configAPI.SetContext(ctx)
			libraryAPI.SetContext(ctx)
			libraryAPI.InitializeLibraryManager()
			crawlerAPI.SetContext(ctx)
		},
		Bind: []interface{}{
			libraryAPI,
			crawlerAPI,
			historyAPI,
			configAPI,
			extractAPI,
			appLogger.NewAPI(),
		},
		LogLevel:                 wlogger.ERROR,
		LogLevelProduction:       wlogger.ERROR,
		EnableDefaultContextMenu: true,
	})

	if err != nil {
		log.Fatal(err)
	}
}
