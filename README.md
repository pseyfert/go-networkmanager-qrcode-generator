# go-networkmanager-qrcode-generator

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![travis Status](https://travis-ci.org/pseyfert/go-networkmanager-qrcode-generator.svg?branch=master)](https://travis-ci.org/pseyfert/go-networkmanager-qrcode-generator)

golang based cli tools and golang packages to interact with NetworkManager for
the generation of QR codes that contain WiFi connection information.

# What's there

 - Run the tool on a Linux computer where WiFi is managed through
   NetworkManager and generate a QR code for one of the saved connections. If
   all goes well, a QR code is generated as PNG file. Scan that file with a device
   that does not have the connection information, but does have a QR code reader
   that understands network settings and add the exchanged network connection
   information to that device.

# What's missing

 - Read QR code -> NetworkManager connection
 - Fixing lots of corner cases (hidden ssids, networks w/o password, handling of unsupported connections)
 - UI for selection of connection
