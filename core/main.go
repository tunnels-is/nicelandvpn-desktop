package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/go-ping/ping"
	"github.com/zveinn/tunnels"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StartService() {
	defer RecoverAndLogToFile()

	CreateLog("loader", "Starting Niceland VPN")

	C = new(Config)
	C.DebugLogging = true

	AdminCheck()
	InitPaths()
	CreateBaseFolder()
	InitLogfile()
	LoadConfig()
	LoadBlockLists()
	getDefaultGateway()

	CreateLog("loader", "Niceland is ready")
	CreateLog("START", "")
}

func AutoReconnect() (connected bool) {
	defer RecoverAndLogToFile()

	// connectingStateChangedLocally := false
	defer func() {
		// if connectingStateChangedLocally {
		// GLOBAL_STATE.Connecting = false
		// }
		// STATE_LOCK.Unlock()
	}()
	// STATE_LOCK.Lock()

	// if !C.AutoReconnect {
	// 	return false
	// }

	// if C.PrevSession == nil {
	// 	return false
	// }

	// if GLOBAL_STATE.Connected || GLOBAL_STATE.Connecting || GLOBAL_STATE.Exiting {
	// 	return false
	// }

	if time.Since(LastConnectionAttemp).Seconds() < 5 {
		return false
	}

	// GLOBAL_STATE.Connecting = true
	// connectingStateChangedLocally = true

	LastConnectionAttemp = time.Now()
	CreateLog("", "Automatic reconnect..")

	// _, err := REF_ConnectToAccessPoint(C.PrevSession, true)
	// if err != nil {
	// 	CreateErrorLog("", "Auto reconnect failed")
	// 	return false
	// }

	// GLOBAL_STATE.LastRouterPing = time.Now()
	CreateLog("", "Auto reconnect success")
	return true
}

func SaveConfig(c *Config) (err error) {
	var config *os.File
	defer func() {
		if config != nil {
			_ = config.Close()
		}
	}()
	defer RecoverAndLogToFile()

	c.Version = GLOBAL_STATE.Version

	cb, err := json.Marshal(c)
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
	}()

	CreateLog("config", "Loading config")

	GLOBAL_STATE.ConfigPath = GLOBAL_STATE.BasePath + "config.json"
	config, err = os.Open(GLOBAL_STATE.ConfigPath)
	if err != nil {

		CreateLog("config", "Generating a new default config")

		NC := new(Config)
		// NC.ManualRouter = false
		NC.DebugLogging = true
		NC.Version = ""
		NC.RouterFilePath = ""
		NC.DisableIPv6OnConnect = false
		NC.LogBlockedDomains = true

		newCon := new(VPNConnectionMETA)
		newCon.ID = primitive.NewObjectID()
		newCon.AutomaticRouter = true
		newCon.IPv4Address = "10.4.4.4"
		newCon.NetMask = "255.255.255.0"
		// The Tag will be immutable
		newCon.Tag = "Default"
		newCon.IFName = "vpnn"
		newCon.TxQueueLen = 3000
		newCon.MTU = 65535
		newCon.RouterProtocol = "tcp"
		newCon.RouterPort = "443"
		newCon.RouterIndex = 0
		newCon.NodeID = ""
		newCon.AutoReconnect = false
		newCon.KillSwitch = false
		newCon.EncryptionProtocol = 3 // AES-256-GCM
		newCon.CustomDNS = false
		newCon.DNS1 = "9.9.9.9"
		newCon.DNS2 = "149.112.112.112"
		newCon.CloseConnectionsOnConnect = false
		newCon.Networks = make([]*ConnectionNetwork, 0)
		newCon.DNS = make(map[string]*ConnectionDNS)
		newCon.Initialize()

		DNS1 := new(ConnectionDNS)
		DNS1.Wildcard = false
		DNS1.TXT = []string{"welcome!"}
		DNS1.CNAME = "welcome"
		DNS1.IP = []string{"10.4.4.4"}

		newCon.DNS["interface.vpn.local"] = DNS1

		newNetwork := &ConnectionNetwork{
			Tag:     "default",
			Network: "",
			Nat:     "",
			Routes:  []*Route{},
		}
		newNetwork.Routes = append(newNetwork.Routes, &Route{
			Address: "default",
			Metric:  "0",
		})
		newCon.Networks = append(newCon.Networks, newNetwork)

		NC.Connections = make([]*VPNConnectionMETA, 0)
		NC.Connections = append(NC.Connections, newCon)

		var cb []byte
		cb, err = json.Marshal(NC)
		if err != nil {
			CreateErrorLog("config", "Unable to turn new config into bytes: ", err)
			return
		}

		config, err = os.Create(GLOBAL_STATE.ConfigPath)
		if err != nil {
			CreateErrorLog("config", "Unable to create new config file: ", err)
			return
		}

		err = os.Chmod(GLOBAL_STATE.ConfigPath, 0o777)
		if err != nil {
			CreateErrorLog("config", "Unable to change ownership of log file: ", err)
			return
		}

		_, err = config.Write(cb)
		if err != nil {
			CreateErrorLog("config", "Unable to write config bytes to new config file: ", err)
			return
		}

		C = NC

	} else {

		var cb []byte
		cb, err = io.ReadAll(config)
		if err != nil {
			CreateErrorLog("config", "Unable to read bytes from config file: ", err)
			return
		}

		err = json.Unmarshal(cb, C)
		if err != nil {
			CreateErrorLog("config", "Unable to turn config file into config object: ", err)
			return
		}

		for i := range C.Connections {
			if C.Connections[i] == nil {
				continue
			}
			C.Connections[i].Initialize()
		}

	}

	CreateLog("config", "Configurations loaded")
	GLOBAL_STATE.C = C
	GLOBAL_STATE.ConfigInitialized = true
}

