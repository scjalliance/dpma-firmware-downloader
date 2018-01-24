package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	fw "github.com/scjalliance/dpmafirmware"
)

const (
	attempts = 5
	delay    = 5 * time.Second
)

func main() {
	log.Println("Starting dpma firmware downloader...")

	var (
		config   = buildConfig()
		manifest fw.Manifest
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

	total := len(manifest.Releases)
	manifest = manifest.IncludeMatch(config.Include.Models)
	matched := len(manifest.Releases)
	log.Printf("Model filter matches %d of %d releases.", matched, total)

	// Print the results for now
	log.Printf("Manifest Data:\n%s", manifest.Summary())

	for _, r := range manifest.Releases {
		process(&config, &manifest.Origin, &r)
	}
}
