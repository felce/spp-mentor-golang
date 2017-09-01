package main

import (
	"flag"
	"fmt"

	"github.com/nranchev/go-libGeoIP"
)

// ./geo_lite_ip  GeoLiteCity.dat 93.35.186.197

func main() {
	var country, countryCode, city, region, postalCode string
	var latitude, longitude float32
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("usage: main DBFILE IPADDRESS\n")
		return
	}

	dbFile := flag.Arg(0)
	ipAddr := flag.Arg(1)

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

		fmt.Println("Country: ", country)
		fmt.Println("Code: ", countryCode)
		fmt.Println("City: ", city)
		fmt.Println("Region: ", region)
		fmt.Println("Postal Code: ", postalCode)
		fmt.Println("Latitude: ", latitude)
		fmt.Println("Longitude: ", longitude)
	}
}

//  GeoLiteCity.dat 81.2.69.142
