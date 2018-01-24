package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gentlemanautomaton/globber"
	fw "github.com/scjalliance/dpmafirmware"
)

// Filter matches models and files.
type Filter struct {
	Models globber.Set  // Comma or whitespace separated, glob matching
	Files  globber.Glob // Glob matching
}

// Config holds configuration data for the DPMA firmware downloader
type Config struct {
	FirmwareDir string `json:"firmwaredir"`
	Manifest    string `json:"manifest"`
	Include     Filter `json:"include"`
	Exclude     Filter `json:"exclude"`
	Flatten     bool   `json:"flatten"`
}

// CacheFile returns the path to the MD5 cache file for the given version.
func (c *Config) CacheFile(v fw.Version) string {
	return filepath.Join(c.FirmwareDir, fmt.Sprintf("%s.md5", v))
}

// DefaultConfig holds the default configuration settings.
var DefaultConfig = Config{
	FirmwareDir: "/var/lib/asterisk/digium_phones/firmware",
	Manifest:    "https://downloads.digium.com/pub/telephony/res_digium_phone/firmware/dpma-firmware.json",
	Include:     Filter{Models: globber.Split(fw.Wildcard), Files: globber.New("*.eff")},
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

	if val := os.Getenv("INCLUDE_MODELS"); val != "" {
		overload = append(overload, "INCLUDE_MODELS")
		config.Include.Models = globber.Split(val)
	}

	if val := os.Getenv("INCLUDE_FILES"); val != "" {
		overload = append(overload, "INCLUDE_FILES")
		config.Include.Files = globber.New(val)
	}

	if val := os.Getenv("EXCLUDE_MODELS"); val != "" {
		overload = append(overload, "EXCLUDE_MODELS")
		config.Exclude.Models = globber.Split(val)
	}

	if val := os.Getenv("EXCLUDE_FILES"); val != "" {
		overload = append(overload, "EXCLUDE_FILES")
		config.Exclude.Files = globber.New(val)
	}

	if val := os.Getenv("FLATTEN"); val != "" {
		overload = append(overload, "FLATTEN")
		config.Flatten, _ = strconv.ParseBool(val)
	}

	if len(overload) > 0 {
		log.Printf("Overriding configuration from environment variables: %v", overload)
	} else {
		log.Println("No environment variables provided for configuration.")
	}

	if len(os.Args) > 1 {
		log.Printf("Overriding configuration from command line arguments.")
		flag.StringVar(&config.Manifest, "url", config.Manifest, "URL of DPMA manifest")
		flag.StringVar(&config.FirmwareDir, "dir", config.FirmwareDir, "Directory in which to save firmware")
		flag.Var(&config.Include.Models, "inc", "Models to include, comma-separated values or globs")
		flag.Var(&config.Include.Files, "incfiles", "Files to include, value or glob")
		flag.Var(&config.Exclude.Models, "exc", "Models to exclude, comma-separated values or globs")
		flag.Var(&config.Exclude.Files, "excfiles", "Files to exclude, value or glob")
		flag.BoolVar(&config.Flatten, "flatten", config.Flatten, "flatten extracted files to single directory")
		flag.Parse()
	}

	log.Printf("Manifest URL:   %s", config.Manifest)
	log.Printf("FirmwareDir:    %s", config.FirmwareDir)
	log.Printf("Include Models: %s", config.Include.Models)
	log.Printf("Exclude Models: %s", config.Exclude.Models)
	log.Printf("Include Files:  %s", config.Include.Files)
	log.Printf("Exclude Files:  %s", config.Exclude.Files)
	log.Printf("Flatten:        %v", config.Flatten)

	return config
}
