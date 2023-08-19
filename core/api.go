package core

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/xlzd/gotp"
)

func ControllerCustomDialer(ctx context.Context, network string, addr string) (net.Conn, error) {
	return OpenProxyTunnelToRouter(ctx)
}

func LocalhostCustomDialer(ctx context.Context, network, addr string) (net.Conn, error) {
	return OpenProxyTunnelToRouter(ctx)
}

func OpenProxyTunnelToRouter(ctx context.Context) (TCP_CONN net.Conn, err error) {

	TCP_CONN, err = net.Dial("tcp", GLOBAL_STATE.ActiveRouter.IP+":443")
	if err != nil {
		CreateErrorLog("", "Could not dial router: ", GLOBAL_STATE.ActiveRouter.IP, err)
		return
	}

	return
}

func ResetEverything() {
	defer RecoverAndLogToFile()

	CreateLog("START", "")
	CleanupWithStateLock()
}

func CleanupWithStateLock() {
	defer STATE_LOCK.Unlock()
	defer RecoverAndLogToFile()
	STATE_LOCK.Lock()

	DisconnectFromRouter(AS)
	_ = SetInterfaceStateToDown()

	RestoreIPv6()
	RestoreDNS()
	InstantlyCleanAllTCPPorts()
	InstantlyCleanAllUDPPorts()

	SetGlobalStateAsDisconnected()
}

