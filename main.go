package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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
	log.Println("Manifest retrieved.")

	// Print the results for now
	log.Printf("Origin: %s", manifest.URL)
	log.Printf("Tarball: %s", manifest.Tarball)

	for _, branch := range manifest.Releases.Branches() {
		log.Printf("Branch %s:", branch.Name)
		for r := range branch.Releases {
			log.Printf("%v", branch.Releases[r])
		}
	}
}
