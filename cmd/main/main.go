package main

import (
	"embed"

	"github.com/tunnels-is/nicelandvpn-desktop/cmd/service"
)

//go:embed dist
var BlockLists embed.FS

func main() {
	service.Start()
}
