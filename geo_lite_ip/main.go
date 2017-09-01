package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	"github.com/nranchev/go-libGeoIP"
)

func getIP(w http.ResponseWriter, r *http.Request, wg *sync.WaitGroup, gi *libgeo.GeoIP) {
	defer wg.Done()
	var ip string
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	ipInfo(ip, w, gi)
}

func ipInfo(ipAddr string, w http.ResponseWriter, gi *libgeo.GeoIP) {

	var country, countryCode, city, region, postalCode string
	var latitude, longitude float32

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

	dbFile := "GeoLiteCity.dat"

	gi, _ := libgeo.Load(dbFile)

	m := martini.Classic()
	m.Map(gi)
	m.Get("/", getIP)
	m.Run()

}
