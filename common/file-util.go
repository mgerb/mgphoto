package common

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/djherbis/times.v1"
)

var (
	// This map is used to define extensions to examine
	knownTypes = map[string][]string{
		"video":   []string{"mp4", "avi", "m4v", "mov", "insv"},
		"photo":   []string{"heic", "jpeg", "jpg", "raw", "arw", "png", "psd", "gpr", "gif", "tiff", "tif", "dng", "insp"},
		"sidecar": []string{"xmp", "on1", "xml"},
		// Don't really need LRV - Low Resolution Video or THM - Thumbnail
	}
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
	if !dryRun {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
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

	t, err := times.Stat(src)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Keep the original mod time
	err = os.Chtimes(dest, t.AccessTime(), t.ModTime())
	if err != nil {
		log.Fatal(err.Error())
	}

	return nil
}

func validFileType(path string) bool {
	return isPhoto(path) || isVideo(path) || (sidecarFiles && isSidecar(path))
}

func isPhoto(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	for _, e := range knownTypes["photo"] {
		if ext == "."+e {
			return true
		}
	}

	return false
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	for _, e := range knownTypes["video"] {
		if ext == "."+e {
			return true
		}
	}

	return false
}

func isSidecar(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	for _, e := range knownTypes["sidecar"] {
		if ext == "."+e {
			return true
		}
	}

	return false
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

		fullPath := path.Join(dir, f.Name())
		if f.IsDir() {
			if f.Name() != "@eaDir" && f.Name() != "thumbnails" {
				filePaths = append(filePaths, getAllFilePaths(fullPath)...)
			}
		} else {
			if validFileType(fullPath) {
				filePaths = append(filePaths, path.Join(fullPath))
			} else {
				Info.Println(fullPath, "\t skipping, unrecognized filetype")
			}
		}
	}

	return filePaths
}

func getFilePathsFromSource(dir string, sourceMedia map[[20]byte]*MediaFile) []string {

	dirlist := make(map[string]struct{})

	for _, med := range sourceMedia {
		dirlist[med.destinationPath(dir)] = struct{}{}
	}

	filePaths := []string{}
	for subdir := range dirlist {

		if _, err := os.Stat(subdir); err == nil { // only process dir if it exists

			files, err := ioutil.ReadDir(subdir)

			if err != nil {
				log.Println(err)
				return filePaths
			}

			for _, f := range files {
				fullPath := path.Join(subdir, f.Name())
				if f.IsDir() {
					if f.Name() != "@eaDir" && f.Name() != "thumbnails" {
						filePaths = append(filePaths, getAllFilePaths(fullPath)...)
					}
				} else {
					if validFileType(fullPath) {
						filePaths = append(filePaths, path.Join(fullPath))
					} else {
						Info.Println(fullPath, "\t skipping, unrecognized filetype")
					}
				}
			}
		}

	}

	return filePaths
}
