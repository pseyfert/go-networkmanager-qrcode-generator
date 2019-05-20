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
	var outputname string
	var connectionId int
	var format string
	flag.StringVar(&outputname, "o", "network.png", "output filename")
	flag.StringVar(&format, "f", "png", "output format (allowed: png, string, plain)")
	flag.IntVar(&connectionId, "i", 30, "network manager connection Id to visualize")
	flag.Parse()
	if !validformat(format) {
		fmt.Printf("ERROR: invalid format requested: %s\n", format)
		os.Exit(8)
	}

	dbusConnection, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("ERROR: couldn't connect to system dbus\n")
		os.Exit(9)
	}

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
	}

	networkSettings, err := nm2qr.GetNetworkSettings(connectionId, dbusConnection)
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
