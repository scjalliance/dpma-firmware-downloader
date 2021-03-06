package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gentlemanautomaton/signaler"
	fw "github.com/scjalliance/dpmafirmware"
)

const (
	attempts = 5
	delay    = 5 * time.Second
)

func main() {
	shutdown := signaler.New().Capture(os.Interrupt, syscall.SIGTERM)
	defer shutdown.Wait()
	defer shutdown.Trigger()

	// Process configuration
	var (
		config   = buildConfig()
		manifest fw.Manifest
	)

	// Retrieve the firmware manifest
	log.Printf("Retrieving DPMA firmware manifest from %s", config.Manifest)

	for i := 0; i < attempts; i++ {
		if shutdown.Signaled() {
			return
		}

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

	// Filter the manifest
	total := len(manifest.Releases)
	manifest = manifest.Filter(fw.ModelMatchFilter(config.Include.Models))
	matched := len(manifest.Releases)
	log.Printf("Filter matches %d of %d releases.", matched, total)

	//log.Printf("Manifest Data:\n%s", manifest.Summary())

	// Keep track of how many versions we've acquired for each model
	var acq AcquisitionMap
	acq.Require(config.Latest)

	// Process each release
	for _, r := range manifest.Releases {
		process(shutdown.Signal, &config, &manifest.Origin, &acq, &r)
	}
}
