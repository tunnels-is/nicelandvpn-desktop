package main

import (
	"context"
	"runtime/debug"
	"time"

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

	for {
		select {
		default:
			time.Sleep(500 * time.Millisecond)
		case ID := <-MONITOR:
			if ID == 1 {
				go core.StateMaintenance(MONITOR)
			} else if ID == 2 {
				go core.ReadFromRouterSocket(MONITOR)
			} else if ID == 4 {
				go core.ReadFromLocalSocket(MONITOR)
			} else if ID == 6 {
				go core.CalculateBandwidth(MONITOR)
			} else if ID == 8 {
				go core.StartLogQueueProcessor(MONITOR)
			} else if ID == 9 {
				go core.CleanPorts(MONITOR)
			}
		}
	}

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
