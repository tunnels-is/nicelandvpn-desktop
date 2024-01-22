package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xlzd/gotp"
	tp "github.com/zveinn/tcpcrypt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		// C.ManualRouter = false

		err := REF_RefreshRouterList()
		if err != nil {
			CreateErrorLog("", "Unable to find the best router for your connection: ", err)
		}

	} else {
		// C.ManualRouter = true

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

	// The domain being used here is an old domain that needs to be replaced.
	// This method uses a custom dialer which does not DNS resolve.
	responseBytes, code, err := SendRequestToControllerProxy(FR.Method, FR.Path, FR.JSONData, "api.atodoslist.net", FR.Timeout)
	if err != nil {
		CreateLog("", "(ROUTER/CONTROLLER) // code: ", code, " // err:", err)
		if code != 0 {
			return nil, code, errors.New(string(responseBytes))
		} else {
			return nil, 500, errors.New("unable to contact controller")
		}
	}

	var respJSON interface{}
	if len(responseBytes) != 0 {
		err = json.Unmarshal(responseBytes, &respJSON)
		if err != nil {
			CreateErrorLog("", "Could not process forward request: ", err)
			return nil, 400, errors.New("unknown error, please try again in a moment")
		}
	}

	return respJSON, code, nil
}

var NEXT_SERVER_REFRESH time.Time

func SetRouterFile(path string) error {
	defer RecoverAndLogToFile()

	C.RouterFilePath = path

	err := SaveConfig(C)
	if err != nil {
		CreateErrorLog("config", "Unable to save config: ", err)
		return errors.New("unable to save config")
	}

	err = REF_RefreshRouterList()
	if err != nil {
		return errors.New(err.Error())
	}

	CreateLog("config", "Router file updated")
	return nil
}

func SetConfig(config *Config) error {
	defer RecoverAndLogToFile()

	fmt.Println(config)
	fmt.Println(config.Connections)

	C.LogBlockedDomains = config.LogBlockedDomains
	if slices.Compare(C.EnabledBlockLists, config.EnabledBlockLists) != 0 {
		C.EnabledBlockLists = config.EnabledBlockLists
		for i := range C.EnabledBlockLists {
			GLOBAL_STATE.BLists[i].Enabled = true
		}
		BuildDomainBlocklist()
	}
	C.EnabledBlockLists = config.EnabledBlockLists

	if !config.DebugLogging {
		for i := range L.LOGS {
			L.LOGS[i] = ""
		}
	}

	GLOBAL_STATE.C = config

	// if dnsWasChanged {
	// 	err := ChangeDNSWhileConnected()
	// 	if err != nil {
	// 		return errors.New("unable to update DNS on tunnel interface")
	// 	}
	// }

	err := SaveConfig(config)
	if err != nil {
		CreateErrorLog("config", "Unable to save config: ", err)
		return errors.New("unable to save config")
	}

	CreateLog("config", "Config update || new config: ", *config)
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

	lines := make([]string, 0)
	for i := len(L.LOGS) - 1; i >= 0; i-- {
		if L.LOGS[i] == "" {
			continue
		}
		lines = append(lines, L.LOGS[i])
	}

	return e.JSON(200, lines)
}

