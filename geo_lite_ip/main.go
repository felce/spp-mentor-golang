package main

import (
	"encoding/json"
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

	w.Header().Set("Content-Type", "application/json")

	var ip string
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	clientInfo, statusCode := ipInfo(ip, data)

	if statusCode == 200 {

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("completed 200 - ok \n"))
		infoJson, _ := json.MarshalIndent(clientInfo, "", "\t")
		w.Write(infoJson)
	} else {

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("completed 404 - ip not found\n"))
		infoJson, _ := json.MarshalIndent(clientInfo, "", "\t")
		w.Write(infoJson)
	}
}

func ipInfo(ipAddr string, data *libgeo.GeoIP) (*ClientInfo, int) {

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
		return clientInfo, 200
	}
	return clientInfo, 404
}

func main() {

	dbFile := "GeoLiteCity.dat"
	data, _ := libgeo.Load(dbFile)

	m := martini.Classic()
	m.Map(data)

	m.Get("/", getIP)
	m.Run()
}
