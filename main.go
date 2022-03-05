package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/ip2location/ip2location-go/v9"
	emoji "github.com/jayco/go-emoji-flag"

	_ "embed"

	"github.com/atotto/clipboard"
)

//go:embed ipdb/IP2LOCATION-LITE-DB1.BIN
var db []byte

var configDir, dbPath string
var dbOpen = false
var ipMenu *systray.MenuItem
var currentIP string

const pingFreq = 30

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
			fmt.Printf("failed copying IP to clipboard: %s", err)
		}
	}()

	setupConfig()
	systray.Run(do, onExit)
}

func setupConfig() {
	cdir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("error retrieving config dir: %v", err)
	}

	configDir = path.Join(cdir, "vpnflag")
	dbPath = path.Join(configDir, "ipdb")

	_, err = os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(configDir, 0755)
		} else {
			log.Fatalf("error creating config dir: %v", err)
		}
	}

	if !dbExists() {
		err := ioutil.WriteFile(dbPath, db, 0644)
		if err != nil {
			log.Fatalf("error writting ipdb: %v", err)
		}
	}

	ip2location.Open(dbPath)

}

func dbExists() bool {
	fi, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatalf("%s stat failed: %v", dbPath, err)
	}

	if fi.IsDir() {
		log.Printf("DB cache %s found but not a regular file\n", dbPath)
		err := os.Remove(dbPath)
		if err != nil {
			log.Fatalf("removing %s failed", dbPath)
		}
	}

	return true
}

func do() {
	for {
		currentIP, err := getIP()
		if err == nil {
			ipMenu.SetTitle("Public IP: " + currentIP)
			t, err := pingTime(pingMeasureURL)
			if err != nil {
				log.Printf("failed testing network speed: %v", err)
				systray.SetTitle("🔴")
			}
			ccode := ip2Loc(currentIP)
			res := fmt.Sprintf("%s %sms", emoji.GetFlag(ccode), t)
			systray.SetTitle(res)
		} else {
			systray.SetTitle("💀")
		}
		time.Sleep(pingFreq * time.Second)
	}
}

func pingTime(url string) (string, error) {
	time_start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	duration := time.Since(time_start)
	return fmt.Sprintf("%d", duration.Milliseconds()), nil
}

func onExit() {
	os.Exit(0)
}

// ip2location.com provider
func ip2Loc(ip string) string {
	results := ip2location.Get_country_short(ip)

	return results.Country_short
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
