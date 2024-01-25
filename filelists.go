package createrepo

import (
	"encoding/xml"
)

// fileLists represents the filelist file in repoedata
type fileLists struct {
	Type         string         `xml:"-"`
	XMLName      xml.Name       `xml:"filelists"`
	Namespace    string         `xml:"xmlns,attr"`
	Count        string         `xml:"packages,attr"`
	Packages     []*packageList `xml:"package"`
	OpenChecksum *checksum      `xml:"-"`
	OpenSize     uint64         `xml:"-"`
}

// XML formats the fileLists to XML
func (f *fileLists) XML() ([]byte, error) {
	return xmlencode(f)
}

func (f *fileLists) String() string {
	b, err := f.XML()
	if err != nil {
		return ""
	}
	return string(b)
}

// writeData returns the metadata and file content of the fileLists section
func (f *fileLists) writeData(baseDir, compressAlgo string) (*data, error) {
	x, err := f.XML()
	if err != nil {
		return nil, err
	}
	compressed, checksum, suffix, err := compress(x, compressAlgo)
	if err != nil {
		return nil, err
	}

	size := uint64(len(compressed))
	href := repoDataDir + "/" + checksum.Data + "-filelists.xml" + suffix
	modTime, err := writeFile(baseDir+"/"+href, compressed)
	if err != nil {
		return nil, err
	}

	data := &data{
		Type:         "filelists",
		Checksum:     checksum,
		OpenChecksum: f.OpenChecksum,
		Location:     &location{Href: href},
		Size:         &size,
		OpenSize:     &f.OpenSize,
		Timestamp:    &modTime,
	}

	return data, nil
}
