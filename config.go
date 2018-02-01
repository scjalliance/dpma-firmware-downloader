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
	Manifest    string `json:"manifest"`
	FirmwareDir string `json:"firmwaredir"`
	CacheDir    string `json:"cachedir"`
	Include     Filter `json:"include"`
	Exclude     Filter `json:"exclude"`
	Latest      int    `json:"latest"`
	Flatten     bool   `json:"flatten"`
}

// CacheFile returns the path to the MD5 cache file for the given version.
func (c *Config) CacheFile(v fw.Version, model string) string {
	return filepath.Join(c.CacheDir, fmt.Sprintf("%s.%s.md5", v, model))
}

// DefaultConfig holds the default configuration settings.
var DefaultConfig = Config{
	FirmwareDir: "fw",
	CacheDir:    "cache",
	Manifest:    "https://downloads.digium.com/pub/telephony/res_digium_phone/firmware/dpma-firmware.json",
	Include:     Filter{Models: globber.NewSet("*"), Files: globber.New("*.eff")},
}

func buildConfig() (config Config) {
	var (
		overload         []string // Slice of provided environment variable names
		configFileLoaded bool
		configFile       = "config.json"
	)
	envString("CONFIG_FILE", &configFile, &overload)

	// Configuration File or Defaults
	data, err := ioutil.ReadFile(configFile)

	switch {
	case os.IsNotExist(err):
		config = DefaultConfig
	case err == nil:
		if err = json.Unmarshal(data, &config); err != nil {
			log.Fatal(err)
		}
		configFileLoaded = true
	default:
		log.Println(err)
	}

	// Environment Variables
	envString("MANIFEST", &config.Manifest, &overload)
	envString("FIRMWARE_DIR", &config.FirmwareDir, &overload)
	envString("CACHE_DIR", &config.CacheDir, &overload)
	envGlobSet("INCLUDE_MODELS", &config.Include.Models, &overload)
	envGlob("INCLUDE_FILES", &config.Include.Files, &overload)
	envGlobSet("EXCLUDE_MODELS", &config.Exclude.Models, &overload)
	envGlob("EXCLUDE_FILES", &config.Exclude.Files, &overload)
	envInt("LATEST", &config.Latest, &overload)
	envBool("FLATTEN", &config.Flatten, &overload)

	// Arguments
	flag.StringVar(&config.Manifest, "url", config.Manifest, "URL of DPMA manifest")
	flag.StringVar(&config.FirmwareDir, "firmwaredir", config.FirmwareDir, "Directory in which to save firmware")
	flag.StringVar(&config.CacheDir, "cachedir", config.CacheDir, "Directory in which to save cache files")
	flag.Var(&config.Include.Models, "include", "Models to include, comma-separated values or globs")
	flag.Var(&config.Include.Files, "includefiles", "Files to include, value or glob")
	flag.Var(&config.Exclude.Models, "exclude", "Models to exclude, comma-separated values or globs")
	flag.Var(&config.Exclude.Files, "excludefiles", "Files to exclude, value or glob")
	flag.IntVar(&config.Latest, "latest", config.Latest, "# of releases to download for each model (0 for unlimited)")
	flag.BoolVar(&config.Flatten, "flatten", config.Flatten, "flatten extracted files to single directory")
	flag.Parse()

	log.Println("Starting dpma firmware downloader...")

	if configFileLoaded {
		log.Printf("Loaded configuration from \"%s\"", configFile)
	} else {
		log.Printf("No configuration file found at \"%s\". Using defaults.", configFile)
	}

	if len(overload) > 0 {
		log.Printf("Loaded configuration from environment variables: %v", overload)
	} else {
		log.Println("No environment variables provided for configuration.")
	}

	if len(os.Args) > 1 {
		log.Printf("Loaded configuration from command line arguments.")
	}

	// Summary
	log.Printf("Manifest URL:   %s", config.Manifest)
	log.Printf("Firmware Dir:   %s", config.FirmwareDir)
	log.Printf("Cache Dir:      %s", config.CacheDir)
	log.Printf("Include Models: %s", config.Include.Models)
	log.Printf("Exclude Models: %s", config.Exclude.Models)
	log.Printf("Include Files:  %s", config.Include.Files)
	log.Printf("Exclude Files:  %s", config.Exclude.Files)
	if config.Latest == 0 {
		log.Printf("Latest:         unlimited")
	} else {
		log.Printf("Latest:         %d", config.Latest)
	}
	log.Printf("Flatten:        %v", config.Flatten)

	return config
}

func envString(name string, value *string, overload *[]string) {
	if env := os.Getenv(name); env != "" {
		*overload = append(*overload, name)
		*value = env
	}
}

func envGlob(name string, value *globber.Glob, overload *[]string) {
	if env := os.Getenv(name); env != "" {
		*overload = append(*overload, name)
		*value = globber.New(env)
	}
}

func envGlobSet(name string, value *globber.Set, overload *[]string) {
	if env := os.Getenv(name); env != "" {
		*overload = append(*overload, name)
		*value = globber.Split(env)
	}
}

func envInt(name string, value *int, overload *[]string) {
	if env := os.Getenv(name); env != "" {
		*overload = append(*overload, name)
		*value, _ = strconv.Atoi(env)
	}
}

func envBool(name string, value *bool, overload *[]string) {
	if env := os.Getenv(name); env != "" {
		*overload = append(*overload, name)
		*value, _ = strconv.ParseBool(env)
	}
}
