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
	XMLName xml.Name `xml:"yml_catalog"`
	Shop    []Shop   `xml:"shop"`
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
	//fileUrl := "https://loveyouhome.ua/index.php?route=extension/feed/unixml/priceru"
	//DownloadFile("C://PriceLove/priceru.xml", fileUrl)
	fileUrl2 := "https://loveyouhome.ua/index.php?route=extension/feed/unixml/tomasby"
	DownloadPrice("./temp/tomasby.xml", fileUrl2)

	ParseXml("./temp/tomasby.xml")
	InfoLogger.Println("tomasby.xml", "Done")
	duration := time.Since(start)
	fmt.Println("Время выполнения: ", duration)

}

func ParseXml(flieUrl string) {

	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	wg := new(sync.WaitGroup)
	results := make(map[string]string)
	structCh := make(chan struct{})

	xmlFile, err := os.Open(flieUrl)
	// if we os.Open returns an error then handle it
	if err != nil {
		ErrorLogger.Println(err)
	}

	fmt.Println("Successfully Opened users.xml")
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

	go func(structCh chan struct{}) {
		defer close(structCh)
		//for i := 0; i < len(users.Shop[0].Offers[0].Offer); i++ {
		for _, offer := range users.Shop[0].Offers[0].Offer {
			catalog := "./temp/" + offer.VendorCode + "_L"
			//for d := 0; d < len(users.Shop[0].Offers[0].Offer[i].Picture); d++ {
			for _, picture := range offer.Picture {
				InfoLogger.Println(offer.VendorCode + "_L")
				results[picture] = catalog
			}
		}
	}(structCh)
	<-structCh
	duration := time.Since(start)
	fmt.Println("Время выполнения 1 горутины: ", duration)
	for picture, catalog := range results {
		wg.Add(1)
		go DownloadPhoto(picture, catalog, tick.C, wg)
	}
	wg.Wait()

}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadPrice(filepath string, url string) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		ErrorLogger.Println(err)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
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
