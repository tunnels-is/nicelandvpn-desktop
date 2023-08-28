package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/go-ping/ping"
)

func StartService(MONITOR chan int) {
	defer RecoverAndLogToFile()

	CreateLog("loader", "Starting Niceland VPN")

	C = new(Config)
	C.DebugLogging = true
	GLOBAL_STATE.NeedsRouterProbe = true

	AdminCheck()
	CreateBaseFolder()
	InitLogfile()
	LoadConfig()

	go StartLogQueueProcessor(MONITOR)
	go StateMaintenance(MONITOR)
	go ReadFromLocalTunnel(MONITOR)
	go ReadFromRouterSocket(MONITOR)
	go CalculateBandwidth(MONITOR)

	CreateLog("loader", "Niceland is ready")
	CreateLog("START", "")
}

func CalculateBandwidth(MONITOR chan int) {
	defer func() {
		MONITOR <- 6
	}()

	defer RecoverAndLogToFile()

	for {
		time.Sleep(980 * time.Millisecond)

		GLOBAL_STATE.UMbps = CURRENT_UBBS * 8
		GLOBAL_STATE.DMbps = CURRENT_DBBS * 8
		CURRENT_DBBS = 0
		CURRENT_UBBS = 0
	}

}

func StateMaintenance(MONITOR chan int) {
	defer RecoverAndLogToFile()
	defer func() {
		time.Sleep(10 * time.Second)
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 1
		}
	}()
	defer RecoverAndLogToFile()

	CleanTCPPorts()
	CleanUDPPorts()

	InterfaceMaintenenceAndBackup()

	if GLOBAL_STATE.NeedsRouterProbe {
		CreateLog("loader", "Starting router probe")
		err := RefreshRouterList()
		if err != nil {
			CreateErrorLog("", "Unable to find the best router for your connection: ", err)
		} else {
			GLOBAL_STATE.NeedsRouterProbe = false
		}
	}

	if ENABLE_INSTERFACE {
		if !GLOBAL_STATE.TunnelInitialized {
			CreateLog("loader", "Preparing the VPN interface")
			err := LaunchPreperation()
			if err != nil {
				CreateErrorLog("", "Could not initialize tunnel tap interface: ", err)
			} else {
				GLOBAL_STATE.TunnelInitialized = true
			}
		}

		if GLOBAL_STATE.TunnelInitialized {
			if !GLOBAL_STATE.NeedsRouterProbe {
				if GLOBAL_STATE.DefaultInterface != nil {
					GLOBAL_STATE.ClientReady = true
				}
			}
		}

	} else {
		GLOBAL_STATE.ClientReady = true
	}

	if AS.TCPTunnelSocket != nil {

		_, err := AS.TCPTunnelSocket.Write([]byte{0, 1})
		if err != nil {
			CreateErrorLog("", "Ping to VPN failed, Disconnecting.")

			if AS.TCPTunnelSocket != nil {
				_ = AS.TCPTunnelSocket.Close()
				AS.TCPTunnelSocket = nil
			}

			SetGlobalStateAsDisconnected()
			var connected bool = false
			if C.AutoReconnect {
				connected = AutoReconnect()
			}

			if !connected {
				if !C.KillSwitch {
					CleanupWithStateLock()
				}
			}

			return
		}

		if time.Since(GLOBAL_STATE.PingReceivedFromRouter).Seconds() > 29 {
			CreateErrorLog("", "VPN has not responded in the last 30 seconds, disconnecting.")
			GLOBAL_STATE.Connected = false

			if AS.TCPTunnelSocket != nil {
				_ = AS.TCPTunnelSocket.Close()
				AS.TCPTunnelSocket = nil
			}

			var connected bool = false
			if C.AutoReconnect {
				connected = AutoReconnect()
			}

			if !connected {
				if !C.KillSwitch {
					CleanupWithStateLock()
				}
			}
		}

	} else {
		if C.AutoReconnect {
			_ = AutoReconnect()
		}
	}

	if BUFFER_ERROR {
		// if AS.TCPTunnelSocket != nil {
		// 	_ = AS.TCPTunnelSocket.Close()
		// 	AS.TCPTunnelSocket = nil
		// }

		SetGlobalStateAsDisconnected()
		BUFFER_ERROR = false
		_ = AutoReconnect()
	}

}

