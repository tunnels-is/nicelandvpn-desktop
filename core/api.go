package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/xlzd/gotp"
	tp "github.com/zveinn/tcpcrypt"
	"github.com/zveinn/tunnels"
)

func ControllerCustomDialer(ctx context.Context, _ string, addr string) (net.Conn, error) {
	return OpenProxyTunnelToRouter()
}

func LocalhostCustomDialer(ctx context.Context, network, addr string) (net.Conn, error) {
	return OpenProxyTunnelToRouter()
}

func OpenProxyTunnelToRouter() (TcpConn net.Conn, err error) {
	TcpConn, err = net.Dial("tcp", GLOBAL_STATE.ActiveRouter.IP+":443")
	if err != nil {
		CreateErrorLog("", "Could not dial router: ", GLOBAL_STATE.ActiveRouter.IP, err)
		return
	}

	return
}

func ResetEverything() {
	defer RecoverAndLogToFile()

	CreateLog("START", "")
	// CleanupWithStateLock("*")
}

// func CleanupWithStateLock(ConnectionName string) {
// defer STATE_LOCK.Unlock()
// defer RecoverAndLogToFile()
// STATE_LOCK.Lock()

// CONNECTIONS[ConnectionName].Disconnect()
// DisconnectFromRouter(AS)
// _ = SetInterfaceStateToDown()
// InstantlyClearPortMaps()

// RestoreIPv6()
// RestoreDNS(false)
//
// SetGlobalStateAsDisconnected()
// }

func REF_SwitchRouter(Tag string) (code int, err error) {
	defer RecoverAndLogToFile()
	fmt.Println("SWITCHING:", Tag)

	if Tag == "" {
		C.ManualRouter = false

		err := REF_RefreshRouterList()
		if err != nil {
			CreateErrorLog("", "Unable to find the best router for your connection: ", err)
		}

	} else {
		C.ManualRouter = true

		for i := range GLOBAL_STATE.RouterList {
			if GLOBAL_STATE.RouterList[i] != nil {
				if GLOBAL_STATE.RouterList[i].Tag == Tag {
					REF_SetActiveRouter(i)
				}
			}
		}
	}

	return 200, nil
}

func SendRawBytesToLocalhostProxy(method string, route string, data []byte, timeoutMS int) ([]byte, int, error) {
	defer RecoverAndLogToFile()

	var req *http.Request

	var err error
	if method == "POST" {
		req, err = http.NewRequest(method, "http://127.0.0.1:1337/"+route, bytes.NewBuffer(data))
	} else if method == "GET" {
		req, err = http.NewRequest(method, "http://127.0.0.1:1337/"+route, nil)
	} else {
		return nil, 0, errors.New("method not supported: " + method)
	}

	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Content-Type", "application/x-binary")

	client := http.Client{
		Timeout: time.Duration(timeoutMS) * time.Millisecond,
		Transport: &http.Transport{
			DialContext: LocalhostCustomDialer,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		} else {
			return nil, 0, err
		}
	}

	client.CloseIdleConnections()
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var x []byte
	x, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return x, resp.StatusCode, nil
}

func SendRequestToLocalhostProxy(method string, route string, data interface{}, timeoutMS int) ([]byte, int, error) {
	defer RecoverAndLogToFile()

	var body []byte
	var err error
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, 0, err
		}
	}

	var req *http.Request

	if method == "POST" {
		req, err = http.NewRequest(method, "http://127.0.0.1:1337/"+route, bytes.NewBuffer(body))
	} else if method == "GET" {
		req, err = http.NewRequest(method, "http://127.0.0.1:1337/"+route, nil)
	} else {
		return nil, 0, errors.New("method not supported:" + method)
	}

	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: time.Duration(timeoutMS) * time.Millisecond,
		Transport: &http.Transport{
			DialContext: LocalhostCustomDialer,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		} else {
			return nil, 0, err
		}
	}
	client.CloseIdleConnections()
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var x []byte
	x, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return x, resp.StatusCode, nil
}

func SendRequestToControllerProxy(method string, route string, data interface{}, domain string, timeoutMS int) ([]byte, int, error) {
	defer RecoverAndLogToFile()

	var body []byte
	var err error
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, 0, err
		}
	}

	var req *http.Request
	if method == "POST" {
		req, err = http.NewRequest(method, "https://"+domain+":443/"+route, bytes.NewBuffer(body))
	} else if method == "GET" {
		req, err = http.NewRequest(method, "https://"+domain+"443:/"+route, nil)
	} else {
		return nil, 0, errors.New("method not supported:" + method)
	}

	if err != nil {
		return nil, 0, err
	}

	req.Host = domain
	req.Header.Add("Host", domain)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: time.Duration(timeoutMS) * time.Millisecond,
		Transport: &http.Transport{
			DialContext: ControllerCustomDialer,
			TLSClientConfig: &tls.Config{
				ServerName:         domain,
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		} else {
			return nil, 0, err
		}
	}
	client.CloseIdleConnections()
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var x []byte
	x, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return x, resp.StatusCode, nil
}

