package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	fw "github.com/scjalliance/dpmafirmware"
)

const downloadSuffx = ".download"

func process(config *Config, origin *fw.Origin, r *fw.Release) {
	source := r.URL(origin)
	cacheFile := config.CacheFile(r.Version)
	cachedMD5Sum, err := readCacheFile(cacheFile)

	prefix := fmt.Sprintf("%-7s: ", r.Version) // Convenient logging prefix

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

	files, err := download(prefix, reader, config.FirmwareDir, downloadSuffx)
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

func download(prefix string, reader *fw.Reader, destDir string, suffix string) (files []string, err error) {
	header, readerErr := reader.Next()

	for readerErr == nil {
		// Determine the destination file with suffix
		path := filepath.Join(destDir, header.Name) + suffix

		// Create the file
		file, createErr := createDownloadFile(path)
		if createErr != nil {
			err = createErr
			return
		}

		// Add the file to the list so that we can move or delete it later
		files = append(files, path)

		// Log the download
		log.Printf("%sDownloading %s (modified: %v, bytes: %d)", prefix, header.Name, header.ModTime, header.Size)

		// Copy the file data from the stream to the destination file
		_, copyErr := io.Copy(file, reader)

		// Close the file
		file.Close()

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
