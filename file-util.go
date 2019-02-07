package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func getFileSuffix(n int) string {
	return "_" + strconv.Itoa(n)
}

// if file already exists
// append _1 to the end.
// Keep incrementing until file
// does not exist.
func renameIfFileExists(path string) string {
	fileSuffix := 1
	for fileExists(path) {
		extension := filepath.Ext(path)
		pathPrefix := path[0 : len(path)-len(extension)]

		previousFileSuffix := getFileSuffix(fileSuffix - 1)
		if strings.HasSuffix(pathPrefix, previousFileSuffix) {
			pathPrefix = pathPrefix[0 : len(pathPrefix)-len(previousFileSuffix)]
		}

		path = pathPrefix + getFileSuffix(fileSuffix) + extension
		fileSuffix++
	}

	return path
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)

	if err != nil {
		return err
	}

	defer srcFile.Close()

	destFile, err := os.Create(dest) // creates if file doesn't exist

	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied

	if err != nil {
		return err
	}

	err = destFile.Sync()

	if err != nil {
		return err
	}

	return nil
}

// recursively read directory and get all file paths
func getAllFilePaths(dir string) []string {

	filePaths := []string{}
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		log.Println(err)
		return filePaths
	}

	for _, f := range files {

		if f.IsDir() {
			filePaths = append(filePaths, getAllFilePaths(path.Join(dir, f.Name()))...)
		} else {

			filePaths = append(filePaths, path.Join(dir, f.Name()))
		}
	}

	return filePaths
}