func AutoReconnect() (connected bool) {
	defer RecoverAndLogToFile()

	connectingStateChangedLocally := false
	defer func() {
		if connectingStateChangedLocally {
			GLOBAL_STATE.Connecting = false
		}
		STATE_LOCK.Unlock()
	}()
	STATE_LOCK.Lock()

	if !C.AutoReconnect {
		return false
	}

	if C.PrevSession == nil {
		return false
	}

	if GLOBAL_STATE.Connected || GLOBAL_STATE.Connecting || GLOBAL_STATE.Exiting {
		return false
	}

	if time.Since(LastConnectionAttemp).Seconds() < 5 {
		return false
	}

	GLOBAL_STATE.Connecting = true
	connectingStateChangedLocally = true

	LastConnectionAttemp = time.Now()
	CreateLog("", "Automatic reconnect..")

	s, _, err := ConnectToAccessPoint(C.PrevSession, true)
	if s == nil || err != nil {
		CreateErrorLog("", "Auto reconnect failed")
		return false
	}

	GLOBAL_STATE.LastRouterPing = time.Now()
	CreateLog("", "Auto reconnect success")
	return true
}

func SaveConfig() (err error) {
	var config *os.File
	defer func() {
		if config != nil {
			_ = config.Close()
		}
	}()
	defer RecoverAndLogToFile()

	// _ = os.Remove(GLOBAL_STATE.ConfigPath)

	FC := new(FileConfig)
	FC.DNS1 = C.DNS1
	FC.DNS1Bytes = C.DNS1Bytes
	FC.DNSIP = C.DNSIP
	FC.DNS2 = C.DNS2
	FC.ManualRouter = C.ManualRouter
	// FC.Region = C.Region
	FC.DebugLogging = C.DebugLogging
	FC.Version = C.Version
	FC.RouterFilePath = C.RouterFilePath
	FC.AutoReconnect = C.AutoReconnect
	FC.KillSwitch = C.KillSwitch

	cb, err := json.Marshal(FC)
	if err != nil {
		CreateErrorLog("", "Unable to turn new config into bytes: ", err)
		return err
	}

	config, err = os.Create(GLOBAL_STATE.ConfigPath)
	if err != nil {
		CreateErrorLog("", "Unable to create new config", err)
		return err
	}

	_, err = config.Write(cb)
	if err != nil {
		CreateErrorLog("", "Unable to write config bytes to new config file: ", err)
		return err
	}

	return
}

// var GLOBAL_STATE.ConfigPath string

