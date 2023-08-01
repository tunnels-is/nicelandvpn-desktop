//go:build darwin

package core

import (
	"os"
)

func GenerateBaseFolderPath() string {
	return GetPWD() + string(os.PathSeparator) + "nicelandvpn" + string(os.PathSeparator)
}

func CreateBaseFolder() {
	GLOBAL_STATE.BasePath = GenerateBaseFolderPath()
	GLOBAL_STATE.BackupPath = GLOBAL_STATE.BasePath

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

func GetPWD() string {
	HOMEPATH := os.Getenv("HOME")
	if HOMEPATH == "" {
		HOMEPATH = "/tmp"
	}
	return HOMEPATH
}

func AdminCheck() error {
	CreateLog("loader", "Admin check")
	GLOBAL_STATE.IsAdmin = true
	return nil
}
