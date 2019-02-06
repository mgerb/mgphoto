package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	inputPath      string
	outputPath     string
	copyDuplicates bool
)

func init() {

	outputPtr := flag.String("o", "./output", "Output path - defaults to ./output")
	dupPtr := flag.Bool("d", false, "Copy duplicates to 'duplicates' folder")

	flag.Parse()

	if len(flag.Args()) < 1 {
		exit(errors.New("Invalid arguments - please supply a source directory"))
	}

	outputPath = *outputPtr
	copyDuplicates = *dupPtr
	inputPath = flag.Args()[0]
}

func main() {

	createDirIfNotExists(outputPath)

	sourceFiles := scanMediaDirectory(inputPath, true)
	destFiles := scanMediaDirectory(outputPath, false)

	for k, val := range sourceFiles {
		val.writeToDestination(outputPath, copyDuplicates && destFiles[k] != nil)
	}
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(0)
}
