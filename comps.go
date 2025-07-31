package createrepo

import (
	"encoding/xml"
	"os"
)

// comps represents the (yum)comps repodata.
type comps struct {
	Type         string              `xml:"-"`
	XMLName      xml.Name            `xml:"comps"`
	Group        []*compsGroup       `xml:"group"`
	Category     []*compsCategory    `xml:"category"`
	Environment  []*compsEnvironment `xml:"environment"`
	OpenChecksum *checksum           `xml:"-"`
	OpenSize     uint64              `xml:"-"`
}

// compsGroup represents a single group in the comps.xml.
type compsGroup struct {
	ID          string             `xml:"id,omitempty"`
	Name        []*nameLang        `xml:"name,omitempty"`
	Description []*descriptionLang `xml:"description,omitempty"`
	Default     bool               `xml:"default"`
	Uservisible bool               `xml:"uservisible"`
	PackageList *compsPackageList  `xml:"packagelist,omitempty"`
}

// compsPackageList represents a package list in a Group.
type compsPackageList struct {
	PackageReqs []*compsPackageReq `xml:"packagereq,omitempty"`
}

// compsPackageReq represents a package requirement in a PackageList.
type compsPackageReq struct {
	Type       string `xml:"type,attr,omitempty"`
	PackageReq string `xml:",chardata"`
}

// compsCategory represents a single category in the comps.xml.
type compsCategory struct {
	ID           string             `xml:"id,omitempty"`
	Name         []*nameLang        `xml:"name,omitempty"`
	Description  []*descriptionLang `xml:"description,omitempty"`
	DisplayOrder string             `xml:"display_order,omitempty"`
	GroupList    *compsGroupList    `xml:"grouplist,omitempty"`
}

// compsGroupList represents a group list in a Category.
type compsGroupList struct {
	GroupID []*compsGroupID `xml:"groupid,omitempty"`
}

// compsGroupID represents a Group ID in a GroupList.
type compsGroupID struct {
	//	Type       string `xml:"type,attr,omitempty"`
	GroupIDEntry string `xml:",chardata"`
}

// compsEnvironment represents a single environment in the comps.xml.
type compsEnvironment struct {
	ID           string             `xml:"id,omitempty"`
	Name         []*nameLang        `xml:"name,omitempty"`
	Description  []*descriptionLang `xml:"description,omitempty"`
	DisplayOrder string             `xml:"display_order,omitempty"`
	GroupList    *compsGroupList    `xml:"grouplist,omitempty"`
	OptionList   *compsGroupList    `xml:"optionlist,omitempty"`
}

// nameLang represents the name in a different language.
type nameLang struct {
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Name string `xml:",chardata"`
}

// descriptionLang represents the description in a different language.
type descriptionLang struct {
	Lang        string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Description string `xml:",chardata"`
}

// XML formats the comps to XML
func (c *comps) XML() ([]byte, error) {
	return xmlencode(c)
}

func (c *comps) String() string {
	b, err := c.XML()
	if err != nil {
		return ""
	}

	return string(b)
}

// readComps reads a comps file and returns its content in
// a Comps struct. If the file doesn't exist, an error will be returned.
func readComps(name string) (*comps, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	comps := &comps{Type: "group"}

	if err := xml.Unmarshal(content, comps); err != nil {
		return nil, err
	}

	compsXML, err := comps.XML()
	if err != nil {
		return nil, err
	}
	comps.OpenSize = uint64(len(compsXML))
	comps.OpenChecksum = getChecksumOfBytes(compsXML)

	return comps, nil
}

// writeData returns the metadata and file content of the comps section
func (c *comps) writeData(baseDir, compressAlgo string) (*data, error) {
	x, err := c.XML()
	if err != nil {
		return nil, err
	}
	compressed, checksum, suffix, err := compress(x, compressAlgo)
	if err != nil {
		return nil, err
	}

	size := uint64(len(compressed))
	href := repoDataDir + "/" + checksum.Data + "-comps.xml" + suffix
	modTime, err := writeFile(baseDir+"/"+href, compressed)
	if err != nil {
		return nil, err
	}

	data := &data{
		Type:         "group",
		Checksum:     checksum,
		OpenChecksum: c.OpenChecksum,
		Location:     &location{Href: href},
		Size:         &size,
		OpenSize:     &c.OpenSize,
		Timestamp:    &modTime,
	}

	return data, nil
}
