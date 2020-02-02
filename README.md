# VPNFlag

A little tool to display VPN country exit flag and "network speed".

Linux 🐧 only for now.

![/images/vpnflag.gif](images/vpnflag.gif)

## Building/installing

VPNFlag relies on https://github.com/getlantern/systray so some dependencies must be satisfied first. See their repository for instructions.

Run `go build` after satisfying deps.

## Notes

Measured network speed is based on the time it takes to make an HTTP GET request to https://api.github.com/zen. While not the best or most comprehensive way to measure "network speed", it's a good enough indication of the network performance when making simple HTTP requests, i.e. browsing the web.

## GeoIP database included

This tool includes IP2Location LITE data available from [https://lite.ip2location.com](https://lite.ip2location.com).
