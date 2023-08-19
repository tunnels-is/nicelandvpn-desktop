package core

import (
	cipher "crypto/cipher"
	"crypto/ecdsa"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/google/gopacket/layers"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var PRODUCTION = false

var ENABLE_INSTERFACE = false

var A = new(Adapter)
var AS = new(AdapterSettings)
var C = new(Config)
var GLOBAL_STATE = new(State)

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567")

// var LastRouterPing = time.Now()
var LastConnectionAttemp = time.Now()

var BUFFER_ERROR bool
var IGNORE_NEXT_BUFFER_ERROR bool

var STATE_LOCK = sync.Mutex{}

var TUNNEL_ADAPTER_NAME = "NicelandVPN"
var TUNNEL_ADAPTER_ADDRESS = "10.4.3.2"
var TUNNEL_ADAPTER_ADDRESS_IP = net.IP{10, 4, 3, 2}

var MAC_CONNECTION_SETTINGS *CONNECTION_SETTINGS

var CURRENT_UBBS int = 0
var CURRENT_DBBS int = 0
var EGRESS_PACKETS uint64 = 0
var INGRESS_PACKETS uint64 = 0

// NETWORKING STUFF
var TCP_MAP = make(map[[4]byte]*IP)
var TCP_MAP_LOCK = sync.RWMutex{}

var UDP_MAP = make(map[[4]byte]*IP)
var UDP_MAP_LOCK = sync.RWMutex{}

var BlockedDomainMap = make(map[string]bool)
var BlockListLock = sync.Mutex{}

type IP struct {
	CurrentPort uint16
	LOCAL       map[uint16]*RemotePort
	REMOTE      map[uint16]*RemotePort
}

type RemotePort struct {
	Local        uint16
	Original     uint16
	Mapped       uint16
	LastActivity time.Time
}

var L = new(Logs)
var LogQueue = make(chan LogItem, 10000)
var TAG_ERROR = "ERROR"
var TAG_GENERAL = "GENERAL"
var LogFile *os.File

type LoggerInterface struct {
}

type Logs struct {
	PING       [100]string
	CONNECT    [100]string
	DISCONNECT [100]string
	SWITCH     [100]string
	GENERAL    [5000]string
}

type LogItem struct {
	Type string
	Line string
}

type LogoutForm struct {
	Email       string
	DeviceToken string
}
type LoginForm struct {
	Email       string
	Password    string
	DeviceName  string
	DeviceToken string
	Digits      string
	Recovery    string
}

type State struct {
	UMbps          int    `json:"UMbps"`
	DMbps          int    `json:"DMbps"`
	UMbpsString    string `json:"UMbpsString"`
	DMbpsString    string `json:"DMbpsString"`
	IngressPackets uint64 `json:"IngressPackets"`
	EgressPackets  uint64 `json:"EgressPackets"`
	ConnectedTimer string `json:"ConnectedTimer"`

	IsAdmin               bool `json:"IsAdmin"`
	NeedsRouterProbe      bool `json:"NeedsRouterProbe"`
	BaseFolderInitialized bool `json:"BaseFolderInitialized"`
	TunnelInitialized     bool `json:"TunnelInitialized"`
	LogFileInitialized    bool `json:"LogFileInitialized"`
	ConfigInitialized     bool `json:"ConfigInitialized"`
	ClientReady           bool `json:"ClientReady"`
	ClientStartupError    bool `json:"ClientStartupError"`
	BufferError           bool `json:"BufferError"`

	Connected  bool `json:"Connected"`
	Connecting bool `json:"Connecting"`
	Exiting    bool `json:"Exiting"`

	C                             *Config   `json:"C"`
	LastRouterPing                time.Time `json:"LastRouterPing"`
	PingReceivedFromRouter        time.Time `json:"PingReceivedFromRouter"`
	SecondsSincePingFromRouter    string    `json:"SecondsSincePingFromRouter"`
	LastAccessPointUpdate         time.Time
	SecondsUntilAccessPointUpdate int
	AvailableCountries            []string        `json:"AvailableCountries"`
	RoutersList                   [2000]*ROUTER   `json:"-"`
	Routers                       []*ROUTER       `json:"Routers"`
	AccessPoints                  []*AccessPoint  `json:"AccessPoints"`
	ActiveRouter                  *ROUTER         `json:"ActiveRouter"`
	ActiveAccessPoint             *AccessPoint    `json:"ActiveAccessPoint"`
	ActiveSession                 *CLIENT_SESSION `json:"ActiveSession"`

	// FILE PATHS
	LogFileName string `json:"LogFileName"`
	LogPath     string `json:"LogPath"`
	ConfigPath  string `json:"ConfigPath"`
	BackupPath  string `json:"BackupPath"`
	BasePath    string `json:"BasePath"`

	DefaultInterface *CONNECTION_SETTINGS `json:"DefaultInterface"`
	// DefaultRouterIP      string
	// DefaultInterfaceName string
	Version string `json:"Version"`
}
type FileConfig struct {
	DNS1           string
	DNS1Bytes      [4]byte
	DNSIP          net.IP
	DNS2           string
	ManualRouter   bool
	Region         string
	DebugLogging   bool
	Version        string
	RouterFilePath string
	AutoReconnect  bool
	KillSwitch     bool
}
type AdapterSettings struct {
	// SleepTrigger bool
	Session *CLIENT_SESSION

	TCPTunnelSocket net.Conn

	RoutingBuffer [8]byte
	PingBuffer    [8]byte

	LastActivity time.Time
	StartPort    uint16
	EndPort      uint16
	VPNIP        net.IP
	UDPHeader    layers.IPv4
	TCPHeader    layers.IPv4
	AEAD         cipher.AEAD
}
type Config struct {
	AutoReconnect  bool
	KillSwitch     bool
	DNS1           string
	DNS1Bytes      [4]byte
	DNSIP          net.IP
	DNS2           string
	ManualRouter   bool
	DebugLogging   bool
	Version        string `json:"-"`
	RouterFilePath string
	// AddBlockLevel  int
	// Region         string
	PrevSession *CONTROLLER_SESSION_REQUEST

	CLI bool `json:"-"`
}

type LOADING_LOGS_RESPONSE struct {
	Lines [100]string
}
type GENERAL_LOGS_RESPONSE struct {
	Lines []string
}
type GeneralLogResponse struct {
	Content  []string
	Time     []string
	Function []string
	Color    []string
}

type DEBUG_OUT struct {
	Lines []string
	File  string
}

type OTK struct {
	Created    time.Time
	Key        [32]byte
	PrivateKey *ecdsa.PrivateKey
	AEAD       cipher.AEAD // used to open client data
}

type OTK_REQUEST struct {
	X *big.Int
	Y *big.Int
}

type OTK_RESPONSE struct {
	X    *big.Int
	Y    *big.Int
	UUID []byte
}

type CHACHA_RESPONSE struct {
	X      *big.Int
	Y      *big.Int
	CHACHA []byte
}

type CONTROLL_PUBLIC_DEVCE_RESPONSE struct {
	Routers      []*ROUTER
	AccessPoints []*AccessPoint
}

type FORWARD_REQUEST struct {
	Path    string
	Method  string
	Timeout int
	// Data     []byte
	JSONData interface{}
}

type CONNECTION_SETTINGS struct {
	DNS1          string
	DNS2          string
	AutoDNS       bool
	IP6Method     string
	IPV6Enabled   bool
	IFName        string
	DefaultRouter string
	AdapterName   string
}

type INTERFACE_SETTINGS struct {
	Index           int
	Flags           net.Flags
	MTU             int
	HardwareAddress net.HardwareAddr
	OIF             net.Interface
	Hop             string
	Metric          int
}

type CONFIG_FORM struct {
	DNS1           string                      `json:"DNS1" bson:"-"`
	DNS2           string                      `json:"DNS2" bson:"-"`
	ManualRouter   bool                        `json:"ManualRouter" bson:"-"`
	Region         string                      `json:"Region" bson:"-"`
	Version        string                      `json:"Version" bson:"-"`
	RouterFilePath string                      `json:"RouterFilePath" bson:"-"`
	DebugLogging   bool                        `json:"DebugLogging" bson:"-"`
	AutoReconnect  bool                        `json:"AutoReconnect" bson:"-"`
	KillSwitch     bool                        `json:"KillSwitch" bson:"-"`
	PrevSession    *CONTROLLER_SESSION_REQUEST `json:"PrevSlot" bson:"-"`
}

type ROUTER struct {
	IP             string `json:"PublicIP"`
	GROUP          uint8  `json:"GROUP"`
	ROUTERID       uint8  `json:"ROUTERID"`
	Tag            string `json:"Tag"`
	MS             uint64 `json:"MS"`
	Online         bool   `json:"Online"`
	Country        string `json:"Country"`
	AvailableMbps  int    `json:"AvailableMbps"`
	Slots          int    `json:"Slots"`
	AvailableSlots int    `json:"AvailableSlots"`

	LastPing  time.Time       `json:"-"`
	PingStats ping.Statistics `json:"-"`

	TCPControllerConnection net.Conn `json:"-"`
	TCPTunnelConnection     net.Conn `json:"-"`

	ROUTER_STATS

	Score int `json:"Score"`
}

type ROUTER_STATS struct {
	AEBP      float64
	AIBP      float64
	CPUP      int
	RAMUsage  int
	DiskUsage int
}

type CONTROLLER_SESSION_REQUEST struct {
	UserID primitive.ObjectID
	ID     primitive.ObjectID

	DeviceToken string `json:",omitempty"`

	SLOTID int
	Type   string `json:",omitempty"`

	Permanent bool `json:",omitempty"`
	Count     int  `json:",omitempty"`

	GROUP     uint8 `json:"GROUP"`
	ROUTERID  uint8 `json:"ROUTERID"`
	SESSIONID uint8 `json:"SESSIONID"`

	XGROUP    uint8 `json:"XGROUP"`
	XROUTERID uint8 `json:"XROUTERID"`
	DEVICEID  uint8 `json:"DEVICEID"`

	// QUICK CONNECT
	Country string `json:",omitempty"`

	TempKey *OTK_REQUEST
}

type CLIENT_SESSION struct {
	Created time.Time
	CONTROLLER_SESSION
	PrivateKey        *ecdsa.PrivateKey `json:"-"`
	StartPort         uint16
	EndPort           uint16
	VPNIP             []byte
	ClientKeyResponse []byte `json:"CKR"`
}

type CONTROLLER_SESSION struct {
	UserID primitive.ObjectID `bson:"UID"`
	ID     primitive.ObjectID `bson:"_id"`

	Permanent bool `bson:"P"`
	Count     int  `bson:"C"`
	SLOTID    int  `bson:"SLOTID"`

	GROUP     uint8 `bson:"G"`
	ROUTERID  uint8 `bson:"RID"`
	SESSIONID uint8 `bson:"SID"`

	XGROUP    uint8 `bson:"XG"`
	XROUTERID uint8 `bson:"XRID"`
	DEVICEID  uint8 `bson:"APID"`

	Assigned     time.Time `bson:"A"`
	ShouldDelete bool      `bson:"-"`
}

type AccessPoint struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id"`

	UID primitive.ObjectID `json:"-" bson:"UID"`
	Tag string             `json:"Tag" bson:"T"`

	GROUP    uint8  `json:"GROUP" bson:"G"`
	ROUTERID uint8  `json:"ROUTERID" bson:"RID"`
	DEVICEID uint8  `json:"DEVICEID" bson:"DID"`
	IP       string `json:"IP" bson:"IP"`

	Access             []*AP_DEVICE_USER_ACCESS `json:"Access" bson:"A"`
	Networks           []*AP_DEVICE_NETWORK_MAP `json:"Networks" bson:"N"`
	Updated            time.Time                `json:"Updated" bson:"U"`
	InternetAccess     bool                     `json:"InternetAccess" bson:"I"`
	LocalNetworkAccess bool                     `json:"LocalNetworkAccess" bson:"LA"`
	Public             bool                     `json:"Public" bson:"P"`

	Online     bool       `json:"Online" bson:"O"`
	LastOnline time.Time  `json:"LastOnline" bson:"LO"`
	GEO        *AP_GEO_DB `json:"GEO,omitempty" bson:"GEO"`

	AvailableSlots int `json:"AvailableSlots" bson:"-"`
	Slots          int `json:"Slots" bson:"-"`
	AvailableMbps  int `json:"AvailableMbps" bson:"ABS"`
	UserMbps       int `json:"UserMbps" bson:"UB"`

	Router *ROUTER `json:"Router"`
}

