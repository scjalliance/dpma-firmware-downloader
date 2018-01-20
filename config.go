package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config holds configuration data for the DPMA firmware downloader
type Config struct {
	DataDir  string `json:"datadir"`
	Manifest string `json:"manifest"`
	Models   string `json:"models"`
}

// DefaultConfig holds the default configuration settings.
var DefaultConfig = Config{
	DataDir:  "/var/lib/asterisk/digium_phones/firmware",
	Manifest: "https://downloads.digium.com/pub/telephony/res_digium_phone/firmware/dpma-firmware.json",
	Models:   "*",
}

func buildConfig() (config Config) {
	// Load configuration settings.
	data, err := ioutil.ReadFile("/etc/config.json")

	switch {
	case os.IsNotExist(err):
		log.Println("No configuration file found. Using defaults.")
		config = DefaultConfig
	case err == nil:
		if err = json.Unmarshal(data, &config); err != nil {
			log.Fatal(err)
		}
		log.Println("Loaded configuration from config.json")
	default:
		log.Println(err)
	}

	var overload []string

	if val := os.Getenv("MANIFEST"); val != "" {
		overload = append(overload, "MANIFEST")
		config.Manifest = val
	}

	if val := os.Getenv("DATA_DIR"); val != "" {
		overload = append(overload, "DATA_DIR")
		config.DataDir = val
	}

	if val := os.Getenv("MODELS"); val != "" {
		overload = append(overload, "MODELS")
		config.Models = val
	}

	if len(overload) > 0 {
		log.Printf("Overriding configuration from environment variables: %v", overload)
	} else {
		log.Println("No environment variables provided for configuration.")
	}

	return config
}