var LAST_PRIVATE_ACCESS_POINT_UPDATE = time.Now()

// func GetPrivateAccessPoints(FR *FORWARD_REQUEST) (interface{}, int, error) {
// 	if GLOBAL_STATE.ActiveRouter == nil {
// 		return nil, 500, errors.New("active router not found, please wait a moment")
// 	}
//
// 	return nil, 0, nil
// }

//	func LoadRoutersUnAuthenticated() (interface{}, int, error) {
//		log.Println("GET ROUTERS UN_AHUTH")
//		GLOBAL_STATE.Routers = nil
//		GLOBAL_STATE.Routers = make([]*ROUTER, 0)
//		for i := range GLOBAL_STATE.RouterList {
//			if GLOBAL_STATE.RouterList[i] == nil {
//				continue
//			}
//
//			GLOBAL_STATE.Routers = append(GLOBAL_STATE.Routers, GLOBAL_STATE.RouterList[i])
//		}
//
//		sort.Slice(GLOBAL_STATE.Routers, func(a, b int) bool {
//			if GLOBAL_STATE.Routers[a] == nil {
//				return false
//			}
//			if GLOBAL_STATE.Routers[b] == nil {
//				return false
//			}
//			if GLOBAL_STATE.Routers[a].Score == GLOBAL_STATE.Routers[b].Score {
//				if GLOBAL_STATE.Routers[a].MS < GLOBAL_STATE.Routers[b].MS {
//					return true
//				}
//			}
//
//			return GLOBAL_STATE.Routers[a].Score > GLOBAL_STATE.Routers[b].Score
//		})
//
//		return nil, 200, nil
//	}
var LAST_ROUTER_AND_ACCESS_POINT_UPDATE = time.Now()

