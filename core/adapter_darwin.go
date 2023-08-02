//go:build darwin

package core

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	water "github.com/songgao/water"
)

type Adapter struct {
	Initialized bool
	Interface   *water.Interface
}

func (A *Adapter) Close() (err error) {
	err = A.Interface.Close()
	return
}

func (A *Adapter) Uninstall() error {
	return nil
}

func GetIPv6Settings(PotentialDefault *CONNECTION_SETTINGS) {
	return
}

func GetDnsSettings(PotentialDefault *CONNECTION_SETTINGS) {

	dnsout, err := exec.Command("networksetup", "-getdnsservers", PotentialDefault.IFName).Output()
	if err != nil {
		CreateErrorLog("", "Unable to find DNS settings >>", err)
		return
	}

	dnsSplit := bytes.Split(dnsout, []byte{10})
	if dnsSplit[0][0] == 84 {
		PotentialDefault.AutoDNS = true
		return
	}

	PotentialDefault.DNS1 = string(dnsSplit[0])

	if len(dnsSplit) > 1 {
		PotentialDefault.DNS2 = string(dnsSplit[1])
	}

	return
}

func ChangeDNS() {
	defer RecoverAndLogToFile()
}
func ChangeDNSWhileConnected() error {
	defer RecoverAndLogToFile()
	return nil
}

func InitializeTunnelAdapter() error {
	defer RecoverAndLogToFile()
	CreateLog("", "INITIALIZING TUNNEL INTERFACE")

	_ = AddRoute(GLOBAL_STATE.ActiveRouter.IP)

	return nil
}

func EnablePacketRouting() error {

	DisableIPv6()
	CreateLog("connect", "Creating default route")

	err := SetInterfaceStateToUp(TUNNEL_ADAPTER_NAME)
	if err != nil {
		return err
	}

	_, err = exec.Command("route", "delete", "default").Output()
	if err != nil {
		CreateErrorLog("", err, "unable to delete default route")
		return err
	}

	_, err = exec.Command("route", "add", "default", "10.4.3.1").Output()
	if err != nil {
		CreateErrorLog("", err, "unable to create default route")
		return err
	}

	return nil
}

func SetInterfaceStateToDown() (err error) {
	return RestoreOriginalDefaultRoute()
}

func RestoreOriginalDefaultRoute() (err error) {
	defer RecoverAndLogToFile()

	_, err = exec.Command("route", "delete", "default").Output()
	if err != nil {
		CreateErrorLog("", err, "unable to delete default route")
		return err
	}
	CreateLog("", "Default route to tunnel interface removed")

	out, err := exec.Command("route", "add", "default", GLOBAL_STATE.DefaultInterface.DefaultRouter).Output()
	if err != nil {
		CreateErrorLog("", err, "unable to create default route to: ", GLOBAL_STATE.DefaultInterface.DefaultRouter)
		return err
	} else {
		CreateLog("", "original default route created >>", string(out))
	}

	return
}

func RestoreIPv6() {
	if GLOBAL_STATE.DefaultInterface.IP6Method == "Manual" {

	} else if GLOBAL_STATE.DefaultInterface.IP6Method == "Automatic" {

		_, err := exec.Command("networksetup", "-setv6automatic", GLOBAL_STATE.DefaultInterface.IFName).Output()
		if err != nil {
			CreateErrorLog("", err, "unable to disable ipv6 on interface", GLOBAL_STATE.DefaultInterface.IFName)
		}

	} else if GLOBAL_STATE.DefaultInterface.IP6Method == "Off" {
		_, err := exec.Command("networksetup", "-setv6off", GLOBAL_STATE.DefaultInterface.IFName).Output()
		if err != nil {
			CreateErrorLog("", err, "unable to disable ipv6 on interface", GLOBAL_STATE.DefaultInterface.IFName)
		}
	}
}

func DisableIPv6() {
	_, err := exec.Command("networksetup", "-setv6off", GLOBAL_STATE.DefaultInterface.IFName).Output()
	if err != nil {
		CreateErrorLog("", err, "unable to disable ipv6 on interface", GLOBAL_STATE.DefaultInterface.IFName)
	}
}

func ResetAfterFailedConnectionAttempt() {
	defer RecoverAndLogToFile()
	RestoreOriginalDefaultRoute()
	RestoreIPv6()
}

func RestoreDNS() {

}

func VerifyAndBackupSettings(PotentialDefault *CONNECTION_SETTINGS) (err error) {
	return
}

