package core

import (
	"errors"
	"os"
)

var DNSCaptureFile *os.File
var DNSCapturMap = make(map[string]bool)

func IsDomainAllowed(domain string) bool {

	_, ok := DNSWhitelist[domain]
	if ok {
		return true
	} else {
		return false
	}
}

func CaptureDNS(domain string) {
	CreateLog("CAPTURE", "CAPTURED: ", domain)
	DNSCapturMap[domain] = true
	return
}

func StartCapturing() {
	GLOBAL_STATE.DNSCaptureEnabled = true
	return
}

func StopCapturing(path string) (err error) {
	defer RecoverAndLogToFile()

	GLOBAL_STATE.DNSCaptureEnabled = false

	DNSCaptureFile, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	if DNSCaptureFile == nil {
		return errors.New("Could not init domain capture file")
	}

	for i := range DNSCapturMap {
		DNSCaptureFile.WriteString(i + "\n")
	}

	_ = DNSCaptureFile.Sync()
	_ = DNSCaptureFile.Close()

	C.DomainWhitelist = path
	DNSWhitelist = DNSCapturMap

	err = SaveConfig()
	if err != nil {
		CreateErrorLog("", "Unable to save config: ", err)
		return err
	}
	return
}
