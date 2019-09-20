package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshsziegler/zfmt"
)

const (
	lastFormatTimeFilePath = ".zfmt-last-format-time"
)

// getInfo returns the file's meta-data struct
func getInfo(fPath string) os.FileInfo {
	fInfo, err := os.Stat(fPath)
	if err != nil {
		log.Fatal(err)
	}
	return fInfo
}

// readFile returns the entire contents of this file
func readFile(fPath string) []byte {
	fContent, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Fatal(err)
	}
	return fContent
}

// getLastFormatTime returns the time of the last format, if the file with that timestamp exists.
// Otherwise, it will return the default/zero time.Time{} struct.
//
// This helps speed up our formatting by only running on files that have been updated since the
// the last run.
func getLastFormatTime() time.Time {
	fContent, err := ioutil.ReadFile(lastFormatTimeFilePath)
	if err != nil {
		return time.Time{}
	}
	fTime, err := time.Parse(time.RFC3339, string(fContent))
	if err != nil {
		return time.Time{}
	}
	return fTime
}

func main() {
	// Store the current time so we can save it later as the start of this format
	thisFormatTime := time.Now().Format(time.RFC3339)
	lastFormatTime := getLastFormatTime()

	// Create a list of files that we we need to check for formatting
	var files []string
	err := filepath.Walk(".", func(fPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		fInfo := getInfo(fPath)
		if fInfo.IsDir() {
			return nil
		}
		// Skip hidden/special files
		if strings.HasPrefix(fPath, ".") {
			return nil
		}
		// Skip the static/dist/ directory
		if strings.HasPrefix(fPath, "static/dist/") {
			return nil
		}
		// Skip if the extensions isn't Go, JS, or CSS
		fExtension := path.Ext(fPath)
		if fExtension != ".go" && fExtension != ".js" && fExtension != ".css" {
			return nil
		}
		// Skip minified JS and CSS files
		if strings.HasSuffix(".min.js") || strings.HasSuffix(".min.css") {
			return nil
		}
		// Skip if this file hasn't been modified since the last format
		if lastFormatTime.After(fInfo.ModTime()) {
			return nil
		}
		files = append(files, fPath)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	// Now format all files in the list we just created
	for _, fPath := range files {
		fContent := readFile(fPath)
		fExtension := path.Ext(fPath)
		// Process the file according to it's path and/or extension
		var result []byte
		switch {
		case fExtension == ".go":
			result = zfmt.FormatGo(fPath, fContent)
		case fExtension == ".js":
			result = zfmt.FormatJS(fPath)
		case fExtension == ".css":
			result = zfmt.FormatCSS(fPath, fContent)
		default:
			continue // Do not process files we don't recognize
		}
		// Only write file to disk if it's different from the old version
		if !bytes.Equal(fContent, result) {
			fmt.Println("  * ", fPath)
			ioutil.WriteFile(fPath, []byte(result), 0644)
		}
	}
	ioutil.WriteFile(lastFormatTimeFilePath, []byte(thisFormatTime), 0644)
}