// func GetRoutersAndAccessPoints(FR *FORWARD_REQUEST) (interface{}, int, error) {
// 	defer RecoverAndLogToFile()
//
// 	if GLOBAL_STATE.ActiveRouter == nil {
// 		return nil, 500, errors.New("active router not found, please wait a moment")
// 	}
//
// 	if !GLOBAL_STATE.LastNodeUpdate.IsZero() {
// 		since := time.Since(GLOBAL_STATE.LastNodeUpdate).Seconds()
// 		GLOBAL_STATE.SecondsUntilNodeUpdate = 55 - int(since)
// 		if since < 55 {
// 			return nil, 200, nil
// 		}
// 	}
//
// 	GLOBAL_STATE.LastNodeUpdate = time.Now()
// 	GLOBAL_STATE.SecondsUntilNodeUpdate = 55
//
// 	responseBytes, code, err := SendRequestToLocalhostProxy("GET", "v1/a", nil, 10000)
// 	if err != nil {
// 		CreateLog("", "(ROUTER/API) // code: ", code, " // err:", err)
// 		if code != 0 {
// 			GLOBAL_STATE.LastNodeUpdate = time.Now().Add(-45 * time.Second)
// 			return nil, code, errors.New(string(responseBytes))
// 		} else {
// 			GLOBAL_STATE.LastNodeUpdate = time.Now().Add(-45 * time.Second)
// 			return nil, code, errors.New("unable to contact controller")
// 		}
// 	}
//
// 	if code != 200 {
// 		return nil, code, errors.New("Unable to fetch access points")
// 	}
//
// 	RoutersAndNodes := new(CONTROLL_PUBLIC_DEVCE_RESPONSE)
//
// 	err = json.Unmarshal(responseBytes, RoutersAndNodes)
// 	if err != nil {
// 		GLOBAL_STATE.LastNodeUpdate = time.Now().Add(-45 * time.Second)
// 		CreateErrorLog("", "Could not process forward request: ", err)
// 		return nil, 400, errors.New("unknown error, please try again in a moment")
// 	}
//
// 	responseBytes, code, err = SendRequestToControllerProxy(FR.Method, FR.Path, FR.JSONData, "api.atodoslist.net", FR.Timeout)
// 	if err != nil {
// 		CreateLog("", "(ROUTER/CONTROLLER) // code: ", code, " // err:", err)
// 		if code != 0 {
// 			return nil, code, errors.New(string(responseBytes))
// 		} else {
// 			return nil, code, errors.New("unable to contact controller")
// 		}
// 	}
//
// 	PrivateNodes := make([]*VPNNode, 0)
// 	if code == 200 {
// 		// CreateLog("", "RESPONSE:", string(responseBytes))
// 		err = json.Unmarshal(responseBytes, &PrivateNodes)
// 		if err != nil {
// 			CreateErrorLog("", "Unable to unmarshal private device list: ", err)
// 			return nil, 0, err
// 		}
// 	}
//
// 	for ii := range RoutersAndNodes.Routers {
// 		RR := RoutersAndNodes.Routers[ii]
//
// 		exists := false
// 		for i := range GLOBAL_STATE.RouterList {
// 			R := GLOBAL_STATE.RouterList[i]
// 			if R == nil {
// 				continue
// 			}
// 			if RR.PublicIP == R.PublicIP {
// 				exists = true
// 			}
// 		}
//
// 		if !exists {
// 			for i := range GLOBAL_STATE.RouterList {
// 				if GLOBAL_STATE.RouterList[i] == nil {
// 					GLOBAL_STATE.RouterList[i] = RoutersAndNodes.Routers[ii]
// 					break
// 				}
// 			}
// 		}
//
// 	}
//
// 	for ii := range RoutersAndNodes.Routers {
// 		RR := RoutersAndNodes.Routers[ii]
//
// 		for i := range GLOBAL_STATE.RouterList {
// 			R := GLOBAL_STATE.RouterList[i]
// 			if R == nil {
// 				continue
// 			}
// 			if RR.PublicIP == R.PublicIP {
// 				// GLOBAL_STATE.RouterList[i].GROUP = RR.GROUP
// 				// GLOBAL_STATE.RouterList[i].ROUTERID = RR.ROUTERID
// 				GLOBAL_STATE.RouterList[i].Country = RR.Country
// 				GLOBAL_STATE.RouterList[i].AvailableMbps = RR.AvailableMbps
// 				GLOBAL_STATE.RouterList[i].Slots = RR.Slots
// 				GLOBAL_STATE.RouterList[i].AvailableSlots = RR.AvailableSlots
// 				GLOBAL_STATE.RouterList[i].AEBP = RR.AEBP
// 				GLOBAL_STATE.RouterList[i].AIBP = RR.AIBP
// 				GLOBAL_STATE.RouterList[i].CPUP = RR.CPUP
// 				GLOBAL_STATE.RouterList[i].RAMUsage = RR.RAMUsage
// 				GLOBAL_STATE.RouterList[i].DiskUsage = RR.DiskUsage
// 				GLOBAL_STATE.RouterList[i].Online = RR.Online
// 				R = GLOBAL_STATE.RouterList[i]
//
// 				includeMSInScoring := true
// 				if R.MS == 31337 {
// 					includeMSInScoring = false
// 				}
//
// 				var baseScore float64 = 10
// 				AIBPScore := 100 / R.AIBP
// 				AEBPScore := 100 / R.AEBP
// 				if AIBPScore < 1 || AIBPScore == 1 {
// 					AIBPScore = 0
// 				} else if AIBPScore > 4 {
// 					AIBPScore = 4
// 				}
// 				if AEBPScore < 1 || AEBPScore == 1 {
// 					AEBPScore = 0
// 				} else if AEBPScore > 4 {
// 					AEBPScore = 4
// 				}
//
// 				var MSScore float64 = 0
// 				if includeMSInScoring {
// 					if R.MS < 20 {
// 						MSScore = 1
// 					} else if R.MS < 100 {
// 						MSScore = 2
// 					} else if R.MS < 200 {
// 						MSScore = 3
// 					} else if R.MS < 300 {
// 						MSScore = 4
// 					} else if R.MS < 400 {
// 						MSScore = 5
// 					}
// 				}
// 				SLOTScore := float64(R.AvailableSlots / R.Slots)
// 				R.Score = int(baseScore - AEBPScore - SLOTScore - AIBPScore - MSScore)
// 			}
// 		}
//
// 	}
//
// 	GLOBAL_STATE.Routers = nil
// 	GLOBAL_STATE.Routers = make([]*ROUTER, 0)
// 	GLOBAL_STATE.AvailableCountries = make([]string, 0)
// 	for i := range GLOBAL_STATE.RouterList {
// 		if GLOBAL_STATE.RouterList[i] == nil {
// 			continue
// 		}
//
// 		countryExists := false
// 		for ii := range GLOBAL_STATE.AvailableCountries {
// 			if GLOBAL_STATE.AvailableCountries[ii] == GLOBAL_STATE.RouterList[i].Country {
// 				countryExists = true
// 			}
// 		}
//
// 		if !countryExists {
// 			GLOBAL_STATE.AvailableCountries = append(GLOBAL_STATE.AvailableCountries, GLOBAL_STATE.RouterList[i].Country)
// 		}
//
// 		GLOBAL_STATE.Routers = append(GLOBAL_STATE.Routers, GLOBAL_STATE.RouterList[i])
// 	}
//
// 	sort.Slice(GLOBAL_STATE.Routers, func(a, b int) bool {
// 		if GLOBAL_STATE.Routers[a] == nil {
// 			return false
// 		}
// 		if GLOBAL_STATE.Routers[b] == nil {
// 			return false
// 		}
// 		if GLOBAL_STATE.Routers[a].Score == GLOBAL_STATE.Routers[b].Score {
// 			if GLOBAL_STATE.Routers[a].MS < GLOBAL_STATE.Routers[b].MS {
// 				return true
// 			}
// 		}
//
// 		return GLOBAL_STATE.Routers[a].Score > GLOBAL_STATE.Routers[b].Score
// 	})
//
// 	GLOBAL_STATE.Nodes = RoutersAndNodes.AccessPoints
// 	for i := range GLOBAL_STATE.Nodes {
// 		A := GLOBAL_STATE.Nodes[i]
//
// 		for ii := range GLOBAL_STATE.RouterList {
// 			R := GLOBAL_STATE.RouterList[ii]
// 			if R == nil {
// 				continue
// 			}
//
// 			if R.GROUP == A.GROUP && R.ROUTERID == A.ROUTERID {
// 				GLOBAL_STATE.Nodes[i].Router = GLOBAL_STATE.RouterList[ii]
// 			}
// 		}
// 	}
//
// 	GLOBAL_STATE.PrivateNodes = PrivateNodes
// 	for i := range GLOBAL_STATE.PrivateNodes {
// 		A := GLOBAL_STATE.PrivateNodes[i]
//
// 		for ii := range GLOBAL_STATE.RouterList {
// 			R := GLOBAL_STATE.RouterList[ii]
// 			if R == nil {
// 				continue
// 			}
//
// 			if R.GROUP == A.GROUP && R.ROUTERID == A.ROUTERID {
// 				GLOBAL_STATE.PrivateNodes[i].Router = GLOBAL_STATE.RouterList[ii]
// 			}
// 		}
// 	}
//
// 	// GLOBAL_STATE.ActiveAccessPoint = GetActiveAccessPointFromActiveSession()
// 	// AS.AP = GLOBAL_STATE.ActiveAccessPoint
//
// 	if len(GLOBAL_STATE.Nodes) == 0 {
// 		GLOBAL_STATE.LastNodeUpdate = time.Now().Add(-45 * time.Second)
// 	}
//
// 	sort.Slice(GLOBAL_STATE.Nodes, func(a, b int) bool {
// 		if GLOBAL_STATE.Nodes[a].Router == nil {
// 			return false
// 		}
// 		if GLOBAL_STATE.Nodes[b].Router == nil {
// 			return false
// 		}
// 		if GLOBAL_STATE.Nodes[a].Router.Score == GLOBAL_STATE.Nodes[b].Router.Score {
// 			if GLOBAL_STATE.Nodes[a].Router.MS < GLOBAL_STATE.Nodes[b].Router.MS {
// 				return true
// 			}
// 		}
// 		return GLOBAL_STATE.Nodes[a].Router.Score > GLOBAL_STATE.Nodes[b].Router.Score
// 	})
//
// 	sort.Slice(GLOBAL_STATE.PrivateNodes, func(a, b int) bool {
// 		if GLOBAL_STATE.PrivateNodes[a].Router == nil {
// 			return false
// 		}
// 		if GLOBAL_STATE.PrivateNodes[b].Router == nil {
// 			return false
// 		}
// 		if GLOBAL_STATE.PrivateNodes[a].Router.Score == GLOBAL_STATE.PrivateNodes[b].Router.Score {
// 			if GLOBAL_STATE.PrivateNodes[a].Router.MS < GLOBAL_STATE.PrivateNodes[b].Router.MS {
// 				return true
// 			}
// 		}
// 		return GLOBAL_STATE.PrivateNodes[a].Router.Score > GLOBAL_STATE.PrivateNodes[b].Router.Score
// 	})
//
// 	fmt.Println("FULL GET ROUTERS CALL")
// 	return nil, code, nil
// }

