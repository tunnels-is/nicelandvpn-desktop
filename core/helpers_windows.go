//go:build windows

package core

import (
	"os"
	"path/filepath"
)

func GenerateBaseFolderPath() string {
	defer RecoverAndLogToFile()

	base := "."
	ex, err := os.Executable()
	if err != nil {
		CreateErrorLog("loader", "Unable to find working directory: ", err.Error())
	} else {
		base = filepath.Dir(ex)
	}

	return base + string(os.PathSeparator) + "files" + string(os.PathSeparator)
}

func CreateBaseFolder() {
	defer RecoverAndLogToFile()

	CreateLog("loader", "Verifying configurations and logging folder")

	_, err := os.Stat(GLOBAL_STATE.BasePath)
	if err != nil {
		err = os.Mkdir(GLOBAL_STATE.BasePath, 0777)
		if err != nil {
			GLOBAL_STATE.ClientStartupError = true
			CreateErrorLog("", "Unable to create base folder: ", err)
			return
		}
	}

	GLOBAL_STATE.BaseFolderInitialized = true
}

// https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func AdminCheck() {
	defer RecoverAndLogToFile()

	CreateLog("loader", "Admin check")
	fd, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		GLOBAL_STATE.IsAdmin = false
		CreateErrorLog("", "nicelandVPN is not running as administrator, please restart as administartor")
		return
	}

	GLOBAL_STATE.IsAdmin = true
	_ = fd.Close()
}
