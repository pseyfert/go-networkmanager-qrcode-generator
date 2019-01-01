package qrcode_for_nm_connection

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus"
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
			fmt.Printf("got from dbus: %v\n", callbody)
			return retval, fmt.Errorf("Could not resolve dbus \"802-11-wireless\" (ini \"wifi\")")
		}
		ssid, found := wifi["ssid"]
		if !found {
			fmt.Printf("got from dbus: %v\n", callbody)
			return retval, fmt.Errorf("Could not resolve ssid")
		}

		retval.Ssid = ssid.Value().([]byte)
	}
	{
		connection, found := resolved["connection"]
		if !found {
			fmt.Printf("got from dbus: %v\n", callbody)
			return retval, fmt.Errorf("Could not resolve \"connection\"")
		}
		id, found := connection["id"]
		if !found {
			fmt.Printf("got from dbus: %v\n", callbody)
			return retval, fmt.Errorf("Could not resolve \"id\"")
		}
		retval.Id = removeQuotes(id.String())
	}
	{
		wifisecurity, found := resolved["802-11-wireless-security"]
		if !found {
			fmt.Printf("got from dbus: %v\n", callbody)
			return retval, fmt.Errorf("Could not resolve dbus \"802-11-wireless-security\" (ini \"wifi-security\")")
		}
		keymgmt, found := wifisecurity["key-mgmt"]
		if !found {
			fmt.Printf("got from dbus: %v\n", callbody)
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

func GetNetworkSettings(settingsId int) (NetworkSetting, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("ERROR: couldn't connect to system dbus\n")
		return NetworkSetting{}, err
	}
	connectionpathstring := fmt.Sprintf("/org/freedesktop/NetworkManager/Settings/%d", settingsId)
	obj := conn.Object("org.freedesktop.NetworkManager", dbus.ObjectPath(connectionpathstring))

	settings := obj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0)
	if e := settings.Err; nil != e {
		fmt.Printf("ERROR: %v\n", e)
		return NetworkSetting{}, e
	}
	networkSettings, err := NewNetworkSetting(settings.Body[0])
	if nil != err {
		return networkSettings, err
	}

	if networkSettings.IsPsk {
		secrets := obj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSecrets", 0, "802-11-wireless-security")
		if e := secrets.Err; nil != e {
			return NetworkSetting{}, fmt.Errorf("ERROR: %v", e)
		}
		networkSettings.AddNetworkSecrets(secrets.Body[0])
		fmt.Printf("Network %s: key is %s\n", networkSettings.Id, networkSettings.Key)
	}
	return networkSettings, nil
}
