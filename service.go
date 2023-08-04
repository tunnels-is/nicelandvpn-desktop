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

func (s *Service) GetRoutersAndAccessPoints() (OUT *ReturnObject) {
	Data, code, err := core.GetRoutersAndAccessPoints()
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
