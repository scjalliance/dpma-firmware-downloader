package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/scjalliance/dpmafirmware"
)

const (
	attempts = 5
	delay    = 5 * time.Second
)

func main() {
	log.Println("Starting dpma firmware downloader...")

	var (
		config   = buildConfig()
		manifest dpmafirmware.Manifest
	)

	log.Printf("Retrieving DPMA firmware manifest from %s", config.Manifest)

	// Retrieve the firmware manifest
	for i := 0; i < attempts; i++ {
		action := "retrieve the DPMA firmware manifest"
		res, err := http.Get(config.Manifest)
		// TODO: Check response status code
		if failed(i, attempts, action, "", err) {
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if failed(i, attempts, action, "", err) {
			continue
		}
		err = json.Unmarshal(body, &manifest)
		if failed(i, attempts, action, "parse the DPMA firmware manifest", err) {
			continue
		}
		break
	}
	log.Println("Manifest retrieved and parsed successfully.")

	// Print the results for now
	log.Printf("Manifest Data:\n%s", manifest.Summary())

	for _, r := range manifest.Releases {
		process(&config, &manifest.Origin, &r)
	}
}

func process(config *Config, origin *dpmafirmware.Origin, r *dpmafirmware.Release) {
	source := r.URL(origin)
	cacheFile := config.CacheFile(r.Version)
	cachedMD5Sum, err := readCache(cacheFile)

	prefix := fmt.Sprintf("%-7s: ", r.Version)

	if err == nil {
		if cachedMD5Sum == r.MD5Sum {
			log.Printf("%sUp to date (md5: %s)", prefix, r.MD5Sum)
			return
		} else {
			log.Printf("%sRevision detected (md5: %s -> %s). Downloading from %s", prefix, cachedMD5Sum, r.MD5Sum, source)
		}
	} else if os.IsNotExist(err) {
		log.Printf("%sMissing (md5: %s). Downloading from %s", prefix, r.MD5Sum, source)
	} else {
		log.Printf("%sFailed to read MD5 from %s: %v", prefix, filepath.Base(cacheFile), err)
		return
	}

	err = download(source, config.FirmwareDir)
	if err != nil {
		log.Printf("%sDownload failed: %v", prefix, err)
		return
	}

	err = writeCacheFile(cacheFile, r.MD5Sum)
	if err != nil {
		log.Printf("%sFailed to write \"%s\" to cache file: %v", prefix, r.MD5Sum, err)
		return
	}
}

func download(source *url.URL, destDir string) error {
	log.Printf("Downloading from %s to %s", source, destDir)
	return nil
}

func readCache(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeCacheFile(path string, md5sum string) error {
	return ioutil.WriteFile(path, []byte(md5sum), 0644)
}
