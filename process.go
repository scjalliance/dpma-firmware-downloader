package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	fw "github.com/scjalliance/dpmafirmware"
)

const downloadSuffx = ".download"

func process(config *Config, origin *fw.Origin, r *fw.Release) {
	source := r.URL(origin)
	cacheFile := config.CacheFile(r.Version)
	cachedMD5Sum, err := readCacheFile(cacheFile)

	wanted := r.Models.Include(config.Include.Models).Exclude(config.Exclude.Models)
	if len(wanted) == 0 {
		// This release doesn't have any firmware that we want
		return
	}

	prefix := fmt.Sprintf("%-7s [%v]: ", r.Version, wanted) // Convenient logging prefix

	// Cache test
	if err == nil {
		if cachedMD5Sum == r.MD5Sum {
			log.Printf("%sUp to date (md5: %s)", prefix, r.MD5Sum)
			return
		}
		log.Printf("%sRevision detected (md5: %s -> %s). Downloading from %s", prefix, cachedMD5Sum, r.MD5Sum, source)
	} else if os.IsNotExist(err) {
		log.Printf("%sMissing (md5: %s). Downloading from %s", prefix, r.MD5Sum, source)
	} else {
		log.Printf("%sFailed to read MD5 from %s: %v", prefix, filepath.Base(cacheFile), err)
		return
	}

	reader, err := r.Get(origin)
	if err != nil {
		log.Printf("%sDownload Failed: %v", prefix, err)
		return
	}
	defer reader.Close()

	files, err := download(config, prefix, reader, wanted, config.FirmwareDir, downloadSuffx)
	if err != nil {
		log.Printf("%sDownload failed: %v", prefix, err)
		return
	}

	var failed []string

	downloadMD5Sum := reader.MD5Sum()
	if downloadMD5Sum == r.MD5Sum {
		log.Printf("%sDownload completed and verified (md5: %s)", prefix, downloadMD5Sum)
		for _, oldpath := range files {
			newpath := strings.TrimSuffix(oldpath, downloadSuffx)
			if mvErr := os.Rename(oldpath, newpath); mvErr != nil {
				failed = append(failed, oldpath)
				log.Printf("%sInstallation failure: %s: %v", prefix, newpath, mvErr)
			} else {
				log.Printf("%sInstalled: %s", prefix, newpath)
			}
		}
	} else {
		log.Printf("%sDownload doesn't match manifest (download md5: %s, manifest md5: %s)", prefix, downloadMD5Sum, r.MD5Sum)
		failed = files
	}

	if len(failed) != 0 {
		log.Printf("%sCleaning up leftover download files...", prefix)
		for _, path := range failed {
			if rmErr := removeDownloadFile(path); rmErr != nil {
				log.Printf("%sUnable to remove %s: %v", prefix, path, rmErr)
			} else {
				log.Printf("%sRemoved %s", prefix, path)
			}
		}
		return
	}

	err = writeCacheFile(cacheFile, r.MD5Sum)
	if err != nil {
		log.Printf("%sFailed to write \"%s\" to cache file: %v", prefix, r.MD5Sum, err)
		return
	}
}

func download(config *Config, prefix string, reader *fw.Reader, models fw.ModelSet, destDir string, suffix string) (files []string, err error) {
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
		log.Printf("%sDownloading %s for %v (modified: %v, bytes: %d)", prefix, header.Name, header.Models, header.ModTime, header.Size)

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
