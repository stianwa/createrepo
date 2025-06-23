package createrepo

import (
	"encoding/xml"
	"os"
	"time"
)

// repoMD represents the main repodata file repomd.xml
type repoMD struct {
	baseDir      string   `xml:"-"`
	XMLName      xml.Name `xml:"repomd"`
	NameSpace    string   `xml:"http://linux.duke.edu/metadata/repo xmlns,attr"`
	NameSpaceRPM string   `xml:"http://linux.duke.edu/metadata/rpm rpm,attr,omitempty"`
	Revision     float64  `xml:"revision"`
	Data         []*data  `xml:"data"`
}

func (r *repoMD) String() string {
	b, err := r.XML()
	if err != nil {
		return ""
	}

	return string(b)
}

// XML formats the repoMD to XML
func (r *repoMD) XML() ([]byte, error) {
	return xmlencode(r)
}

// Write repomd.xml to disk
func (r *repoMD) Write() error {
	content, err := r.XML()
	if err != nil {
		return err
	}

	if _, err := writeFile(r.baseDir+"/"+repoMDXML, content); err != nil {
		return err
	}

	return nil
}

// readRepoMD returns a RepoMD from the current repomd.xml if it
// exists. If not, a nill is returned. Upon other errors, the error
// will be set.
func (r *Repo) readRepoMD() (*repoMD, error) {
	content, err := os.ReadFile(r.baseDir + "/" + repoMDXML)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	repomd := &repoMD{baseDir: r.baseDir}

	if err := xml.Unmarshal(content, repomd); err != nil {
		return nil, err
	}

	return repomd, nil
}

// newRepoMD returns a new *RepoMD, the named file will be used when
// writing to disk.
func newRepoMD(baseDir string) *repoMD {
	return &repoMD{
		baseDir:  baseDir,
		Revision: float64(time.Now().Unix()),
	}
}
