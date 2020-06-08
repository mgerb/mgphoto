package common

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var exiftoolExecutable = "exiftool"

// check if exiftool is installed
func init() {
	checkForExifToolInstallation()
}

func checkForExifToolInstallation() {
	cmd := exec.Command(exiftoolExecutable)
	err := cmd.Run()
	if err != nil {
		println("----------------------------------------")
		println("It looks like Exiftool is not installed. For more accurate timestamp readings,\nit is recommended to install exiftool and make sure it exists in your path: https://exiftool.org/install.html")
		println("----------------------------------------\n")
	}
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
		tags[k] = v
	}

	return tags, nil
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
