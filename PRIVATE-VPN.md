# Private VPN

# NOTES
- It might take a couple of minutes to see updates to the Private VPN inside the NicelandVPN app after creating or updating it.
- The api.atodoslist.net domain is temporary, we will be moving away from this setup soon. The reason for this domain was to create obscurity behind our product. But eventually we found a better way to accomplish this. 

# Creating/Updating your Private VPN
1. Create your VPN config in a .json file ( See examples of VPN configs below)
2. Create the VPN<br/>
A list of router IPs can be found here: https://raw.githubusercontent.com/tunnels-is/info/master/all
```bash
curl -v -H "Content-Type: application/json" -X POST https://api.atodoslist.net/v2/device/create --resolve 'api.atodoslist.net:443:167.235.34.77' -d @vpn.json


# Example vpn.json (detailed documentation further down)
{
	"UID": "64a582dd8d9eb9e39599b522",
	"IP": "185.186.76.193",
	"Tag": "office-network",
	"InternetAccess": true,
	"LocalNetworkAccess": true,
	"AvailableMbps": 1000,
	"UserMbps": 5,
	"InterfaceIP": "185.186.76.193",
	"RouterIP": "51.89.206.24",
	"APIKey": "4c8aa0eb-87df-44d9-9a5b-a2a26fc4b122",
	"StartPort": 2000,
	"EndPort": 62000,
	"NAT": [{
		"Tag": "local-to-11",
		"Network": "185.186.76.0/24",
		"Nat": "11.11.12.0/24"
	}],
	"DNS": {
		"cname.meow.com": {
			"CNAME": "meow.com"
		},
		"meow.com": {
			"Wildcard": true,
			"IP": ["1.1.1.1", "33.33.33.33", "44.44.44.44"],
			"TXT": ["This is a text record", "second record", "third!"],
		},
		"txt.meow.com": {
			"TXT": ["This is a text record only for txt.meow.com"],
		}
	},
	"Access": [{
		"UID": "6501ba548a32a75e4a309911",
		"Tag": "User-1"
	}],
}

```
3. <b>When you create a VPN your will receive an JSON Response from the endpoint. That response contains a NEW `APIKey` and `_id` (DeviceID). This key will now be the authentication key for updating your VPN.</b>

4. <b>Save the NEW `APIKey` and `_id` variables, you will need those when updating your VPN in the future.</b> (We recommend updating your VPN .json config with these new variables)


5. Updating your VPN (Remember to replace the `APIKey` and `_id` variables)
```bash
curl -v -H "Content-Type: application/json" -X POST https://api.atodoslist.net/v2/device/update --resolve 'api.atodoslist.net:443:167.235.34.77' -d @vpn.json

# Example vpn.json (detailed documentation further down)
{
  "_id": "[ REPLCE ]", // NEW: this is the _id returned from step 2
	"UID": "64a582dd8d9eb9e39599b522",
	"IP": "185.186.76.193",
	"Tag": "office-network",
	"InternetAccess": true,
	"LocalNetworkAccess": true,
	"AvailableMbps": 1000,
	"UserMbps": 5,
	"InterfaceIP": "185.186.76.193",
	"RouterIP": "51.89.206.24",
	"APIKey": "[ REPLACE ]", // NEW: Replace the User APIKey with the APIKey returned from step 2
	"StartPort": 2000,
	"EndPort": 62000,
	"NAT": [{
		"Tag": "local-to-11",
		"Network": "185.186.76.0/24",
		"Nat": "11.11.12.0/24"
	}],
	"DNS": {
		"cname.meow.com": {
			"CNAME": "meow.com"
		},
		"meow.com": {
			"Wildcard": true,
			"IP": ["1.1.1.1", "33.33.33.33", "44.44.44.44"],
			"TXT": ["This is a text record", "second record", "third!"],
		},
		"txt.meow.com": {
			"TXT": ["This is a text record only for txt.meow.com"],
		}
	},
	"Access": [{
		"UID": "6501ba548a32a75e4a309911",
		"Tag": "User-1"
	}],
}
```
6. Deleting the VPN
```bash
```bash
curl -v -H "Content-Type: application/json" -X POST https://api.atodoslist.net/v2/device/delete --resolve 'api.atodoslist.net:443:167.235.34.77' -d @vpn.json

