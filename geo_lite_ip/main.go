package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/nranchev/go-libGeoIP"
)

func getIP(w http.ResponseWriter, r *http.Request) {

	var ip string
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	ipInfo(ip, w)
}

func ipInfo(ipAddr string, w http.ResponseWriter) {

	var country, countryCode, city, region, postalCode string
	var latitude, longitude float32

	dbFile := "GeoLiteCity.dat"

	gi, err := libgeo.Load(dbFile)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	loc := gi.GetLocationByIP(ipAddr)
	if loc != nil {
		country = loc.CountryName
		countryCode = loc.CountryCode
		city = loc.City
		region = loc.Region
		postalCode = loc.PostalCode
		latitude = loc.Latitude
		longitude = loc.Longitude
		fmt.Fprintf(w, "IP: %s \n\n", ipAddr)

		fmt.Fprintf(w, "Country: %s \n", country)
		fmt.Fprintf(w, "Code:  %s \n", countryCode)
		fmt.Fprintf(w, "City:  %s \n", city)
		fmt.Fprintf(w, "Region:  %s \n", region)
		fmt.Fprintf(w, "Postal Code:  %s \n", postalCode)
		fmt.Fprintf(w, "Latitude:  %f \n", latitude)
		fmt.Fprintf(w, "Longitude:  %f \n", longitude)
	}
}

func main() {

	m := martini.Classic()

	m.Get("/", getIP)
	m.Run()
}
