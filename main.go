package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"
	"text/tabwriter"
	"time"
)

var (
	inputPath            string
	outputPath           string
	copyDuplicates       bool
	mvDuplicates         bool
	tinyFiles            bool
	dryRun               bool
	analyze              bool
	fullDestScan         bool
	logPath              string
	version              = "undefined"
	reDateTime           = regexp.MustCompile(`(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})`)
	errMissingCreateTime = errors.New(`Missing create time`)
	Info                 *log.Logger
	Warn                 *log.Logger
	Error                *log.Logger
	wg                   sync.WaitGroup
	maplock              sync.RWMutex
	workercount          int   = 100
	minBytes             int64 = 50000
)

func init() {
	if version != "undefined" {
		println("mgphoto ", version, "\n")
	}

	outputPtr := flag.String("out", "./photos", "Output path - defaults to ./photos")
	logPtr := flag.String("log", "./transfer.log", "Log path - defaults to ./transfer.log")
	dupPtr := flag.Bool("copy-dupes", false, "Copy duplicates to 'duplicates' folder")
	mvPtr := flag.Bool("move-dupes", false, "Move duplicates to their correct location")
	tinyPtr := flag.Bool("copy-tiny", false, "Copy really small images (<5kb)")
	dryPtr := flag.Bool("dryrun", false, "Don't actually do anything")
	analyzePtr := flag.Bool("analyze", false, "Track how long operations are taking")
	fullDestPtr := flag.Bool("full-scan", false, "Scan the entire Destination for duplicates")

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
	dryRun = *dryPtr
	analyze = *analyzePtr
	fullDestScan = *fullDestPtr

	inputPath = flag.Args()[0]
}

func main() {

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", logPath, ":", err)
	}

	wr := tabwriter.NewWriter(logFile, 10, 8, 3, ' ', 0)
	multiWarn := io.MultiWriter(wr, ioutil.Discard)
	multiErr := io.MultiWriter(wr, os.Stderr)

	Info = log.New(wr, "INFO:  ", log.Ldate|log.Ltime)
	Warn = log.New(multiWarn, "WARN:  ", log.Ldate|log.Ltime)
	Error = log.New(multiErr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info.Println("************************************************")
	if dryRun {
		Info.Println(" * * * *            DRY RUN             * * * * ")
	} else {
		Info.Println(" > > > >            NEW RUN             < < < < ")
	}
	Info.Println("************************************************")
	defer wr.Flush()

	createDirIfNotExists(outputPath)

	sourceFiles := getAllFilePaths(inputPath)

	println("Processing source files...")
	sourceMediaFiles := getMediaFiles(sourceFiles, true)

	if !tinyFiles {
		for k, f := range sourceMediaFiles {
			if (f.isPhoto() || f.isVideo()) && f.size < minBytes {
				f.Info("skipping too small photo")
				delete(sourceMediaFiles, k)
			}
		}
	}

	var destFiles []string

	if fullDestScan {
		destFiles = getAllFilePaths(outputPath)
	} else { // Only get paths from directories we're placing things into
		destFiles = getFilePathsFromSource(outputPath, sourceMediaFiles)
	}

	println("Scanning destination for duplicates...")
	destMediaFiles := getMediaFiles(destFiles, mvDuplicates)

	dupeDestFiles := make(map[[20]byte]*MediaFile)
	originalMediaFiles := make(map[[20]byte]*MediaFile)

	// if we are not copying and not moving duplicates omit them
	if !copyDuplicates || mvDuplicates {
		for k := range sourceMediaFiles {
			if destMediaFiles[k] != nil { // file exists in src & dest && has same hash (of first 2k bytes)
				if mvDuplicates {
					dupeDestFiles[k] = destMediaFiles[k]
					originalMediaFiles[k] = sourceMediaFiles[k]
				}
				if sourceMediaFiles[k].size > destMediaFiles[k].size { // file in destination may not be complete
					sourceMediaFiles[k].Info("is larger than duplicate, replacing", destMediaFiles[k].path)
					sourceMediaFiles[k].replace = true
				} else {
					sourceMediaFiles[k].Info("Duplicate of", destMediaFiles[k].path)
					delete(sourceMediaFiles, k)
				}
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
	count := len(paths)

	if count < 1 {
		return outputMap
	}

	progressBar := NewProgressBar(count)
	jobs := make(chan pathBool, count)
	results := make(chan *MediaFile, count)

	for w := 1; w <= workercount; w++ {
		go worker(jobs, results)
	}

	for _, path := range paths {
		jobs <- pathBool{path: path, processMetaData: processMetaData}
	}
	close(jobs)

	for r := 1; r <= count; r++ {
		mediaFile := <-results

		if mediaFile != nil {
			maplock.Lock()
			outputMap[mediaFile.sha1] = mediaFile
			maplock.Unlock()
		}
		progressBar.increment()
	}
	progressBar.wait()

	return outputMap
}

type pathBool struct {
	path            string
	processMetaData bool
}

func worker(jobs <-chan pathBool, results chan<- *MediaFile) {
	for j := range jobs {
		mediaFile := NewMediaFile(j.path, j.processMetaData)
		results <- mediaFile
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
