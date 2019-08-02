package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

// The struct to store the config variables
type Config struct {
	URL    string `json:"url"`
	Ports  int64  `json:"ports"`
	APIKey string `json:"apiKey"`
}

// To store the response of the endpoint
type Response struct {
	Server     string `json:"server"`
	Port       int64  `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Direction  string `json:"direction"`
	Target     string `json:"target"`
	TargetPort int64  `json:"targetPort"`
	SourcePort int64  `json:"sourcePort"`
}

// This method loads reads the configuration file
func loadConfig(configFile string) Config {
	log.Printf("Loading config file %v\n", configFile)
	cnfg, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
	var c Config
	json.Unmarshal(cnfg, &c)
	return c
}

// This builds the url from the config file
func buildURL(c Config) string {
	baseURL, err := url.Parse(c.URL)
	if err != nil {
		log.Fatal("Malformed URL: ", err.Error())
	}

	baseURLWithAddr := net.JoinHostPort(baseURL.Hostname(), strconv.FormatInt(c.Ports, 10))

	// Parameters
	params := url.Values{}
	params.Add("apiKey", c.APIKey)

	var newQuery *url.URL
	newQuery, err = url.Parse(fmt.Sprintf("%v://%v", baseURL.Scheme, baseURLWithAddr))
	if err != nil {
		log.Fatal("Malformed URL: ", err.Error())
	}

	newQuery.RawQuery = params.Encode()
	return newQuery.String()
}

// Main function
func main() {
	c := loadConfig("config.json")

	url := buildURL(c)
	for {
		log.Printf("Processing %v\n", url)

		var getRestResponse []Response

		request := gorequest.New()
		_, body, _ := request.Get(url).EndStruct(&getRestResponse)
		log.Printf("Body:\n%v\n", body)
		time.Sleep(time.Minute * 1)
	}
}
