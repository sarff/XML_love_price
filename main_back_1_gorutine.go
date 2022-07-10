package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

type Priceru_feed struct {
	XMLName xml.Name
	Shop    []Shop `xml:"shop"`
}

type Shop struct {
	XMLName xml.Name `xml:"shop"`
	Offers  []Offers `xml:"offers"`
}

type Offers struct {
	XMLName xml.Name `xml:"offers"`
	Offer   []Offer  `xml:"offer"`
}

type Offer struct {
	XMLName    xml.Name `xml:"offer"`
	Id         string   `xml:"id,attr"`
	VendorCode string   `xml:"vendorCode"`
	Picture    []string `xml:"picture"`
}

type xmlname struct {
	name         string
	fileUrl      string
	path         string
	catalogPhoto string
}

func init() {
	//InfoLogger.Println("Starting the application...")
	//InfoLogger.Println("Something noteworthy happened")
	//WarningLogger.Println("There is something you should know about")
	//ErrorLogger.Println("Something went wrong")

	file, err := os.OpenFile("logs2.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	start := time.Now()

	first := xmlname{
		name:         "priceru",
		fileUrl:      "https://loveyouhome.ua/index.php?route=extension/feed/unixml/priceru",
		path:         "./temp/priceru.xml",
		catalogPhoto: "./temp/",
	}
	second := xmlname{
		name:         "tomasby",
		fileUrl:      "https://loveyouhome.ua/index.php?route=extension/feed/unixml/tomasby",
		path:         "./temp2/tomasby.xml",
		catalogPhoto: "./temp2/",
	}

	//first.DownloadPrice()
	//second.DownloadPrice()

	//first.ParseXml()
	runParse(first)
	InfoLogger.Println(first.name, "Done")

	//second.ParseXml()
	runParse(second)
	InfoLogger.Println(second.name, "Done")

	duration := time.Since(start)
	fmt.Println("Время выполнения: ", duration)

}

func runParse(p xmlname) {
	p.DownloadPrice()
	p.ParseXml()
}

func (x *xmlname) ParseXml() {
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	wg := new(sync.WaitGroup)
	xmlFile, err := os.Open(x.path)
	// if we os.Open returns an error then handle it
	if err != nil {
		ErrorLogger.Println(err)
	}

	fmt.Printf("Successfully Opened %s", x.path)
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our Users array
	var users Priceru_feed
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &users)

	// we iterate through every user within our users array and
	// print out the user Type, their name, and their facebook url
	// as just an example
	start := time.Now()
	for _, offer := range users.Shop[0].Offers[0].Offer {
		catalog := x.catalogPhoto + offer.VendorCode + "_L"

		for _, picture := range offer.Picture {
			wg.Add(1)
			InfoLogger.Println(offer.VendorCode + "_L")
			go DownloadPhoto(picture, catalog, tick.C, wg)
		}
		wg.Wait()
	}
	duration := time.Since(start)
	fmt.Println("Время выполнения цикла: ", duration)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (x *xmlname) DownloadPrice() {

	// Get the data
	resp, err := http.Get(x.fileUrl)
	if err != nil {
		ErrorLogger.Println(err)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(x.path)
	if err != nil {
		ErrorLogger.Println(err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		ErrorLogger.Println(err)
	}
}

func DownloadPhoto(url string, catalog string, limit <-chan time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	<-limit
	fileName := catalog + "/" + url[strings.LastIndex(url, "/")+1:]
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(catalog, 0755)
			output, err := os.Create(fileName)
			defer output.Close()

			response, err := http.Get(url)
			if err != nil {
				ErrorLogger.Println(err)
				return
			}
			defer response.Body.Close()
			io.Copy(output, response.Body)
		}
	}
}