func ForwardToRouter(FR *FORWARD_REQUEST) (interface{}, int, error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.ActiveRouter == nil {
		return nil, 500, errors.New("tctive router not found, please wait a moment")
	}

	responseBytes, code, err := SendRequestToLocalhostProxy(FR.Method, FR.Path, FR.JSONData, FR.Timeout)
	if err != nil {
		CreateLog("", "(ROUTER/API) // code: ", code, " // err:", err)
		if code != 0 {
			return nil, code, errors.New(string(responseBytes))
		} else {
			return nil, code, errors.New("unable to contact controller")
		}
	}

	var respJSON interface{}
	err = json.Unmarshal(responseBytes, &respJSON)
	if err != nil {
		CreateErrorLog("", "Could not process forward request: ", err)
		return nil, 400, errors.New("unknown error, please try again in a moment")
	}

	return respJSON, code, nil
}

func ForwardToController(FR *FORWARD_REQUEST) (interface{}, int, error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.ActiveRouter == nil {
		return nil, 500, errors.New("active router not found, please wait a moment")
	}

	log.Println("FR:", FR)
	// The domain being used here is an old domain that needs to be replaced.
	// This method uses a custom dialer which does not DNS resolve.
	responseBytes, code, err := SendRequestToControllerProxy(FR.Method, FR.Path, FR.JSONData, "api.atodoslist.net", FR.Timeout)
	if err != nil {
		CreateLog("", "(ROUTER/CONTROLLER) // code: ", code, " // err:", err)
		if code != 0 {
			return nil, code, errors.New(string(responseBytes))
		} else {
			return nil, code, errors.New("unable to contact controller")
		}
	}

	var respJSON interface{}
	err = json.Unmarshal(responseBytes, &respJSON)
	if err != nil {
		CreateErrorLog("", "Could not process forward request: ", err)
		return nil, 400, errors.New("unknown error, please try again in a moment")
	}

	return respJSON, code, nil
}

