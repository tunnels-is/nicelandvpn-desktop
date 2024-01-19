package main

import (
	"context"
	"runtime/debug"

	"github.com/tunnels-is/nicelandvpn-desktop/cmd/service"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) IsProduction() bool {
	return PRODUCTION
}

func (a *App) startup(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			core.CreateErrorLog("", r, string(debug.Stack()))
			if a != nil {
				go a.startup(a.ctx)
			}
		}
	}()
	a.ctx = ctx

	service.Start()
}

func (a *App) shutdown(ctx context.Context) {
	core.CleanupOnClose()
}

func (a *App) CloseApp() {
	runtime.Quit(a.ctx)
}

func (a *App) CopyToClipboard(text string) {
	_ = runtime.ClipboardSetText(a.ctx, text)
}

func (a *App) SetTitle(title string) {
	runtime.WindowSetTitle(a.ctx, title)
}

func (a *App) OpenURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}
