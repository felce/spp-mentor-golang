package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-martini/martini"
	"github.com/nranchev/go-libGeoIP"
	// log "github.com/sirupsen/logrus"
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

type LogInfo struct {
	Req         string
	QueryString string
	Resp        string
}

func checkError(err error) {

	if err != nil {
		log.Fatal(err)
	}
}

func getIP(w http.ResponseWriter, r *http.Request, data *libgeo.GeoIP) {

	w.Header().Set("Content-Type", "application/json")

	qs := r.URL.Query()
	var infoJson []byte
	var ip string

	ip = qs.Get("ip")
	if ip == "" {
		if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
			ip = ipProxy
		} else {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		}
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

	requestBody := requestInfo(r)
	q, _ := json.MarshalIndent(qs, "", "\t")
	logInfo := &LogInfo{Req: requestBody, QueryString: string(q[:]), Resp: string(infoJson[:])}

	logginToFile(logInfo)
}

func requestInfo(r *http.Request) string {

	requestDump, _ := httputil.DumpRequest(r, true)

	return string(requestDump)
}

func dailyLogFile() string {

	var lastFileDate string

	currentDate := time.Now().Format("02_01_2006")

	files, err := ioutil.ReadDir("log/")
	if err != nil {
		log.Fatal(err)
	}

	n := len(files)

	if n == 0 {
		return "0001_" + currentDate + ".log"
	}

	lastFileDate = files[n-1].Name()[5:15]

	if currentDate != lastFileDate {
		newNumb := fmt.Sprintf("%04s", strconv.Itoa(n+1))
		return newNumb + "_" + currentDate + ".log"
	}

	return files[n-1].Name()
}

func logginToFile(logInfo *LogInfo) {

	logPath := dailyLogFile()

	lf, err := os.OpenFile("log/"+logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

	if err != nil {
		log.Fatal("OpenLogfile: os.OpenFile:", err)
	}
	defer lf.Close()

	log.SetOutput(lf)
	log.Println("\nrequest body:\n", logInfo.Req, "\nquery string:\n", logInfo.QueryString, "\nresponse:\n", logInfo.Resp, "\n")
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