func FindDefaultInterfaceAndGateway() (PotentialDefault *CONNECTION_SETTINGS, err error) {

	cmd := exec.Command("netstat", "-nr", "-f", "inet")
	routeList, err := cmd.CombinedOutput()
	if err != nil {
		CreateErrorLog("", "Unable to list network service >> ", err)
		return nil, err
	}

	split := bytes.Split(routeList, []byte{10})
	for i, v := range split {
		vstring := string(v)
		if i < 4 {
			continue
		}

		fields := strings.Fields(vstring)
		if len(fields) < 1 {
			continue
		}

		if fields[0] == "default" {
			PotentialDefault = new(CONNECTION_SETTINGS)
			PotentialDefault.AdapterName = fields[3]
			PotentialDefault.DefaultRouter = fields[1]
			break
		}
	}
	if PotentialDefault == nil {
		CreateErrorLog("", "unable to find potential default interface from routers: ", err)
		return nil, err
	}

	netList, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		CreateErrorLog("", "Unable to list network service >> ", err)
		return nil, err
	}

	splitList := bytes.Split(netList, []byte{10})
	for i, v := range splitList {
		if strings.Contains(string(v), "Hardware Port") {

			serviceName := strings.Split(string(v), ":")[1]
			serviceName = strings.Replace(serviceName, " ", "", 1)

			adapterName := strings.Split(string(splitList[i+1]), ":")[1]
			adapterName = strings.Replace(adapterName, " ", "", 1)
			if adapterName == PotentialDefault.AdapterName {
				PotentialDefault.IFName = serviceName
				break
			}
		}
	}

	if PotentialDefault == nil {
		return nil, errors.New("")
	}

	ifStats, err := exec.Command("networksetup", "-getinfo", PotentialDefault.IFName).Output()
	if err != nil {
		CreateErrorLog("", "Unable to get network service info >> ", err)
		return nil, err
	}

	splitList = bytes.Split(ifStats, []byte{10})
	splitList[0] = []byte{}
	for _, v := range splitList {
		vSplit := strings.Split(string(v), ": ")
		if vSplit[0] == "IPv6" {
			PotentialDefault.IP6Method = vSplit[1]
			if PotentialDefault.IP6Method == "Off" {
				PotentialDefault.IPV6Enabled = false
			} else {
				PotentialDefault.IPV6Enabled = true
			}
		}
	}

	return
}

func LaunchPreperation() (err error) {

	A.Interface, err = water.New(water.Config{
		DeviceType: water.TUN,
	})

	if err != nil {
		CreateErrorLog("", err, "unable to create tunnel interface")
		return err
	}

	TUNNEL_ADAPTER_NAME = A.Interface.Name()

	CreateLog("", "Initializing link/up on device "+A.Interface.Name())

	err = SetInterfaceStateToUp(TUNNEL_ADAPTER_NAME)
	if err != nil {
		return err
	}

	ipOut, err := exec.Command("ifconfig", A.Interface.Name(), "mtu", "65535").Output()
	if err != nil {
		CreateErrorLog("", err, "Unable to change mtu", "STDOUT", string(ipOut))
		return err
	}

	GLOBAL_STATE.TunnelInitialized = true
	return
}

func SetInterfaceStateToUp(name string) error {

	ipOut, err := exec.Command("ifconfig", A.Interface.Name(), "10.4.3.2", "10.4.3.1", "up").Output()

	if err != nil {
		CreateErrorLog("", err, "unable to bring up tunnel adapter ", "STDOUT", string(ipOut))
		return err
	}

	return nil
}

func DeleteTunnelInterfaceRoutes(IP string) (err error) {
	_, err = exec.Command("route", "-n", "delete", "-net", IP, "10.4.3.1").Output()
	if err != nil {
		CreateErrorLog("", err, "unable to add route to IP", IP)
		return err
	}
	CreateLog("", "added route to IP", IP)

	return
}

func AddRouteToTunnelInterface(IP string) (err error) {
	_, err = exec.Command("route", "-n", "add", "-net", IP, "10.4.3.1").Output()
	if err != nil {
		CreateErrorLog("", err, "unable to add route to IP", IP)
		return err
	}
	CreateLog("", "added route to IP", IP)

	return
}

func AddRoute(IP string) (err error) {
	_ = DeleteRoute(IP, false)
	_, err = exec.Command("route", "-n", "add", "-net", IP, GLOBAL_STATE.DefaultInterface.DefaultRouter).Output()
	if err != nil {
		CreateErrorLog("", "unable to add route to IP", IP, err)
		return err
	}

	return
}

func DeleteRoute(IP string, ignoreActiveRouterIP bool) (err error) {
	if !ignoreActiveRouterIP {
		if GLOBAL_STATE.ActiveRouter != nil && GLOBAL_STATE.ActiveRouter.IP == IP {
			return
		}
	}

	_, err = exec.Command("route", "-n", "delete", "-net", IP).Output()
	if err != nil {
		CreateErrorLog("", err, "unable to delete route for IP", IP)
		return
	}

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

	CreateLog("", "NEW DEFAULT INTERFACE >> ", GLOBAL_STATE.DefaultInterface)

	BackupSettingsToFile(PotentialDefault)

	return
}

func RestoreSettingsFromFile() {

}

func PrintInterfaces() (error, []byte) {
	out, err := exec.Command("ifconfig").Output()
	if err != nil {
		return err, nil
	}
	return nil, out
}

func PrintRouters() (error, []byte) {
	out, err := exec.Command("netstat", "-rn").Output()
	if err != nil {
		return err, nil
	}
	return nil, out
}

func PrintDNS() (error, []byte) {
	var out = make([]byte, 0)
	dnsout, err := exec.Command("networksetup", "-getdnsservers", GLOBAL_STATE.DefaultInterface.IFName).Output()
	if err != nil {
		out = append(out, []byte("Error: "+err.Error())...)
		return nil, nil
	}
	out = append(out, []byte(GLOBAL_STATE.DefaultInterface.IFName)...)
	out = append(out, dnsout...)
	out = append(out, []byte{13, 10}...)
	return nil, out
}
