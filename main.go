package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	URL    string `json:"url"`
	Ports  int64  `json:"ports"`
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
func main() {
	c := loadConfig("config.json")
	log.Println(c.APIKey)
}
