package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gentlemanautomaton/signaler"
	fw "github.com/scjalliance/dpmafirmware"
)

const downloadSuffx = ".download"

func process(shutdown signaler.Signal, config *Config, origin *fw.Origin, acquired *AcquisitionMap, r *fw.Release) {
	if shutdown.Signaled() {
		return
	}

	// Apply our model filters
	matched := r.Models.Include(config.Include.Models).Exclude(config.Exclude.Models)

	// Exclude models that have already reached their quota
	wanted := matched.Exclude(acquired)

	// Record the matched models in the acquistion map
	acquired.Add(matched...)

	// Return if we don't want any firmware from this release
	if len(wanted) == 0 {
		return
	}

	// Determine cache status
	status := determineCacheStatus(config, r.MD5Sum, r.Version, wanted)
	needed := status.Needed()

	// Logging prefix for convenience
	versionPrefix := fmt.Sprintf("%-7s ", r.Version)
	neededPrefix := fmt.Sprintf("%s[%v]: ", versionPrefix, needed)

	// Report cache status
	for _, line := range status.Summary() {
		log.Printf("%s%s", versionPrefix, line)
	}

	// Exit if there's no work to be done
	if len(needed) == 0 {
		return
	}

	if shutdown.Signaled() {
		return
	}

	// Proceed with downloading the needed files
	//prefix = fmt.Sprintf("%s [%v]: ", prefix, needed)
	source := r.URL(origin)
	log.Printf("%sDownloading  url: %s", neededPrefix, source)

	reader, err := r.Get(origin)
	if err != nil {
		log.Printf("%sFailed     error: %v", neededPrefix, err)
		return
	}

	// Close the reader when we're done or if a shutdown is initiated
	stop := shutdown.Derive()
	stopped := stop.Then(func() { reader.Close() })
	defer stopped.Wait()
	defer stop.Trigger()

	files, err := download(config, versionPrefix, reader, needed, config.FirmwareDir, downloadSuffx)
	var failed []string
	if err != nil {
		log.Printf("%sFailed     error: %v", neededPrefix, err)
		failed = files
	} else {
		downloadMD5Sum := reader.MD5Sum()
		if downloadMD5Sum == r.MD5Sum {
			log.Printf("%sVerified     md5: %s", neededPrefix, downloadMD5Sum)
			for _, oldpath := range files {
				newpath := strings.TrimSuffix(oldpath, downloadSuffx)
				if mvErr := os.Rename(oldpath, newpath); mvErr != nil {
					failed = append(failed, oldpath)
					log.Printf("%sInstallation failed: %s: %v", neededPrefix, newpath, mvErr)
				} else {
					log.Printf("%sInstalled   path: %s", neededPrefix, newpath)
				}
			}
		} else {
			log.Printf("%sDownload doesn't match manifest (download md5: %s, manifest md5: %s)", neededPrefix, downloadMD5Sum, r.MD5Sum)
			failed = files
		}
	}

	if len(failed) != 0 {
		log.Printf("%sCleaning up leftover download files...", neededPrefix)
		for _, path := range failed {
			if rmErr := removeDownloadFile(path); rmErr != nil {
				log.Printf("%sUnable to remove %s: %v", neededPrefix, path, rmErr)
			} else {
				log.Printf("%sRemoved     path: %s", neededPrefix, path)
			}
		}
		return
	}

	for _, model := range needed {
		cacheFile := config.CacheFile(r.Version, model)
		err = writeCacheFile(cacheFile, r.MD5Sum)
		if err != nil {
			log.Printf("%s[%v]: Failed to write \"%s\" to cache file: %v", versionPrefix, model, r.MD5Sum, err)
			return
		}
	}
}

func download(config *Config, versionPrefix string, reader *fw.Reader, models fw.ModelSet, destDir string, suffix string) (files []string, err error) {
	header, readerErr := reader.Next()

	for readerErr == nil {
		if !shouldDownload(config, models, header) {
			header, readerErr = reader.Next()
			continue
		}

		// Determine the destination file with suffix
		var path string
		if config.Flatten {
			path = header.Name
		} else {
			path = header.Path
		}
		path = filepath.Join(destDir, path) + suffix

		// Create the file
		file, createErr := createDownloadFile(path)
		if createErr != nil {
			err = createErr
			return
		}

		// Add the file to the list so that we can move or delete it later
		files = append(files, path)

		// Log the download
		modelPrefix := fmt.Sprintf("%s[%v]: ", versionPrefix, header.Models)
		log.Printf("%sProcessing  file: %s (modified: %v, bytes: %d)", modelPrefix, header.Name, header.ModTime, header.Size)

		// Copy the file data from the stream to the destination file
		_, copyErr := io.Copy(file, reader)

		// Close the file
		file.Close()

		// Update the file timestamp
		os.Chtimes(path, time.Now(), header.ModTime)

		// Check for a failure during the copy
		if copyErr != nil {
			err = copyErr
			return
		}

		// Move on to the next file
		header, readerErr = reader.Next()
	}

	if readerErr != io.EOF {
		err = readerErr
	}

	return
}

func determineCacheStatus(config *Config, md5Sum string, version fw.Version, models fw.ModelSet) (status CacheStatus) {
	status.md5Sum = md5Sum

	for _, model := range models {
		cacheFile := config.CacheFile(version, model)
		cachedMD5Sum, err := readCacheFile(cacheFile)
		if err != nil {
			if os.IsNotExist(err) {
				status.Missing = append(status.Missing, model)
			} else {
				status.Failed = append(status.Failed, model)
			}
		} else {
			if cachedMD5Sum != md5Sum {
				status.Revised = append(status.Revised, model)
			} else {
				status.Current = append(status.Current, model)
			}
		}
	}

	return
}

func shouldDownload(config *Config, models fw.ModelSet, header *fw.Header) bool {
	if len(header.Models) > 0 && !models.Map().Contains(header.Models...) {
		// The file doesn't pertain to one of the models we want
		return false
	}

	if inc := config.Include.Files; inc.String() != "" && !inc.Match(header.Path) {
		// This isn't one of the files we've been asked to include
		return false
	}

	if exc := config.Exclude.Files; exc.String() != "" && exc.Match(header.Path) {
		// This is one of the files we've been asked to exclude
		return false
	}

	return true
}