func LoadConfig() {
	defer RecoverAndLogToFile()

	var config *os.File
	var err error
	defer func() {

		if config != nil {
			_ = config.Close()
		}

		if err != nil {
			GLOBAL_STATE.ClientStartupError = true
		}
	}()

	CreateLog("loader", "Loading config")

	GLOBAL_STATE.ConfigPath = GLOBAL_STATE.BasePath + "config"
	config, err = os.Open(GLOBAL_STATE.ConfigPath)
	if err != nil {

		CreateErrorLog("", "Unable to open config: ", err)
		CreateLog("", "Generating a new default config")

		NC := new(FileConfig)
		NC.DNS1Bytes = [4]byte{1, 1, 1, 1}
		NC.DNS1 = "1.1.1.1"
		NC.DNS2 = "8.8.8.8"
		NC.DNSIP = net.IP{NC.DNS1Bytes[0], NC.DNS1Bytes[1], NC.DNS1Bytes[2], NC.DNS1Bytes[3]}
		NC.ManualRouter = false
		NC.Region = ""
		NC.DebugLogging = true
		NC.Version = ""
		NC.RouterFilePath = ""
		NC.AutoReconnect = true
		NC.KillSwitch = false

		var cb []byte
		cb, err = json.Marshal(NC)
		if err != nil {
			GLOBAL_STATE.ClientStartupError = true
			CreateErrorLog("", "Unable to turn new config into bytes: ", err)
			return
		}

		config, err = os.Create(GLOBAL_STATE.ConfigPath)
		if err != nil {
			GLOBAL_STATE.ClientStartupError = true
			CreateErrorLog("", "Unable to create new config file: ", err)
			return
		}

		err = os.Chmod(GLOBAL_STATE.ConfigPath, 0777)
		if err != nil {
			GLOBAL_STATE.ClientStartupError = true
			CreateErrorLog("", "Unable to change ownership of log file: ", err)
			return
		}

		_, err = config.Write(cb)
		if err != nil {
			GLOBAL_STATE.ClientStartupError = true
			CreateErrorLog("", "Unable to write config bytes to new config file: ", err)
			return
		}

		FC := new(Config)
		FC.DNS1 = NC.DNS1
		FC.DNS1Bytes = NC.DNS1Bytes
		FC.DNSIP = NC.DNSIP
		FC.DNS2 = NC.DNS2
		FC.ManualRouter = NC.ManualRouter
		// FC.Region = NC.Region
		FC.DebugLogging = NC.DebugLogging
		FC.Version = NC.Version
		FC.RouterFilePath = NC.RouterFilePath
		FC.AutoReconnect = NC.AutoReconnect
		FC.KillSwitch = NC.KillSwitch
		// CONFIG_INITIALIZED = true
		C = FC

	} else {

		var cb []byte
		cb, err = io.ReadAll(config)
		if err != nil {
			CreateErrorLog("", "Unable to read bytes from config file: ", err)
			return
		}

		err = json.Unmarshal(cb, C)
		if err != nil {
			CreateErrorLog("", "Unable to turn config file into config object: ", err)
			return
		}

		C.DNSIP = net.ParseIP(C.DNS1).To4()
		C.DNS1Bytes = [4]byte{C.DNSIP[0], C.DNSIP[1], C.DNSIP[2], C.DNSIP[3]}

		// CONFIG_INITIALIZED = true
	}

	CreateLog("loader", "Configurations loaded")
	GLOBAL_STATE.C = C
	GLOBAL_STATE.ConfigInitialized = true
}

func CleanupOnClose() {
	defer RecoverAndLogToFile()

	GLOBAL_STATE.Exiting = true
	SetGlobalStateAsDisconnected()
	CleanupWithStateLock()

	_ = LogFile.Close()
}

func PingRouter(IP string) (*ping.Statistics, error) {
	defer RecoverAndLogToFile()

	// CreateLog("", "PING >> ", IP)
	pinger, err := ping.NewPinger(IP)
	if err != nil {
		return nil, err
	}

	routeAdded := false
	if GLOBAL_STATE.Connected || GLOBAL_STATE.Connecting {
		err = AddRoute(IP)
		if err != nil {
			CreateErrorLog("", err)
		} else {
			routeAdded = true
		}
	}

	pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = time.Second * 3
	err = pinger.Run()
	if err != nil {
		return nil, err
	}

	if routeAdded {
		err = DeleteRoute(IP, false)
		if err != nil {
			CreateErrorLog("", err)
		}
	}

	defer pinger.Stop()
	return pinger.Statistics(), nil
}

func GetRoutersFromLocalFile() ([][]byte, error) {
	defer RecoverAndLogToFile()

	if C.RouterFilePath == "" {
		// CreateErrorLog("loader", "Router file path not present in config")
		return nil, errors.New("")
	}

	file, err := os.Open(C.RouterFilePath)
	if err != nil {
		CreateErrorLog("loader", "Unable to open local router file: ", C.RouterFilePath, " || err: ", err)
		return nil, err
	}
	defer file.Close()

	bodyBytes, err := io.ReadAll(file)
	if err != nil {
		CreateErrorLog("loader", "Unable to read from local router file: ", err)
		return nil, err
	}

	lineSplit := bytes.Split(bodyBytes, []byte{13, 10})
	return lineSplit, nil
}

