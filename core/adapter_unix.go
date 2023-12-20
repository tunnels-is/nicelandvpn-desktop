//go:build freebsd || linux || openbsd

package core

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	water "github.com/songgao/water"
)

type Adapter struct {
	Interface *water.Interface
}

// func (A *Adapter) Close() (err error) {
// 	if A.Interface != nil {
// 		err = A.Interface.Close()
// 	}
// 	return
// }
//
// func (A *Adapter) Uninstall() (err error) {
// 	return
// }

// func GetIPv6Settings(PotentialDefault *CONNECTION_SETTINGS) {
// 	out, err := exec.Command("bash", "-c", "cat /proc/sys/net/ipv6/conf/"+PotentialDefault.IFName+"/disable_ipv6").CombinedOutput()
// 	if err != nil {
// 		CreateErrorLog("", err, "Error getting ipv6 settings for interface: ", PotentialDefault.IFName, " || msg: ", err, " || output: ", string(out))
// 		return
// 	}
//
// 	outString := string(out)
// 	outString = strings.TrimSpace(outString)
//
// 	if outString == "0" {
// 		PotentialDefault.IPV6Enabled = false
// 	} else {
// 		PotentialDefault.IPV6Enabled = true
// 	}
// }

// func VerifyAndBackupSettings(PotentialDefault *CONNECTION_SETTINGS) (err error) {
// 	GetIPv6Settings(PotentialDefault)
// 	return
// }

// func FindDefaultInterfaceAndGatewayDuringStartup() (err error) {
// 	PotentialDefault, err := FindDefaultInterfaceAndGateway()
// 	if err != nil {
// 		CreateErrorLog("", "Could not find default interface and gateway >> ", err)
// 		return errors.New("")
// 	}
//
// 	if PotentialDefault == nil {
// 		CreateErrorLog("", "Could not find default interface and gateway >> Potential Default is nil")
// 		return errors.New("")
// 	}
//
// 	if PotentialDefault.DefaultRouter == "" {
// 		CreateErrorLog("", "Default interface had invalid Default router", PotentialDefault, " >> ", err)
// 		return errors.New("")
// 	}
//
// 	VerifyAndBackupSettings(PotentialDefault)
//
// 	GLOBAL_STATE.DefaultInterface = PotentialDefault
//
// 	CreateLog("", "NEW DEFAULT INTERFACE >> ", MAC_CONNECTION_SETTINGS)
//
// 	BackupSettingsToFile(PotentialDefault)
// 	return
// }

// func RestoreSettingsFromFile(PotentialDefault *CONNECTION_SETTINGS) {
// 	CreateLog("", "RESTORING SETTINGS FROM FILE")
//
// 	backupFile, err := os.Open(GenerateBaseFolderPath() + PotentialDefault.IFName + "_backup")
// 	if err != nil {
// 		CreateErrorLog("", "Unable to open backup file, please restart the application. If this problem persists contact customer support")
// 		return
// 	}
//
// 	backupBytes, err := io.ReadAll(backupFile)
// 	if err != nil {
// 		CreateErrorLog("", "Unable to parse read file, please restart the application. If this problem persists contact customer support")
// 		return
// 	}
//
// 	CS := new(CONNECTION_SETTINGS)
// 	err = json.Unmarshal(backupBytes, CS)
// 	if err != nil {
// 		CreateErrorLog("", "Unable to parse read file, please restart the application. If this problem persists contact customer support")
// 		return
// 	}
//
// 	GLOBAL_STATE.DefaultInterface = CS
// 	RestoreIPv6()
// }

// func LaunchPreperation() (err error) {
// 	return
// }

// func ResetAfterFailedConnectionAttempt() {
// 	defer RecoverAndLogToFile()
// 	CreateLog("connect", "RESETTING EVERYTHING AFTER A FAILED CONNECT")
// 	RestoreIPv6()
// 	// RestoreDNS()
// }

// func ChangeDNS() error {
// 	return nil
// }
//
// func RestoreDNS(force bool) {
// }
//
// func ChangeDNSWhileConnected() error {
// 	return nil
// }
//
// func ChangeDNSOnTunnelInterface() error {
// 	return nil
// }

// func EnablePacketRouting() (err error) {
// 	defer RecoverAndLogToFile()

// _ = DisableIPv6()

// err = SetInterfaceStateToUp()
// if err != nil {
// 	CreateErrorLog("", "Unable to bring default interface up")
// 	return err
// }

// out, errx := exec.Command("ip", "route", "add", "default", "via", TUNNEL_ADAPTER_ADDRESS, "dev", TUNNEL_ADAPTER_NAME, "metric", "0").CombinedOutput()
// if errx != nil {
// 	if !strings.Contains(string(out), "File exists") {
// 		CreateErrorLog("", "IP || Unable to add default route || msg: ", errx, " || output: ", string(out))
// 		return errx
// 	}
// }

// 	return
// }

// func InitializeTunnelInterface() (err error) {
// err = AdjustRoutersForTunneling()
// if err != nil {
// 	CreateErrorLog("", "Unable to fix route metrics: ", err)
// 	return err
// }