func LoadDNSWhitelist() (err error) {
	defer RecoverAndLogToFile()

	if C.DomainWhitelist == "" {
		return nil
	}

	WFile, err := os.OpenFile(C.DomainWhitelist, os.O_RDWR|os.O_CREATE, 0o777)
	if err != nil {
		return err
	}
	defer WFile.Close()

	scanner := bufio.NewScanner(WFile)

	WhitelistMap := make(map[string]bool)
	for scanner.Scan() {
		domain := scanner.Text()
		if domain == "" {
			continue
		}
		WhitelistMap[domain] = true
	}

	err = scanner.Err()
	if err != nil {
		CreateErrorLog("loader", "Unable to load domain whitelist: ", err)
		return err
	}

	DNSWhitelist = WhitelistMap

	return nil
}

func CleanupOnClose() {
	defer RecoverAndLogToFile()

	// CleanupWithStateLock()

	_ = LogFile.Close()
}

func REF_PingRouter(routerIP, gateway string) (*ping.Statistics, error) {
	defer RecoverAndLogToFile()

	// CreateLog("", "PING >> ", IP)
	pinger, err := ping.NewPinger(routerIP)
	if err != nil {
		return nil, err
	}

	routeAdded := false
	err = tunnels.IP_AddRoute(routerIP, gateway, "2")
	if err != nil {
		CreateErrorLog("", err)
	} else {
		routeAdded = true
	}

	pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = time.Second * 3
	err = pinger.Run()
	if err != nil {
		return nil, err
	}

	if routeAdded {
		// err = DeleteRoute(IP, false)
		err = tunnels.IP_DelRoute(routerIP, gateway, "2")
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

	newCountryList := make(map[string]struct{})
	fmt.Println("LINES RETURNED:", lines)
	fmt.Println("LINES RETURNED:", len(lines))

	for index, line := range lines {
		fmt.Println(string(line))
		text := string(line)
		if text == "" {
			continue
		}

		lineSplit := strings.Split(text, "__")
		if len(lineSplit) < 6 {
			CreateErrorLog("", "invalid line in router file")
			continue
		}

		Type, err := strconv.Atoi(lineSplit[0])
		if err == nil {
		} else {
			CreateErrorLog("", "invalid status in router file", lineSplit[0])
			continue
		}

		Status, err := strconv.Atoi(lineSplit[1])
		if err == nil {
		} else {
			CreateErrorLog("", "invalid status in router file", lineSplit[0])
			continue
		}

		UserMbps, err := strconv.Atoi(lineSplit[2])
		if err == nil {
		} else {
			CreateErrorLog("", "invalid availableMbps in router file", lineSplit[2])
			continue
		}

		Mbps, err := strconv.Atoi(lineSplit[3])
		if err == nil {
		} else {
			CreateErrorLog("", "invalid availableMbps in router file", lineSplit[3])
			continue
		}

		Country := lineSplit[4]
		PublicIP := lineSplit[5]
		Tag := lineSplit[6]

		Slots := Mbps / UserMbps

		if Type == 1 {
			count++
			newCountryList[Country] = struct{}{}

			NR := new(ROUTER)
			NR.ListIndex = index
			NR.Tag = Tag
			NR.IP = PublicIP
			NR.Country = Country
			NR.AvailableMbps = Mbps
			NR.Slots = Slots
			NR.Status = Status
			NR.AvailableUserMbps = UserMbps

			router := GLOBAL_STATE.RouterList[index]
			if router == nil {
				GLOBAL_STATE.RouterList[index] = NR
				NR.MS = 9999
			} else {
				GLOBAL_STATE.RouterList[index].Tag = NR.Tag
				GLOBAL_STATE.RouterList[index].IP = NR.IP
				GLOBAL_STATE.RouterList[index].Country = NR.Country
				GLOBAL_STATE.RouterList[index].AvailableMbps = NR.AvailableMbps
				GLOBAL_STATE.RouterList[index].Status = NR.Status
				GLOBAL_STATE.RouterList[index].AvailableUserMbps = NR.AvailableUserMbps
				GLOBAL_STATE.RouterList[index].Slots = NR.Slots
				GLOBAL_STATE.RouterList[index].Status = NR.Status
			}

		}

	}

	GLOBAL_STATE.AvailableCountries = make([]string, 0)
	for i := range newCountryList {
		GLOBAL_STATE.AvailableCountries = append(GLOBAL_STATE.AvailableCountries, i)
	}

	return
}

func DownloadRoutersFromOnlineSource() ([][]byte, error) {
	defer RecoverAndLogToFile()

	client := new(http.Client)
	resp, err := client.Get("https://raw.githubusercontent.com/tunnels-is/info/master/infra")
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

	lineSplit := bytes.Split(bodyBytes, []byte{10})
	return lineSplit, nil
}

func REF_PingAllRouters() {
	defer RecoverAndLogToFile()

	for i := range GLOBAL_STATE.RouterList {
		if GLOBAL_STATE.RouterList[i] == nil {
			continue
		}

		stats, err := REF_PingRouter(GLOBAL_STATE.RouterList[i].IP, DEFAULT_GATEWAY.String())
		if err != nil {
			CreateErrorLog("loader", "Could not ping router: ", GLOBAL_STATE.RouterList[i].IP, " // msg: ", err)
			continue
		}

		if stats.AvgRtt.Microseconds() == 0 {
			CreateErrorLog("loader", "0 Microseconds ping, assuming router is offline: ", GLOBAL_STATE.RouterList[i].IP)
			GLOBAL_STATE.RouterList[i].PingStats = *stats
			GLOBAL_STATE.RouterList[i].MS = 9999
		} else {
			GLOBAL_STATE.RouterList[i].PingStats = *stats
			GLOBAL_STATE.RouterList[i].MS = uint64(stats.AvgRtt.Milliseconds())
			CreateLog("loader", GLOBAL_STATE.RouterList[i].IP, " // Avarage latency: ", GLOBAL_STATE.RouterList[i].PingStats.AvgRtt)
		}

	}
}

func REF_RefreshRouterList() (err error) {
	defer RecoverAndLogToFile()

	if DEFAULT_GATEWAY == nil {
		CreateLog("", "no default gateway found")
		return
	}

	if time.Since(LAST_ROUTER_PROBE).Milliseconds() < int64(ROUTER_PROBE_TIMEOUT_MS) {
		// fmt.Println("TIMEOUT", time.Since(LAST_ROUTER_PROBE).Milliseconds() > int64(ROUTER_PROBE_TIMEOUT_MS))
		// fmt.Println("TIMEOUT", time.Since(LAST_ROUTER_PROBE).Milliseconds())
		return
	}

	var fileLines [][]byte
	fileLines, err = GetRoutersFromLocalFile()
	if err != nil {
		// fmt.Println("ONLINE DOWNLOAD!")
		fileLines, err = DownloadRoutersFromOnlineSource()
		if err != nil {
			fmt.Println("BIG ERR:", err)
			return err
		}
	}

	routerCount := ParseRoutersFromRawDataToMemory(fileLines)
	fmt.Println("COUNT:", routerCount)
	if routerCount == 0 {
		CreateErrorLog("loader", "No routers found during probe")
		return errors.New("no routers found")
	}

	CreateLog("loader", "Starting ping check")

	REF_PingAllRouters()

	index, err := GetLowestLatencyRouter()
	if err != nil {
		CreateErrorLog("loader", "Could not find lowest latency router")
		return err
	}

	REF_SetActiveRouter(index)
	LAST_ROUTER_PROBE = time.Now()
	CreateLog("loader", "Done probing for routers")
	return nil
}

func REF_SetActiveRouter(index int) {
	_ = tunnels.IP_AddRoute(GLOBAL_STATE.RouterList[index].IP, DEFAULT_GATEWAY.String(), "0")
	GLOBAL_STATE.ActiveRouter = GLOBAL_STATE.RouterList[index]
	CreateLog("loader", "Active router changed >> ", GLOBAL_STATE.ActiveRouter.IP, " >> Latency is ", GLOBAL_STATE.ActiveRouter.MS, " MS")
}

func GetLowestLatencyRouter() (int, error) {
	defer RecoverAndLogToFile()

	lowestIndex := 0
	lowestMS := 999999999
	foundLowest := false

	for i := range GLOBAL_STATE.RouterList {
		if GLOBAL_STATE.RouterList[i] == nil {
			continue
		}

		if GLOBAL_STATE.RouterList[i].MS == 9999 {
			continue
		}

		if GLOBAL_STATE.RouterList[i].MS < uint64(lowestMS) {
			foundLowest = true
			lowestMS = int(GLOBAL_STATE.RouterList[i].MS)
			lowestIndex = i
		}

	}

	if foundLowest {
		return lowestIndex, nil
	}

	return 0, errors.New("no routers")
}

func REF_ConnectToRouter(EntryIndex int, proto, port string) (TUNNEL net.Conn, err error) {
	defer RecoverAndLogToFile()

	// HARDCODED FOR NOW
	// - might switch this up later
	port = "443"
	proto = "tcp"

	var routerIP string
	r := GLOBAL_STATE.RouterList[EntryIndex]
	if r != nil {
		routerIP = r.IP
	}

	if routerIP == "" {
		return nil, errors.New("unable to find router")
	}

	gw, err := tunnels.FindGateway()
	if err != nil {
		return nil, errors.New("unable to find default gateway")
	}

	err = tunnels.IP_AddRoute(routerIP, gw.To4().String(), "0")
	if err != nil {
		return nil, errors.New("unable to route to router via default gateway")
	}

	dialer := net.Dialer{Timeout: time.Duration(10 * time.Second)}
	TUNNEL, err = dialer.Dial(proto, routerIP+":"+port)
	if err != nil {
		return nil, err
	}

	return
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

	err = os.Chmod(GLOBAL_STATE.LogFileName, 0o777)
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

// func FindAllInterfaces() (IFList map[string]*INTERFACE_SETTINGS) {
// 	defer RecoverAndLogToFile()
//
// 	IF, err := net.Interfaces()
// 	if err != nil {
// 		CreateErrorLog("", "Could not find network interfaces || msg: ", err)
// 	}
//
// 	IFList = make(map[string]*INTERFACE_SETTINGS)
//
// 	for _, v := range IF {
// 		if v.Name == TUNNEL_ADAPTER_NAME || v.Name == "lo" {
// 			continue
// 		}
//
// 		netif, ok := IFList[v.Name]
// 		if !ok {
// 			IFList[v.Name] = new(INTERFACE_SETTINGS)
// 			netif = IFList[v.Name]
// 		}
//
// 		netif.Index = v.Index
// 		netif.Flags = v.Flags
// 		netif.MTU = v.MTU
// 		netif.HardwareAddress = v.HardwareAddr
// 		netif.OIF = v
// 	}
//
// 	return
// }
//
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

type Event struct {
	Tag    string
	Action func()
}

var GLOBAL_EventQueue = make(chan func(), 100)

func EventAndStateManager() {
	for event := range GLOBAL_EventQueue {
		log.Println(GET_FUNC(0))
		event()
	}
}