func ParseRoutersFromRawDataToMemory(lines [][]byte) (count int) {
	defer RecoverAndLogToFile()

	for _, v := range lines {
		lineSplit := bytes.Split(v, []byte{44})
		if len(lineSplit) < 2 {
			continue
		}

		IP := string(lineSplit[1])
		TAG := string(lineSplit[0])

		exists := false
		for ri := range GLOBAL_STATE.RoutersList {
			if GLOBAL_STATE.RoutersList[ri] == nil {
				continue
			} else if GLOBAL_STATE.RoutersList[ri].IP == IP {
				exists = true
				GLOBAL_STATE.RoutersList[ri].IP = IP
				GLOBAL_STATE.RoutersList[ri].Tag = TAG
				count++
				break
			}
		}

		if !exists {
			for ri := range GLOBAL_STATE.RoutersList {
				if GLOBAL_STATE.RoutersList[ri] == nil {
					count++
					GLOBAL_STATE.RoutersList[ri] = new(ROUTER)
					GLOBAL_STATE.RoutersList[ri].IP = IP
					GLOBAL_STATE.RoutersList[ri].Tag = TAG
					CreateLog("", "New Router Discovered: ", GLOBAL_STATE.RoutersList[ri].IP)
					break
				}
			}
		}
	}

	return

}