// CreateLog("", "Tunnel interface enabled")
// var ipOut []byte
//
// CreateLog("", "adding IP/CIDR "+TUNNEL_ADAPTER_ADDRESS+"/24 to tunnel interface")
//
// ipOut, err = exec.Command("ip", "addr", "add", TUNNEL_ADAPTER_ADDRESS+"/32", "dev", TUNNEL_ADAPTER_NAME).Output()
// if err != nil {
// 	if !strings.Contains(string(ipOut), "File exists") {
// 		CreateErrorLog("", "IP || Unable to add IP/CIDR range to tunnel interface || msg: ", err, " || output: ", string(ipOut))
// 		return err
// 	}
// }
//
// ipOut, err = exec.Command("ip", "link", "set", TUNNEL_ADAPTER_NAME, "mtu", "65535").Output()
// if err != nil {
// 	CreateErrorLog("", "IP || unable to set txqueuelen on tunnel interface || msg: ", err, " || output: ", string(ipOut))
// 	return err
// }
//
// ipOut, err = exec.Command("ip", "link", "set", TUNNEL_ADAPTER_NAME, "txqueuelen", "3000").Output()
// if err != nil {
// 	CreateErrorLog("", "IP || unable to set txqueuelen on tunnel interface || msg: ", err, " || output: ", string(ipOut))
// 	return err
// }
//
// err = SetInterfaceStateToUp()
// if err != nil {
// 	return err
// }
//
// 	return nil
// }

// func SetInterfaceStateToUp() (err error) {
// 	CreateLog("connect", "Initializing link/up on device: "+TUNNEL_ADAPTER_NAME)
//
// 	ipOut, err := exec.Command("ip", "link", "set", "dev", TUNNEL_ADAPTER_NAME, "up").Output()
// 	if err != nil {
// 		CreateErrorLog("", "IP || unable to bring the tunnel interface up (link up) || msg: ", err, " || output: ", string(ipOut))
// 		return err
// 	}
//
// 	return
// }
//
// func SetInterfaceStateToDown() (err error) {
// 	ipOut, err := exec.Command("ip", "link", "set", "dev", TUNNEL_ADAPTER_NAME, "down").Output()
// 	if err != nil {
// 		CreateErrorLog("", "IP || unable to bring the tunnel interface down (link down) || msg: ", err, " || output: ", string(ipOut))
// 		return err
// 	}
//
// 	return
// }

func AdjustRoutersForTunneling() (err error) {
	CreateLog("connect", "Adjusting route metrics")

	var out []byte
	out, err = exec.Command("ip", "route").Output()
	if err != nil {
		CreateErrorLog("", "IP || unable to list routes || msg: ", err, " || output: ", string(out))
		return err
	}
	split := strings.Split(string(out), "\n")
	DefaultRoutes := make(map[string]string)
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
			fmt.Println("MAKE: ", cmdMake)

			out, err := exec.Command("bash", "-c", cmdMake).Output()
			if err != nil {
				CreateErrorLog("", "IP || unable adjust route: ", cmdMake, " || msg: ", err, " || output: ", string(out))
				return err
			}

			fmt.Println("MAKE: ", cmddel)
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

// func FindDefaultInterfaceAndGateway() (POTENTIAL_DEFAULT *CONNECTION_SETTINGS, err error) {
// 	INTERFACE_SETTINGS := FindAllInterfaces()
//
// 	var out []byte
// 	out, err = exec.Command("ip", "route").Output()
// 	if err != nil {
// 		CreateLog("general", err, "could not get route list")
// 		return nil, err
// 	}
//
// 	split := strings.Split(string(out), "\n")
// 	defaultName := ""
// 	lowestMetric := 999999
// 	for _, v := range split {
// 		fields := strings.Fields(v)
// 		if len(fields) < 1 {
// 			continue
// 		}
// 		if fields[0] == "default" {
// 			var metricInt int = 0
// 			if fields[len(fields)-2] == "metric" {
// 				metricInt, err = strconv.Atoi(fields[len(fields)-1])
// 				if err != nil {
// 					CreateErrorLog("", "Unable to parse interface metric", fields)
// 					return nil, err
// 				}
// 			} else {
// 				metricInt = 0
// 			}
//
// 			for i := range INTERFACE_SETTINGS {
// 				if INTERFACE_SETTINGS[i] == nil {
// 					continue
// 				}
//
// 				if fields[4] == TUNNEL_ADAPTER_NAME {
// 					continue
// 				}
//
// 				if i == fields[4] {
// 					INTERFACE_SETTINGS[i].Metric = metricInt
// 					INTERFACE_SETTINGS[i].Hop = fields[2]
// 					if metricInt < lowestMetric {
// 						defaultName = fields[4]
// 						lowestMetric = metricInt
// 					}
// 				}
// 			}
//
// 		}
// 	}
//
// 	for i := range INTERFACE_SETTINGS {
// 		if INTERFACE_SETTINGS[i] == nil {
// 			continue
// 		}
//
// 		if i == defaultName {
//
// 			POTENTIAL_DEFAULT = new(CONNECTION_SETTINGS)
// 			POTENTIAL_DEFAULT.IFName = i
// 			POTENTIAL_DEFAULT.DefaultRouter = INTERFACE_SETTINGS[i].Hop
// 			return
//
// 		}
// 	}
//
// 	if lowestMetric == 999999 || defaultName == "" {
// 		return nil, errors.New("")
// 	}
//
// 	return
// }

func PrintInterfaces() (error, []byte) {
	out, err := exec.Command("bash", "-c", "ip a").Output()
	if err != nil {
		return err, nil
	}
	return nil, out
}

func PrintRoutes() (error, []byte) {
	out, err := exec.Command("bash", "-c", "ip route").Output()
	if err != nil {
		return err, nil
	}
	return nil, out
}

func PrintDNS() (error, []byte) {
	return nil, []byte(".. DNS already printed via resolv.conf and NetworkManager settings")
}
