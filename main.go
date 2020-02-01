package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/ip2location/ip2location-go"
	emoji "github.com/jayco/go-emoji-flag"
	"github.com/oschwald/geoip2-golang"
	"github.com/rdegges/go-ipify"
)

func main() {
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	systray.Run(do, onExit)
}

func do() {
	for {

		c2 := make(chan string, 1)
		go func() {
			ip, err := ipify.GetIp()
			if err == nil {
				gh := pingTime("https://api.github.com/zen")
				ccode := countryFromIP(ip)
				c2 <- fmt.Sprintf("%s %sms", emoji.GetFlag(ccode), gh)
			}
		}()
		select {
		case res := <-c2:
			systray.SetTitle(res)
		case <-time.After(2 * time.Second):
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
		panic(err)
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
	ip2location.Open("ipdb/IP2LOCATION-LITE-DB1.BIN")
	defer ip2location.Close()
	results := ip2location.Get_all(ip)
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
