package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	"github.com/nranchev/go-libGeoIP"
)

type ClientInfo struct {
	Ip          string
	Country     string
	CountryCode string
	City        string
	Region      string
	PostalCode  string
	Latitude    float32
	Longitude   float32
}

func getIP(w http.ResponseWriter, r *http.Request, data *libgeo.GeoIP, mutex sync.Mutex) {
	var ip string
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	info := ClientInfo{}

	info.ipInfo(ip, w, data, mutex)
}

func (clientInfo *ClientInfo) ipInfo(ipAddr string, w http.ResponseWriter, data *libgeo.GeoIP, mutex sync.Mutex) {

	var country, countryCode, city, region, postalCode string
	var latitude, longitude float32

	mutex.Lock()

	loc := data.GetLocationByIP(ipAddr)
	if loc != nil {
		func() {
			defer mutex.Unlock()
			country = loc.CountryName
			countryCode = loc.CountryCode
			city = loc.City
			region = loc.Region
			postalCode = loc.PostalCode
			latitude = loc.Latitude
			longitude = loc.Longitude
		}()

		clientInfo.Ip = ipAddr
		clientInfo.Country = country
		clientInfo.CountryCode = countryCode
		clientInfo.City = city
		clientInfo.Region = region
		clientInfo.PostalCode = postalCode
		clientInfo.Latitude = latitude
		clientInfo.Longitude = longitude

		rankingsJson, _ := json.MarshalIndent(clientInfo, "", "  ")

		fmt.Fprintf(w, "%s \n\n", string(rankingsJson))
	}
}

func main() {

	var mutex sync.Mutex
	dbFile := "GeoLiteCity.dat"
	data, _ := libgeo.Load(dbFile)

	m := martini.Classic()
	m.Map(data)
	m.Map(mutex)

	m.Get("/", getIP)
	m.Run()
}
