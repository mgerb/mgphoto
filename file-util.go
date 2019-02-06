package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

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

// recursively read files in directory
func readFiles(dir string, processMetaData bool) []*MediaFile {

	mediaFiles := []*MediaFile{}
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return mediaFiles
	}

	for _, f := range files {

		if f.IsDir() {
			mediaFiles = append(mediaFiles, readFiles(path.Join(dir, f.Name()), processMetaData)...)
		} else {

			mediaFile := NewMediaFile(path.Join(dir, f.Name()), processMetaData)

			if mediaFile != nil {
				mediaFiles = append(mediaFiles, mediaFile)
			}
		}
	}

	return mediaFiles
}

func scanMediaDirectory(path string, processMetaData bool) map[[20]byte]*MediaFile {
	mediaFiles := readFiles(path, processMetaData)

	outputMap := map[[20]byte]*MediaFile{}

	for _, m := range mediaFiles {
		outputMap[m.sha1] = m
	}

	return outputMap
}