# Example vpn.json (detailed documentation further down)
{
  "_id": "[ REPLCE ]", // NEW: this is the _id returned from step 2
	"UID": "64a582dd8d9eb9e39599b522",
	"APIKey": "[ REPLACE ]", // NEW: Replace the User APIKey with the APIKey returned from step 2
}
```

# Installation
 1. Download the binary (niceland-network) here: https://drive.google.com/file/d/1l6zSu5f-9tXqMgdgyyXmWKzZ5Cs33thY/view?usp=drive_link
 2. Apply iptables rules

 ```bash
  $ iptables -I OUTPUT 1 --src [InterfaceIP] -p tcp --tcp-flags RST RST -j DROP
 ```
 3. Update the sysctl config (/etc/sysctl.conf) with these lines (make sure to replace `StartPort` and `Endport`)

 ```bash
net.ipv6.conf.default.disable_ipv6=1
net.ipv6.conf.all.disable_ipv6=1
# Example: 2000-62000
net.ipv4.ip_local_reserved_ports = [StartPort]-[EndPort]
# Make sure the local_port_range end port is LOWER then the VPN StartPort
net.ipv4.ip_local_port_range = 1024 1999
 ```
 Then apply the sysctl configurations
 ```bash
  $ sysctl -p
 ```
 4. Run the binary
 ```bash
  $ ./niceland-network -deviceID=[_id from the .json config] -apiKey=[APIKey form the .json config] -routerURL=https://raw.githubusercontent.com/tunnels-is/info/master/all
 ```




# VPN .json OBJECT with documentation
```json
{
  "_id":"64f084dbbb7f1d8e2e02f922", // The VPN Device ID
	"UID": "6501ba548a32a75e4a309922", // Your Account ID
	"Tag": "office-network", // Your VPN network name/tag
	"APIKey": "f1606dc9-4587-4084-8396-155100d705aa", // Your account API Key

	"RouterIP": "51.89.206.24", // The Router IP Address you want your VPN to be connected to

	"AvailableMbps": 1000, // Total available Mbps on the server. We recommend allocating about 80-90% of the available bandwidth torwards the VPN.
	"UserMbps": 5, // The minimum guaranteed bandwidth for each user on the VPN. 
  // There is no upper-limit to bandwidth, but if rate-limiting kicks in, it will make sure that users do not get rate-limited below this point.
	"StartPort": 2000, // The first port in the VPN port range
	"EndPort": 62000, // The last port in the VPN port range
  
  // HOW PORT ALLOCATION WORKS
  // The VPN will calculate how many ports (per ip / per network protocol ) each user has based on AvailableMbps and UserMbps
  //
  // MaxUsers = AvailableMbps / UserMbps
  // TotalPorts = EndPort - StartPort  
  // PortPerUser = TotalPorts / MaxUsers

	"InterfaceIP": "185.186.76.193", // The local interface IP that the VPN uses to listen for network packets
	"IP": "185.186.76.193", // The internet IP of your network
	"InternetAccess": true,  // Enables or Disables internet access from the VPN
	"LocalNetworkAccess": true, // Enables or Disables local network access from the VPN

   ** OPTIONAL
  // The Private VPN offers 3 Octet NAT for network sizes up to /12
	"NAT": [{
		"Tag": "local-to-11", // Tag for the current NAT configurations
		"Network": "185.186.76.0/24", // Local Network CIDR
		"Nat": "11.11.12.0/24" // NAT CIDR
	}],

  ** OPTIONAL
  // The Private VPN supports custom CNAME, A and TXT records. Multiple A and TXT records can be defined per domain.
	"DNS": {
		"cname.meow.com": {
			"CNAME": "meow.com"
		},
		"meow.com": {
			"Wildcard": true, // If wildcard is enabled all subdomains not matching other custom domain mappings will receive responses from this custom DNS mapping
			"IP": ["1.1.1.1", "33.33.33.33", "44.44.44.44"],
			"TXT": ["This is a text record", "second record", "third!"],
		},
		"txt.meow.com": {
			"TXT": ["This is a text record only for txt.meow.com"],
		}
	},

   ** OPTIONAL
  // The Private VPN supports multi-user access to VPN Networks. 
  // Add a users ID and a custom Tag to the Access list to enable user access to the VPN Network.
	"Access": [{
		"UID": "6501ba548a32a75e4a309911",
		"Tag": "User-1"
	}],
}

```