package core

import (
	cipher "crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/google/uuid"
	"github.com/zveinn/tcpcrypt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var PRODUCTION = false

var ENABLE_INSTERFACE = false

var (
	// A = new(Adapter)
	// AS           = new(AdapterSettings)
	C            = new(Config)
	GLOBAL_STATE = new(State)
)

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567")

// var LastRouterPing = time.Now()
var LastConnectionAttemp = time.Now()

var (
	BUFFER_ERROR             bool
	IGNORE_NEXT_BUFFER_ERROR bool
)

// var STATE_LOCK = sync.Mutex{}

var (
	TUNNEL_ADAPTER_NAME       = "nvpn"
	TUNNEL_ADAPTER_ADDRESS    = "10.0.0.3"
	TUNNEL_ADAPTER_ADDRESS_IP = net.IP{10, 0, 0, 3}
)

var MAC_CONNECTION_SETTINGS *CONNECTION_SETTINGS

var (
	CURRENT_UBBS    int    = 0
	CURRENT_DBBS    int    = 0
	EGRESS_PACKETS  uint64 = 0
	INGRESS_PACKETS uint64 = 0
)

// NETWORKING STUFF
var (
	TCP_MAP      = make(map[[4]byte]*IP)
	TCP_MAP_LOCK = sync.RWMutex{}
)

var (
	UDP_MAP      = make(map[[4]byte]*IP)
	UDP_MAP_LOCK = sync.RWMutex{}
)

var DNSWhitelist = make(map[string]bool)

var GLOBAL_BLOCK_LIST = make(map[string]bool)

type IP struct {
	LOCAL  map[uint16]*RemotePort
	REMOTE map[uint16]*RemotePort
}

type RemotePort struct {
	Local        uint16
	Original     uint16
	Mapped       uint16
	LastActivity time.Time
}

var (
	L           = new(Logs)
	LogQueue    = make(chan LogItem, 10000)
	TAG_ERROR   = "ERROR"
	TAG_GENERAL = "GENERAL"
	LogFile     *os.File
)

type LoggerInterface struct{}

type Logs struct {
	LOGS [2000]string
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

	C                             *Config              `json:"C"`
	DefaultInterface              *CONNECTION_SETTINGS `json:"DefaultInterface"`
	LastRouterPing                time.Time            `json:"LastRouterPing"`
	PingReceivedFromRouter        time.Time            `json:"PingReceivedFromRouter"`
	SecondsSincePingFromRouter    string               `json:"SecondsSincePingFromRouter"`
	LastAccessPointUpdate         time.Time
	SecondsUntilAccessPointUpdate int
	AvailableCountries            []string        `json:"AvailableCountries"`
	RoutersList                   [2000]*ROUTER   `json:"-"`
	Routers                       []*ROUTER       `json:"Routers"`
	AccessPoints                  []*VPNNode      `json:"AccessPoints"`
	PrivateAccessPoints           []*VPNNode      `json:"PrivateAccessPoints"`
	ActiveRouter                  *ROUTER         `json:"ActiveRouter"`
	ActiveAccessPoint             *VPNNode        `json:"ActiveAccessPoint"`
	ActiveSession                 *CLIENT_SESSION `json:"ActiveSession"`

	// FILE PATHS
	LogFileName   string `json:"LogFileName"`
	LogPath       string `json:"LogPath"`
	ConfigPath    string `json:"ConfigPath"`
	BackupPath    string `json:"BackupPath"`
	BlockListPath string `json:"BlockListPath"`
	BasePath      string `json:"BasePath"`

	Version string `json:"Version"`

	// BLOCKING AND PARENTAL CONTROLS
	BLists []*List `json:"BlockLists"`
	// DNSCaptureEnabled bool    `json:"DNSCaptureEnabled"`
	// DNSWhitelistEnabled bool    `json:"DNSWhitelistEnabled"`
}

type List struct {
	FullPath string
	Tag      string
	Enabled  bool
	Domains  string
}

type CONFIG_FORM struct {
	DNS1                      string                      `json:"DNS1"`
	DNS2                      string                      `json:"DNS2"`
	ManualRouter              bool                        `json:"ManualRouter"`
	Region                    string                      `json:"Region"`
	Version                   string                      `json:"Version"`
	RouterFilePath            string                      `json:"RouterFilePath"`
	DebugLogging              bool                        `json:"DebugLogging"`
	AutoReconnect             bool                        `json:"AutoReconnect"`
	KillSwitch                bool                        `json:"KillSwitch"`
	PrevSession               *CONTROLLER_SESSION_REQUEST `json:"PrevSlot"`
	DisableIPv6OnConnect      bool                        `json:"DisableIPv6OnConnect"`
	CloseConnectionsOnConnect bool                        `json:"CloseConnectionsOnConnect"`
	CustomDNS                 bool                        `json:"CustomDNS"`
	EnabledBlockLists         []string                    `json:"EnabledBlockLists"`
	LogBlockedDomains         bool                        `json:"LogBlockedDomains"`
}

var CONNECTIONS = make(map[string]*VPNConnection, 100)

type VPNConnection struct {
	Name string

	// TUN/TAP
	Tun          *Adapter
	Address      string
	AddressNetIP net.IP
	Routes       []string

	// ??????
	Session *CLIENT_SESSION
	EVPNS   *tcpcrypt.SocketWrapper

	// STATES
	PingReceived time.Time
	Connected    bool
	Connecting   bool
	Exiting      bool

	// VPN NODE
	Node       *VPNNode
	NodeSrcIP  net.IP
	PingBuffer [8]byte
	StartPort  uint16
	EndPort    uint16

	// Stats
	EgressBytes    int
	EgressPackets  int
	IngressBytes   int
	IngressPackets int

	// DNS
	PrevDNS  net.IP
	DNSBytes [4]byte
	DNSIP    net.IP

	//  NAT
	NAT_CACHE         map[[4]byte][4]byte `json:"-"`
	REVERSE_NAT_CACHE map[[4]byte][4]byte `json:"-"`

	// PORT MAPPING
	TCP_MAP [256]*O1
	UDP_MAP [256]*O1

	//  PACKET MANIPULATION
	EP_Version  byte
	EP_Protocol byte

	EP_DstIP [4]byte

	EP_IPv4HeaderLength byte
	EP_IPv4Header       []byte
	EP_TPHeader         []byte

	EP_SrcPort    [2]byte
	EP_DstPort    [2]byte
	EP_MappedPort *RP

	EP_NAT_IP [4]byte
	EP_NAT_OK bool

	EP_RST byte

	EP_DNS_Response         []byte
	EP_DNS_OK               bool
	EP_DNS_Port_Placeholder [2]byte
	EP_DNS_Packet           []byte

	// This IP gets over-written on connect
	EP_VPNSrcIP [4]byte

	EP_NEW_RST  int
	PREV_DNS_IP [4]byte
	IS_UNIX     bool

	IP_Version  byte
	IP_Protocol byte

	IP_DstIP [4]byte
	IP_SrcIP [4]byte

	IP_IPv4HeaderLength byte
	IP_IPv4Header       []byte
	IP_TPHeader         []byte

	IP_SrcPort    [2]byte
	IP_DstPort    [2]byte
	IP_MappedPort *RP

	IP_NAT_IP [4]byte
	IP_NAT_OK bool

	// This IP gets over-written on connect
	// IP_VPNSrcIP [4]byte
	IP_InterfaceIP [4]byte
}

func (V *VPNConnection) Disconnect() {
	// TODO
	// 1. clean port mappings
	// 2. clean nat nat cache
	// 3. disconnect tunnel
	// 4. delete tuntap ???????
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
	Version        string
	RouterFilePath string

	DomainWhitelist           string
	EnabledBlockLists         []string
	LogBlockedDomains         bool
	DisableIPv6OnConnect      bool
	CloseConnectionsOnConnect bool
	CustomDNS                 bool

	PrevSession *CONTROLLER_SESSION_REQUEST `json:"-"`
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

type FINAL_SEAL struct {
	Created    time.Time
	Key        []byte
	PrivateKey *N_PrivateKey
	PublicKey  *N_PublicKey
	AEAD       cipher.AEAD
}

func (S *FINAL_SEAL) ECDH() (err error) {
	S.Key, err = S.PrivateKey.KEY.ECDH(S.PublicKey.KEY)
	return
}

func (S *FINAL_SEAL) PublicKeyFromRequest(R *N_KeyRequest) (err error) {
	S.PublicKey = new(N_PublicKey)
	S.PublicKey.KEY, err = ecdh.P521().NewPublicKey(R.PB)
	if err != nil {
		return
	}
	S.PublicKey.UUID = string(R.UUID)
	return
}

func NewNicelandPrivateKey() (PK *N_PrivateKey, err error) {
	PK = new(N_PrivateKey)
	PK.KEY, err = ecdh.P521().GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	PK.UUID = uuid.NewString()
	return
}

func NewNicelandPublicKeyFromBytes(b []byte) (PK *N_PublicKey, err error) {
	PK = new(N_PublicKey)
	PK.KEY, err = ecdh.P521().NewPublicKey(b)
	if err != nil {
		return
	}
	PK.UUID = uuid.NewString()
	return
}

type N_KeyRequest struct {
	PB   []byte
	UUID []byte
}

type N_PrivateKey struct {
	KEY  *ecdh.PrivateKey
	UUID string
}

func (NP *N_PrivateKey) PublicKeyToRequest() *N_KeyRequest {
	return &N_KeyRequest{
		PB:   NP.KEY.PublicKey().Bytes(),
		UUID: []byte(NP.UUID),
	}
}

type N_PublicKey struct {
	KEY  *ecdh.PublicKey
	UUID string
}

func (NP *N_PublicKey) ToRequest() *N_KeyRequest {
	return &N_KeyRequest{
		PB:   NP.KEY.Bytes(),
		UUID: []byte(NP.UUID),
	}
}

type OTK_REQUEST struct {
	X *big.Int
	Y *big.Int
	// B []byte
}

type OTK_RESPONSE struct {
	X *big.Int
	Y *big.Int
	// B    []byte
	UUID []byte
}

type CHACHA_RESPONSE struct {
	X      *big.Int
	Y      *big.Int
	CHACHA []byte
}

type CONTROLL_PUBLIC_DEVCE_RESPONSE struct {
	Routers      []*ROUTER
	AccessPoints []*VPNNode
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

	Proto     string `json:"Proto"`
	Port      string `json:"Port"`
	GROUP     uint8  `json:"GROUP"`
	ROUTERID  uint8  `json:"ROUTERID"`
	SESSIONID uint8  `json:"SESSIONID"`

	XGROUP    uint8 `json:"XGROUP"`
	XROUTERID uint8 `json:"XROUTERID"`
	DEVICEID  uint8 `json:"DEVICEID"`

	// QUICK CONNECT
	Country string `json:",omitempty"`

	// NEW REF
	Name string `json:"-"`
}

type CLIENT_SESSION struct {
	Created time.Time
	CONTROLLER_SESSION
	StartPort uint16
	EndPort   uint16
	VPNIP     []byte
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

type VPNNode struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id"`

	UID primitive.ObjectID `json:"-" bson:"UID"`
	Tag string             `json:"Tag" bson:"T"`

	GROUP    uint8  `json:"GROUP" bson:"G"`
	ROUTERID uint8  `json:"ROUTERID" bson:"RID"`
	DEVICEID uint8  `json:"DEVICEID" bson:"DID"`
	IP       string `json:"IP" bson:"IP"`

	Access             []*DeviceUserRegistration `json:"Access" bson:"A"`
	Updated            time.Time                 `json:"Updated" bson:"U"`
	InternetAccess     bool                      `json:"InternetAccess" bson:"I"`
	LocalNetworkAccess bool                      `json:"LocalNetworkAccess" bson:"LA"`
	Public             bool                      `json:"Public" bson:"P"`

	Online     bool       `json:"Online" bson:"O"`
	LastOnline time.Time  `json:"LastOnline" bson:"LO"`
	GEO        *AP_GEO_DB `json:"GEO,omitempty" bson:"GEO"`

	AvailableSlots int `json:"AvailableSlots" bson:"-"`
	Slots          int `json:"Slots" bson:"-"`
	AvailableMbps  int `json:"AvailableMbps" bson:"ABS"`
	UserMbps       int `json:"UserMbps" bson:"UB"`

	Country     string `json:"Country" bson:"Country"`
	CountryFull string `json:"CountryFull" bson:"CountryFull"`
	// MIGHT USE
	NAT             []*DeviceNatRegistration          `json:"NAT" bson:"NAT"`
	BlockedNetworks []string                          `json:"BlockedNetworks" bson:"BlockedNetworks"`
	DNS             map[string]*DeviceDNSRegistration `json:"DNS" bson:"DNS"`
	DNSWhiteList    bool                              `json:"DNSWhiteList" bson:"DNSWhiteList"`

	Router *ROUTER `json:"Router"`

	// NAT_CACHE         map[[4]byte][4]byte `json:"-"`
	// REVERSE_NAT_CACHE map[[4]byte][4]byte `json:"-"`
}

// type DNSMap struct {
// 	IP       string
// 	Wildcard bool
// }

type DeviceDNSRegistration struct {
	Wildcard bool     `json:"Wildcard" bson:"Wildcard"`
	IP       []string `json:"IP" bson:"IP"`
	TXT      []string `json:"TXT" bson:"TXT"`
	CNAME    string   `json:"CNAME" bson:"CNAME"`
}

type DeviceNatRegistration struct {
	Tag     string `json:"Tag" bson:"T"`
	Network string `json:"Network" bson:"N"`
	Nat     string `json:"Nat" bson:"L"`
}

type DeviceUserRegistration struct {
	UID primitive.ObjectID `json:"UID" bson:"UID"`
	Tag string             `json:"Tag" bson:"T"`
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

// var CurrentOpenSockets []*OpenSockets

//	type OpenSockets struct {
//		RemoteAddress string  `json:"RemoteAddress"`
//		RemoteIP      [4]byte `json:"-"`
//		LocalPort     uint16  `json:"LocalPort"`
//		RemotePort    uint16  `json:"RemotePort"`
//	}
type MIB_TCPROW_OWNER_PID struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

type MIB_TCPTABLE_OWNER_PID struct {
	dwNumEntries uint32
	table        [30000]MIB_TCPROW_OWNER_PID
}

// Device token struct need for the login respons from user scruct
type DEVICE_TOKEN struct {
	DT      string    `bson:"DT"`
	N       string    `bson:"N"`
	Created time.Time `bson:"C"`
}

// use struct you get from the login request
type User struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	APIKey           string             `bson:"AK" json:"APIKey"`
	Email            string             `bson:"E"`
	TwoFactorEnabled bool               `json:"TwoFactorEnabled" bson:"TFE"`
	Disabled         bool               `bson:"D" json:"Disabled"`
	Tokens           []*DEVICE_TOKEN    `json:"Tokens" bson:"T"`
	DeviceToken      *DEVICE_TOKEN      `json:",omitempty" bson:"-"`

	CashCode      int       `bson:"CSC" json:"CashCode"`
	Affiliate     string    `bson:"AF"`
	SubLevel      int       `bson:"SUL"`
	SubExpiration time.Time `bson:"SE"`
	TrialStarted  time.Time `bson:"TrialStarted" json:"TrialStarted"`

	CancelSub bool `json:"CancelSub" bson:"CS"`

	Version string `json:"Version" bson:"-"`
}
