package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/scjalliance/dpmafirmware"
)

// Config holds configuration data for the DPMA firmware downloader
type Config struct {
	FirmwareDir string `json:"firmwaredir"`
	Manifest    string `json:"manifest"`
	Models      string `json:"models"`
}

// CacheFile returns the path to the MD5 cache file for the given version.
func (c *Config) CacheFile(v dpmafirmware.Version) string {
	return filepath.Join(c.FirmwareDir, fmt.Sprintf("%s.md5", v))
}

// DefaultConfig holds the default configuration settings.
var DefaultConfig = Config{
	FirmwareDir: "/var/lib/asterisk/digium_phones/firmware",
	Manifest:    "https://downloads.digium.com/pub/telephony/res_digium_phone/firmware/dpma-firmware.json",
	Models:      "*",
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

	if val := os.Getenv("FIRMWARE_DIR"); val != "" {
		overload = append(overload, "FIRMWARE_DIR")
		config.FirmwareDir = val
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
