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

package ux

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/godbus/dbus"
	nm2qr "github.com/pseyfert/go-networkmanager-qrcode-generator/qrcode_for_nm_connection"
	fuzzy "github.com/schollz/closestmatch"
	// fuzzy "github.com/schollz/closestmatch/levenshtein"
)

type outernode struct {
	XMLName xml.Name    `xml:"node"`
	Nodes   []innernode `xml:"node"`
}
type innernode struct {
	XMLName xml.Name `xml:"node"`
	Name    string   `xml:"name,attr"`
	// Id      int
}

func ConnectionIDs(conn *dbus.Conn) ([]int, error) {
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")

	settings := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0)
	if e := settings.Err; nil != e {
		fmt.Printf("ERROR: %v\n", e)
		return []int{}, e
	}

	introspection := settings.Body[0].(string)

	var n outernode
	xml.Unmarshal([]byte(introspection), &n)

	fmt.Printf("Got %d connections\n", len(n.Nodes))
	retval := make([]int, len(n.Nodes))
	errors := 0
	for i, _ := range n.Nodes {
		// n.Nodes[i].Id, _ = strconv.Atoi(n.Nodes[i].Name)
		id, err := strconv.Atoi(n.Nodes[i].Name)
		if err != nil {
			fmt.Printf("Unexpeced connection setting id: %s\n", n.Nodes[i])
			errors += 1
		} else {
			retval[i-errors] = id
		}
	}

	return retval[0 : len(n.Nodes)-errors], nil
}

func AllConnections(conn *dbus.Conn) ([]nm2qr.NetworkSetting, error) {
	ids, err := ConnectionIDs(conn)

	if err != nil {
		return []nm2qr.NetworkSetting{}, err
	}

	networks := make([]nm2qr.NetworkSetting, 0, len(ids))
	for _, id := range ids {
		networkSettings, err := nm2qr.GetNetworkSettings(id, conn)
		if err == nil {
			networks = append(networks, networkSettings)
		}
	}
	return networks, nil
}

func BestMatch(connectionName string, dbusConnection *dbus.Conn) (nm2qr.NetworkSetting, error) {
	networks, err := AllConnections(dbusConnection)
	if err != nil {
		var retval nm2qr.NetworkSetting
		return retval, err
	}
	networkNames := make([]string, 0, 2*len(networks))
	networkMaps := make(map[string]nm2qr.NetworkSetting)
	for _, networkSettings := range networks {
		networkNames = append(networkNames, networkSettings.Id)
		networkNames = append(networkNames, string(networkSettings.Ssid))
		networkMaps[string(networkSettings.Ssid)] = networkSettings
		networkMaps[networkSettings.Id] = networkSettings
	}
	cm := fuzzy.New(networkNames, []int{2, 3, 4})
	best := cm.Closest(connectionName)
	networkSettings := networkMaps[best]
	return networkSettings, nil
}
