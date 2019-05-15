package main

import (
	"flag"
	"os"
)

var (
	inputPath      string
	outputPath     string
	copyDuplicates bool
	version        = "undefined"
)

func init() {
	if version != "undefined" {
		println("mgphoto ", version, "\n")
	}

	outputPtr := flag.String("o", "./photos", "Output path - defaults to ./photos")
	dupPtr := flag.Bool("d", false, "Copy duplicates to 'duplicates' folder")

	flag.Parse()

	if len(flag.Args()) < 1 {
		println("Invalid arguments - please supply a source directory")
		os.Exit(0)
	}

	outputPath = *outputPtr
	copyDuplicates = *dupPtr
	inputPath = flag.Args()[0]
}

func main() {

	createDirIfNotExists(outputPath)

	sourceFiles := getAllFilePaths(inputPath)
	destFiles := getAllFilePaths(outputPath)

	println("Processing source files...")
	sourceMediaFiles := getMediaFiles(sourceFiles, true)

	println("Scanning destination for duplicates...")
	destMediaFiles := getMediaFiles(destFiles, false)

	// if we are not copying duplicates omit them
	if !copyDuplicates {
		for k := range sourceMediaFiles {
			if destMediaFiles[k] != nil {
				delete(sourceMediaFiles, k)
			}
		}
	}

	if len(sourceMediaFiles) == 0 {
		println("No new files to copy.")
		return
	}

	println("Copying new files to destination...")
	progressBar := NewProgressBar(len(sourceMediaFiles))
	for k, val := range sourceMediaFiles {
		val.writeToDestination(outputPath, copyDuplicates && destMediaFiles[k] != nil)
		progressBar.increment()
	}

	progressBar.wait()
}

// get media file objects from file path list
func getMediaFiles(paths []string, processMetaData bool) map[[20]byte]*MediaFile {

	outputMap := map[[20]byte]*MediaFile{}

	if len(paths) < 1 {
		return outputMap
	}

	progressBar := NewProgressBar(len(paths))

	for _, path := range paths {
		mediaFile := NewMediaFile(path, processMetaData)

		if mediaFile != nil {
			outputMap[mediaFile.sha1] = mediaFile
		}
		progressBar.increment()
	}

	progressBar.wait()

	return outputMap
}
