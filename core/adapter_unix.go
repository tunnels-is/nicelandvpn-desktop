//go:build freebsd || linux || openbsd

package core

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	water "github.com/songgao/water"
)

type Adapter struct {
	Interface *water.Interface
}

func (A *Adapter) Close() (err error) {
	if A.Interface != nil {
		err = A.Interface.Close()
	}
	return
}
func (A *Adapter) Uninstall() (err error) {
	return
}

func GetIPv6Settings(PotentialDefault *CONNECTION_SETTINGS) {
	out, err := exec.Command("bash", "-c", "sysctl net.ipv6.conf."+PotentialDefault.IFName+".disable_ipv6").Output()
	if err != nil {
		CreateErrorLog("", err, "SYSCTL || Error getting ipv6 settings for interface: ", PotentialDefault.IFName, " || msg: ", err, " || output: ", string(out))
		return
	}

	outSplit := strings.Split(string(out), " ")
	if len(outSplit) < 3 {
		CreateErrorLog("", err, "SYSCTL || Error getting ipv6 settings for interface: ", PotentialDefault.IFName, " || output: ", string(out))
		return
	}

	if outSplit[2] == "0" {
		PotentialDefault.IPV6Enabled = false
	} else {
		PotentialDefault.IPV6Enabled = true
	}
}

func VerifyAndBackupSettings(PotentialDefault *CONNECTION_SETTINGS) (err error) {
	GetIPv6Settings(PotentialDefault)
	return
}

func FindDefaultInterfaceAndGatewayDuringStartup() (err error) {

	PotentialDefault, err := FindDefaultInterfaceAndGateway()
	if err != nil {
		CreateErrorLog("", "Could not find default interface and gateway >> ", err)
		return errors.New("")
	}

	if PotentialDefault == nil {
		CreateErrorLog("", "Could not find default interface and gateway >> Potential Default is nil")
		return errors.New("")
	}

	if PotentialDefault.DefaultRouter == "" {
		CreateErrorLog("", "Default interface had invalid Default router", PotentialDefault, " >> ", err)
		return errors.New("")
	}

	VerifyAndBackupSettings(PotentialDefault)

	GLOBAL_STATE.DefaultInterface = PotentialDefault

	CreateLog("", "NEW DEFAULT INTERFACE >> ", MAC_CONNECTION_SETTINGS)

	BackupSettingsToFile(PotentialDefault)
	return
}

func RestoreSettingsFromFile(PotentialDefault *CONNECTION_SETTINGS) {
	CreateLog("", "RESTORING SETTINGS FROM FILE")

	backupFile, err := os.Open(GenerateBaseFolderPath() + PotentialDefault.IFName + "_backup")
	if err != nil {
		CreateErrorLog("", "Unable to open backup file, please restart the application. If this problem persists contact customer support")
		return
	}

	backupBytes, err := io.ReadAll(backupFile)
	if err != nil {
		CreateErrorLog("", "Unable to parse read file, please restart the application. If this problem persists contact customer support")
		return
	}

	CS := new(CONNECTION_SETTINGS)
	err = json.Unmarshal(backupBytes, CS)
	if err != nil {
		CreateErrorLog("", "Unable to parse read file, please restart the application. If this problem persists contact customer support")
		return
	}

	GLOBAL_STATE.DefaultInterface = CS
	RestoreIPv6()
}

func LaunchPreperation() (err error) {
	err = InitializeTunnelInterface()
	return
}

func ResetAfterFailedConnectionAttempt() {
	defer RecoverAndLogToFile()
	CreateLog("connect", "RESETTING EVERYTHING AFTER A FAILED CONNECT")
	RestoreIPv6()
	// RestoreDNS()
}

func ChangeDNS() error {
	return nil
}

func RestoreDNS() {
}

func ChangeDNSWhileConnected() error {
	return nil
}

func ChangeDNSOnTunnelInterface() error {
	return nil
}