var NEXT_SERVER_REFRESH time.Time

func SetRouterFile(path string) error {
	// defer STATE_LOCK.Unlock()
	defer RecoverAndLogToFile()
	// STATE_LOCK.Lock()

	CreateLog("START", "")

	C.RouterFilePath = path
	C.ManualRouter = false

	err := SaveConfig()
	if err != nil {
		CreateErrorLog("", "Unable to save config: ", err)
		return errors.New("unable to save config")
	}

	err = REF_RefreshRouterList()
	if err != nil {
		return errors.New(err.Error())
	}

	CreateLog("loader", "Router file updated")
	return nil
}

func SetConfig(SF *CONFIG_FORM) error {
	// defer STATE_LOCK.Unlock()
	defer RecoverAndLogToFile()
	// STATE_LOCK.Lock()

	CreateLog("START", "")

	if C.CustomDNS != SF.CustomDNS {
		CreateLog("loader", "unable to change custom DNS state while connected")
		return errors.New("unable to change custom DNS state while connected")
	}

	if SF.Version != "" {
		C.Version = SF.Version
	}

	C.RouterFilePath = SF.RouterFilePath
	C.DebugLogging = SF.DebugLogging
	C.AutoReconnect = SF.AutoReconnect
	C.KillSwitch = SF.KillSwitch
	C.DisableIPv6OnConnect = SF.DisableIPv6OnConnect
	C.CloseConnectionsOnConnect = SF.CloseConnectionsOnConnect
	C.CustomDNS = SF.CustomDNS

	C.LogBlockedDomains = SF.LogBlockedDomains
	if slices.Compare(C.EnabledBlockLists, SF.EnabledBlockLists) != 0 {
		C.EnabledBlockLists = SF.EnabledBlockLists
		for i := range C.EnabledBlockLists {
			GLOBAL_STATE.BLists[i].Enabled = true
		}
		BuildDomainBlocklist()
	}
	// C.EnabledBlockLists = SF.EnabledBlockLists
	// if SF.PrevSession != nil {
	// 	C.PrevSession = SF.PrevSession
	// }

	if !SF.DebugLogging {
		for i := range L.LOGS {
			L.LOGS[i] = ""
		}
	}

	// var dnsWasChanged bool = false
	if SF.DNS1 != "" {
		if C.DNS1 != SF.DNS1 {
			C.DNS1 = SF.DNS1
			C.DNSIP = net.ParseIP(C.DNS1).To4()
			if len(C.DNSIP) < 4 {
				return errors.New("DNS1 is invalid or empty")
			}
			C.DNS1Bytes = [4]byte{C.DNSIP[0], C.DNSIP[1], C.DNSIP[2], C.DNSIP[3]}
			// dnsWasChanged = true
		}
	} else {

		CreateLog("loader", "Error while updating config || DNS1 is invalid: ", SF.DNS1)
		return errors.New("DNS1 is invalid or empty")
	}

	if C.DNS2 != SF.DNS2 {
		C.DNS2 = SF.DNS2
		// dnsWasChanged = true
	}

	// if dnsWasChanged {
	// 	err := ChangeDNSWhileConnected()
	// 	if err != nil {
	// 		return errors.New("unable to update DNS on tunnel interface")
	// 	}
	// }

	err := SaveConfig()
	if err != nil {
		CreateErrorLog("", "Unable to save config: ", err)
		return errors.New("unable to save config")
	}

	CreateLog("", "Config update || new config: ", *C)
	return nil
}

