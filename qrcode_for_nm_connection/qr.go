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
	"os"
	"strings"
	"unicode/utf8"

	qrcode "github.com/skip2/go-qrcode"
)

func NetworkCode(ns NetworkSetting) string {
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
	return setupcode
}

func QRNetworkCode(ns NetworkSetting) (qrcode.QRCode, error) {
	setupcode := NetworkCode(ns)

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

// ████ ▄▄▄▄▄ █ ▀▀▄ ▀▀▄  ▀ ▄▄▀█ █ ▄▄▄▄▄ ████
// ████ █   █ ███ ▄▄█ ▄▄ ▀██▀ █ █ █   █ ████
// ████ █▄▄▄█ █ ▄▄ █▄▄ █   █ ▀▄██ █▄▄▄█ ████
// ████▄▄▄▄▄▄▄█ █ ▀ █ ▀ █▄█ █▄▀ █▄▄▄▄▄▄▄████
// ████  ▄▀▀▄▄▀▄  ██  █▄▄▀ ▄▄▀▄█ ▄ ▄▀▀  ████
// ████▀ █▄▄ ▄█▄▀█▄█▀█▄▄██ ▄██▀    █ ▀█ ████
// ████▀▄ ▄▀ ▄▀▄▄█▄ ▄█ ██▄ ▄▄▀ ██ █  █▀ ████

func CompressQR(in string) string {
	lines := strings.Split(in, "\n")
	var out string
	if len(lines[0])%2 == 1 {
		os.Exit(1337)
	}
	for r := 0; r < len(lines)-1; r += 2 {
		for iu, il, wu, wl := 0, 0, 0, 0; il < len(lines[r+1]) && iu < len(lines[r]); iu, il = iu+wu, il+wl {
			upperrune, wu := utf8.DecodeRuneInString(lines[r][iu:])
			upperrune_, wu_ := utf8.DecodeRuneInString(lines[r][iu+wu:])
			lowerrune, wl := utf8.DecodeRuneInString(lines[r+1][il:])
			lowerrune_, wl_ := utf8.DecodeRuneInString(lines[r+1][il+wl:])
			if (lowerrune != lowerrune_) || (upperrune != upperrune_) {
				panic("consecutive runes should be equal")
			}

			if upperrune == rune('\u2588') && lowerrune == rune('\u2588') {
				fmt.Printf("%q", rune('\u2588'))
				// out = fmt.Sprintf(out, "%s%s", out, rune('\u2588'))
			}
			if upperrune == rune('\u0020') && lowerrune == rune('\u0020') {
				fmt.Printf("%q", rune('\u0020'))
				// out = fmt.Sprintf(out, "%s%s", out, rune('\u0020'))
			}
			if upperrune == rune('\u2588') && lowerrune == rune('\u0020') {
				fmt.Printf("%q", rune('\u2580'))
				// out = fmt.Sprintf(out, "%s%s", out, rune('\u2580'))
			}
			if upperrune == rune('\u0020') && lowerrune == rune('\u2588') {
				fmt.Printf("%q", rune('\u2584'))
				// out = fmt.Sprintf(out, "%s%s", out, rune('\u2584'))
			}

			iu += wu + wu_
			il += wl + wl_
		}
		fmt.Print("\n")
		// out += "\n"

		// >>> print('{:x}'.format(ord('▄')))
		// 2584
		// >>> print('{:x}'.format(ord('▀')))
		// 2580
		// >>> print('{:x}'.format(ord('█')))
		// 2588
		// >>> print('{:x}'.format(ord(' ')))
		// 20
	}

	// if 1 == (len(lines) % 2) {
	// 	for c := 0; c < len(lines[len(lines)-1]); c += 2 {
	// 		out += "▀"
	// 	}
	// }
	return out
}
