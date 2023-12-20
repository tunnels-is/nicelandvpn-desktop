package main

import (
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

var (
	user *core.User
	PAFR core.FORWARD_REQUEST
)

var (
	app_state_str = [10]string{
		"VPN List Update", "Ready to Connect", "Version", "VPN Tunnel Ready",
		"Launched as Admin", "Config Loaded", "Base Folder Created",
		"Log File Created", "Buffer Error", "Launch Error",
	}
	interface_str     = [3]string{"Name", "IPv6 Enabled", "Gateway"}
	connection_str    = [3]string{"Entr Router", "ms", "QoS"}
	network_stats_str = [5]string{"Connected", "Download", "Packets", "Upload", "Packets"}
)
