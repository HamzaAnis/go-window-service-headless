package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
)

type Config struct {
	URL    string `json:"url"`
	Ports  string `json:"ports"`
	APIKey string `json:"apiKey"`
}

func loadConfig(configFile string) Config {
	cnfg, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
	var c Config
	json.Unmarshal(cnfg, &c)
	return c
}

func buildURL(c Config) string {
	baseUrl, err := url.Parse(c.URL)
	if err != nil {
		log.Fatal("Malformed URL: ", err.Error())
	}
	baseUrlWithAddr := net.JoinHostPort(baseUrl.Hostname(), "8000")

	// Parameters
	params := url.Values{}
	params.Add("apiKey", c.APIKey)
	if baseUrl.Scheme != "" {
		log.Println("There is scheme")
	} else {
		log.Println("There is no scheme")
	}
	var newQuery *url.URL
	newQuery, err = url.Parse(fmt.Sprintf("%v://%v", baseUrl.Scheme, baseUrlWithAddr))
	if err != nil {
		log.Fatal("1 Malformed URL: ", err.Error())
	}

	newQuery.RawQuery = params.Encode()
	return newQuery.String()
}
func main() {
	c := loadConfig("config.json")

	url := buildURL(c)
	log.Println(url)
}
