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

func getIP(w http.ResponseWriter, r *http.Request, data *libgeo.GeoIP) {

	var ip string
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	clientInfo := ipInfo(ip, data)

	if clientInfo != nil {
		rankingsJson, _ := json.MarshalIndent(clientInfo, "", "  ")
		fmt.Fprintf(w, "%s \n\n", string(rankingsJson))
	} else {
		fmt.Fprintf(w, "%s \n\n", "IP not found!")
	}
}

func ipInfo(ipAddr string, data *libgeo.GeoIP) *ClientInfo {

	clientInfo := &ClientInfo{}
	clientInfo.Ip = ipAddr
	mutex.Lock()
	defer mutex.Unlock()

	loc := data.GetLocationByIP(ipAddr)
	if loc != nil {

		clientInfo.Country = loc.CountryName
		clientInfo.CountryCode = loc.CountryCode
		clientInfo.City = loc.City
		clientInfo.Region = loc.Region
		clientInfo.PostalCode = loc.PostalCode
		clientInfo.Latitude = loc.Latitude
		clientInfo.Longitude = loc.Longitude
		return clientInfo
	}
	return nil
}

func main() {

	dbFile := "GeoLiteCity.dat"
	data, _ := libgeo.Load(dbFile)

	m := martini.Classic()
	m.Map(data)

	m.Get("/", getIP)
	m.Run()
}
