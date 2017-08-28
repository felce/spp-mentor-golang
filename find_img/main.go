package main

import (
	"bufio"
	"fmt"
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
)

var numberOfRezizedFiles int

type AboutImg struct {
	fileName    string
	indexInPath string
	format      string
	resp        io.ReadCloser
}

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

func download(waitGroup *sync.WaitGroup, dirname string, chanelForIndex chan int, size string,
	chanelForFiles chan string, chanelForResizeFile chan AboutImg) {
	defer waitGroup.Done()
	for url := range chanelForFiles {

		var options AboutImg

		indexInFile := strconv.Itoa(<-chanelForIndex)

		response, err := http.Get(string(url))
		checkError(err)

		defer response.Body.Close()

		newFileName := parseFileName(url)
		var fileFormat string
		fileNameLen := len(newFileName)

		checkError(err)
		for i := fileNameLen - 1; i > 0; i-- {
			if byte(newFileName[i]) == '.' {
				fileFormat = newFileName[i:fileNameLen]
				break
			}
		}
		options.fileName = newFileName
		options.indexInPath = indexInFile
		options.format = fileFormat
		options.resp = response.Body

		fmt.Println("finish download", indexInFile)
		chanelForResizeFile <- options
	}
}

func downloadImg(urlsArray []string, dirname string, numberOfG, size string, chanelForResizeFile chan AboutImg, chanelForIndex chan int,
	chanelForFiles chan string) {

	numberOfGoroutines, err := strconv.Atoi(numberOfG)
	checkError(err)
	waitGroup := new(sync.WaitGroup)

	for i := 0; i < numberOfGoroutines; i++ {

		waitGroup.Add(1)

		go download(waitGroup, dirname, chanelForIndex, size, chanelForFiles, chanelForResizeFile)

	}

	for index, url := range urlsArray {
		chanelForFiles <- url
		chanelForIndex <- index
	}

	defer close(chanelForFiles)
	defer close(chanelForIndex)

	waitGroup.Wait()

}

func resizeImg(size, dir string, chanelForResizeFile chan AboutImg, waitGroup *sync.WaitGroup, numberOfFiles int) {
	defer waitGroup.Done()

	for options := range chanelForResizeFile {

		sizeOfNewImg, err := strconv.Atoi(size)
		checkError(err)
		var zIndex uint = uint(sizeOfNewImg)

		fileName := options.fileName
		fileFormat := options.format
		resp := options.resp
		indexInFile := options.indexInPath

		fmt.Println("start resize ", indexInFile)
		var img image.Image
		if fileFormat == ".jpg" || fileFormat == ".jpeg" {
			img, err = jpeg.Decode(resp)
		} else if fileFormat == ".png" {
			img, err = png.Decode(resp)
		}
		resizeThisFormat(err, zIndex, img, dir, fileName, fileFormat, resp, indexInFile, chanelForResizeFile, numberOfFiles)
	}
}

func resizeThisFormat(err error, zIndex uint, img image.Image, dir, fileName, fileFormat string,
	file io.ReadCloser, indexInFile string, chanelForResizeFile chan AboutImg, numberOfFiles int) {

	if err == nil {

		height := float64(img.Bounds().Size().Y)
		width := float64(img.Bounds().Size().X)
		newWidth := int(zIndex)
		newHeight := int((height / width) * float64(zIndex))

		xscale := float64(width) / float64(newWidth)
		yscale := float64(height) / float64(newHeight)

		dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

		for x := 0; x < newWidth; x++ {
			for y := 0; y < newHeight; y++ {
				srcx := int(float64(x) * xscale)
				srcy := int(float64(y) * yscale)
				dst.Set(x, y, img.At(srcx, srcy))
			}
		}

		out, err := os.Create(dir + "/" + indexInFile + "_" + fileName)
		checkError(err)

		defer out.Close()
		if fileFormat == ".jpg" || fileFormat == ".jpeg" {
			jpeg.Encode(out, dst, nil)
		} else if fileFormat == ".png" {
			png.Encode(out, dst)

		}
	} else {

		empty_file, _ := os.Create(dir + "/" + indexInFile + "_" + fileName)
		io.Copy(empty_file, file)
		empty_file.Close()
	}
	numberOfRezizedFiles++
	defer func() {
		if numberOfRezizedFiles >= numberOfFiles {
			os.Exit(0)
		}
	}()
}

func startResize(chanelForResizeFile chan AboutImg, size, dirname string, numberOfGoroutinesForResize, numberOfFiles int) {

	waitGroup := new(sync.WaitGroup)

	for i := 0; i < numberOfGoroutinesForResize; i++ {

		waitGroup.Add(1)
		go resizeImg(size, dirname, chanelForResizeFile, waitGroup, numberOfFiles)

	}

	waitGroup.Wait()
}

func main() {

	numberOfRezizedFiles = 0
	fileWithUrls := os.Args[1]
	dirname := os.Args[2]
	size := os.Args[3]
	numberOfGoroutinesForDownload := os.Args[4]
	numberOfGoroutinesForResize, err := strconv.Atoi(os.Args[5])
	checkError(err)

	os.MkdirAll(dirname, os.ModePerm)

	urlsArray := initArrayOfUrl(fileWithUrls)
	chanelForIndex := make(chan int)

	chanelForFiles := make(chan string)
	chanelForResizeFile := make(chan AboutImg)
	numberOfFiles := len(urlsArray)

	go downloadImg(urlsArray, dirname, numberOfGoroutinesForDownload, size, chanelForResizeFile, chanelForIndex, chanelForFiles)

	startResize(chanelForResizeFile, size, dirname, numberOfGoroutinesForResize, numberOfFiles)
}