func REF_ConnectToAccessPoint(ConnectRequest *ConnectionRequest) (code int, errm error) {
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
	fmt.Println(ConnectRequest.ID)
	fmt.Println(ConnectRequest.UserID)
	fmt.Println(ConnectRequest.DeviceToken)
	fmt.Println("COUNTRY", ConnectRequest.Country)

	CreateLog("connect", "Creating a route to VPN")

	var VpnMeta *VPNConnectionMETA
	VPNConnection := new(VPNConnection)
	if ConnectRequest.Country != "" {
		for i, v := range GLOBAL_STATE.C.Connections {
			if v.Tag == "Default" {
				VpnMeta = GLOBAL_STATE.C.Connections[i]
			}
		}
	} else {
		for i, v := range GLOBAL_STATE.C.Connections {
			if v.ID == ConnectRequest.ID {
				VpnMeta = GLOBAL_STATE.C.Connections[i]
			}
		}
	}
	if VpnMeta == nil {
		CreateErrorLog("", "vpn connection metadata not found for tag: ", ConnectRequest.ID)
		return 500, errors.New("error in router tunnel")
	}
	VPNConnection.Meta = VpnMeta
	VpnMeta.Initialize()
	fmt.Println("NODEID:", VpnMeta.NodeID)

	var err error

	// -------------------------------------------
	// -------------------------------------------
	//
	// ROUTER CONNECTION INITIALIZATION
	//
	// -------------------------------------------
	// -------------------------------------------
	CreateLog("connect", "Connecting to router")

	ConnectRequest.NodeID, err = primitive.ObjectIDFromHex(VpnMeta.NodeID)
	if err != nil {
		CreateErrorLog("connect", "node id format is invalid: ", err)
		return 400, errors.New("invalid node id")
	}

	var routerIndexForConnection int
	if VpnMeta.AutomaticRouter {
		routerIndexForConnection = GLOBAL_STATE.ActiveRouter.ListIndex
	} else {
		routerIndexForConnection = VpnMeta.RouterIndex
	}

	if ConnectRequest.Country == "" {
		ConnectRequest.Country = VpnMeta.Country
	}

	ARS, err := REF_ConnectToRouter(
		routerIndexForConnection,
		VpnMeta.RouterProtocol,
		VpnMeta.RouterPort,
	)
	if err != nil {
		CreateErrorLog("", "Unable to open tunnel to active router: ", err)
		return 500, errors.New("error in router tunnel")
	}

	// 3 == encryption protocol
	_, err = ARS.Write([]byte{CODE_ConnectToNode, 0, 1, 3})
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

	ConnectRequestBytes, err := json.Marshal(ConnectRequest)
	if err != nil {
		CreateErrorLog("connect", "Unable to marshal hello response: ", err)
		return 500, errors.New("")
	}

	fmt.Println("SEND:", string(ConnectRequestBytes))
	// fmt.Println("NN:", len(SessionFromUserBytes))
	_, err = EARS.Write(ConnectRequestBytes)
	if err != nil {
		CreateErrorLog("connect", "unable to send session to router:", err)
		return 500, errors.New("")
	}

	// -------------------------------------------
	// -------------------------------------------
	//
	// SESSION COMES BACK FROM CONTROLLER
	// - technically the router writes is back to us
	// but the controller populated the data.
	//
	// -------------------------------------------
	// -------------------------------------------
	_, bytesFromController, err := EARS.Read()
	if err != nil {
		CreateErrorLog("connect", "unable to receive connect request from controller", err)
		return 500, errors.New("")
	}
	err = json.Unmarshal(bytesFromController, ConnectRequest)
	if err != nil {
		CreateErrorLog("connect", "data from controller is corrupt, please contact custom support if this problem persists", err)
		return 500, errors.New("unable to parse response from controller")
	}

	if ConnectRequest.ErrorCode != 0 {
		fmt.Println("RESPONSE:", ConnectRequest)
		CreateErrorLog("connect", "error message from controller:", ConnectRequest.Error, ConnectRequest.ErrorCode)
		return 500, errors.New(ConnectRequest.Error)
	}

	// -------------------------------------------
	// -------------------------------------------
	//
	// EXIT NODE CONNECTION INITIALIZATION
	//
	// -------------------------------------------
	// -------------------------------------------

	VPNConnection.EVPNS, err = tp.NewSocketWrapper(ARS, tp.EncType(VpnMeta.EncryptionProtocol))
	if err != nil {
		CreateErrorLog("connect", "unable to create encryption seal for vpn endpoint", err)
		return 500, errors.New("")
	}
	err = VPNConnection.EVPNS.InitHandshake()
	if err != nil {
		CreateErrorLog("connect", "unable to handshake with VPN endpoint", err)
		return 500, errors.New("")
	}

	// encryptedBytes := make([]byte, math.MaxUint16)
	// decryptedBytes := make([]byte, math.MaxUint16)
	_, responseBytes, err := VPNConnection.EVPNS.Read()
	if err != nil {
		CreateErrorLog("connect", "unable to receive session from router:", err)
		return 500, errors.New("")
	}

	fmt.Println("RSP", string(responseBytes))

	VPNConnection.Session = new(CLIENT_SESSION)
	err = json.Unmarshal(responseBytes, VPNConnection.Session)
	if err != nil {
		CreateErrorLog("connect", "controller sent a non-json response", err)
		return 500, errors.New(string(responseBytes))
	}
	VPNConnection.Session.Created = time.Now()

	VPNConnection.AddressNetIP = net.ParseIP(VpnMeta.IPv4Address).To4()
	VPNConnection.StartPort = VPNConnection.Session.StartPort
	VPNConnection.EndPort = VPNConnection.Session.EndPort

	// THIS IS THE IP USED WHEN CHANGING PACKETS
	fmt.Println("IFIP")
	fmt.Println(VPNConnection.Session.InterfaceIP)
	fmt.Println(VPNConnection.AddressNetIP)

	to4 := VPNConnection.Session.InterfaceIP.To4()
	VPNConnection.EP_VPNSrcIP[0] = to4[0]
	VPNConnection.EP_VPNSrcIP[1] = to4[1]
	VPNConnection.EP_VPNSrcIP[2] = to4[2]
	VPNConnection.EP_VPNSrcIP[3] = to4[3]

	VPNConnection.PingReceived = time.Now()

	VPNConnection.Tun, err = CB_CreateNewTunnelInterface(
		VpnMeta.IFName,
		VpnMeta.IPv4Address,
		VpnMeta.NetMask,
		VpnMeta.TxQueueLen,
		VpnMeta.MTU,
		VpnMeta.Persistent,
	)
	if err != nil {
		CreateErrorLog("connect", "unable to create tunnel interface", err)
		return 500, errors.New("")
	}

	err = VPNConnection.BuildNATMap()
	if err != nil {
		CreateErrorLog("connect", "unable to build NAT map", err)
		return 500, errors.New("")
	}

	err = VPNConnection.Tun.PreConnect(VPNConnection.Meta)
	if err != nil {
		CreateErrorLog("connect", "unable to configure tunnel interface", err)
		return 500, errors.New("")
	}

	err = VPNConnection.Tun.Connect(VPNConnection.Meta)
	if err != nil {
		CreateErrorLog("connect", "unable to configure tunnel interface", err)
		return 500, errors.New("")
	}

	CONNECTIONS.RemoveConnection(VPNConnection.Meta.Tag)
	CONNECTIONS.AddConnection(VPNConnection)

	VPNConnection.Connected = true
	go VPNConnection.ReadFromLocalSocket()
	go VPNConnection.ReadFromRouterSocket()

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
