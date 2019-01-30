package main

// MediaFile - contains file information
type MediaFile struct {
	name string
	md5  string
}

// NewMediaFile - generate new file and process meta data
func NewMediaFile(data []byte) *MediaFile {

	// TODO:
	// file, err := ioutil.ReadFile(filePath)

	return &MediaFile{}
}

func (m *MediaFile) processMetaData() {

}
