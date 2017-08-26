package main

import (
	"bufio"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/nfnt/resize"
)

func checkError(err error) {

	if err != nil {
		log.Fatal(err)
	}
}

func initArrayOfUrl(fileName string) []string {

	var urlsArray []string
	file, err := os.Open(fileName)
	checkError(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		url := scanner.Text()
		urlsArray = append(urlsArray, url)
	}

	return urlsArray
}

func parseFileName(url string) string {

	lengthUrl := len(url)
	slashIndex := 0
	for i := lengthUrl - 1; i > 0; i-- {
		if byte(url[i]) == '/' {
			slashIndex = i
			break
		}
	}

	imgName := url[slashIndex+1 : lengthUrl]

	for i := range imgName {
		if !unicode.IsLetter(rune(imgName[i])) && imgName[i] != '.' {
			imgName = strings.Replace(imgName, string(imgName[i]), "_", -1)
		}
		if imgName[i] == '.' {
			strings.ToLower(imgName[i+1:])
			break
		}
	}
	return imgName
}

func download(chanelForUrls chan string, waitGroup *sync.WaitGroup,
	dirname string, chanelForIndex chan int, size, numberOfGoroutinesForResize string) {

	defer waitGroup.Done()

	for url := range chanelForUrls {

		indexInFile := strconv.Itoa(<-chanelForIndex)

		if url != "" {
			response, err := http.Get(string(url))
			checkError(err)

			defer response.Body.Close()

			newFileName := parseFileName(url)
			var fileFormat string
			fileNameLen := len(newFileName)

			sizeOfNewImg, err := strconv.Atoi(size)
			checkError(err)

			var zIndex uint = uint(sizeOfNewImg)

			for i := fileNameLen - 1; i > 0; i-- {
				if byte(newFileName[i]) == '.' {
					fileFormat = newFileName[i:fileNameLen]
					break
				}
			}

			if fileFormat == ".jpg" || fileFormat == ".jpeg" {
				img, err := jpeg.Decode(response.Body)
				resizeThisFormat(err, zIndex, img, dirname, newFileName, fileFormat, response.Body, indexInFile)
			} else if fileFormat == ".png" {
				img, err := png.Decode(response.Body)
				resizeThisFormat(err, zIndex, img, dirname, newFileName, fileFormat, response.Body, indexInFile)
			}
		}

	}

}

func downloadImg(urlsArray []string, dirname string, numberOfG, size, numberOfGoroutinesForResize string) {

	numberOfGoroutines, err := strconv.Atoi(numberOfG)
	checkError(err)

	chanelForUrls := make(chan string)
	chanelForIndex := make(chan int)
	waitGroup := new(sync.WaitGroup)

	for i := 0; i < numberOfGoroutines; i++ {
		waitGroup.Add(1)
		go download(chanelForUrls, waitGroup, dirname, chanelForIndex, size, numberOfGoroutinesForResize)
	}

	for index, i := range urlsArray {
		chanelForUrls <- i
		chanelForIndex <- index
	}

	close(chanelForUrls)
	close(chanelForIndex)
	waitGroup.Wait()
}

func resizeThisFormat(err error, zIndex uint, img image.Image, dir, fileName, fileFormat string, file io.ReadCloser, indexInFile string) {

	if err == nil {
		m := resize.Resize(zIndex, 0, img, resize.Lanczos3)
		file.Close()
		os.Remove(dir + "/" + fileName)

		out, err := os.Create(dir + "/" + indexInFile + "_" + fileName)
		checkError(err)

		defer out.Close()
		if fileFormat == ".jpg" || fileFormat == ".jpeg" {
			jpeg.Encode(out, m, nil)
		} else if fileFormat == ".png" {
			png.Encode(out, m)
		}
	} else {
		file_blanc, _ := os.Create(dir + "/" + indexInFile + "_" + fileName)
		io.Copy(file_blanc, file)
		file.Close()
	}
}

func main() {

	fileWithUrls := os.Args[1]
	dirname := os.Args[2]
	size := os.Args[3]
	numberOfGoroutinesForDownload := os.Args[4]
	numberOfGoroutinesForResize := os.Args[5]

	os.MkdirAll(dirname, os.ModePerm)

	urlsArray := initArrayOfUrl(fileWithUrls)
	downloadImg(urlsArray, dirname, numberOfGoroutinesForDownload, size, numberOfGoroutinesForResize)
}
