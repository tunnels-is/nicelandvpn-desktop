package main

/*
#include <stdlib.h>
int requestElevatedPrivileges();
*/
import "C"

import (
	// "fmt"
	"log"
	// "os"
	// "os/exec"
	// "syscall"
)

// func main() {

// 	path, _ := os.Executable()
// 	fullPath := strings.Replace(path, "launcher", "NicelandVPN", 1)

// 	cmd := exec.Command("osascript", "-e", `do shell script "sudo '`+fullPath+`' > /dev/null 2>&1 &" with prompt "Start Niceland VPN with elevated privileges" with administrator privileges`)
// 	err := cmd.Start()

// 	// https://stackoverflow.com/questions/48856158/change-icon-of-notification-when-using-osascript-e-display-notification/49079025#49079025
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	os.Exit(0)
// }

func main() {

}

func elevate() {
	// Check if the program is already running with elevated privileges
	// if syscall.Geteuid() == 0 {
	// 	// Your code that requires elevated privileges goes here
	// 	fmt.Println("Running with elevated privileges")
	// } else {
	// Request elevated privileges
	status := C.requestElevatedPrivileges()
	log.Println("ELEVATE STATUS:", status)
	log.Println("ELEVATE STATUS:", status)
	log.Println("ELEVATE STATUS:", status)
	log.Println("ELEVATE STATUS:", status)
	log.Println("ELEVATE STATUS:", status)
	log.Println("ELEVATE STATUS:", status)

	// If privileges were obtained successfully, execute the program again with elevated privileges
	// cmd := exec.Command("networksetup")
	// cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 0, Gid: 0}, Setpgid: true, Setsid: true, Foreground: true}
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// cmd.Stdin = os.Stdin
	// argv0 := C.CString("fdslkfjskld")
	// defer C.free(unsafe.Pointer(argv0))
	// C.system(argv0)
	// C.fflush(C.stdout)
	FP := C.ExecuteStuff()
	log.Println(FP)
	// log.Println(C.system(argv0))

	// if err := cmd.Start(); err != nil {
	// 	fmt.Println("Error running with elevated privileges:", err)
	// 	os.Exit(1)
	// }
}
