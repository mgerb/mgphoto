package main

import (
	"crypto/sha1"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

// MediaFile - contains file information
type MediaFile struct {
	name string
	path string
	date *time.Time
	sha1 [20]byte
}

// NewMediaFile - generate new file and process meta data
// returns nil if file cannot be handled
func NewMediaFile(path string, processMetaData bool) *MediaFile {

	file, err := os.Open(path)

	if err != nil {
		log.Println(err)
		return nil
	}

	defer file.Close()

	// read bytes from file
	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		log.Println(err)
		return nil
	}

	mediaFile := &MediaFile{
		path: path,
		name: filepath.Base(file.Name()),
		sha1: sha1.Sum(bytes),
	}

	if processMetaData {
		mediaFile.processMetaData(file)
	}

	return mediaFile
}

func (m *MediaFile) unknownCreation(file *os.File) bool {
	return m.date == nil
}

func (m *MediaFile) processMetaData(file *os.File) {

	// make sure file starts at beginning
	file.Seek(0, 0)

	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(file)

	if err != nil {
		return
	}

	t, err := x.DateTime()

	if err != nil {
		return
	}

	m.date = &t
}

func (m *MediaFile) writeToDestination(dest string, copyDuplicates bool) error {

	dir := dest

	if copyDuplicates {
		dir = path.Join(dest, "duplicates")
	}

	if m.date != nil {
		year := m.date.Format("2006")
		month := m.date.Format("2006-01-02")
		dir = path.Join(dir, year, month)
	} else {
		dir = path.Join(dir, "unknown")
	}

	createDirIfNotExists(dir)

	fullPath := renameIfFileExists(path.Join(dir, m.name))

	err := copyFile(m.path, fullPath)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