func PrepareState(e echo.Context) (err error) {
	defer RecoverAndLogToFile()
	form := new(FORWARD_REQUEST)
	err = e.Bind(form)
	if err != nil {
		return e.JSON(400, err)
	}

	// if form.Authed {
	// 	_, code, err := GetRoutersAndAccessPoints(form)
	// 	if err != nil {
	// 		fmt.Println("GET INFO ERROR:", err)
	// 	}
	// 	if code != 200 {
	// 		fmt.Println("GET INFO CODE:", code)
	// 	}
	//
	// } else {
	// 	_, _, _ = LoadRoutersUnAuthenticated()
	// }

	// GLOBAL_STATE.EgressPackets = EGRESS_PACKETS
	// GLOBAL_STATE.IngressPackets = INGRESS_PACKETS
	// ubps := GLOBAL_STATE.UMbps
	// utext := "bps"
	// dbps := GLOBAL_STATE.DMbps
	// dtext := "bps"
	// if ubps > 1100000 {
	// 	utext = "Mbps"
	// 	ubps = ubps / 1000000
	// } else if ubps > 1100 {
	// 	utext = "Kbps"
	// 	ubps = ubps / 1000
	// }
	// GLOBAL_STATE.UMbpsString = fmt.Sprintf("%d %s", ubps, utext)
	//
	// if dbps > 1100000 {
	// 	dtext = "Mbps"
	// 	dbps = dbps / 1000000
	// } else if dbps > 1100 {
	// 	dtext = "Kbps"
	// 	dbps = dbps / 1000
	// }

	// GLOBAL_STATE.DMbpsString = fmt.Sprintf("%d %s", dbps, dtext)
	// GLOBAL_STATE.SecondsSincePingFromRouter = fmt.Sprintf("%.0f seconds", time.Since(GLOBAL_STATE.PingReceivedFromRouter).Seconds())

	// if GLOBAL_STATE.ActiveSession != nil {
	// 	seconds := time.Since(GLOBAL_STATE.ActiveSession.Created).Seconds()
	// 	label := "seconds"
	//
	// 	if seconds > 60 && seconds < 120 {
	// 		label = "minute"
	// 		seconds = seconds / 60
	// 	} else if seconds >= 120 {
	// 		label = "minutes"
	// 		seconds = seconds / 60
	// 	}
	//
	// 	GLOBAL_STATE.ConnectedTimer = fmt.Sprintf("%.0f %s", seconds, label)
	// }

	// if GLOBAL_STATE.ActiveSession != nil {
	// 	S := GLOBAL_STATE.ActiveSession
	// 	for i := range GLOBAL_STATE.AccessPoints {
	// 		A := GLOBAL_STATE.AccessPoints[i]
	// 		if A.GROUP == S.XGROUP && A.ROUTERID == S.XROUTERID {
	// 			GLOBAL_STATE.ActiveAccessPoint = GLOBAL_STATE.AccessPoints[i]
	// 		}
	// 	}
	// }
	// GLOBAL_STATE.ActiveAccessPoint = GetActiveAccessPointFromActiveSession()
	return
}

//func REF_GetNodeFromSession(S *CLIENT_SESSION) *VPNNode {
//	for i := range GLOBAL_STATE.Nodes {
//		A := GLOBAL_STATE.Nodes[i]
//		// CreateLog("", "AAP: ", A.GROUP, " - ", S.XGROUP, " - ", A.ROUTERID, " - ", S.XROUTERID, " - ", A.DEVICEID, " - ", S.DEVICEID)
//		if A.GROUP == S.XGROUP && A.ROUTERID == S.XROUTERID && A.DEVICEID == S.DEVICEID {
//			// GLOBAL_STATE.ActiveAccessPoint = GLOBAL_STATE.AccessPoints[i]
//			return GLOBAL_STATE.Nodes[i]
//		}
//	}
//
//	for i := range GLOBAL_STATE.PrivateNodes {
//		A := GLOBAL_STATE.PrivateNodes[i]
//		// CreateLog("", "AAP: ", A.GROUP, " - ", S.XGROUP, " - ", A.ROUTERID, " - ", S.XROUTERID, " - ", A.DEVICEID, " - ", S.DEVICEID)
//		if A.GROUP == S.XGROUP && A.ROUTERID == S.XROUTERID && A.DEVICEID == S.DEVICEID {
//			// GLOBAL_STATE.ActiveAccessPoint = GLOBAL_STATE.AccessPoints[i]
//			return GLOBAL_STATE.PrivateNodes[i]
//		}
//	}
//
//	return nil
//}