func EnablePacketRouting() (err error) {
	defer RecoverAndLogToFile()

	DisableIPv6()

	err = SetInterfaceStateToUp()
	if err != nil {
		CreateErrorLog("", "Unable to bring default interface up")
		return err
	}

	out, errx := exec.Command("ip", "route", "add", "default", "via", TUNNEL_ADAPTER_ADDRESS, "dev", TUNNEL_ADAPTER_NAME, "metric", "0").CombinedOutput()
	if errx != nil {
		if !strings.Contains(string(out), "File exists") {
			CreateErrorLog("", "IP || Unable to add default route || msg: ", errx, " || output: ", string(out))
			return errx
		}
	}

	return
}

func InitializeTunnelInterface() (err error) {

	err = AdjustRoutersForTunneling()
	if err != nil {
		CreateErrorLog("", "Unable to fix route metrics: ", err)
		return err
	}

	IF, err := net.Interfaces()
	if err != nil {
		CreateErrorLog("", "Could not find network interfaces || msg: ", err)
	}
	interfaceAlreadyExists := false
	for _, v := range IF {
		if v.Name == TUNNEL_ADAPTER_NAME {
			interfaceAlreadyExists = true
		}

	}

	if !interfaceAlreadyExists {
		config := water.Config{
			DeviceType: water.TUN,
		}
		config.Name = TUNNEL_ADAPTER_NAME

		A.Interface, err = water.New(config)
		if err != nil {
			CreateErrorLog("", "Unable to create tunnel inteface || msg: ", err)
			return err
		}
	}

	CreateLog("", "Tunnel interface enabled")
	var ipOut []byte

	CreateLog("", "adding IP/CIDR "+TUNNEL_ADAPTER_ADDRESS+"/24 to tunnel interface")

	ipOut, err = exec.Command("ip", "addr", "add", TUNNEL_ADAPTER_ADDRESS+"/24", "dev", TUNNEL_ADAPTER_NAME).Output()
	if err != nil {
		if !strings.Contains(string(ipOut), "File exists") {
			CreateErrorLog("", "IP || Unable to add IP/CIDR range to tunnel interface || msg: ", err, " || output: ", string(ipOut))
			return err
		}
	}

	ipOut, err = exec.Command("ip", "link", "set", TUNNEL_ADAPTER_NAME, "mtu", "65535").Output()
	if err != nil {
		CreateErrorLog("", "IP || unable to set txqueuelen on tunnel interface || msg: ", err, " || output: ", string(ipOut))
		return err
	}

	ipOut, err = exec.Command("ip", "link", "set", TUNNEL_ADAPTER_NAME, "txqueuelen", "3000").Output()
	if err != nil {
		CreateErrorLog("", "IP || unable to set txqueuelen on tunnel interface || msg: ", err, " || output: ", string(ipOut))
		return err
	}

	err = SetInterfaceStateToUp()
	if err != nil {
		return err
	}

	return nil
}

func SetInterfaceStateToUp() (err error) {
	CreateLog("connect", "Initializing link/up on device: "+TUNNEL_ADAPTER_NAME)

	ipOut, err := exec.Command("ip", "link", "set", "dev", TUNNEL_ADAPTER_NAME, "up").Output()
	if err != nil {
		CreateErrorLog("", "IP || unable to bring the tunnel interface up (link up) || msg: ", err, " || output: ", string(ipOut))
		return err
	}

	return
}

func SetInterfaceStateToDown() (err error) {

	ipOut, err := exec.Command("ip", "link", "set", "dev", TUNNEL_ADAPTER_NAME, "down").Output()
	if err != nil {
		CreateErrorLog("", "IP || unable to bring the tunnel interface down (link down) || msg: ", err, " || output: ", string(ipOut))
		return err
	}

	return
}

