package main

import (
	"embed"
	"os"
	"strings"

	"github.com/tunnels-is/nicelandvpn-desktop/core"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

const VERSION = "1.1.3"
const PRODUCTION = true
const ENABLE_INSTERFACE = true

var MONITOR = make(chan int, 200)

var APP = NewApp()
var SERVICE = NewService()

func main() {

	var WebViewGPUPolicy = 0
	for i := range os.Args {
		argToLower := strings.ToLower(os.Args[i])
		if argToLower == "-disablegpu" {
			WebViewGPUPolicy = 2
		}
	}

	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INSTERFACE
	core.GLOBAL_STATE.Version = VERSION

	go core.StartService(MONITOR)

	logger := new(core.LoggerInterface)

	err := wails.Run(&options.App{
		Title: "Niceland VPN",

		Width:  1050,
		Height: 650,

		Frameless:     false,
		Fullscreen:    false,
		AlwaysOnTop:   false,
		DisableResize: false,

		Logger: logger,
		Windows: &windows.Options{
			// Do not enable zoom controls
			// They are kind of broken inside webview and cause bad scaling
			IsZoomControlEnabled: false,

			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewGpuIsDisabled:              false,
		},

		Linux: &linux.Options{
			WebviewGpuPolicy:    linux.WebviewGpuPolicy(WebViewGPUPolicy),
			WindowIsTranslucent: false,
		},

		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 true,
				HideToolbarSeparator:       false,
			},
			Appearance: "NSAppearanceNameDarkAqua",
			About: &mac.AboutInfo{
				Title:   "Niceland",
				Message: "Support: support@nicelandvpn.is",
			},
		},

		CSSDragProperty: "--wails-draggable",
		CSSDragValue:    "drag",

		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		OnStartup:  APP.startup,
		OnShutdown: APP.shutdown,

		Bind: []interface{}{
			APP,
			SERVICE,
		},

		Debug: options.Debug{
			OpenInspectorOnStartup: !PRODUCTION,
		},
	})

	if err != nil {
		if core.LogFile != nil {
			_, _ = core.LogFile.WriteString("Unable to start the GUI(wails.io): " + err.Error())
			core.CreateErrorLog("", "Unable to start the GUI(wails.io): ", err.Error())
		}
	}
}
