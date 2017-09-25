package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-martini/martini"
	"github.com/nranchev/go-libGeoIP"
	"github.com/sirupsen/logrus"
)

var (
	FileWithLogs       *os.File
	CurrentDate        string
	NumberOfFiles      int
	NewNumberOfFile    string
	currentLogFileName string
	mutex              sync.Mutex
)

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
	QueryString map[string][]string
	Info        *ClientInfo
	Error       *ErrorInfo
}

func updCurrentDate() {

	CurrentDate = time.Now().Format("02_01_2006")
}

func updNewNumberOfFile() {

	NewNumberOfFile = fmt.Sprintf("%04s", strconv.Itoa(NumberOfFiles))
}

func openLogFile(fileName string) {

	FileWithLogs, _ = os.OpenFile("log/"+fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
}

func createAndOpenNewFile() {

	FileWithLogs.Close()
	NumberOfFiles++
	updNewNumberOfFile()
	fileName := NewNumberOfFile + "_" + CurrentDate + ".log"
	openLogFile(fileName)

}

func getNameOfLOgsFileAfterServerStart() string {

	files, err := ioutil.ReadDir("log/")
	if err != nil {

		os.MkdirAll("log", os.ModePerm)
	}

	NumberOfFiles = len(files)

	if NumberOfFiles == 0 {

		return "0001_" + CurrentDate + ".log"
	}
	lastFileDate := files[NumberOfFiles-1].Name()[5:15]

	if CurrentDate != lastFileDate {

		NumberOfFiles++
		updNewNumberOfFile()
		return NewNumberOfFile + "_" + CurrentDate + ".log"
	}

	return files[NumberOfFiles-1].Name()
}

func getIP(w http.ResponseWriter, r *http.Request, data *libgeo.GeoIP, log *logrus.Logger) {

	updCurrentDate()
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
	clientInfo, errInfo := ipInfo(ip, data)

	if clientInfo != nil {

		w.WriteHeader(http.StatusOK)
		infoJson, _ = json.MarshalIndent(clientInfo, "", "\t")
	} else {

		w.WriteHeader(http.StatusNotFound)
		infoJson, _ = json.MarshalIndent(errInfo, "", "\t")
	}
	w.Write(infoJson)

	requestBody := requestInfo(r)
	logInfo := &LogInfo{Req: requestBody, QueryString: qs, Info: clientInfo, Error: errInfo}

	writeLogToFile(logInfo, log)
}

func requestInfo(r *http.Request) string {

	var request []string
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	return strings.Join(request, "\n")
}

func writeLogToFile(logInfo *LogInfo, log *logrus.Logger) {

	log.Formatter = new(logrus.JSONFormatter)
	log.Out = FileWithLogs

	log.WithFields(logrus.Fields{
		"REQUEST_BODY": logInfo.Req,
		"QUERY_STRING": logInfo.QueryString,
		"RESPONSE":     logInfo.Info,
		"ERROR":        logInfo.Error,
	}).Info()
}

func ipInfo(ipAddr string, data *libgeo.GeoIP) (*ClientInfo, *ErrorInfo) {

	mutex.Lock()
	defer mutex.Unlock()

	loc := data.GetLocationByIP(ipAddr)
	if loc == nil {
		errInfo := &ErrorInfo{Ip: ipAddr, Error: "geo info for ip not found"}
		return nil, errInfo
	}
	clientInfo := &ClientInfo{Ip: ipAddr, Country: loc.CountryName, CountryCode: loc.CountryCode,
		City: loc.City, Region: loc.Region, PostalCode: loc.PostalCode,
		Latitude: loc.Latitude, Longitude: loc.Longitude}

	return clientInfo, nil
}

func main() {
	go func() {

		dateNow := time.Now().Format("02_01_2006")
		for {
			newDate := time.Now().Format("02_01_2006")

			if dateNow != newDate {

				dateNow = newDate
				updCurrentDate()
				createAndOpenNewFile()
			}
			time.Sleep(time.Minute)
		}
	}()

	updCurrentDate()
	currentLogFileName = getNameOfLOgsFileAfterServerStart()

	dbFile := "GeoLiteCity.dat"
	data, err := libgeo.Load(dbFile)
	var log = logrus.New()
	if err != nil {
		log.Fatal(err)
	}
	openLogFile(currentLogFileName)

	m := martini.Classic()
	m.Map(data)
	m.Map(log)

	m.Get("/", getIP)
	m.Run()
}
