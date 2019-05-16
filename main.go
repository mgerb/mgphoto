package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

var (
	inputPath            string
	outputPath           string
	copyDuplicates       bool
	mvDuplicates         bool
	tinyFiles            bool
	logPath              string
	version              = "undefined"
	reDateTime           = regexp.MustCompile(`(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})`)
	errMissingCreateTime = errors.New(`Missing create time`)
	Info                 *log.Logger
	Warn                 *log.Logger
	Error                *log.Logger
)

func init() {
	if version != "undefined" {
		println("mgphoto ", version, "\n")
	}

	outputPtr := flag.String("o", "./photos", "Output path - defaults to ./photos")
	logPtr := flag.String("l", "./transfer.log", "Log path - defaults to ./transfer.log")
	dupPtr := flag.Bool("d", false, "Copy duplicates to 'duplicates' folder")
	mvPtr := flag.Bool("m", false, "Move duplicates to their correct location")
	tinyPtr := flag.Bool("t", false, "Copy really small images (<5kb)")

	flag.Parse()

	if len(flag.Args()) < 1 {
		println("Invalid arguments - please supply a source directory")
		os.Exit(0)
	}

	outputPath = *outputPtr
	copyDuplicates = *dupPtr
	mvDuplicates = *mvPtr
	tinyFiles = *tinyPtr
	logPath = *logPtr
	inputPath = flag.Args()[0]

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", output, ":", err)
	}

	multiWarn := io.MultiWriter(file, os.Stdout)
	multiErr := io.MultiWriter(file, os.Stderr)

	Info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(multiWarn, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(multiErr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {

	createDirIfNotExists(outputPath)

	sourceFiles := getAllFilePaths(inputPath)
	destFiles := getAllFilePaths(outputPath)

	println("Processing source files...")
	sourceMediaFiles := getMediaFiles(sourceFiles, true)

	if !tinyFiles {
		for k, f := range sourceMediaFiles {
			if f.isPhoto() && f.size < 5000 {
				delete(sourceMediaFiles, k)
			}
		}
	}

	println("Scanning destination for duplicates...")
	destMediaFiles := getMediaFiles(destFiles, mvDuplicates)

	dupeDestFiles := make(map[[20]byte]*MediaFile)
	originalMediaFiles := make(map[[20]byte]*MediaFile)

	// if we are not copying and not moving duplicates omit them
	if !copyDuplicates || mvDuplicates {
		for k := range sourceMediaFiles {
			if destMediaFiles[k] != nil {
				if mvDuplicates {
					dupeDestFiles[k] = destMediaFiles[k]
					originalMediaFiles[k] = sourceMediaFiles[k]
				}
				delete(sourceMediaFiles, k)
			}
		}
	}

	if len(sourceMediaFiles) == 0 && len(dupeDestFiles) == 0 {
		println("No new files to copy or move.")
		return
	}

	if len(sourceMediaFiles) > 0 {
		println("Copying new files to destination...")
		progressBar := NewProgressBar(len(sourceMediaFiles))
		for k, val := range sourceMediaFiles {
			val.writeToDestination(outputPath, copyDuplicates && destMediaFiles[k] != nil)
			progressBar.increment()
		}

		progressBar.wait()
	}

	if mvDuplicates && len(dupeDestFiles) > 0 {
		fmt.Println("Moving existing files to the correct destination...")
		dupeProgressBar := NewProgressBar(len(dupeDestFiles))
		for k, val := range dupeDestFiles {
			val.moveToDestination(outputPath, originalMediaFiles[k])
			dupeProgressBar.increment()
		}
		dupeProgressBar.wait()
	}
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
