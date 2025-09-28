package createrepo

import (
	"fmt"
	"os"
)

// data represents a data set in repomd.xml
type data struct {
	Type         string    `xml:"type,attr"`
	Checksum     *checksum `xml:"checksum,omitempty"`
	OpenChecksum *checksum `xml:"open-checksum,omitempty"`
	Location     *location `xml:"location"`
	Timestamp    *uint64   `xml:"timestamp"`
	Size         *uint64   `xml:"size"`
	OpenSize     *uint64   `xml:"open-size,omitempty"`
}

// sameChecksumAndExists returns true if the two data elements have the same
// type, checksums and that their href exists.
func (d *data) sameChecksumAndExists(checksum *checksum, baseDir string) bool {
	if checksum == nil || d.OpenChecksum == nil ||
		d.OpenChecksum.Data != checksum.Data {
		return false
	}

	_, err := os.Stat(baseDir + "/" + d.Location.Href)

	return err == nil
}

// dataSet represents all the repository data gathered from the
// RPMs
type dataSet struct {
	primary   *primary
	fileLists *fileLists
	comps     *comps
}

// writeData writes meta data to disk and returns an repoMD upon success
func (r *dataSet) writeData(baseDir, compressAlgo string) (*repoMD, error) {
	ret := newRepoMD(baseDir)

	cleanUp := true
	defer func() {
		if cleanUp {
			for _, c := range ret.Data {
				os.Remove(c.Location.Href)
			}
		}
	}()

	pmeta, err := r.primary.writeData(baseDir, compressAlgo)
	if err != nil {
		return nil, err
	}
	ret.Data = append(ret.Data, pmeta)

	fmeta, err := r.fileLists.writeData(baseDir, compressAlgo)
	if err != nil {
		return nil, err
	}
	ret.Data = append(ret.Data, fmeta)

	if r.comps != nil {
		cmeta, err := r.comps.writeData(baseDir, compressAlgo)
		if err != nil {
			return nil, err
		}
		ret.Data = append(ret.Data, cmeta)
	}

	cleanUp = false

	return ret, nil
}

// getData returns datasets for primary, filelists and comps (if specified)
func (r *Repo) getData() (*dataSet, error) {

	ls, err := getRPMFiles(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("getRPMFileNames: %v", err)
	}

	var packages []*rpmPackage
	var files []*packageList
	for _, name := range ls {
		p, f, err := getPackage(r.baseDir, name)
		if err != nil {
			return nil, fmt.Errorf("getPackage: %s: %v", name, err)
		}
		packages = append(packages, p)
		files = append(files, f)
	}

	meta := &dataSet{
		primary: &primary{
			Type:         "primary",
			Namespace:    "http://linux.duke.edu/metadata/common",
			NamespaceRPM: "http://linux.duke.edu/metadata/rpm",
			Count:        fmt.Sprintf("%d", len(packages)),
			Packages:     packages},
		fileLists: &fileLists{
			Type:      "filelists",
			Namespace: "http://linux.duke.edu/metadata/filelists",
			Count:     fmt.Sprintf("%d", len(packages)),
			Packages:  files,
		},
	}

	if r.config.CompsFile != "" {
		c, err := readComps(r.config.CompsFile)
		if err != nil {
			return nil, fmt.Errorf("comps: %v", err)
		}
		meta.comps = c
	}

	if b, err := meta.primary.XML(); err == nil {
		meta.primary.OpenChecksum = getChecksumOfBytes(b)
		meta.primary.OpenSize = uint64(len(b))
	} else {
		return nil, err
	}

	if b, err := meta.fileLists.XML(); err == nil {
		meta.fileLists.OpenChecksum = getChecksumOfBytes(b)
		meta.fileLists.OpenSize = uint64(len(b))
	} else {
		return nil, err
	}

	return meta, nil
}
