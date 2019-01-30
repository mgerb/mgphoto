package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

var (
	inputPath  string
	outputPath string
)

func init() {

	outputPtr := flag.String("o", "./output", "Output path - defaults to ./output")

	flag.Parse()

	if len(flag.Args()) < 1 {
		exit(errors.New("Invalid arguments - please supply a source directory"))
	}

	outputPath = *outputPtr
	inputPath = flag.Args()[0]
}

func main() {

	createDirIfNotExists(outputPath)

	mediaFiles := readFiles(inputPath)

	for _, f := range mediaFiles {
		fmt.Println(f)
	}
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(0)
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
}

// recursively read files in directory
func readFiles(dir string) []*MediaFile {

	mediaFiles := []*MediaFile{}
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return mediaFiles
	}

	for _, f := range files {

		if f.IsDir() {
			mediaFiles = append(mediaFiles, readFiles(path.Join(dir, f.Name()))...)
		} else {
			mediaFiles = append(mediaFiles, &MediaFile{name: f.Name()})
		}
	}

	return mediaFiles
}