func AdjustRoutersForTunneling() (err error) {
	CreateLog("connect", "Adjusting route metrics")

	var out []byte
	out, err = exec.Command("ip", "route").Output()
	if err != nil {
		CreateErrorLog("", "IP || unable to list routes || msg: ", err, " || output: ", string(out))
		return err
	}
	split := strings.Split(string(out), "\n")
	var DefaultRoutes = make(map[string]string)
	for _, v := range split {
		if strings.Contains(v, "default") {
			// log.Println(v)
			metricSplit := strings.Split(v, "metric ")

			if len(metricSplit) > 1 {
				finalRoute := strings.Replace(v, "metric "+metricSplit[1], "", 1)
				DefaultRoutes[metricSplit[1]] = finalRoute
			} else {
				DefaultRoutes["0"] = metricSplit[0]
			}

			break
		}
	}

	keys := make([]string, 0, len(DefaultRoutes))
	for k := range DefaultRoutes {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	_, ok := DefaultRoutes["0"]
	if ok {
		for _, k := range keys {
			v := DefaultRoutes[k]
			iStipped := strings.Replace(k, " ", "", -1)
			cmddel := "ip route del " + v + " metric " + iStipped
			intMetric, err := strconv.Atoi(iStipped)
			if err != nil {
				CreateErrorLog("", "unable to convert route metric value to string || msg: ", err, " || value: ", iStipped)
				return err
			}
			intMetric++
			cmdMake := "ip route add " + v + " metric " + strconv.Itoa(intMetric)

			out, err := exec.Command("bash", "-c", cmdMake).Output()
			if err != nil {
				CreateErrorLog("", "IP || unable adjust route: ", cmdMake, " || msg: ", err, " || output: ", string(out))
				return err
			}

			out, err = exec.Command("bash", "-c", cmddel).Output()
			if err != nil {
				CreateErrorLog("", "IP || unable adjust route: ", cmdMake, " || msg: ", err, " || output: ", string(out))
				return err
			}

			CreateLog("connect", "changed route: ", v+" || old metric: "+iStipped, " || new metric value: ", strconv.Itoa(intMetric))
		}
	}

	return
}

func InitializeTunnelAdapter() (err error) {
	defer RecoverAndLogToFile()

	err = DisableIPv6()
	if err != nil {
		return err
	}

	GLOBAL_STATE.TunnelInitialized = true
	CreateLog("connect", "Niceland VPN tunnel initialized")
	return
}

func DeleteTunnelInterfaceRoutes(IP string) (err error) {

	out, err := exec.Command("ip", "route", "del", IP, "via", TUNNEL_ADAPTER_ADDRESS, "metric", "0").Output()
	if err != nil {
		CreateErrorLog("", "IP || Unable to delete route: ", IP, " || Gateway: ", TUNNEL_ADAPTER_ADDRESS, " || msg: ", err, " || output: ", string(out))
		return err
	}

	return
}

func AddRouteToTunnelInterface(IP string) (err error) {

	out, err := exec.Command("ip", "route", "add", IP, "via", TUNNEL_ADAPTER_ADDRESS, "metric", "0").Output()
	if err != nil {

		CreateErrorLog("", "IP || Unable to add routea to: ", IP, " || Gateway: ", TUNNEL_ADAPTER_ADDRESS, " || msg: ", err, " || output: ", string(out))
		return err
	}

	return
}

func AddRoute(IP string) (err error) {
	if GLOBAL_STATE.DefaultInterface == nil {
		CreateLog("", "Not adding route, no default interface")
		return errors.New("no default interface")
	}

	_ = DeleteRoute(IP, false)

	out, err := exec.Command("ip", "route", "add", IP, "via", GLOBAL_STATE.DefaultInterface.DefaultRouter, "metric", "0").Output()
	if err != nil {
		CreateErrorLog("", "IP || Unable to add route to: ", IP, " || Gateway: ", TUNNEL_ADAPTER_ADDRESS, " || msg: ", err, " || output: ", string(out))
		return err
	}

	return
}

func DeleteRoute(IP string, ignoreActiveRouter bool) (err error) {

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateLog("", "Not deleting route, no default interface")
		return errors.New("no default interface")
	}

	if !ignoreActiveRouter {
		if GLOBAL_STATE.ActiveRouter != nil && GLOBAL_STATE.ActiveRouter.IP == IP {
			return
		}
	}

	out, err := exec.Command("ip", "route", "del", IP).Output()
	if err != nil {
		CreateErrorLog("", "IP || Unable to delete route: ", IP, " || Gateway: ", TUNNEL_ADAPTER_ADDRESS, " || msg: ", err, " || output: ", string(out))
		return
	}

	return
}