func GetLogsForCLI() (*GeneralLogResponse, error) {
	defer RecoverAndLogToFile()

	R := &GeneralLogResponse{
		Content: make([]string, 0),
		Time:    make([]string, 0),
		Color:   make([]string, 0),
	}

	for i := range L.LOGS {
		if L.LOGS[i] == "" {
			continue
		}

		splitLine := strings.Split(L.LOGS[i], " || ")

		R.Content = append(R.Content, strings.Join(splitLine[2:], " "))
		R.Time = append(R.Time, splitLine[0])
		R.Function = append(R.Function, splitLine[1])
	}

	return R, nil
}

func HTTPS_GetLogs(e echo.Context) (err error) {
	defer RecoverAndLogToFile()

	Count := 0
	for i := range L.LOGS {
		if L.LOGS[i] != "" {
			Count++
		}
	}

	R := &GeneralLogResponse{
		Content: make([]string, 0),
		Time:    make([]string, 0),
		Color:   make([]string, 0),
	}

	for i := len(L.LOGS) - 1; i >= 0; i-- {
		if L.LOGS[i] == "" {
			continue
		}

		if strings.Contains(L.LOGS[i], "ERR") {
			R.Color = append(R.Color, "error")
		} else if strings.Contains(L.LOGS[i], "ERROR") {
			R.Color = append(R.Color, "error")
		} else if strings.Contains(L.LOGS[i], "err") {
			R.Color = append(R.Color, "error")
		} else if strings.Contains(L.LOGS[i], "error") {
			R.Color = append(R.Color, "error")
		} else {
			R.Color = append(R.Color, "")
		}

		splitLine := strings.Split(L.LOGS[i], " || ")

		R.Content = append(R.Content, strings.Join(splitLine[2:], " "))
		R.Time = append(R.Time, splitLine[0])
		R.Function = append(R.Function, splitLine[1])

	}

	return e.JSON(200, R)
}

