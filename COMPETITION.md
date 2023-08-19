# TUI Competition
The TUI competition will run from 7th August 2023 to 7th October 2023.

# Discord
 - Contact: sveinn ( keyb1nd )
 - https://discord.gg/7Ts3PCnCd9
 - https://twitter.com/nicelandvpn

# Rewards
 - 1st place: 6 months Niclend VPN subscription + Credits on the websites
	- First place will also receive a budget of 230 USD to order any kind of equipment they want.
 - 2nd place: 6 months Niceland VPN subscription.

# How to get started
1. Fork the nicelandvpn-desktop repository.
2. Checkout the tui-competition branch.

You will be doing all your work on the tui-competition branch. Once you feel satisfied with your work you can create a pull reques again the tui-competition branch on the original repository.

# Rules
 - Please follow the general development guide.
 - Participants must use the BubbleTea package. Any package from the [Bubbles Library](https://github.com/charmbracelet/bubbles) and [Lipgloss](https://github.com/charmbracelet/lipgloss) are allowed to be used for this competition. Additional packages are generally not allowed. If you want to add additional packages then contact us on discord or twitter.
 - No refactoring of internal functions and methods. If you feel like a refactor could be benefitial to the TUI you can contact us on discord or twitter to discuss it.
 - The TUI should not include texts, messages or content that is not in-line with the context of the NicelandVPN app. 
 - Creating fun easter-eggs is allowed, asuming they follow the previously mentioned rule about TUI content and are not disruptive to the apps functionality.
 - Using the same color scheme as the GUI is prefferred, but not required.

# Judging Criteria
## Basic Methods implementations
 - (2 Points)Login
 - (2 Points)Router List / Router Selection
 - (2 Points)VPN List / Connect
 - (2 Points)Disconnect
 - (2 Points)Loading log feed
 - (2 Points)General log view
 - (2 Points)Application state overview
 - (2 Points)Stats Window (can be be merged with Application state overview like in the GUI)
 - (2 Points)Logout

## Bonus Point Methods
 - (2 Point)Anything that is inside the Settings menu in the GUI

## Other Criteria
 - (5 Points) for using the same color scheme as GUI
 - (1-10 Points) Design (Colors, Layout, Progress Bars, etc..)
 - (1-10 Points) Usability (Navigation, Interaction, Notifications, etc..)
 - (1-10 Points) Internals (Stability, Readable Code, Code Design, etc..)
 - (10 Points) MacOS Compatibility
 - (10 Points) Windows Compatibility
 - (100 Points) Linux Compatibility
 - (-100 Points) Spaghetti code


# Functionality of NicelandVPN 
# Design 
 - Colors can be found inside the variables.scss file

# App Flow

## The GLOBAL_SATE
```golang
type State struct {
  // MOST IMPORTANT SATATES
	IsAdmin               bool `json:"IsAdmin"` // Tells you if the app was started as admin or not. The app cannot connect without having admin rights
	ClientReady           bool `json:"ClientReady"` // Tells you if the client is ready to connect. The app cannot connect unless the client is ready
	ClientStartupError    bool `json:"ClientStartupError"` // Tells you if there was an error during startup
	C                             *Config   `json:"C"` // This is the current configuration for the app

  // STATES THAT ARE RELEVANT TO THE TUI
	Connected  bool `json:"Connected"` // Tells you if your are currently connected
	UMbpsString    string `json:"UMbpsString"` // Formatted upload speed string
	DMbpsString    string `json:"DMbpsString"` // Formatted download speed string
	IngressPackets uint64 `json:"IngressPackets"` // Ingress packet count
	EgressPackets  uint64 `json:"EgressPackets"` // Egress packet count
	ConnectedTimer string `json:"ConnectedTimer"` // Connection timer - tells you how long the current session has been connected
	BaseFolderInitialized bool `json:"BaseFolderInitialized"`  // The path to the base folder
	LogFileInitialized    bool `json:"LogFileInitialized"` // Tells you if the log file was initialized
	ConfigInitialized     bool `json:"ConfigInitialized"` // Tells you if the config file was initialized
	TunnelInitialized     bool `json:"TunnelInitialized"` // Tells you if the tunnel/tap interface has been enabled. The app cannot connect without having the tunnel/tap inteface
	DefaultInterface *CONNECTION_SETTINGS `json:"DefaultInterface"` // This is the current default interface on your compter

	SecondsSincePingFromRouter    string    `json:"SecondsSincePingFromRouter"` // This is a formatted string that shows the time since the last received ping in seconds
	SecondsUntilAccessPointUpdate int // This displays how many seconds are until the next AccessPoint list update (note: this also updates the Router list)
	Routers                       []*ROUTER       `json:"Routers"` // This is the router list used to display routers in the TUI and GUI
	AccessPoints                  []*AccessPoint  `json:"AccessPoints"` // This is the currently active AccessPoint list
	ActiveRouter                  *ROUTER         `json:"ActiveRouter"` // This is the currently active entry router (entry point) to the routing mesh-network
	ActiveAccessPoint             *AccessPoint    `json:"ActiveAccessPoint"` // This is the currently active/connected AccessPoint/vpn
	ActiveSession                 *CLIENT_SESSION `json:"ActiveSession"` // This is the currently active session object
	LogFileName string `json:"LogFileName"` // This is the log file name
	LogPath     string `json:"LogPath"` // This is the log file path
	ConfigPath  string `json:"ConfigPath"` // This is the config path
	BackupPath  string `json:"BackupPath"` // This is the backup path where we store interface configuration backups
	BasePath    string `json:"BasePath"` // This is the base folder path for all generated files
	Version string `json:"Version"` // This is the current version of the app

  // STATES THAT YOU DO NOT NEED TO WORRY ABOUT
	UMbps          int    `json:"UMbps"` // Upload speed in Megabits per second
	DMbps          int    `json:"DMbps"` // Download speed in Mbits per second

	NeedsRouterProbe      bool `json:"NeedsRouterProbe"` // Tells you if the app needs to find new routers - you do not need to interact with this variable
	BufferError    bool   `json:"BufferError"` // Tells you if there was an error in the read/write buffers used for the routing.

	Connecting bool `json:"Connecting"` // Tells you if you are in the process of connecting
	Exiting    bool `json:"Exiting"` // Tells you if you are in the process of exiting the app

	LastRouterPing                time.Time `json:"LastRouterPing"` // This tell you when you send the last ping to the router
  PingReceivedFromRouter        time.Time `json:"PingReceivedFromRouter"` // This tells you when you received the last ping from the target router/vpn accesspoint
	LastAccessPointUpdate         time.Time // This shows when you last received an update to the AccessPoint list
	AvailableCountries            []string        `json:"AvailableCountries"` // This shows the available countries based on the router list
	RoutersList                   [2000]*ROUTER   `json:"-"` // This is the currently active router list for the core
}

```


## Initialization
The initialization process has already been coded for the TUI
```golang

const PRODUCTION = false // if the production flag it "true" the app won't log anything.
const ENABLE_INSTERFACE = false // this trigger will determine if the application will start the actual tunnel/tap interface. Without it you cannot connect. We recommend keeping it disabled until you are ready to test the connect/switch/disconnect functionality of the TUI

var MONITOR = make(chan int, 200)
var TUI *tea.Program

func main() {

  // This is just to initialize the globals inside the core.
	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INSTERFACE

  // These two function are essential to maintaining the VPNS core operations
  // DO NOT CHANGE.
	go RoutineMonitor()
	go core.StartService(MONITOR)

  // This routine is a "ticker" update for the TUI, it forces it to
  // re-render every second. You can disable this or change it however you like.
	go TimedUIUpdate(MONITOR)

  // This is where the TUI starts
	TUI = tea.NewProgram(initialModel())
	if _, err := TUI.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

```
We do not want there to be more code inside the "main.go" file. You are however allowed to create the TUI folder/file structure however you want.  We have made some boilerplate code inside "tui.go". You do not have to use the boilerplate code at all, only if you want to. 

## Login
Login requires these fields: "email", "password", "devicename". the field "digits" is an optional two-factor authentication field and "recovery" is an optional field for two-factor recovery login. However, do not worry about the recovery or two-factor authentication unless you intend to implement the enabling of two-factor authentication as well. 
</br>
The login method uses a function called core.ForwardToController. This function has a specific format

```golang
data, code, err := core.ForwardToController(FR)

type FORWARD_REQUEST struct {
	Path    string // Path: /v2/user/login
	Method  string // Method: POST
	Timeout int // Timeout: 20000
	JSONData interface{} // structs.LoginForm{}
}

```

## Selecting Routers
The app will automatically select the lowest latency router once it starts up, but you can switch routers manually using the Tag field from the selected router. If the Tag field is empty, the app will go back to automatic router selection mode.
- NOTE: Once the router selection process is finished, you can find the new router as: GLOBAL_STATE.ActiveRouter.
- NOTE: The SwitchRouter function generates logs under the "loader" tag.

```golang
code, err := core.SwitchRouter(Tag)
```
## Connecting / Switching
 - NOTE: the GUI has a non-advanced mode where users only see the country flags, however we do not want to implement that kind of selection for the TUI since we asume that TUI users are advanced users. 
 - NOTE: The current active router can be found here: GLOBAL_STATE.ActiveRouter.
 - NOTE: The connect method generates logs using the tag "connect".

There are two seperate methods for Connecting and Switching routers in the service.go file. However, the only difference between them is if the core should start the routing process or not. 
```golang
data, code, err := core.Connect(NS, true)
```
The second argument to the connect function indicates if the routing process needs to be started or not. If the user is already connected to a VPN the second argument should be "false". 
</br>
The connection request is as follows:

```golang
type CONTROLLER_SESSION_REQUEST struct {
	UserID primitive.ObjectID // The ID of the user
	DeviceToken string `json:",omitempty"`

	SLOTID int // Slot is configured by the core
	Type   string `json:",omitempty"` // Type is configured by the core

	GROUP     uint8 `json:"GROUP"` // This is the GLOBAL_STATE.ActiveRouter.GROUP field
	ROUTERID  uint8 `json:"ROUTERID"` // This is the entry GLOBAL_STATE.ActiveRouter.ROUTERID field

	XGROUP    uint8 `json:"XGROUP"` // This is the [Selected AccessPoint/VPN].GROUPID
	XROUTERID uint8 `json:"XROUTERID"`// This is the [Selected Accesspoint/VPN].ROUTERID
	DEVICEID  uint8 `json:"DEVICEID"`
  // This is the [Selected AccessPoint/VPN].DEVICEID

	// QUICK CONNECT
	Country string `json:",omitempty"` // This field is not used while running in TUI mode

	TempKey *OTK_REQUEST // This field is only used by the core
}
```


## Disconnecting
Disconnecting is simple, just call this method:
```golang
core.Disconnect()
```
## Loging out
When logging out you need to call the disconnect method and send a request to the controller to clear the device login.
```golang
core.Disconnect()

data, code, err := core.ForwardToController(FR)

type FORWARD_REQUEST struct {
	Path    string // Path: /v2/user/logout
	Method  string // Method: POST
	Timeout int // Timeout: 20000
	JSONData interface{} // structs.LogoutForm{}
}

```

## Functions
You can find all the needed functionality inside the "service.go" file located in the root folder of the project. However, we do not want to use the service layer inside the TUI since it's specifically meant for Wails.io but it will show you which functions are used by the GUI. If you are wondering how to use those function you can always check the front end code.
</br>
The TUI should only be interacting with the "core" module.