func DownloadRoutersFromOnlineSource() ([][]byte, error) {
	defer RecoverAndLogToFile()

	client := new(http.Client)
	resp, err := client.Get("https://raw.githubusercontent.com/tunnels-is/info/master/all")
	if err != nil {
		CreateErrorLog("loader", "Unable to get routers from online file: ", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		CreateErrorLog("loader", "Unable to get routers from online file // status: ", resp.Status, " // err: ", err)
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		CreateErrorLog("loader", "Unable to parse response from online file: ", err)
		return nil, err
	}

	lineSplit := bytes.Split(bodyBytes, []byte{13, 10})
	return lineSplit, nil
}

func PingAllRouters() {
	defer RecoverAndLogToFile()

	GLOBAL_STATE.LastRouterPing = time.Now()

	for i := range GLOBAL_STATE.RoutersList {
		if GLOBAL_STATE.RoutersList[i] == nil {
			continue
		}

		stats, err := PingRouter(GLOBAL_STATE.RoutersList[i].IP)
		if err != nil {
			CreateErrorLog("loader", "Could not ping router: ", GLOBAL_STATE.RoutersList[i].IP, " // msg: ", err)
			continue
		}

		if stats.AvgRtt.Microseconds() == 0 {
			CreateErrorLog("loader", "0 Microseconds ping, assuming router is offline: ", GLOBAL_STATE.RoutersList[i].IP)
			GLOBAL_STATE.RoutersList[i].PingStats = *stats
			GLOBAL_STATE.RoutersList[i].MS = 31337
		} else {
			GLOBAL_STATE.RoutersList[i].PingStats = *stats
			GLOBAL_STATE.RoutersList[i].MS = uint64(stats.AvgRtt.Milliseconds())
			CreateLog("loader", GLOBAL_STATE.RoutersList[i].IP, " // Avarage latency: ", GLOBAL_STATE.RoutersList[i].PingStats.AvgRtt)
		}

	}
}

func RefreshRouterList() (err error) {
	defer RecoverAndLogToFile()

	var fileLines [][]byte
	fileLines, err = GetRoutersFromLocalFile()

	if err != nil {
		fileLines, err = DownloadRoutersFromOnlineSource()
		if err != nil {
			return err
		}
	}

	routerCount := ParseRoutersFromRawDataToMemory(fileLines)
	if routerCount == 0 {
		CreateErrorLog("loader", "No routers found during probe")
		return errors.New("no routers found")
	}

	CreateLog("loader", "Starting ping check")

	PingAllRouters()

	if !GLOBAL_STATE.Connected && !GLOBAL_STATE.Connecting {
		index, err := GetLowestLatencyRouter()
		if err != nil {
			CreateErrorLog("loader", "Could not find lowest latency router")
			return err
		}
		SetActiveRouter(index)
	}

	CreateLog("loader", "Done probing for routers")
	return nil
}

func SetActiveRouter(index int) {
	GLOBAL_STATE.ActiveRouter = GLOBAL_STATE.RoutersList[index]
	_ = AddRoute(GLOBAL_STATE.ActiveRouter.IP)
	CreateLog("loader", "Active router changed >> ", GLOBAL_STATE.ActiveRouter.IP, " >> Latency is ", GLOBAL_STATE.ActiveRouter.MS, " MS")
}

func GetLowestLatencyRouter() (int, error) {
	defer RecoverAndLogToFile()

	lowestIndex := 0
	lowestMS := 999999999
	foundLowest := false

	for i := range GLOBAL_STATE.RoutersList {
		if GLOBAL_STATE.RoutersList[i] == nil {
			continue
		}

		if GLOBAL_STATE.RoutersList[i].MS == 31337 {
			continue
		}

		if GLOBAL_STATE.RoutersList[i].MS < uint64(lowestMS) {
			foundLowest = true
			lowestMS = int(GLOBAL_STATE.RoutersList[i].MS)
			lowestIndex = i
		}

	}

	if foundLowest {
		return lowestIndex, nil
	}

	return 0, errors.New("no routers")
}

func ConnectToActiveRouter(RoutingBuffer [8]byte) (TUNNEL net.Conn, err error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.ActiveRouter != nil {

		dialer := net.Dialer{Timeout: time.Duration(10 * time.Second)}
		TUNNEL, err = dialer.Dial("tcp", GLOBAL_STATE.ActiveRouter.IP+":443")
		if err != nil {
			return nil, err
		}

		_, err = TUNNEL.Write(RoutingBuffer[:])
		if err != nil {
			return nil, err
		}

	} else {
		return nil, errors.New("no active router found")
	}

	return
}

func BackupSettingsToFile(NewDefault *CONNECTION_SETTINGS) {
	defer RecoverAndLogToFile()

	_ = os.Remove(GLOBAL_STATE.BasePath + NewDefault.IFName + "_backup")
	backupFile, err := os.Create(GLOBAL_STATE.BasePath + NewDefault.IFName + "_backup")
	if err != nil {
		CreateErrorLog("", "Unable to open setting-backup file, please restart the application. If this problem persists contact customer support")
		return
	}

	err = os.Chmod(GLOBAL_STATE.LogFileName, 0777)
	if err != nil {
		CreateErrorLog("", "Unable to change ownership of log file: ", err)
		return
	}

	defer func() {

		if backupFile != nil {
			_ = backupFile.Close()
		}
	}()
	defer RecoverAndLogToFile()

	sb, err := json.Marshal(NewDefault)
	if err != nil {
		CreateErrorLog("", "Unable to marshal adapter backup settings: ", err)
		return
	}

	_, err = backupFile.Write(sb)
	if err != nil {
		CreateErrorLog("", "Unable to write backup settings to file: ", err)
		return
	}

}

func InterfaceMaintenenceAndBackup() {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.Connected || GLOBAL_STATE.Connecting || GLOBAL_STATE.Exiting {
		return
	}

	PotentialDefault, err := FindDefaultInterfaceAndGateway()
	if err != nil {
		return
	}

	if PotentialDefault == nil {
		return
	}

	if PotentialDefault.DefaultRouter == "" {
		return
	}

	err = VerifyAndBackupSettings(PotentialDefault)
	if err != nil {
		return
	}

	BackupSettingsToFile(PotentialDefault)
	GLOBAL_STATE.DefaultInterface = PotentialDefault
}

func FindAllInterfaces() (IFList map[string]*INTERFACE_SETTINGS) {
	defer RecoverAndLogToFile()

	IF, err := net.Interfaces()
	if err != nil {
		CreateErrorLog("", "Could not find network interfaces || msg: ", err)
	}

	IFList = make(map[string]*INTERFACE_SETTINGS)

	for _, v := range IF {
		if v.Name == TUNNEL_ADAPTER_NAME || v.Name == "lo" {
			continue
		}

		netif, ok := IFList[v.Name]
		if !ok {
			IFList[v.Name] = new(INTERFACE_SETTINGS)
			netif = IFList[v.Name]
		}

		netif.Index = v.Index
		netif.Flags = v.Flags
		netif.MTU = v.MTU
		netif.HardwareAddress = v.HardwareAddr
		netif.OIF = v
	}

	return
}

// KEEPING THIS CODE HERE, MIGHT WANT TO USE IT LATER.
// func GET_SERVER_LIST_FROM_DNS(batch int) (error, []*ROUTER) {

// 	CreateLog("loader", "Getting servers using DNS: "+C.Region+"-"+strconv.Itoa(batch)+".nicelandvpn.is.")

// 	var newRouters = make([]*ROUTER, 0)
// 	m := new(dns.Msg)
// 	m.SetQuestion(C.Region+"-"+strconv.Itoa(batch)+".nicelandvpn.is.", 16)
// 	msg, err := dns.Exchange(m, "1.1.1.1:53")
// 	if err != nil {
// 		CreateErrorLog("loader", "Unable to query 1.1.1.1 when looking for routers: ", err)
// 		msg, err = dns.Exchange(m, "8.8.8.8:53")
// 		if err != nil {
// 			CreateErrorLog("loader", "Unable to query 8.8.8.8 when looking for routers: ", err)
// 			msg, err = dns.Exchange(m, "9.9.9.9:53")
// 			if err != nil {
// 				CreateErrorLog("loader", "Unable to query 9.9.9.9 when looking for routers: ", err)
// 				return errors.New("Unable to fetch server list from DNS"), nil
// 			}
// 		}
// 	}

// 	if len(msg.Answer) == 0 {
// 		CreateLog("loader", "No anser to DNS request")
// 		return nil, newRouters
// 	}
// 	for _, v := range msg.Answer {
// 		a, ok := v.(*dns.TXT)
// 		if ok {
// 			splitIPS := strings.Split(a.Txt[0], ",")
// 			for _, v := range splitIPS {
// 				NR := new(ROUTER)
// 				NR.IP = v
// 				NR.MS = 0
// 				newRouters = append(newRouters, NR)
// 			}
// 		}
// 	}

// 	return nil, newRouters
// }
// func DEBUG() *DEBUG_OUT {
// 	defer RecoverAndLogToFile()

// 	var outPrint = make([]string, 0)

// 	// admin check
// 	outPrint = append(outPrint, "=========== Checking for Admin Mode")
// 	ADMIN_CHECK()

// 	// find running modules

// 	// check for more then one background service
// 	outPrint = append(outPrint, "=========== Looking for background processes")
// 	processes, _ := process.Processes()
// 	for _, process := range processes {
// 		name, _ := process.Name()

// 		if strings.Contains(name, CLIENT_PROCESS_NAME) {
// 			if process.Pid == int32(os.Getpid()) {
// 				outPrint = append(outPrint, "Current process ID found: "+fmt.Sprint(process.Pid))
// 			} else {
// 				outPrint = append(outPrint, "Duplicate background process found with ID: "+fmt.Sprint(process.Pid))
// 			}
// 		}
// 	}

// 	outPrint = append(outPrint, "=========== Looking for network interfaces")
// 	IF, err := net.Interfaces()
// 	if err != nil {
// 		outPrint = append(outPrint, "Could not find network interfaces")
// 	}

// 	// find interfaces
// 	for _, v := range IF {
// 		if v.Name == "lo" {
// 			continue
// 		}

// 		outPrint = append(outPrint, fmt.Sprint("FOUND IF: ", v.Name, " || index: ", v.Index, " || name: ", v.Name, " || MTU: ", v.MTU, " || HWADDR: ", v.HardwareAddr))
// 	}

// 	// find default interface
// 	outPrint = append(outPrint, "=========== Looking for default network settings ")
// 	if WIN_ADAPTER_SETTINGS != nil {
// 		outPrint = append(outPrint, fmt.Sprint(*WIN_ADAPTER_SETTINGS))
// 	}
// 	if LINUX_INTERFACE_SETTINGS != nil {
// 		outPrint = append(outPrint, fmt.Sprint(*LINUX_INTERFACE_SETTINGS))
// 	}
// 	if NM_CONNECTION_SETTINGS != nil {
// 		outPrint = append(outPrint, fmt.Sprint(*NM_CONNECTION_SETTINGS))
// 	}
// 	if MAC_CONNECTION_SETTINGS != nil {
// 		outPrint = append(outPrint, fmt.Sprint(*MAC_CONNECTION_SETTINGS))
// 	}

// 	// for i, _ := range MAC_CONNECTION_SETTINGS {
// 	// }
// 	// get default router ip
// 	outPrint = append(outPrint, "=========== Looking for default route IP/Gateway ")
// 	outPrint = append(outPrint, "DEFAULT IP/GATEWAY: "+DEFAULT_ROUTER_IP)
// 	outPrint = append(outPrint, "DEFAULT IF NAME: "+DEFAULT_INTERFACE_NAME)

// 	// print backup files
// 	outPrint = append(outPrint, "=========== Looking for network settings backup file")

// 	if WIN_ADAPTER_SETTINGS != nil {
// 		backupFile, err := os.Open(GenerateBaseFolderPath() + WIN_ADAPTER_SETTINGS.IFName + "_backup")
// 		if err != nil {
// 			outPrint = append(outPrint, "Couldn't find network backup file: "+err.Error())
// 		}

// 		if backupFile != nil {
// 			defer backupFile.Close()
// 		}

// 		backupBytes, err := io.ReadAll(backupFile)
// 		if err != nil {
// 			outPrint = append(outPrint, "Couldn't read network backup file: "+err.Error())
// 		}
// 		outPrint = append(outPrint, "Backup file contents: "+string(backupBytes))
// 	}
// 	if NM_CONNECTION_SETTINGS != nil {
// 		backupFile, err := os.Open(GenerateBaseFolderPath() + NM_CONNECTION_SETTINGS.IFName + "_backup")
// 		if err != nil {
// 			outPrint = append(outPrint, "Couldn't find network backup file: "+err.Error())
// 		}

// 		if backupFile != nil {
// 			defer backupFile.Close()
// 		}

// 		backupBytes, err := io.ReadAll(backupFile)
// 		if err != nil {
// 			outPrint = append(outPrint, "Couldn't read network backup file: "+err.Error())
// 		}
// 		outPrint = append(outPrint, "NetworkManager Backup file contents: "+string(backupBytes))
// 	}
// 	if LINUX_INTERFACE_SETTINGS != nil {

// 		resolvOut, err := exec.Command("bash", "-c", `cat /etc/resolv.conf`).Output()
// 		if err != nil {
// 			outPrint = append(outPrint, "Could not read resolv.conf: "+err.Error())
// 		} else {
// 			outPrint = append(outPrint, "resolv.conf contents: "+string(resolvOut))
// 		}

// 	}

// 	outPrint = append(outPrint, "=========== Print interfaces")

// 	err, out := PrintInterfaces()
// 	if err != nil {
// 		outPrint = append(outPrint, "Could not print routing table: "+err.Error())
// 	} else {
// 		outSplit := bytes.Split(out, []byte{13, 10})
// 		for _, v := range outSplit {
// 			outPrint = append(outPrint, string(v))
// 		}
// 	}

// 	outPrint = append(outPrint, "=========== Print routing table")

// 	err, out = PrintRouters()
// 	if err != nil {
// 		outPrint = append(outPrint, "Could not print routing table: "+err.Error())
// 	} else {
// 		outSplit := bytes.Split(out, []byte{13, 10})
// 		for _, v := range outSplit {
// 			outPrint = append(outPrint, string(v))
// 		}
// 	}
// 	outPrint = append(outPrint, "=========== Print DNS settings")

// 	err, out = PrintDNS()
// 	if err != nil {
// 		outPrint = append(outPrint, "Could not print DNS: "+err.Error())
// 	} else {
// 		outSplit := bytes.Split(out, []byte{13, 10})
// 		for _, v := range outSplit {
// 			outPrint = append(outPrint, string(v))
// 		}
// 	}

// 	// ping routers
// 	outPrint = append(outPrint, "=========== About to ping routers")
// 	for _, v := range ROUTERS {
// 		if v == nil {
// 			continue
// 		}
// 		err, stats := PingRouter(v.IP, true)
// 		if err != nil {
// 			outPrint = append(outPrint, "Unable to ping router: "+v.IP)
// 			continue
// 		}
// 		outPrint = append(outPrint, "PING FROM: "+v.IP+" || MinRTT: "+fmt.Sprint(stats.MinRtt))
// 	}

// 	OUT := new(DEBUG_OUT)
// 	FILENAME := GenerateBaseFolderPath() + "nicelandVPN-DEBUG-" + time.Now().Format("2006-01-02-15-04-05")

// 	debugFile, err := os.Create(FILENAME)
// 	if err != nil {
// 		OUT.File = "unable to create debug file"
// 		return OUT
// 	}
// 	if debugFile != nil {
// 		defer debugFile.Close()
// 	}

// 	for _, v := range outPrint {
// 		debugFile.WriteString(v + "\n")
// 		OUT.Lines = append(OUT.Lines, v)
// 	}

// 	OUT.File = FILENAME
// 	outPrint = nil

// 	return OUT
// }
