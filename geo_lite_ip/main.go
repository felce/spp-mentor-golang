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

var mutex sync.Mutex

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

func New() *ClientInfo {
	return &ClientInfo{}
}

func getIP(w http.ResponseWriter, r *http.Request, data *libgeo.GeoIP) {
	var ip string
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	info := New()

	info.ipInfo(ip, w, data)
}

func (clientInfo *ClientInfo) ipInfo(ipAddr string, w http.ResponseWriter, data *libgeo.GeoIP) {

	var country, countryCode, city, region, postalCode string
	var latitude, longitude float32

	mutex.Lock()
	defer mutex.Unlock()

	loc := data.GetLocationByIP(ipAddr)
	if loc != nil {
		country = loc.CountryName
		countryCode = loc.CountryCode
		city = loc.City
		region = loc.Region
		postalCode = loc.PostalCode
		latitude = loc.Latitude
		longitude = loc.Longitude

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

	dbFile := "GeoLiteCity.dat"

	data, _ := libgeo.Load(dbFile)

	m := martini.Classic()
	m.Map(data)

	m.Get("/", getIP)
	m.Run()
}
