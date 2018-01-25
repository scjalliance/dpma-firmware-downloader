package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func readCacheFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeCacheFile(path string, md5sum string) error {
	// Prepare the directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("unable to prepare directory for download: %v", err)
	}

	// Write the file
	return ioutil.WriteFile(path, []byte(md5sum), 0644)
}

func createDownloadFile(path string) (*os.File, error) {
	// Prepare the directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, fmt.Errorf("unable to prepare directory for download: %v", err)
	}

	// Prepare the file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare file for download: %v", err)
	}

	return file, nil
}

func removeDownloadFile(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return os.Remove(path)
	}
	return nil
}
