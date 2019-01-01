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

	nm2qr "github.com/pseyfert/go-networkmanager-qrcode-generator/qrcode_for_nm_connection"
)

func main() {
	var outputname string
	var connectionId int
	flag.StringVar(&outputname, "o", "network.png", "output filename")
	flag.IntVar(&connectionId, "i", 30, "network manager connection Id to visualize")
	flag.Parse()
	networkSettings, err := nm2qr.GetNetworkSettings(connectionId)
	if nil != err {
		fmt.Printf("something went wrong in network setting retrival, %v\n", err)
		os.Exit(1)
	}

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
