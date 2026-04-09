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
	"ImageMaster/core/meta"
	sourceapi "ImageMaster/core/source"

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

	historyAPI := history.NewAPI(AppName)
	configAPI := config.NewAPI(AppName)
	libraryAPI := library.NewAPI(configAPI)
	extractAPI := archiveapi.NewAPI(configAPI)
	metaAPI := meta.NewAPI(AppVersion, BuildCommit, BuildTime)
	sourceAPI := sourceapi.NewAPI(configAPI, historyAPI.GetStore())
	crawlerAPI := crawlerapi.NewCrawlerAPI(configAPI, historyAPI.GetStore())

	err := wails.Run(&options.App{
		Title:  "ImageMaster",
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
			sourceAPI.SetContext(ctx)
		},
		Bind: []interface{}{
			libraryAPI,
			crawlerAPI,
			historyAPI,
			configAPI,
			extractAPI,
			metaAPI,
			sourceAPI,
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
