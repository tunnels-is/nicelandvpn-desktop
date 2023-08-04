//go:build aix || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris

package core

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func GenerateBaseFolderPath() string {
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
	GLOBAL_STATE.BasePath = GenerateBaseFolderPath()
	GLOBAL_STATE.BackupPath = GLOBAL_STATE.BasePath

	_, err := os.Stat(GLOBAL_STATE.BasePath)
	if err != nil {
		err = os.Mkdir(GLOBAL_STATE.BasePath, 0777)
		if err != nil {
			CreateErrorLog("", "Unable to create base folder: ", err)
			return
		}
	}

	GLOBAL_STATE.BaseFolderInitialized = true
}

func AdminCheck() {

	CreateLog("loader", "Admin check")
	isAdmin := getProcessOwner()
	if isAdmin == "root" {
		GLOBAL_STATE.IsAdmin = true
	} else {
		GLOBAL_STATE.IsAdmin = false
		CreateErrorLog("", "nicelandVPN is not running as administrator, please restart as administartor")
	}
}

func getProcessOwner() string {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		// fmt.Println(err)
		return "X"
	}
	stdout = bytes.Replace(stdout, []byte{10}, nil, -1)
	return string(stdout)
}
