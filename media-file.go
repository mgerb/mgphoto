package main

import (
	"bytes"
	"crypto/sha1"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"gopkg.in/djherbis/times.v1"
)

// MediaFile - contains file information
type MediaFile struct {
	name     string
	path     string
	date     *time.Time
	sha1     [20]byte
	filetype string
	size     int64
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

	fi, err := file.Stat()
	if err != nil {
		log.Println(path, "not accessible")
		return nil
	}

	startSHA := time.Now()
	bytes := make([]byte, 4000000)

	// Only read the first 4 MB of large files
	if fi.Size() > 4000000 {
		if _, err = io.ReadFull(file, bytes); err != nil {
			log.Println(err)
			return nil
		}
	} else {
		// read bytes from file
		if bytes, err = ioutil.ReadAll(file); err != nil {
			log.Println(err)
			return nil
		}
	}
	if analyze {
		timeTrack(startSHA, "SHA Generation")
	}

	mediaFile := &MediaFile{
		path: path,
		name: filepath.Base(file.Name()),
		sha1: sha1.Sum(bytes),
		size: fi.Size(),
	}

	if processMetaData {
		mediaFile.processMetaData(file)
	}

	return mediaFile
}

func (m *MediaFile) unknownCreation(file *os.File) bool {
	return m.date == nil
}

func (m *MediaFile) isPhoto() bool {
	return isPhoto(m.path)
}

func (m *MediaFile) isVideo() bool {
	return isVideo(m.path)
}

func (m *MediaFile) isSidecar() bool {
	return isSidecar(m.path)
}

func (m *MediaFile) Info(input ...string) {
	var wrap []interface{} = make([]interface{}, len(input)+1)
	wrap[0] = m.path + "\t"
	for i, d := range input {
		wrap[i+1] = d
	}
	Info.Println(wrap...)
}

func (m *MediaFile) Warn(input ...string) {
	var wrap []interface{} = make([]interface{}, len(input)+1)
	wrap[0] = m.path + "\t"
	for i, d := range input {
		wrap[i+1] = d
	}
	Warn.Println(wrap...)
}

func (m *MediaFile) Error(input ...string) {
	var wrap []interface{} = make([]interface{}, len(input)+1)
	wrap[0] = m.path + "\t"
	for i, d := range input {
		wrap[i+1] = d
	}
	Error.Println(wrap)
}

func (m *MediaFile) processMetaData(file *os.File) {
	if analyze {
		defer timeTrack(time.Now(), "EXIF analysis")
	}
	// fmt.Println(m.path)

	var d *time.Time
	if m.isVideo() {
		d = m.getExifDateExifTool()
	}

	if m.isPhoto() {
		d = getExifDate(file)
		if d == nil {
			d = m.getExifDateExifTool()
		}
	}

	// No Exif Data found
	if d == nil {
		m.Warn("No EXIF data found, using file mod time")
		d = m.getFileTime()
	}

	if d == nil {
		m.Error("unable to find date")
	}

	m.date = d
}

func (m *MediaFile) getFileTime() *time.Time {
	t, err := times.Stat(m.path)
	if err != nil {
		log.Fatal(err.Error())
	}

	if t.HasBirthTime() {
		cr := t.BirthTime()
		mod := t.ModTime()
		if cr.Before(mod) {
			return &cr
		} else {
			return &mod
		}
	} else {
		d := t.ModTime()
		return &d
	}
}

func (m *MediaFile) getExifDateExifTool() *time.Time {
	tags, err := getTagsViaExifTool(m.path)

	if err != nil {
		return nil
	}
	date, err := getExifCreateDateFromTags(tags)
	if err != nil {
		return nil
	}
	return &date
}

func getTagsViaExifTool(file string) (map[string]string, error) {
	var out bytes.Buffer
	cmd := exec.Command("exiftool", file)

	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	tags := make(map[string]string)

	data := strings.Trim(out.String(), " \r\n")
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		k, v := strings.Replace(strings.TrimSpace(line[0:32]), " ", "", -1), strings.TrimSpace(line[33:])
		// k = normalizeEXIFTag(k)
		tags[k] = v
	}

	return tags, nil
}

func getExifDate(file *os.File) *time.Time {
	// make sure file starts at beginning
	file.Seek(0, 0)

	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(file)

	if err != nil {
		return nil
	}

	t, err := x.DateTime()

	if err != nil {
		return nil
	}

	return &t
}

// getExifCreateDate attempts to get the given file's original creation date
// from its EXIF tags.
func getExifCreateDateFromTags(tags map[string]string) (time.Time, error) {
	// Looking for the first tag that sounds like a date.
	dateTimeFields := []string{
		"DateAndTimeOriginal",
		"DateTimeOriginal",
		"Date/TimeOriginal",
		"DateTaken",
		"CreateDate",
		"MediaCreateDate",
		"TrackCreateDate",
		"ModifyDate",
		"FileModificationDateTime",
		"FileAccessDateTime",
		"EncodedDate",
		"TaggedDate",
	}

	toInt := func(s string) (i int) {
		i, _ = strconv.Atoi(s)
		return
	}

	for _, field := range dateTimeFields {
		taken, ok := tags[field]
		if !ok {
			continue
		}

		all := reDateTime.FindAllStringSubmatch(taken, -1)

		if len(all) < 1 || len(all[0]) < 6 {
			return time.Time{}, errMissingCreateTime
		}

		y := toInt(all[0][1])
		if y == 0 {
			continue
		}

		t := time.Date(
			y,
			time.Month(toInt(all[0][2])),
			toInt(all[0][3]),
			toInt(all[0][4]),
			toInt(all[0][5]),
			toInt(all[0][6]),
			0,
			time.Local,
		)

		if t.IsZero() {
			continue
		}

		return t, nil
	}

	return time.Time{}, errMissingCreateTime
}

func (m *MediaFile) writeToDestination(dest string, copyDuplicates bool) error {
	dir := dest

	if copyDuplicates {
		dir = path.Join(dir, "duplicates")
	}

	dir = m.destinationPath(dir)

	createDirIfNotExists(dir)

	fullPath := renameIfFileExists(path.Join(dir, m.name))

	m.Info("copying to\t", fullPath)
	if !dryRun {
		err := copyFile(m.path, fullPath)

		if err != nil {
			m.Error(err.Error())
			return err
		}
	}

	return nil
}

func (m *MediaFile) destinationPath(dest string) string {
	dir := dest

	if m.date != nil {
		year := m.date.Format("2006")
		month := m.date.Format("01")
		day := m.date.Format("02")
		dir = path.Join(dir, year, month, day)
	} else {
		dir = path.Join(dir, "unknown")
	}

	return dir
}

func (m *MediaFile) moveToDestination(dest string, original *MediaFile) error {
	dir := m.destinationPath(dest)

	createDirIfNotExists(dir)

	if path.Join(dir, m.name) == m.path && m.sha1 == original.sha1 {
		m.Info("is already in the correct location")
		return nil
	}

	fullPath := renameIfFileExists(path.Join(dir, m.name))

	m.Info("Moving to\t", fullPath)
	if !dryRun {
		err := os.Rename(m.path, fullPath)

		if err != nil {
			m.Error(err.Error())
			return err
		}

	}

	return nil
}
