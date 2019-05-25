# go-networkmanager-qrcode-generator

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![travis Status](https://travis-ci.org/pseyfert/go-networkmanager-qrcode-generator.svg?branch=master)](https://travis-ci.org/pseyfert/go-networkmanager-qrcode-generator)

golang based cli tools and golang packages to interact with NetworkManager for
the generation of QR codes that contain WiFi connection information.

## What's there

 - Run the tool `github.com/pseyfert/go-networkmanager-qrcode-generator` on a
   Linux computer where WiFi is managed through NetworkManager and generate a
   QR code for one of the saved connections. If all goes well, a QR code is
   generated. Scan that file with a device that does not have the connection
   information, but does have a QR code reader that understands network settings
   and add the exchanged network connection information to that device. The QR
   code can be thrown on the terminal or saved as png.
 - Run the tool `github.com/pseyfert/go-networkmanager-qrcode-generator/tui` on
   a Linux computer where WiFi is managed through NetworkManager and browse
   through network connections and generate a QR code on the terminal for them.

## What's missing (functionality)

 - Read QR code -> NetworkManager connection
 - Fixing lots of corner cases (hidden ssids, networks w/o password, handling of unsupported connections)
 - UI improvments

## What's missing (infrastructure)

 - tests
 - docs
 - review
