package createrepo

import (
	"encoding/xml"
)

// primary represents the primary repodata
type primary struct {
	Type         string        `xml:"-"`
	XMLName      xml.Name      `xml:"metadata"`
	Namespace    string        `xml:"xmlns,attr"`
	NamespaceRPM string        `xml:"xmlns:rpm,attr"`
	Count        string        `xml:"packages,attr"`
	Packages     []*rpmPackage `xml:"package"`
	OpenChecksum *checksum     `xml:"-"`
	OpenSize     uint64        `xml:"-"`
}

// XML formats the primary to XML
func (p *primary) XML() ([]byte, error) {
	return xmlencode(p)
}

func (p *primary) String() string {
	b, err := p.XML()
	if err != nil {
		return ""
	}
	return string(b)
}

// writeData returns the metadata and file content of the primary
func (p *primary) writeData(baseDir, compressAlgo string) (*data, error) {
	x, err := p.XML()
	if err != nil {
		return nil, err
	}
	compressed, checksum, suffix, err := compress(x, compressAlgo)
	if err != nil {
		return nil, err
	}

	size := uint64(len(compressed))
	href := repoDataDir + "/" + checksum.Data + "-primary.xml" + suffix
	modTime, err := writeFile(baseDir+"/"+href, compressed)
	if err != nil {
		return nil, err
	}

	data := &data{
		Type:         "primary",
		Checksum:     checksum,
		OpenChecksum: p.OpenChecksum,
		Location:     &location{Href: href},
		Size:         &size,
		OpenSize:     &p.OpenSize,
		Timestamp:    &modTime,
	}

	return data, nil
}