type AP_DEVICE_USER_ACCESS struct {
	UID primitive.ObjectID `json:"UID" bson:"UID"`
	Tag string             `json:"Tag" bson:"T"`
}

type AP_DEVICE_NETWORK_MAP struct {
	Tag          string `json:"Tag" bson:"T"`
	Network      string `json:"Network" bson:"N"`
	LocalNetwork string `json:"LocalNetwork" bson:"L"`
}

type AP_GEO_DB struct {
	Updated     time.Time `json:"Updated" bson:"U"`
	IPV         string    `bson:"IPV" json:"-"`
	Country     string    `bson:"Country" json:"Country"`
	CountryFull string    `bson:"CountryFull" json:"CountryFull"`
	City        string    `bson:"City" json:"City"`
	// ASN     string `bson:"ASN" json:"ASN"`
	ISP   string `bson:"ISP" json:"-"`
	Proxy bool   `bson:"Proxy" json:"Proxy"`
	Tor   bool   `bson:"Tor" json:"Tor"`
}

var PS_IFLIST []*PS_DEFAULT_ROUTES

type PS_DEFAULT_ROUTES struct {
	// CimClass struct {
	// 	CimSuperClassName string `json:"CimSuperClassName,omitempty"`
	// 	CimSuperClass     struct {
	// 		CimSuperClassName   string `json:"CimSuperClassName"`
	// 		CimSuperClass       string `json:"CimSuperClass"`
	// 		CimClassProperties  string `json:"CimClassProperties"`
	// 		CimClassQualifiers  string `json:"CimClassQualifiers"`
	// 		CimClassMethods     string `json:"CimClassMethods"`
	// 		CimSystemProperties string `json:"CimSystemProperties"`
	// 	} `json:"CimSuperClass,omitempty"`
	// 	CimClassProperties  []string `json:"CimClassProperties,omitempty"`
	// 	CimClassQualifiers  []string `json:"CimClassQualifiers,omitempty"`
	// 	CimClassMethods     []string `json:"CimClassMethods,omitempty"`
	// 	CimSystemProperties struct {
	// 		Namespace  string      `json:"Namespace"`
	// 		ServerName string      `json:"ServerName"`
	// 		ClassName  string      `json:"ClassName"`
	// 		Path       interface{} `json:"Path"`
	// 	} `json:"CimSystemProperties,omitempty"`
	// } `json:"CimClass,omitempty"`
	// CimInstanceProperties []struct {
	// 	Name            string      `json:"Name"`
	// 	Value           interface{} `json:"Value"`
	// 	CimType         int         `json:"CimType"`
	// 	Flags           string      `json:"Flags"`
	// 	IsValueModified bool        `json:"IsValueModified"`
	// } `json:"CimInstanceProperties,omitempty"`
	// CimSystemProperties struct {
	// 	Namespace  string      `json:"Namespace"`
	// 	ServerName string      `json:"ServerName"`
	// 	ClassName  string      `json:"ClassName"`
	// 	Path       interface{} `json:"Path"`
	// } `json:"CimSystemProperties,omitempty"`
	// Publish            int         `json:"Publish"`
	// Protocol           int         `json:"Protocol"`
	// Store              int         `json:"Store"`
	// AddressFamily      int         `json:"AddressFamily"`
	// State              int         `json:"State"`
	// IfIndex int `json:"ifIndex"`
	// Caption            interface{} `json:"Caption"`
	// Description        interface{} `json:"Description"`
	// ElementName        interface{} `json:"ElementName"`
	// InstanceID         string      `json:"InstanceID"`
	// AdminDistance      interface{} `json:"AdminDistance"`
	// DestinationAddress interface{} `json:"DestinationAddress"`
	// IsStatic           interface{} `json:"IsStatic"`
	RouteMetric int `json:"RouteMetric"`
	// TypeOfRoute        int         `json:"TypeOfRoute"`
	// CompartmentID      int         `json:"CompartmentId"`
	DestinationPrefix string `json:"DestinationPrefix"`
	InterfaceAlias    string `json:"InterfaceAlias"`
	InterfaceIndex    int    `json:"InterfaceIndex"`
	InterfaceMetric   int    `json:"InterfaceMetric"`
	NextHop           string `json:"NextHop"`
	// PreferredLifetime  struct {
	// 	Ticks             int64   `json:"Ticks"`
	// 	Days              int     `json:"Days"`
	// 	Hours             int     `json:"Hours"`
	// 	Milliseconds      int     `json:"Milliseconds"`
	// 	Minutes           int     `json:"Minutes"`
	// 	Seconds           int     `json:"Seconds"`
	// 	TotalDays         float64 `json:"TotalDays"`
	// 	TotalHours        float64 `json:"TotalHours"`
	// 	TotalMilliseconds int64   `json:"TotalMilliseconds"`
	// 	TotalMinutes      float64 `json:"TotalMinutes"`
	// 	TotalSeconds      float64 `json:"TotalSeconds"`
	// } `json:"PreferredLifetime"`
	// ValidLifetime struct {
	// 	Ticks             int64   `json:"Ticks"`
	// 	Days              int     `json:"Days"`
	// 	Hours             int     `json:"Hours"`
	// 	Milliseconds      int     `json:"Milliseconds"`
	// 	Minutes           int     `json:"Minutes"`
	// 	Seconds           int     `json:"Seconds"`
	// 	TotalDays         float64 `json:"TotalDays"`
	// 	TotalHours        float64 `json:"TotalHours"`
	// 	TotalMilliseconds int64   `json:"TotalMilliseconds"`
	// 	TotalMinutes      float64 `json:"TotalMinutes"`
	// 	TotalSeconds      float64 `json:"TotalSeconds"`
	// } `json:"ValidLifetime"`
	// PSComputerName interface{} `json:"PSComputerName"`
}

type TWO_FACTOR_CONFIRM struct {
	Email  string
	Code   string
	Digits string
}

type QR_CODE struct {
	Value string
	// Recovery string
}
