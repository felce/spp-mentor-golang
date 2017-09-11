package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	"github.com/nranchev/go-libGeoIP"
)

var mutex sync.Mutex

type ErrorInfo struct {
	Ip    string
	Error string
}

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

func checkError(err error) {

	if err != nil {
		log.Fatal(err)
	}
}

func getIP(w http.ResponseWriter, r *http.Request, data *libgeo.GeoIP) {

	w.Header().Set("Content-Type", "application/json")

	var ip string
	var infoJson []byte
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	clientInfo := ipInfo(ip, data)

	if clientInfo != nil {

		w.WriteHeader(http.StatusOK)
		infoJson, _ = json.MarshalIndent(clientInfo, "", "\t")
	} else {

		w.WriteHeader(http.StatusNotFound)
		errInfo := &ErrorInfo{Ip: ip, Error: "geo info for ip not found"}
		infoJson, _ = json.MarshalIndent(errInfo, "", "\t")
	}
	w.Write(infoJson)
}

func ipInfo(ipAddr string, data *libgeo.GeoIP) *ClientInfo {

	mutex.Lock()
	defer mutex.Unlock()

	loc := data.GetLocationByIP(ipAddr)
	if loc == nil {
		return nil
	}
	clientInfo := &ClientInfo{Ip: ipAddr, Country: loc.CountryName, CountryCode: loc.CountryCode,
		City: loc.City, Region: loc.Region, PostalCode: loc.PostalCode,
		Latitude: loc.Latitude, Longitude: loc.Longitude}

	return clientInfo
}

func main() {

	dbFile := "GeoLiteCity.dat"
	data, err := libgeo.Load(dbFile)
	checkError(err)

	m := martini.Classic()
	m.Map(data)

	m.Get("/", getIP)
	m.Run()
}
