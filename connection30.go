package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/godbus/dbus"
	qrcode "github.com/skip2/go-qrcode"
)

type NetworkSetting struct {
	Ssid     []byte
	Id       string
	Sec      string // WPA or WEP
	IsPsk    bool
	IsHidden bool
	Key      string
	// contents from dbus are:
	// 802-11-wireless: map[mac-address:@ay [0xa0, 0x88, …] mac-address-blacklist:@as [] mode:"infrastructure" security:"802-11-wireless-security" ssid:@ay [0x50, …]]
	// connection: map[permissions:["user:…"] type:"802-11-wireless" uuid:"c3…" id:"P…"]
	// 802-11-wireless-security: map[auth-alg:"open" key-mgmt:"wpa-psk"]
}

func removeQuotes(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s[1 : len(s)-1]
	}
	return s
}

func (ns *NetworkSetting) AddNetworkSecrets(callbody interface{}) error {
	networkSecrets := callbody.(map[string]map[string]dbus.Variant)

	wifisecurity, found := networkSecrets["802-11-wireless-security"]
	if !found {
		return fmt.Errorf("No 802-11-wireless-security block in network Secrets")
	}
	networkKey, found := wifisecurity["psk"]
	if !found {
		return fmt.Errorf("No key in 802-11-wireless-security block")
	}
	ns.Key = removeQuotes(networkKey.String())
	return nil
}

func NewNetworkSetting(callbody interface{}) (NetworkSetting, error) {
	var retval NetworkSetting
	resolved := callbody.(map[string]map[string]dbus.Variant)
	{
		wifi, found := resolved["802-11-wireless"]
		if !found {
			return retval, fmt.Errorf("Could not resolve dbus \"802-11-wireless\" (ini \"wifi\")")
		}
		ssid, found := wifi["ssid"]
		if !found {
			return retval, fmt.Errorf("Could not resolve ssid")
		}

		retval.Ssid = ssid.Value().([]byte)
	}
	{
		connection, found := resolved["connection"]
		if !found {
			return retval, fmt.Errorf("Could not resolve \"connection\"")
		}
		id, found := connection["id"]
		if !found {
			return retval, fmt.Errorf("Could not resolve \"id\"")
		}
		retval.Id = removeQuotes(id.String())
	}
	{
		wifisecurity, found := resolved["802-11-wireless-security"]
		if !found {
			return retval, fmt.Errorf("Could not resolve dbus \"802-11-wireless-security\" (ini \"wifi-security\")")
		}
		keymgmt, found := wifisecurity["key-mgmt"]
		if !found {
			return retval, fmt.Errorf("Could not resolve key-mgmt")
		}
		keymgmt_string := removeQuotes(keymgmt.String())
		if strings.HasPrefix(keymgmt_string, "wpa") {
			retval.Sec = "WPA"
		} else if strings.HasPrefix(keymgmt_string, "wep") {
			retval.Sec = "WEP"
		} else {
			retval.Sec = "unknown"
		}
		retval.IsPsk = strings.HasSuffix(keymgmt_string, "-psk")
	}
	retval.IsHidden = false // TODO: implement

	return retval, nil
}

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}
	obj := conn.Object("org.freedesktop.NetworkManager",
		"/org/freedesktop/NetworkManager/Settings/30")

	fmt.Printf("Got Object with path %s\n", obj.Path())
	settings := obj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0)
	if e := settings.Err; nil != e {
		fmt.Printf("ERROR: %v", e)
		os.Exit(2)
	}
	networkSettings, err := NewNetworkSetting(settings.Body[0])
	if nil != err {
		os.Exit(666)
	}

	if networkSettings.IsPsk {
		secrets := obj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSecrets", 0, "802-11-wireless-security")
		if e := secrets.Err; nil != e {
			fmt.Printf("ERROR: %v", e)
			os.Exit(2)
		}
		networkSettings.AddNetworkSecrets(secrets.Body[0])
		fmt.Printf("Network %s: key is %s", networkSettings.Id, networkSettings.Key)
	}

	qr, err := QRNetworkCode(networkSettings)
	if nil != err {
		fmt.Printf("AHHHH, %v\n", err)
		os.Exit(777)
	}
	qr.WriteFile(-5, "network.png")

}

func QRNetworkCode(ns NetworkSetting) (qrcode.QRCode, error) {
	var setupcode string
	setupcode += "WIFI:"
	if ns.IsPsk {
		if ns.Sec == "WPA" {
			setupcode += "T:WPA;"
		} else if ns.Sec == "WEP" {
			setupcode += "T:WEP;"
		}
		setupcode += "P:\"" + ns.Key + "\";"
	}
	setupcode += "S:\""
	setupcode += string(ns.Ssid)
	setupcode += "\";"
	if ns.IsHidden {
		setupcode += "H:true;"
	}
	setupcode += ";"

	code, err := qrcode.New(setupcode, qrcode.Medium)
	if nil != err {
		return qrcode.QRCode{}, err
	}
	return *code, err
}

// this QR code documentation is taken from
// https://github.com/zxing/zxing/wiki/Barcode-Contents
// accessible under https://github.com/zxing/zxing.wiki.git
// The zxing project is under the Apache License v2.0 http://www.apache.org/licenses/LICENSE-2.0.html
//
// Wi-Fi Network config (Android, iOS 11+)
//
// We propose a syntax like "MECARD" for specifying wi-fi configuration. Scanning such a code would, after prompting the user, configure the device's Wi-Fi accordingly. Example:
//
// WIFI:T:WPA;S:mynetwork;P:mypass;;
//
// Parameter 	Example 	Description
// T 	WPA 	Authentication type; can be WEP or WPA, or 'nopass' for no password. Or, omit for no password.
// S 	mynetwork 	Network SSID. Required. Enclose in double quotes if it is an ASCII name, but could be interpreted as hex (i.e. "ABCD")
// P 	mypass 	Password, ignored if T is "nopass" (in which case it may be omitted). Enclose in double quotes if it is an ASCII name, but could be interpreted as hex (i.e. "ABCD")
// H 	true 	Optional. True if the network SSID is hidden.
//
// Order of fields does not matter. Special characters "", ";", "," and ":" should be escaped with a backslash ("") as in MECARD encoding. For example, if an SSID was literally "foo;bar\baz" (with double quotes part of the SSID name itself) then it would be encoded like: WIFI:S:\"foo\;bar\\baz\";;
