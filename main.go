package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Priceru_feed struct {
	XMLName xml.Name `xml:"priceru_feed"`
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
	XMLName xml.Name `xml:"offer"`
	Id      string   `xml:"id,attr"`
	Picture []string `xml:"picture"`
}

func download(url string, catalog string) {
	fileName := catalog + "/" + url[strings.LastIndex(url, "/")+1:]
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(catalog, 0755)
			output, err := os.Create(fileName)
			defer output.Close()

			response, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer response.Body.Close()
			io.Copy(output, response.Body)
		}
	}
}

func main() {
	fileUrl := "https://loveyouhome.ua/index.php?route=extension/feed/unixml/priceru"
	DownloadFile("C://PriceLove/priceru.xml", fileUrl)
	fileUrl2 := "https://loveyouhome.ua/index.php?route=extension/feed/unixml/tomasby"
	DownloadFile("C://PriceLove/tomasby.xml", fileUrl2)

	xmlFile, err := os.Open("priceru.xml")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
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
	for i := 0; i < len(users.Shop[0].Offers[0].Offer); i++ {
		//fmt.Println("User Type: " + users.Shop[0].Offers[0].Offer[i].Id)
		catalog := "C:/PriceYUG/photo/" + users.Shop[0].Offers[0].Offer[i].Id
		for d := 0; d < len(users.Shop[0].Offers[0].Offer[i].Picture); d++ {
			//fmt.Println("User Type: " + users.Shop[0].Offers[0].Offer[i].Picture[d])
			download(users.Shop[0].Offers[0].Offer[i].Picture[d], catalog)
		}
		//fmt.Println("User Name: " + users.Shop[i].)
		//fmt.Println("Facebook Url: " + users.Offer[i].Social.Facebook)
	}
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