func FindDefaultInterfaceAndGateway() (POTENTIAL_DEFAULT *CONNECTION_SETTINGS, err error) {

	INTERFACE_SETTINGS := FindAllInterfaces()

	var out []byte
	out, err = exec.Command("ip", "route").Output()
	if err != nil {
		CreateLog("general", err, "could not get route list")
		return nil, err
	}

	split := strings.Split(string(out), "\n")
	defaultName := ""
	lowestMetric := 999999
	for _, v := range split {
		fields := strings.Fields(v)
		if len(fields) < 1 {
			continue
		}
		if fields[0] == "default" {
			var metricInt int = 0
			if fields[len(fields)-2] == "metric" {
				metricInt, err = strconv.Atoi(fields[len(fields)-1])
				if err != nil {
					CreateErrorLog("", "Unable to parse interface metric", fields)
					log.Println(err)
					return nil, err
				}
			} else {
				metricInt = 0
			}

			for i := range INTERFACE_SETTINGS {
				if INTERFACE_SETTINGS[i] == nil {
					continue
				}

				if fields[4] == TUNNEL_ADAPTER_NAME {
					continue
				}

				if i == fields[4] {
					INTERFACE_SETTINGS[i].Metric = metricInt
					INTERFACE_SETTINGS[i].Hop = fields[2]
					if metricInt < lowestMetric {
						defaultName = fields[4]
						lowestMetric = metricInt
					}
				}
			}

		}
	}

	for i := range INTERFACE_SETTINGS {
		if INTERFACE_SETTINGS[i] == nil {
			continue
		}

		if i == defaultName {

			POTENTIAL_DEFAULT = new(CONNECTION_SETTINGS)
			POTENTIAL_DEFAULT.IFName = i
			POTENTIAL_DEFAULT.DefaultRouter = INTERFACE_SETTINGS[i].Hop
			return

		}
	}

	if lowestMetric == 999999 || defaultName == "" {
		return nil, errors.New("")
	}

	return
}

func RestoreIPv6() {
	defer RecoverAndLogToFile()
	if GLOBAL_STATE.DefaultInterface == nil {
		CreateErrorLog("", "Failed to restore IPv6 settings, interface settings not found")
		return
	}

	if GLOBAL_STATE.DefaultInterface.IPV6Enabled {

		out, err := exec.Command("sysctl", "-w", "net.ipv6.conf."+GLOBAL_STATE.DefaultInterface.IFName+".disable_ipv6=1").Output()
		if err != nil {
			CreateErrorLog("", err, "SYSCTL || Error resting IPv6 settings || msg: ", err, " || output: ", string(out))
		}

	} else {
		out, err := exec.Command("sysctl", "-w", "net.ipv6.conf."+GLOBAL_STATE.DefaultInterface.IFName+".disable_ipv6=0").Output()
		if err != nil {
			CreateErrorLog("", err, "SYSCTL || Error resting IPv6 settings || msg: ", err, " || output: ", string(out))
		}
	}

	return

}

func DisableIPv6() error {
	defer RecoverAndLogToFile()

	CreateLog("connect", "Disabling IPv6 on interface: ", GLOBAL_STATE.DefaultInterface.IFName)

	out, err := exec.Command("sysctl", "-w", "net.ipv6.conf."+GLOBAL_STATE.DefaultInterface.IFName+".disable_ipv6=1").Output()
	if err != nil {
		CreateErrorLog("", err, "SYSCTL || Unable to turn off IPv6 support || msg: ", err, " || output: ", string(out))
		return err
	}

	return nil
}

func PrintInterfaces() (error, []byte) {
	out, err := exec.Command("bash", "-c", "ip a").Output()
	if err != nil {
		return err, nil
	}
	return nil, out
}

func PrintRouters() (error, []byte) {
	out, err := exec.Command("bash", "-c", "ip route").Output()
	if err != nil {
		return err, nil
	}
	return nil, out
}

func PrintDNS() (error, []byte) {
	return nil, []byte(".. DNS already printed via resolv.conf and NetworkManager settings")
}
