package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {

	path, _ := os.Executable()
	fullPath := strings.Replace(path, "launcher", "NicelandVPN", 1)

	cmd := exec.Command("osascript", "-e", `do shell script "sudo '`+fullPath+`' > /dev/null 2>&1 &" with prompt "Start Niceland VPN with elevated privileges" with administrator privileges`)
	err := cmd.Start()

	// https://stackoverflow.com/questions/48856158/change-icon-of-notification-when-using-osascript-e-display-notification/49079025#49079025
	if err != nil {
		log.Println(err)
	}

	os.Exit(0)
}