func SwitchRouter(Tag string) (code int, err error) {
	defer STATE_LOCK.Unlock()
	defer RecoverAndLogToFile()
	STATE_LOCK.Lock()

	if GLOBAL_STATE.ClientStartupError {
		return 400, errors.New("there is a problem with the background service, please check your logs")
	}

	CreateLog("START", "")

	if GLOBAL_STATE.Connecting {
		CreateLog("loader", "unable to change routing while nicelandVPN is connecting")
		return 400, errors.New("unable to change routing while connecting")
	} else if GLOBAL_STATE.Connected {
		CreateLog("loader", "unable to change routing while nicelandVPN is connected")
		return 400, errors.New("unable to change routing while connected")
	} else if GLOBAL_STATE.Exiting {
		CreateLog("loader", "unabel to change routing while nicelandVPN is exiting")
		return 400, errors.New("unable to change routing while exiting")
	}

	if Tag == "" {
		C.ManualRouter = false
		if GLOBAL_STATE.LastRouterPing.IsZero() {
			PingAllRouters()
		}

		if time.Since(GLOBAL_STATE.LastRouterPing).Seconds() > 120 {
			PingAllRouters()
		}

		index, err := GetLowestLatencyRouter()
		if err != nil {
			CreateErrorLog("loader", "Could not find lowest latency router")
			return 400, errors.New("unable to find lowest latency router")
		}
		SetActiveRouter(index)

	} else {
		C.ManualRouter = true

		for i := range GLOBAL_STATE.RoutersList {
			if GLOBAL_STATE.RoutersList[i] != nil {
				if GLOBAL_STATE.RoutersList[i].Tag == Tag {
					SetActiveRouter(i)
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
				ServerName: domain,
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

var LAST_ROUTER_AND_ACCESS_POINT_UPDATE = time.Now()

func GetRoutersAndAccessPoints() (interface{}, int, error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.ActiveRouter == nil {
		return nil, 500, errors.New("active router not found, please wait a moment")
	}

	if !GLOBAL_STATE.LastAccessPointUpdate.IsZero() {
		since := time.Since(GLOBAL_STATE.LastAccessPointUpdate).Seconds()
		GLOBAL_STATE.SecondsUntilAccessPointUpdate = 55 - int(since)
		if since < 55 {
			return nil, 200, nil
		}
	}

	GLOBAL_STATE.LastAccessPointUpdate = time.Now()
	GLOBAL_STATE.SecondsUntilAccessPointUpdate = 55

	responseBytes, code, err := SendRequestToLocalhostProxy("GET", "v1/a", nil, 10000)
	if err != nil {
		CreateLog("", "(ROUTER/API) // code: ", code, " // err:", err)
		if code != 0 {
			GLOBAL_STATE.LastAccessPointUpdate = time.Now().Add(-45 * time.Second)
			return nil, code, errors.New(string(responseBytes))
		} else {
			GLOBAL_STATE.LastAccessPointUpdate = time.Now().Add(-45 * time.Second)
			return nil, code, errors.New("unable to contact controller")
		}
	}

	RoutersAndAccessPoints := new(CONTROLL_PUBLIC_DEVCE_RESPONSE)

	err = json.Unmarshal(responseBytes, RoutersAndAccessPoints)
	if err != nil {
		GLOBAL_STATE.LastAccessPointUpdate = time.Now().Add(-45 * time.Second)
		CreateErrorLog("", "Could not process forward request: ", err)
		return nil, 400, errors.New("unknown error, please try again in a moment")
	}

	for ii := range RoutersAndAccessPoints.Routers {
		RR := RoutersAndAccessPoints.Routers[ii]

		exists := false
		for i := range GLOBAL_STATE.RoutersList {
			R := GLOBAL_STATE.RoutersList[i]
			if R == nil {
				continue
			}
			if RR.IP == R.IP {
				exists = true
			}
		}

		if !exists {
			for i := range GLOBAL_STATE.RoutersList {
				if GLOBAL_STATE.RoutersList[i] == nil {
					GLOBAL_STATE.RoutersList[i] = RoutersAndAccessPoints.Routers[ii]
					break
				}
			}
		}

	}

	for ii := range RoutersAndAccessPoints.Routers {
		RR := RoutersAndAccessPoints.Routers[ii]

		for i := range GLOBAL_STATE.RoutersList {
			R := GLOBAL_STATE.RoutersList[i]
			if R == nil {
				continue
			}
			if RR.IP == R.IP {
				GLOBAL_STATE.RoutersList[i].GROUP = RR.GROUP
				GLOBAL_STATE.RoutersList[i].ROUTERID = RR.ROUTERID
				GLOBAL_STATE.RoutersList[i].Country = RR.Country
				GLOBAL_STATE.RoutersList[i].AvailableMbps = RR.AvailableMbps
				GLOBAL_STATE.RoutersList[i].Slots = RR.Slots
				GLOBAL_STATE.RoutersList[i].AvailableSlots = RR.AvailableSlots
				GLOBAL_STATE.RoutersList[i].AEBP = RR.AEBP
				GLOBAL_STATE.RoutersList[i].AIBP = RR.AIBP
				GLOBAL_STATE.RoutersList[i].CPUP = RR.CPUP
				GLOBAL_STATE.RoutersList[i].RAMUsage = RR.RAMUsage
				GLOBAL_STATE.RoutersList[i].DiskUsage = RR.DiskUsage
				GLOBAL_STATE.RoutersList[i].Online = RR.Online
				R = GLOBAL_STATE.RoutersList[i]

				includeMSInScoring := true
				if R.MS == 31337 {
					includeMSInScoring = false
				}

				var baseScore float64 = 10
				AIBPScore := 100 / R.AIBP
				AEBPScore := 100 / R.AEBP
				if AIBPScore < 1 || AIBPScore == 1 {
					AIBPScore = 0
				} else if AIBPScore > 4 {
					AIBPScore = 4
				}
				if AEBPScore < 1 || AEBPScore == 1 {
					AEBPScore = 0
				} else if AEBPScore > 4 {
					AEBPScore = 4
				}

				var MSScore float64 = 0
				if includeMSInScoring {
					if R.MS < 20 {
						MSScore = 1
					} else if R.MS < 100 {
						MSScore = 2
					} else if R.MS < 200 {
						MSScore = 3
					} else if R.MS < 300 {
						MSScore = 4
					} else if R.MS < 400 {
						MSScore = 5
					}
				}
				var SLOTScore float64 = float64(R.AvailableSlots / R.Slots)
				R.Score = int(baseScore - AEBPScore - SLOTScore - AIBPScore - MSScore)
			}
		}

	}

	GLOBAL_STATE.Routers = nil
	GLOBAL_STATE.Routers = make([]*ROUTER, 0)
	GLOBAL_STATE.AvailableCountries = make([]string, 0)
	for i := range GLOBAL_STATE.RoutersList {
		if GLOBAL_STATE.RoutersList[i] == nil {
			continue
		}

		countryExists := false
		for ii := range GLOBAL_STATE.AvailableCountries {
			if GLOBAL_STATE.AvailableCountries[ii] == GLOBAL_STATE.RoutersList[i].Country {
				countryExists = true
			}
		}

		if !countryExists {
			GLOBAL_STATE.AvailableCountries = append(GLOBAL_STATE.AvailableCountries, GLOBAL_STATE.RoutersList[i].Country)
		}

		GLOBAL_STATE.Routers = append(GLOBAL_STATE.Routers, GLOBAL_STATE.RoutersList[i])
	}

	sort.Slice(GLOBAL_STATE.Routers, func(a, b int) bool {
		if GLOBAL_STATE.Routers[a] == nil {
			return false
		}
		if GLOBAL_STATE.Routers[b] == nil {
			return false
		}
		return GLOBAL_STATE.Routers[a].Score > GLOBAL_STATE.Routers[b].Score
	})

	GLOBAL_STATE.AccessPoints = RoutersAndAccessPoints.AccessPoints
	for i := range GLOBAL_STATE.AccessPoints {
		A := GLOBAL_STATE.AccessPoints[i]

		for ii := range GLOBAL_STATE.RoutersList {
			R := GLOBAL_STATE.RoutersList[ii]
			if R == nil {
				continue
			}

			if R.GROUP == A.GROUP && R.ROUTERID == A.ROUTERID {
				GLOBAL_STATE.AccessPoints[i].Router = GLOBAL_STATE.RoutersList[ii]
			}
		}
	}

	if len(GLOBAL_STATE.AccessPoints) == 0 {
		GLOBAL_STATE.LastAccessPointUpdate = time.Now().Add(-45 * time.Second)
	}

	sort.Slice(GLOBAL_STATE.AccessPoints, func(a, b int) bool {
		if GLOBAL_STATE.AccessPoints[a].Router == nil {
			return false
		}
		if GLOBAL_STATE.AccessPoints[b].Router == nil {
			return false
		}
		return GLOBAL_STATE.AccessPoints[a].Router.Score > GLOBAL_STATE.AccessPoints[b].Router.Score
	})

	return nil, code, nil
}

func ForwardToRouter(FR *FORWARD_REQUEST) (interface{}, int, error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.ClientStartupError {
		return nil, 500, errors.New("there is a problem with the background service, please check your logs")
	}

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

	if GLOBAL_STATE.ClientStartupError {
		return nil, 500, errors.New("there is a problem with the background service, please check your logs")
	}

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
	defer STATE_LOCK.Unlock()
	defer RecoverAndLogToFile()
	STATE_LOCK.Lock()

	CreateLog("START", "")

	if GLOBAL_STATE.ClientStartupError {
		return errors.New("there is a problem with the background service, please check your logs")
	}

	if GLOBAL_STATE.Connecting {
		CreateLog("loader", "Unable to change routers while nicelandVPN is connecting")
		return errors.New("unable to change routers while nicelandVPN is connecting")
	} else if GLOBAL_STATE.Connected {
		CreateLog("loader", "Unable to change routers while nicelandVPN is connected")
		return errors.New("unable to change routers while nicelandVPN is connected")
	} else if GLOBAL_STATE.Exiting {
		CreateLog("loader", "Unable to change routers while nicelandVPN is exiting")
		return errors.New("unable to change routers while nicelandVPN is exiting")
	}

	C.RouterFilePath = path
	C.ManualRouter = false

	err := SaveConfig()
	if err != nil {
		CreateErrorLog("", "Unable to save config: ", err)
		return errors.New("unable to save config")
	}

	GLOBAL_STATE.LastRouterPing = time.Now().AddDate(0, 0, 1)

	err = RefreshRouterList()
	if err != nil {
		return errors.New(err.Error())
	}

	CreateLog("loader", "Router file updated")
	return nil
}

func SetConfig(SF *CONFIG_FORM) error {
	defer STATE_LOCK.Unlock()
	defer RecoverAndLogToFile()
	STATE_LOCK.Lock()

	CreateLog("START", "")

	if GLOBAL_STATE.Connecting {
		CreateLog("loader", "unable to change config while nicelandVPN is connecting")
		return errors.New("unable to change config while nicelandVPN is connecting")
	} else if GLOBAL_STATE.Exiting {
		CreateLog("loader", "unable to change config while nicelandVPN is exiting")
		return errors.New("unable to change config while nicelandVPN is exiting")
	}

	if SF.Version != "" {
		C.Version = SF.Version
	}

	C.RouterFilePath = SF.RouterFilePath
	C.DebugLogging = SF.DebugLogging
	C.AutoReconnect = SF.AutoReconnect
	C.KillSwitch = SF.KillSwitch

	if SF.PrevSession != nil {
		C.PrevSession = SF.PrevSession
	}

	if !SF.DebugLogging {
		DumpLoadingLogs(L)
		for i := range L.GENERAL {
			L.GENERAL[i] = ""
		}
	}

	var dnsWasChanged bool = false
	if SF.DNS1 != "" {
		if C.DNS1 != SF.DNS1 {
			C.DNS1 = SF.DNS1
			C.DNSIP = net.ParseIP(C.DNS1).To4()
			if len(C.DNSIP) < 4 {
				return errors.New("DNS1 is invalid or empty")
			}
			C.DNS1Bytes = [4]byte{C.DNSIP[0], C.DNSIP[1], C.DNSIP[2], C.DNSIP[3]}
			dnsWasChanged = true
		}
	} else {

		CreateLog("loader", "Error while updating config || DNS1 is invalid: ", SF.DNS1)
		return errors.New("DNS1 is invalid or empty")
	}

	if C.DNS2 != SF.DNS2 {
		C.DNS2 = SF.DNS2
		dnsWasChanged = true
	}

	if dnsWasChanged && GLOBAL_STATE.Connected {
		err := ChangeDNSWhileConnected()
		if err != nil {
			return errors.New("unable to update DNS on tunnel interface")
		}
	}

	err := SaveConfig()
	if err != nil {
		CreateErrorLog("", "Unable to save config: ", err)
		return errors.New("unable to save config")
	}

	CreateLog("", "Config update || new config: ", *C)
	return nil
}

func PrepareState() {
	defer RecoverAndLogToFile()

	GLOBAL_STATE.EgressPackets = EGRESS_PACKETS
	GLOBAL_STATE.IngressPackets = INGRESS_PACKETS
	ubps := GLOBAL_STATE.UMbps
	utext := "bps"
	dbps := GLOBAL_STATE.DMbps
	dtext := "bps"
	if ubps > 1100000 {
		utext = "Mbps"
		ubps = ubps / 1000000
	} else if ubps > 1100 {
		utext = "Kbps"
		ubps = ubps / 1000
	}
	GLOBAL_STATE.UMbpsString = fmt.Sprintf("%d %s", ubps, utext)

	if dbps > 1100000 {
		dtext = "Mbps"
		dbps = dbps / 1000000
	} else if dbps > 1100 {
		dtext = "Kbps"
		dbps = dbps / 1000
	}

	GLOBAL_STATE.DMbpsString = fmt.Sprintf("%d %s", dbps, dtext)
	GLOBAL_STATE.SecondsSincePingFromRouter = fmt.Sprintf("%.0f seconds", time.Since(GLOBAL_STATE.PingReceivedFromRouter).Seconds())

	if GLOBAL_STATE.ActiveSession != nil {
		seconds := time.Since(GLOBAL_STATE.ActiveSession.Created).Seconds()
		label := "seconds"

		if seconds > 60 && seconds < 120 {
			label = "minute"
			seconds = seconds / 60
		} else if seconds >= 120 {
			label = "minutes"
			seconds = seconds / 60
		}

		GLOBAL_STATE.ConnectedTimer = fmt.Sprintf("%.0f %s", seconds, label)
	}

	if GLOBAL_STATE.ActiveSession != nil {
		S := GLOBAL_STATE.ActiveSession
		for i := range GLOBAL_STATE.AccessPoints {
			A := GLOBAL_STATE.AccessPoints[i]
			if A.GROUP == S.XGROUP && A.ROUTERID == S.XROUTERID {
				GLOBAL_STATE.ActiveAccessPoint = GLOBAL_STATE.AccessPoints[i]
			}
		}
	}

}

func GetLogsForCLI() (*GeneralLogResponse, error) {
	defer RecoverAndLogToFile()

	R := &GeneralLogResponse{
		Content: make([]string, 0),
		Time:    make([]string, 0),
		Color:   make([]string, 0),
	}

	for i := range L.GENERAL {
		if L.GENERAL[i] == "" {
			continue
		}

		splitLine := strings.Split(L.GENERAL[i], " || ")

		R.Content = append(R.Content, strings.Join(splitLine[2:], " "))
		R.Time = append(R.Time, splitLine[0])
		R.Function = append(R.Function, splitLine[1])
	}

	return R, nil
}

func GetLogs(lengthFromJavascript int) (*GeneralLogResponse, error) {
	defer RecoverAndLogToFile()

	Count := 0
	for i := range L.GENERAL {
		if L.GENERAL[i] != "" {
			Count++
		}
	}

	if lengthFromJavascript == Count {
		return nil, nil
	}

	R := &GeneralLogResponse{
		Content: make([]string, 0),
		Time:    make([]string, 0),
		Color:   make([]string, 0),
	}

	for i := len(L.GENERAL) - 1; i >= 0; i-- {
		if L.GENERAL[i] == "" {
			continue
		}

		if strings.Contains(L.GENERAL[i], "ERR") {
			R.Color = append(R.Color, "error")
		} else if strings.Contains(L.GENERAL[i], "ERROR") {
			R.Color = append(R.Color, "error")
		} else if strings.Contains(L.GENERAL[i], "err") {
			R.Color = append(R.Color, "error")
		} else if strings.Contains(L.GENERAL[i], "error") {
			R.Color = append(R.Color, "error")
		} else {
			R.Color = append(R.Color, "")
		}

		splitLine := strings.Split(L.GENERAL[i], " || ")

		R.Content = append(R.Content, strings.Join(splitLine[2:], " "))
		R.Time = append(R.Time, splitLine[0])
		R.Function = append(R.Function, splitLine[1])

	}

	return R, nil
}

func GetLoadingLogs(t string) (Logs *LOADING_LOGS_RESPONSE, err error) {
	defer RecoverAndLogToFile()

	switch t {
	case "connect":
		return &LOADING_LOGS_RESPONSE{Lines: L.CONNECT}, nil
	case "disconnect":
		return &LOADING_LOGS_RESPONSE{Lines: L.DISCONNECT}, nil
	case "switch":
		return &LOADING_LOGS_RESPONSE{Lines: L.SWITCH}, nil
	case "loader":
		return &LOADING_LOGS_RESPONSE{Lines: L.PING}, nil
	default:
		CreateLog("", "Log TYPE not valid || type: ", t)
	}

	return nil, nil
}

func Disconnect() {
	defer RecoverAndLogToFile()

	CreateLog("START", "")
	C.PrevSession = nil
	CleanupWithStateLock()
	EGRESS_PACKETS = 0
	INGRESS_PACKETS = 0
	GLOBAL_STATE.ActiveAccessPoint = nil
	GLOBAL_STATE.ActiveSession = nil
	GLOBAL_STATE.IngressPackets = 0
	GLOBAL_STATE.EgressPackets = 0
	GLOBAL_STATE.UMbps = 0
	GLOBAL_STATE.DMbps = 0
	GLOBAL_STATE.DMbpsString = ""
	GLOBAL_STATE.UMbpsString = ""
	GLOBAL_STATE.ConnectedTimer = ""
}

func DisconnectFromRouter(AdapterSettings *AdapterSettings) {
	defer RecoverAndLogToFile()

	if AdapterSettings == nil {
		GLOBAL_STATE.Connected = false
		return
	}

	if AdapterSettings.TCPTunnelSocket != nil {
		_ = AdapterSettings.TCPTunnelSocket.Close()
		AdapterSettings.TCPTunnelSocket = nil
	}

	AdapterSettings.Session = nil
	AdapterSettings = nil
	GLOBAL_STATE.ActiveSession = nil
	SetGlobalStateAsDisconnected()

	CreateLog("connect", "VPN disconnected")
}

func ConnectToAccessPoint(NS *CONTROLLER_SESSION_REQUEST, startRouting bool) (S *CLIENT_SESSION, code int, errm error) {
	defer RecoverAndLogToFile()

	var router_shared_key [32]byte
	OTKResp := new(OTK_RESPONSE)
	OTKReq := new(OTK_REQUEST)
	CCCR := new(CHACHA_RESPONSE)
	var NSRespBytes []byte
	var AESKeyb *big.Int
	var AESKey [32]byte
	var CCDec []byte
	CC_DATA := new(OTK_REQUEST)

	var FINAL_OTK = new(OTK)
	var FINAL_OTKR = new(OTK_REQUEST)
	defer func() {

		if S != nil {
			S.PrivateKey = nil
		}

		router_shared_key = [32]byte{}
		OTKResp = nil
		OTKReq = nil
		CCCR = nil
		NSRespBytes = nil
		AESKeyb = nil
		AESKey = [32]byte{}
		CCDec = nil
		CC_DATA = nil
		FINAL_OTK = nil
		FINAL_OTKR = nil
	}()

	if GLOBAL_STATE.ActiveRouter == nil {
		return nil, 400, errors.New("no active router has been found, please wait for a few seconds")
	}

	if !GLOBAL_STATE.ConfigInitialized {
		return nil, 400, errors.New("the application is still initializing default configurations, please wait a few seconds")
	}

	if !GLOBAL_STATE.ClientReady {
		return nil, 400, errors.New("the VPN is not ready to connect, please wait a moment and try again")
	}

	start := time.Now()

	CreateLog("connect", "Starting Session")

	if NS.SLOTID == 0 {
		NS.SLOTID = 1
	}
	if NS.Country == "" {
		NS.Type = "connect-specific"
	} else {
		NS.Type = "connect"
	}

	CreateLog("connect", "Creating a route to VPN")
	_ = AddRoute(GLOBAL_STATE.ActiveRouter.IP)

	E := elliptic.P521()
	S = new(CLIENT_SESSION)
	S.Created = time.Now()

	var err error

	CreateLog("connect", "Generating Encryption Keys")
	S.PrivateKey, OTKReq, err = GenerateEllipticCurveAndPrivateKey()
	if err != nil {
		CreateErrorLog("connect", "Unable to generate encryption keys: ", err)
		return nil, 500, errors.New("unable to generate encryption keys")
	}

	FINAL_OTK.PrivateKey, FINAL_OTKR, err = GenerateEllipticCurveAndPrivateKey()
	if err != nil {
		CreateErrorLog("connect", "Unable to generate encryption keys: ", err)
		return nil, 500, errors.New("unable to exchange encryption keys with the server")
	}

	responseBytes, code, err := SendRequestToLocalhostProxy("POST", "v1/api", OTKReq, 10000)
	if code != http.StatusOK {
		return nil, 500, errors.New(string(responseBytes))
	}
	if err != nil {
		CreateErrorLog("connect", "Unable to exchange encryption keys: ", code, err)
		return nil, 500, errors.New("unknown Error")
	}

	err = json.Unmarshal(responseBytes, OTKResp)
	if err != nil {
		CreateErrorLog("connect", "unable to parse encryption key response from router: ", err)
		return nil, 500, errors.New("unable to exchange encryption keys with router, please try again in a moment")
	}

	a, _ := E.ScalarMult(OTKResp.X, OTKResp.Y, S.PrivateKey.D.Bytes())
	router_shared_key = sha256.Sum256(a.Bytes())

	UUID := Decrypt(OTKResp.UUID, router_shared_key[:])

	CreateLog("connect", "Key exchange complete")

	NS.TempKey = FINAL_OTKR

	NSbytes, err := json.Marshal(NS)
	if err != nil {
		CreateErrorLog("connect", "Unable to marshal hello response: ", err)
		return nil, 500, errors.New("unable to exchange encryption keys with the VPN")
	}

	NSEncBytes := Encrypt(NSbytes, router_shared_key[:])

	data := append(UUID, NSEncBytes...)

	responseBytes, code, err = SendRawBytesToLocalhostProxy("POST", "v2/api", data, 30000)
	if code != http.StatusOK {
		CreateErrorLog("connect", "Error code from router during proxy request: ", code)
		return nil, code, errors.New(string(responseBytes))
	}

	if err != nil {
		CreateErrorLog("connect", " >> Unable to create session: ", err)
		return nil, 500, errors.New("unknown Error")
	}

	CreateLog("connect", "OK from Router")
	NSRespBytes = Decrypt(responseBytes, router_shared_key[:])

	err = json.Unmarshal(NSRespBytes, S)
	if err != nil {
		CreateErrorLog("connect", "Unable to parse response from router: ", err)
		return nil, 500, errors.New("unable to create a session")
	}

	err = json.Unmarshal(S.ClientKeyResponse, CCCR)
	if err != nil {
		CreateErrorLog("connect", "Unable to parse encryption key exchange: ", err)
		return nil, 500, errors.New("unable to create a session")
	}

	AESKeyb, _ = FINAL_OTK.PrivateKey.Curve.ScalarMult(CCCR.X, CCCR.Y, FINAL_OTK.PrivateKey.D.Bytes())
	AESKey = sha256.Sum256(AESKeyb.Bytes())

	CCDec = Decrypt(CCCR.CHACHA, AESKey[:])
	err = json.Unmarshal(CCDec, CC_DATA)
	if err != nil {
		CreateErrorLog("connect", "Unable to parse encryption key exchange: ", err)
		return nil, 500, errors.New("unable to create a session")
	}

	NewAdapterSettings := new(AdapterSettings)
	NewAdapterSettings.AEAD, err = GenerateAEADFromPrivateKey(FINAL_OTK.PrivateKey, CC_DATA)
	if err != nil {
		CreateErrorLog("connect", "Unable to exchange encryption keys with the VPN: ", err)
		return nil, 500, errors.New("unable to create a session")
	}

	NewAdapterSettings.EndPort = S.EndPort
	NewAdapterSettings.StartPort = S.StartPort
	NewAdapterSettings.VPNIP = net.IP{S.VPNIP[0], S.VPNIP[1], S.VPNIP[2], S.VPNIP[3]}
	NewAdapterSettings.TCPHeader = layers.IPv4{
		SrcIP:    NewAdapterSettings.VPNIP,
		DstIP:    net.IP{0, 0, 0, 0},
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}

	NewAdapterSettings.UDPHeader = layers.IPv4{
		SrcIP:    NewAdapterSettings.VPNIP,
		DstIP:    net.IP{0, 0, 0, 0},
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
	}

	NewAdapterSettings.RoutingBuffer = CreateMETABuffer(
		CODE_CLIENT_connect_tunnel_with_handshake,
		S.GROUP,
		S.ROUTERID,
		S.SESSIONID,
		S.DEVICEID,
		0,
		0,
	)

	NewAdapterSettings.PingBuffer = CreateMETABuffer(
		CODE_CLIENT_ping,
		S.GROUP,
		S.ROUTERID,
		S.SESSIONID,
		0,
		0,
		0,
	)

	CreateLog("connect", "Connecting to router")
	ROUTER_TUNNEL, err := ConnectToActiveRouter(NewAdapterSettings.RoutingBuffer)
	if err != nil {
		CreateErrorLog("", "Unable to open tunnel to active router: ", err)
		return nil, 500, errors.New("error in router tunnel")
	}

	NewAdapterSettings.RoutingBuffer[0] = 0

	CreateLog("connect", "Session is ready - it took ", fmt.Sprintf("%.0f", math.Abs(time.Since(start).Seconds())), " seconds to connect")

	if startRouting {
		err = EnablePacketRouting()
		if err != nil {
			if !GLOBAL_STATE.Connected {
				ResetAfterFailedConnectionAttempt()
				DisconnectFromRouter(NewAdapterSettings)
			}
			return nil, 500, errors.New("unable to start routing")
		}
	}

	IGNORE_NEXT_BUFFER_ERROR = true
	if AS.TCPTunnelSocket != nil {
		_ = AS.TCPTunnelSocket.Close()
	}

	GLOBAL_STATE.ActiveRouter.TCPTunnelConnection = ROUTER_TUNNEL
	NewAdapterSettings.TCPTunnelSocket = ROUTER_TUNNEL

	// Client key response needs to be removed before
	// the session can be returned to the GUI
	S.ClientKeyResponse = nil

	AS = NewAdapterSettings
	AS.Session = S
	GLOBAL_STATE.ActiveSession = AS.Session
	AS.LastActivity = time.Now()
	GLOBAL_STATE.Connected = true
	BUFFER_ERROR = false
	C.PrevSession = NS
	GLOBAL_STATE.PingReceivedFromRouter = time.Now()

	CreateLog("connect", "VPN connection ready")

	return S, 200, nil
}

func Connect(NS *CONTROLLER_SESSION_REQUEST, initializeRouting bool) (S *CLIENT_SESSION, code int, err error) {
	defer func() {
		GLOBAL_STATE.Connecting = false
		STATE_LOCK.Unlock()
	}()
	defer RecoverAndLogToFile()
	STATE_LOCK.Lock()

	if GLOBAL_STATE.Connecting {
		return nil, 400, errors.New("the app is already trying to connect, please wait a moment")
	}

	GLOBAL_STATE.Connecting = true

	if GLOBAL_STATE.ClientStartupError {
		return nil, 400, errors.New("the app did not start normally, please review the logs to identify the issue. If the issue persists then please contact customer support")
	}

	CreateLog("START", "")

	S, CODE, err := ConnectToAccessPoint(NS, initializeRouting)
	if err != nil {
		return nil, CODE, err
	}

	return S, CODE, nil
}

func GetQRCode(LF *TWO_FACTOR_CONFIRM) (QR *QR_CODE, err error) {
	if LF.Email == "" {
		return nil, errors.New("email missing")
	}

	// According to golang 1.20 this is deprecated
	// Remove once we upgrade to 1.20
	rand.Seed(time.Now().UnixNano())

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
