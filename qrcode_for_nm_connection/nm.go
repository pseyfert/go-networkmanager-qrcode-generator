/*
 * Copyright (C) 2019 Paul Seyfert
 * Author: Paul Seyfert <pseyfert.mathphys@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

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
	DbusId   int
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
			return retval, fmt.Errorf("Could not resolve dbus \"802-11-wireless\" (ini \"wifi\"): %v", callbody)
		}
		ssid, found := wifi["ssid"]
		if !found {
			return retval, fmt.Errorf("Could not resolve ssid. got from dbus: %v", callbody)
		}

		retval.Ssid = ssid.Value().([]byte)
	}
	{
		connection, found := resolved["connection"]
		if !found {
			return retval, fmt.Errorf("Could not resolve \"connection\". got from dbus: %v", callbody)
		}
		id, found := connection["id"]
		if !found {
			return retval, fmt.Errorf("Could not resolve \"id\". got from dbus: %v", callbody)
		}
		retval.Id = removeQuotes(id.String())
	}
	{
		wifisecurity, found := resolved["802-11-wireless-security"]
		if !found {
			return retval, fmt.Errorf("Could not resolve dbus \"802-11-wireless-security\" (ini \"wifi-security\"). got from dbus: %v\n", callbody)
		}
		keymgmt, found := wifisecurity["key-mgmt"]
		if !found {
			return retval, fmt.Errorf("Could not resolve key-mgmt. got from dbus: %v\n", callbody)
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

func GetNetworkSettings(settingsId int, conn *dbus.Conn) (NetworkSetting, error) {
	connectionpathstring := fmt.Sprintf("/org/freedesktop/NetworkManager/Settings/%d", settingsId)
	obj := conn.Object("org.freedesktop.NetworkManager", dbus.ObjectPath(connectionpathstring))

	settings := obj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0)
	if e := settings.Err; nil != e {
		return NetworkSetting{}, e
	}
	networkSettings, err := NewNetworkSetting(settings.Body[0])
	networkSettings.DbusId = settingsId
	if nil != err {
		return networkSettings, err
	}

	if networkSettings.IsPsk {
		secrets := obj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSecrets", 0, "802-11-wireless-security")
		if e := secrets.Err; nil != e {
			return NetworkSetting{}, fmt.Errorf("ERROR: %v", e)
		}
		networkSettings.AddNetworkSecrets(secrets.Body[0])
	}
	return networkSettings, nil
}
