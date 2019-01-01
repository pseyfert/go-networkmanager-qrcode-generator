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
