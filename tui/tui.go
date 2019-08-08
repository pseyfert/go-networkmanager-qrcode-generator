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
	"fmt"
	"log"
	"sort"
	"strings"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"github.com/godbus/dbus"
	nm2qr "github.com/pseyfert/go-networkmanager-qrcode-generator/qrcode_for_nm_connection"
	ux "github.com/pseyfert/go-networkmanager-qrcode-generator/ux"
	"github.com/skip2/go-qrcode"
)

func dbusmap(cons []nm2qr.NetworkSetting) map[int]nm2qr.NetworkSetting {
	retval := make(map[int]nm2qr.NetworkSetting)
	for _, con := range cons {
		retval[con.DbusId] = con
	}
	return retval
}

func sortedids(cons map[int]nm2qr.NetworkSetting) []int {
	keys := make([]int, 0, len(cons))
	for _, con := range cons {
		if con.IsPsk {
			keys = append(keys, con.DbusId)
		}
	}
	sort.Ints(keys)
	return keys
}

func main() {
	dbusConnection, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("couldn't connect to system dbus: %v", err)
	}
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	cons, err := ux.AllConnections(dbusConnection)
	if err != nil {
		log.Fatalf("couldn't obtain connections: %v", err)
	}
	conmap := dbusmap(cons)
	sortedkeys := sortedids(conmap)

	networklist := widgets.NewList()
	networklist.Title = "known connections"
	networklist.Rows = make([]string, 0, len(cons))

	for _, id := range sortedkeys {
		s := fmt.Sprintf("[%d] %s (%s)", id, conmap[id].Id, conmap[id].Ssid)
		networklist.Rows = append(networklist.Rows, s)
	}
	networklist.TextStyle = ui.NewStyle(ui.ColorCyan)
	networklist.BorderStyle = ui.NewStyle(ui.ColorBlack, ui.ColorWhite)
	networklist.TitleStyle = ui.NewStyle(ui.ColorBlack, ui.ColorWhite)
	networklist.SelectedRowStyle = ui.NewStyle(ui.ColorMagenta)
	networklist.WrapText = false
	networklist.SetRect(0, 0, 30, 35)

	code := widgets.NewTable()
	code.SetRect(30, 0, 90, 35)
	code.Title = "QR code"
	code.RowSeparator = false
	code.Rows = [][]string{[]string{"QR code to appear here"},
		[]string{"USAGE:"},
		[]string{"q, ^C, ESC: quit"},
		[]string{"RET:        generate code"},
		[]string{"s:          save code as png (/tmp/nm2qr_<name>.png)"},
	}
	code.TextStyle = ui.NewStyle(ui.ColorBlue)

	ui.Render(networklist, code)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		getqr := func() (qrcode.QRCode, string) {
			sel := networklist.SelectedRow
			id := sortedkeys[sel]
			qr, err := nm2qr.QRNetworkCode(conmap[id])
			if nil != err {
				log.Fatalf("something went wrong in qr code generation, %v", err)
			}
			return qr, conmap[id].Id
		}
		switch e.ID {
		case "q", "<C-c>", "<Escape>":
			return
		case "j", "<Down>":
			networklist.ScrollDown()
		case "k", "<Up>":
			networklist.ScrollUp()
		case "<C-d>":
			networklist.ScrollHalfPageDown()
		case "<C-u>":
			networklist.ScrollHalfPageUp()
		case "<C-f>":
			networklist.ScrollPageDown()
		case "<C-b>":
			networklist.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				networklist.ScrollTop()
			}
		case "<Home>":
			networklist.ScrollTop()
		case "G", "<End>":
			networklist.ScrollBottom()
		case "<Enter>":
			{
				qr, title := getqr()
				qrcode := qr.ToSmallString(false)
				rows := strings.Split(qrcode, "\n")
				rrows := make([][]string, 0, len(rows))
				for _, row := range rows {
					rrows = append(rrows, []string{row})
				}

				code.Rows = rrows
				code.Title = title
				code.TextStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack)
			}
		case "s":
			{
				qr, title := getqr()
				fname := fmt.Sprintf("/tmp/nm2qr_%s.png", title)
				qr.WriteFile(-5, fname)

				code.Rows = [][]string{[]string{fmt.Sprintf("saved as %s", fname)}}
				code.TextStyle = ui.NewStyle(ui.ColorBlue)
			}
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(networklist, code)
	}

}