func REF_ConnectToAccessPoint(ConnectionFromUser *ConnectionRequest) (code int, errm error) {
	defer RecoverAndLogToFile()
	start := time.Now()

	defer func() {
		runtime.GC()
	}()

	if !GLOBAL_STATE.ConfigInitialized {
		return 400, errors.New("the application is still initializing default configurations, please wait a few seconds")
	}

	// if !GLOBAL_STATE.ClientReady {
	// 	return 400, errors.New("the VPN is not ready to connect, please wait a moment and try again")
	// }

	fmt.Println("CR")
	fmt.Println(ConnectionFromUser)
	fmt.Println(ConnectionFromUser.RouterIndex)
	fmt.Println(ConnectionFromUser.NodeIndex)
	fmt.Println(ConnectionFromUser.UserID)
	fmt.Println(ConnectionFromUser.DeviceToken)

	CreateLog("connect", "Creating a route to VPN")
	_ = tunnels.IP_AddRoute(GLOBAL_STATE.ActiveRouter.IP, DEFAULT_GATEWAY.String(), "0")

	VPNC := new(VPNConnection)
	VPNC.ID = uuid.NewString()
	var err error

	// -------------------------------------------
	// -------------------------------------------
	//
	// ROUTER CONNECTION INITIALIZATION
	//
	// -------------------------------------------
	// -------------------------------------------
	CreateLog("connect", "Connecting to router")
	ARS, err := REF_ConnectToRouter(
		ConnectionFromUser.RouterIndex,
		ConnectionFromUser.RouterProtocol,
		ConnectionFromUser.RouterPort,
	)
	if err != nil {
		CreateErrorLog("", "Unable to open tunnel to active router: ", err)
		return 500, errors.New("error in router tunnel")
	}

	_, err = ARS.Write([]byte{28, 0, 1, 3})
	if err != nil {
		CreateErrorLog("connect", "unable to send initialization code to router", err)
		return 500, errors.New("")
	}

	EARS, err := tp.NewSocketWrapper(ARS, tp.AES256)
	if err != nil {
		CreateErrorLog("connect", "unable to create encryption seal", err)
		return 500, errors.New("")
	}

	err = EARS.InitHandshake()
	if err != nil {
		CreateErrorLog("connect", "unable to handshake with router:", err)
		return 500, errors.New("")
	}

	SessionFromUserBytes, err := json.Marshal(ConnectionFromUser)
	if err != nil {
		CreateErrorLog("connect", "Unable to marshal hello response: ", err)
		return 500, errors.New("")
	}

	fmt.Println("SEND:", string(SessionFromUserBytes))
	// fmt.Println("NN:", len(SessionFromUserBytes))
	_, err = EARS.Write(SessionFromUserBytes)
	if err != nil {
		CreateErrorLog("connect", "unable to send session to router:", err)
		return 500, errors.New("")
	}
	// fmt.Println("N:", n)

	// -------------------------------------------
	// -------------------------------------------
	//
	// EXIT NODE CONNECTION INITIALIZATION
	//
	// -------------------------------------------
	// -------------------------------------------

	VPNC.EVPNS, err = tp.NewSocketWrapper(ARS, tp.AES256)
	if err != nil {
		CreateErrorLog("connect", "unable to create encryption seal for vpn endpoint", err)
		return 500, errors.New("")
	}
	err = VPNC.EVPNS.InitHandshake()
	if err != nil {
		CreateErrorLog("connect", "unable to handshake with VPN endpoint", err)
		return 500, errors.New("")
	}

	// encryptedBytes := make([]byte, math.MaxUint16)
	// decryptedBytes := make([]byte, math.MaxUint16)
	_, responseBytes, err := VPNC.EVPNS.Read()
	if err != nil {
		CreateErrorLog("connect", "unable to receive session from router:", err)
		return 500, errors.New("")
	}

	fmt.Println("RSP", string(responseBytes))

	VPNC.Session = new(CLIENT_SESSION)
	err = json.Unmarshal(responseBytes, VPNC.Session)
	if err != nil {
		CreateErrorLog("connect", "Unable to parse response from router: ", err)
		return 500, errors.New("")
	}
	VPNC.Session.Created = time.Now()

	// TODO
	VPNC.Name = ConnectionFromUser.IFName
	VPNC.Address = ConnectionFromUser.IPv4Address
	VPNC.AddressNetIP = net.ParseIP(VPNC.Address)

	VPNC.StartPort = VPNC.Session.StartPort
	VPNC.EndPort = VPNC.Session.EndPort

	// THIS IS THE IP USED WHEN CHANGING PACKETS
	VPNC.EP_VPNSrcIP[0] = VPNC.Session.InterfaceIP[0]
	VPNC.EP_VPNSrcIP[1] = VPNC.Session.InterfaceIP[1]
	VPNC.EP_VPNSrcIP[2] = VPNC.Session.InterfaceIP[2]
	VPNC.EP_VPNSrcIP[3] = VPNC.Session.InterfaceIP[3]

	VPNC.PingReceived = time.Now()

	VPNC.Tun, err = CB_CreateNewTunnelInterface(
		VPNC.Name,
		VPNC.Address,
		"255.255.255.0",
		ConnectionFromUser.MTU,
		ConnectionFromUser.TxQueueLen,
		ConnectionFromUser.Persistent,
	)
	if err != nil {
		CreateErrorLog("connect", "unable to create tunnel interface", err)
		return 500, errors.New("")
	}

	// err = VPNC.BuildNATMap(VPNC.Node)
	// if err != nil {
	// 	CreateErrorLog("connect", "unable to build NAT map", err)
	// 	return 500, errors.New("")
	// }

	err = VPNC.Tun.PreConnect()
	if err != nil {
		CreateErrorLog("connect", "unable to configure tunnel interface", err)
		return 500, errors.New("")
	}

	// err = VPNC.Tun.Connect()
	// if err != nil {
	// 	CreateErrorLog("connect", "unable to configure tunnel interface", err)
	// 	return 500, errors.New("")
	// }

	CT_LOCK.Lock()
	for i, v := range CONNECTIONS {
		if v.Name != VPNC.Name {
			continue
		}
		_ = v.EVPNS.SOCKET.Close()
		_ = v.Tun.Close()
		delete(CONNECTIONS, i)
	}

	CONNECTIONS[VPNC.ID] = VPNC
	CT_LOCK.Unlock()

	VPNC.Connected = true
	go VPNC.ReadFromLocalSocket()
	go VPNC.ReadFromRouterSocket()

	CreateLog("connect", "Session is ready - it took ", fmt.Sprintf("%.0f", math.Abs(time.Since(start).Seconds())), " seconds to connect")

	return 200, nil
}

func GetQRCode(LF *TWO_FACTOR_CONFIRM) (QR *QR_CODE, err error) {
	if LF.Email == "" {
		return nil, errors.New("email missing")
	}

	// According to golang 1.20 this is deprecated
	// Remove once we upgrade to 1.20

	b := make([]rune, 16)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	TOTP := strings.ToUpper(string(b))

	authenticatorAppURL := gotp.NewDefaultTOTP(TOTP).ProvisioningUri(LF.Email, "NicelandVPN")

	QR = new(QR_CODE)
	QR.Value = authenticatorAppURL

	return QR, nil
}
