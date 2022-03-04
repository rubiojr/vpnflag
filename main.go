//go:generate statik -src=ipdb
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/ip2location/ip2location-go/v9"
	emoji "github.com/jayco/go-emoji-flag"
	"github.com/oschwald/geoip2-golang"

	"github.com/atotto/clipboard"
	"github.com/rakyll/statik/fs"
	_ "github.com/rubiojr/vpnflag/statik" // TODO: Replace with the absolute import path
)

var configDir, dbPath string
var dbOpen = false
var ipMenu *systray.MenuItem
var currentIP string

const ipURL = "https://am.i.mullvad.net/ip"
const pingMeasureURL = "https://api.github.com/zen"

func main() {
	ipMenu = systray.AddMenuItem("Public IP", "Public IP address")
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	go func() {
		<-ipMenu.ClickedCh
		err := clipboard.WriteAll(currentIP)
		if err != nil {
			fmt.Printf("Error copying IP to clipboard: %s", err)
		}
	}()

	setupConfig()
	dbPath = path.Join(configDir, "ipdb")
	systray.Run(do, onExit)
}

func setupConfig() {
	cdir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configDir = path.Join(cdir, "vpnflag")
	_, err = os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(configDir, 0755)
		} else {
			log.Fatalf("Error creating config dir: %v", err)
		}
	}

}

func dbExists() bool {
	fi, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	if fi.IsDir() {
		log.Fatalf("DB cache %s found but not a regular file.", dbPath)
	}
	return true
}

func do() {
	for {
		c2 := make(chan string, 1)
		go func() {
			currentIP, err := getIP()
			if err == nil {
				ipMenu.SetTitle("Public IP: " + currentIP)
				gh := pingTime(pingMeasureURL)
				ccode := countryFromIP(currentIP)
				c2 <- fmt.Sprintf("%s %sms", emoji.GetFlag(ccode), gh)
			}
		}()
		select {
		case res := <-c2:
			systray.SetTitle(res)
		case <-time.After(10 * time.Second):
			fmt.Println("Getting the IP timed out.")
			systray.SetTitle("💀")
		}
		time.Sleep(5 * time.Second)
	}
}

func pingTime(url string) string {
	time_start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error testing network speed: %v", err)
		return "🔴"
	}
	defer resp.Body.Close()

	duration := time.Since(time_start)
	return fmt.Sprintf("%d", duration.Milliseconds())
}

func onExit() {
	os.Exit(0)
}

func countryFromIP(ip string) string {
	return ip2Loc(ip)
}

// ip2location.com provider
func ip2Loc(ip string) string {
	if !dbExists() {
		statikFS, err := fs.New()
		if err != nil {
			log.Fatal(err)
		}

		// Access individual files by their paths.
		r, err := statikFS.Open("/IP2LOCATION-LITE-DB1.BIN")
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()
		contents, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(dbPath, contents, 0644)
		if err != nil {
			log.Fatal("Error writing IP database cache.")
		}
	}
	if !dbOpen {
		ip2location.Open(dbPath)
	} else {
		dbOpen = true
	}
	results := ip2location.Get_country_short(ip)
	return results.Country_short
}

// Maxmind GeoIP provider, not currently used
func maxmindGeoIP(ipstr string) string {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ip := net.ParseIP(ipstr)
	record, err := db.City(ip)
	if err != nil {
		log.Fatal(err)
	}
	return record.Country.IsoCode
}

func getIP() (string, error) {
	resp, err := http.Get(ipURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ip)), nil
}
