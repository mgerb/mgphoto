package common

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/tabwriter"
)

func initLogger() *tabwriter.Writer {
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

	return wr
}
