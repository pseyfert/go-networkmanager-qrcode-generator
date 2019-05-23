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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/godbus/dbus"
	nm2qr "github.com/pseyfert/go-networkmanager-qrcode-generator/qrcode_for_nm_connection"
	ux "github.com/pseyfert/go-networkmanager-qrcode-generator/ux"
)

func validformat(s string) bool {
	return s == "png" || s == "plain" || s == "string"
}

func main() {
	dbusConnection, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("ERROR: couldn't connect to system dbus\n")
		os.Exit(9)
	}

	var outputname string
	var connectionId int
	var connectionName string
	var format string
	var exactMatch bool
	var listConnections bool
	flag.StringVar(&outputname, "o", "network.png", "output filename")
	flag.StringVar(&format, "f", "png", "output format (allowed: png, string, plain)")
	flag.IntVar(&connectionId, "i", -1, "network manager connection Id to visualize")
	flag.StringVar(&connectionName, "n", "", "network manager connection name to visualize")
	flag.BoolVar(&exactMatch, "e", false, "matches by name must be exact (fuzzy by default)")
	flag.BoolVar(&listConnections, "l", false, "list connection names and quit")

	flag.Parse()
	if !validformat(format) {
		fmt.Printf("ERROR: invalid format requested: %s\n", format)
		os.Exit(8)
	}
	if listConnections {
		fmt.Printf("the following connections are known:\n")
		cons, err := ux.AllConnections(dbusConnection)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(8)
		}
		for _, con := range cons {
			fmt.Printf("%s:\tSSID %s\n", con.Id, con.Ssid)
		}
		os.Exit(0)
	}
	if connectionId < 0 && connectionName == "" {
		fmt.Printf("ERROR: specify either a connection ID or a connection name")
		os.Exit(7)
	}

	var networkSettings nm2qr.NetworkSetting
	if connectionId >= 0 {
		ids, err := ux.ConnectionIDs(dbusConnection)
		if nil != err {
			fmt.Printf("could not obtain list of connections: %v\n", err)
			fmt.Print("continuing\n")
		} else {
			found := false
			for _, id := range ids {
				if id == connectionId {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("%d is not in the list of known connections. trying anyway.\n", connectionId)
			}
			networkSettings, err = nm2qr.GetNetworkSettings(connectionId, dbusConnection)
		}
	} else {
		networkSettings, err = ux.BestMatch(connectionName, dbusConnection)
		if exactMatch && !(networkSettings.Id == connectionName || string(networkSettings.Ssid) == connectionName) {
			fmt.Printf("%s is not in the list of known connections.\n", connectionName)
			os.Exit(3)
		}
	}

	if nil != err {
		fmt.Printf("something went wrong in network setting retrival, %v\n", err)
		os.Exit(1)
	}

	if format == "plain" {
		qr := nm2qr.NetworkCode(networkSettings)
		fmt.Printf("QR code should contain:\n%s\n", qr)
	} else if format == "string" {
		qr, err := nm2qr.QRNetworkCode(networkSettings)
		if nil != err {
			fmt.Printf("something went wrong in qr code generation, %v\n", err)
			os.Exit(2)
		}
		// fmt.Printf("QR code as string:\n%s\n", nm2qr.CompressQR(qr.ToString(false)))
		fmt.Printf("QR code as string:\n%s\n", qr.ToString(false))
		fmt.Printf("QR code as string:\n%s\n", qr.ToSmallString(false))
	} else {
		qr, err := nm2qr.QRNetworkCode(networkSettings)
		if nil != err {
			fmt.Printf("something went wrong in qr code generation, %v\n", err)
			os.Exit(2)
		}
		err = qr.WriteFile(-5, outputname)
		if nil != err {
			fmt.Printf("something went wrong in qr code storing, %v\n", err)
			os.Exit(3)
		}
	}
}
