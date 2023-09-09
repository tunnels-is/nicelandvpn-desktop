package main

import (
	"log"
	"syscall"
	"unsafe"
)

func main() {

	// xlist, _ := process.Processes()

	// for i := range xlist {
	// 	// log.Println(xlist[i].Connections())
	// 	// xlist[i].
	// 	// ff, _ := xlist[i].OpenFiles()
	// 	// for ii := range ff {
	// 	// 	log.Println(ff[ii])

	// 	// }

	// 	cc, _ := xlist[i].Connections()
	// 	for ii := range cc {
	// 		log.Println(cc[ii])

	// 	}
	// }

	DLL := syscall.NewLazyDLL("Ws2_32.dll")
	DLL.Load()

	var host = make([]byte, 100)
	lenx := len(host)

	proc2 := DLL.NewProc("gethostname")
	r1, r2, err := syscall.SyscallN(
		proc2.Addr(),
		uintptr(unsafe.Pointer(&host[0])),
		uintptr(unsafe.Pointer(&lenx)),
	)

	log.Println(r1, r2, err)

	log.Println(string(host))

	// var mod = syscall.NewLazyDLL("user32.dll")
	// var proc = mod.NewProc("MessageBoxW")
	// var MB_YESNOCANCEL = 0x00000003

	// ret, _, _ := proc.Call(0,
	// 	uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("This test is Done."))),
	// 	uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Done Title"))),
	// 	uintptr(MB_YESNOCANCEL))
	// fmt.Printf("Return: %d\n", ret)
}
