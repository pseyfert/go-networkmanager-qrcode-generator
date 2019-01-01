package qrcode_for_nm_connection

import (
	qrcode "github.com/skip2/go-qrcode"
)

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
