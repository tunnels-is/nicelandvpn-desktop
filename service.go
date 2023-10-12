package main

import (
	"github.com/tunnels-is/nicelandvpn-desktop/core"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Service struct {
	// ctx context.Context
}

func NewService() *Service {
	return &Service{}
}

type ReturnObject struct {
	Data interface{}
	Code int
	Err  string
}

func (s *Service) GetState() (OUT *ReturnObject) {
	core.PrepareState()
	return CreateReturnData(200, core.GLOBAL_STATE)
}

func CreateReturnData(code int, data interface{}) *ReturnObject {
	return &ReturnObject{
		Data: data,
		Code: code,
		Err:  "",
	}
}

func CreateReturnError(code int, message string) *ReturnObject {
	return &ReturnObject{
		Data: nil,
		Code: code,
		Err:  message,
	}
}

func (s *Service) SwitchRouter(Tag string) (OUT *ReturnObject) {
	code, err := core.SwitchRouter(Tag)
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, nil)
	return
}

func (s *Service) Connect(NS *core.CONTROLLER_SESSION_REQUEST) (OUT *ReturnObject) {
	Data, code, err := core.Connect(NS, true)
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, Data)
	return
}

func (s *Service) Switch(NS *core.CONTROLLER_SESSION_REQUEST) (OUT *ReturnObject) {
	Data, code, err := core.Connect(NS, false)
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, Data)
	return
}

func (s *Service) Disconnect() (OUT *ReturnObject) {
	core.Disconnect()
	OUT = CreateReturnData(200, nil)
	return
}

func (s *Service) ResetEverything() (OUT *ReturnObject) {
	core.ResetEverything()
	OUT = CreateReturnData(200, nil)
	return
}

func (s *Service) SetConfig(FORM *core.CONFIG_FORM) (OUT *ReturnObject) {
	err := core.SetConfig(FORM)
	if err != nil {
		OUT = CreateReturnError(400, err.Error())
		return
	}
	OUT = CreateReturnData(200, nil)
	return
}

func (s *Service) GetQRCode(FORM *core.TWO_FACTOR_CONFIRM) (OUT *ReturnObject) {
	QR, err := core.GetQRCode(FORM)
	if err != nil {
		OUT = CreateReturnError(400, err.Error())
		return
	}
	OUT = CreateReturnData(200, QR)
	return
}

func (s *Service) GetLoadingLogs(Type string) (OUT *ReturnObject) {
	Logs, err := core.GetLoadingLogs(Type)
	if err != nil {
		OUT = CreateReturnError(400, err.Error())
		return
	}
	OUT = CreateReturnData(200, Logs)
	return
}

func (s *Service) GetLogs(lengthFromJavascript int) (OUT *ReturnObject) {
	Logs, err := core.GetLogs(lengthFromJavascript)
	if err != nil {
		OUT = CreateReturnError(400, err.Error())
		return
	}
	OUT = CreateReturnData(200, Logs)
	return
}

func (s *Service) LoadRoutersUnAuthenticated() (OUT *ReturnObject) {
	Data, code, err := core.LoadRoutersUnAuthenticated()
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, Data)
	return
}

func (s *Service) GetRoutersAndAccessPoints(FR *core.FORWARD_REQUEST) (OUT *ReturnObject) {
	Data, code, err := core.GetRoutersAndAccessPoints(FR)
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, Data)
	return
}

func (s *Service) ForwardToController(FR *core.FORWARD_REQUEST) (OUT *ReturnObject) {
	Data, code, err := core.ForwardToController(FR)
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, Data)
	return
}

func (s *Service) ForwardToRouter(FR *core.FORWARD_REQUEST) (OUT *ReturnObject) {
	Data, code, err := core.ForwardToRouter(FR)
	if err != nil {
		OUT = CreateReturnError(code, err.Error())
		return
	}
	OUT = CreateReturnData(code, Data)
	return
}

func (s *Service) OpenFileDialogForRouterFile(clearFile bool) string {
	core.CreateLog("START", "")

	path := ""
	var err error
	if !clearFile {
		path, err = runtime.OpenFileDialog(APP.ctx, runtime.OpenDialogOptions{
			ShowHiddenFiles:      true,
			CanCreateDirectories: false,
		})
		if err != nil {
			return ""
		}
	}

	err = core.SetRouterFile(path)
	if err != nil {
		core.CreateErrorLog("loader", "Unable to update router file: ", err.Error())
		return ""
	}

	return path
}

func (s *Service) EnableDNSWhitelist() {
	core.GLOBAL_STATE.DNSWhitelistEnabled = true
}

func (s *Service) DisableDNSWhitelist() {
	core.GLOBAL_STATE.DNSWhitelistEnabled = false
}

func (s *Service) StartDNSCapture() {
	core.StartCapturing()
	return
}

func (s *Service) StopDNSCapture() string {
	core.CreateLog("START", "")

	path := ""
	var err error
	path, err = runtime.SaveFileDialog(APP.ctx, runtime.SaveDialogOptions{
		Title:                "Create A File",
		DefaultFilename:      "allowed_websites",
		ShowHiddenFiles:      true,
		CanCreateDirectories: true,
	})

	if err != nil {
		core.CreateErrorLog("", "Unable to save capture file", err.Error())
		return err.Error()
	}

	err = core.StopCapturing(path)
	if err != nil {
		core.CreateErrorLog("", "Unable to save capture file", err.Error())
		return err.Error()
	}

	return ""
}

func (s *Service) RebuildDomainBlocklist() {
	core.BuildDomainBlocklist()

	core.C.EnabledBlockLists = make([]string, 0)
	for i := range core.GLOBAL_STATE.BLists {
		if core.GLOBAL_STATE.BLists[i].Enabled {
			core.C.EnabledBlockLists = append(core.C.EnabledBlockLists, core.GLOBAL_STATE.BLists[i].Tag)
		}
	}

	_ = core.SaveConfig()
}
func (s *Service) DisableAllBlocklists() {

	for i := range core.GLOBAL_STATE.BLists {
		core.GLOBAL_STATE.BLists[i].Enabled = false
	}

}

func (s *Service) EnableAllBlocklists() {

	for i := range core.GLOBAL_STATE.BLists {
		core.GLOBAL_STATE.BLists[i].Enabled = true
	}

}

func (s *Service) DisableBlocklist(tag string) {

	for i := range core.GLOBAL_STATE.BLists {
		if core.GLOBAL_STATE.BLists[i].Tag == tag {
			core.GLOBAL_STATE.BLists[i].Enabled = false
		}
	}

}

func (s *Service) EnableBlocklist(tag string) {

	for i := range core.GLOBAL_STATE.BLists {
		if core.GLOBAL_STATE.BLists[i].Tag == tag {
			core.GLOBAL_STATE.BLists[i].Enabled = true
		}
	}

}
