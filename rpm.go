package createrepo

import (
	"fmt"
	"github.com/cavaliergopher/rpm"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// packageList represents the RPM package file list
type packageList struct {
	PkgID   string   `xml:"pkgid,attr"`
	Name    string   `xml:"name,attr"`
	Arch    string   `xml:"arch,attr"`
	Version *version `xml:"version"`
	Files   []*file  `xml:"file"`
}

// file represents a file in the RPM package file list
type file struct {
	Type string `xml:"type,attr,omitempty"`
	Path string `xml:",chardata"`
}

// rpmPackage represents the RPM package in primary meta
type rpmPackage struct {
	Type        string    `xml:"type,attr"`
	Name        string    `xml:"name"`
	Arch        string    `xml:"arch"`
	Version     *version  `xml:"version"`
	Checksum    *checksum `xml:"checksum"`
	Summary     string    `xml:"summary"`
	Description string    `xml:"description"`
	Packager    string    `xml:"packager"`
	URL         string    `xml:"url"`
	Time        *tm       `xml:"time,omitempty"`
	Size        *size     `xml:"size"`
	Location    *location `xml:"location"`
	Format      *format   `xml:"format"`
}

// tm represents file and build times
type tm struct {
	File  string `xml:"file,attr,omitempty"`
	Build string `xml:"build,attr,omitempty"`
}

// size represents package sizes
type size struct {
	Package   string `xml:"package,attr,omitempty"`
	Installed string `xml:"installed,attr,omitempty"`
	Archive   string `xml:"archive,attr,omitempty"`
}

// version represents the NEVR in the RPM
type version struct {
	Epoch   int    `xml:"epoch,attr"`
	Version string `xml:"ver,attr"`
	Release string `xml:"rel,attr"`
}

// checksum represents the file's checksum
type checksum struct {
	Type  string `xml:"type,attr"`
	PkgID string `xml:"pkgid,attr,omitempty"`
	Data  string `xml:",chardata"`
}

func (c *checksum) String() string {
	return c.Type + " " + c.Data
}

// location represents the file path relative to repository directory
type location struct {
	Href string `xml:"href,attr"`
}

// format represents package format meta
type format struct {
	License   *license   `xml:"rpm:license"`
	Vendor    *vendor    `xml:"rpm:vendor"`
	Group     *group     `xml:"rpm:group"`
	BuildHost *buildHost `xml:"rpm:buildhost"`
	SourceRPM *sourceRPM `xml:"rpm:sourcerpm"`
	Provides  []*entry   `xml:"rpm:provides>rpm:entry,omitempty"`
	Requires  []*entry   `xml:"rpm:requires>rpm:entry,omitempty"`
	Obsoletes []*entry   `xml:"rpm:obsoletes>rpm:entry,omitempty"`
}

// license represents package license
type license struct {
	License string `xml:",chardata"`
}

// vendor represents package vendor
type vendor struct {
	Vendor string `xml:",chardata"`
}

// group represents package group
type group struct {
	Group string `xml:",chardata"`
}

// buildHost represents package buildhost
type buildHost struct {
	BuildHost string `xml:",chardata"`
}

// sourceRPM represents package sourcerpm
type sourceRPM struct {
	SourceRPM string `xml:",chardata"`
}

// entry represents attributes to an entry
type entry struct {
	Name    string `xml:"name,attr,omitempty"`
	Flags   string `xml:"flags,attr,omitempty"`
	Epoch   string `xml:"epoch,attr,omitempty"`
	Version string `xml:"ver,attr,omitempty"`
	Release string `xml:"rel,attr,omitempty"`
	Pre     string `xml:"pre,attr,omitempty"`
}

func getPackage(dir, name string) (*rpmPackage, *packageList, error) {
	path := filepath.Clean(dir + "/" + name)

	fi, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	var checksum *checksum
	if c, ok := getXattrChecksum(path); ok {
		checksum = c
	} else if c, err := getChecksumOfFile(path); err == nil {
		checksum = c
		if err := setXattrChecksum(path, checksum); err != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}

	pkg, err := rpm.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %v", path,err)
	}

	group := &group{Group: "Unspecified"}
	groups := pkg.Groups()
	if len(groups) > 0 {
		group.Group = groups[0]
	}

	provides, providesMap := getDependencies(pkg.Provides(), nil)
	requires, _ := getDependencies(pkg.Requires(), providesMap)
	obsoletes, _ := getDependencies(pkg.Obsoletes(), nil)

	p := &rpmPackage{
		Type: "rpm",
		Name: pkg.Name(),
		Arch: pkg.Architecture(),
		Version: &version{
			Epoch:   pkg.Epoch(),
			Version: pkg.Version(),
			Release: pkg.Release(),
		},
		Summary:     pkg.Summary(),
		Description: pkg.Description(),
		Packager:    pkg.Packager(),
		URL:         pkg.URL(),
		Time: &tm{
			File:  fmt.Sprintf("%d", fi.ModTime().Unix()),
			Build: fmt.Sprintf("%d", pkg.BuildTime().Unix()),
		},
		Checksum: checksum,
		Location: &location{Href: name},
		Size: &size{
			Package:   fmt.Sprintf("%d", fi.Size()),
			Installed: fmt.Sprintf("%d", pkg.Size()),
			Archive:   fmt.Sprintf("%d", pkg.ArchiveSize()),
		},
		Format: &format{
			License:   &license{License: pkg.License()},
			Vendor:    &vendor{Vendor: pkg.Vendor()},
			Group:     group,
			BuildHost: &buildHost{BuildHost: pkg.BuildHost()},
			SourceRPM: &sourceRPM{SourceRPM: pkg.SourceRPM()},
			Provides:  provides,
			Requires:  requires,
			Obsoletes: obsoletes,
		},
	}

	var files []*file
	for _, pfile := range pkg.Files() {
		f := &file{
			Path: pfile.Name(),
		}
		if pfile.IsDir() {
			f.Type = "dir"
		}
		files = append(files, f)
	}
	f := &packageList{
		Name:  pkg.Name(),
		Arch:  pkg.Architecture(),
		PkgID: checksum.Data,
		Version: &version{
			Epoch:   pkg.Epoch(),
			Version: pkg.Version(),
			Release: pkg.Release(),
		},
		Files: files,
	}

	return p, f, nil
}

func getDependencies(deps []rpm.Dependency, provides map[entry]bool) ([]*entry, map[entry]bool) {
	var ents []*entry

	if provides == nil {
		provides = make(map[entry]bool)
	}

	m := make(map[entry]bool)

	// keep only highest version of libc version
	var libc rpm.Dependency

	for _, d := range deps {
		// Compare versins if dependency starts with "libc.so.6"
		if strings.HasPrefix(d.Name(), "libc.so.6") {
			if libc == nil {
				libc = d
				continue
			}
			n := compareLibC(libc.Name(), d.Name())
			if n == 0 || n == 1 {
				continue
			} else if n == 2 {
				libc = d
				continue
			}
			// n must now be -1 (error) include this dependency
		}

		// Skip names beginning with 'rpmlib('
		if strings.HasPrefix(d.Name(), "rpmlib(") {
			continue
		}

		var epoch, ver, rel string
		ver = d.Version()
		if ver != "" {
			if n := strings.Index(ver, ":"); n > 0 {
				epoch = ver[0:n]
				ver = ver[n+1:]
			}
			if n := strings.Index(ver, "-"); n > 1 {
				rel = ver[n+1:]
				ver = ver[0:n]
			} else {
				rel = d.Release()
			}
			if epoch == "" && d.Epoch() != 0 {
				epoch = fmt.Sprintf("%d", d.Epoch())
			}
		}

		fmt.Printf("e:%s v:%s r:%s\n", epoch, ver, rel)
		
		flags, pre := getFlag(d.Flags())

		c := &entry{
			Name:    d.Name(),
			Epoch:   epoch,
			Version: ver,
			Release: rel,
			Flags:   flags,
			Pre:     pre,
		}

		// Skip duplicates
		if _, ok := m[*c]; ok {
			continue
		}

		// Skip own provides
		if _, ok := provides[*c]; ok {
			continue
		}

		m[*c] = true
		ents = append(ents, c)
	}

	if libc != nil {
		ents = append(ents, &entry{Name: libc.Name()})
	}

	return ents, m
}

// Return values: 0 - same; 1 - first is bigger; 2 - second is bigger,
// * -1 - error
func compareLibC(c1, c2 string) int {
	if c1 == c2 {
		return 0
	}

	b1 := strings.Index(c1, "(")
	b2 := strings.Index(c2, "(")

	if b1 < 0 && b2 < 0 {
		return 0
	}

	if b1 < 0 {
		return 2
	}
	if b2 < 0 {
		return 1
	}

	g1, ok1 := readParenthesis(c1[b1:])
	g2, ok2 := readParenthesis(c2[b2:])

	if !(ok1 && ok2) ||
		len(g1) == 0 || len(g2) == 0 ||
		len(g1) > 2 || len(g2) > 2 ||
		len(g1) == 2 && g1[1] != "64bit" ||
		len(g2) == 2 && g2[1] != "64bit" {
		return -1
	}

	if len(g1) == 0 && len(g2) == 0 ||
		len(g1) == 1 && len(g2) == 1 && g1[0] == g2[0] ||
		len(g1) == 2 && len(g2) == 2 && g1[0] == g2[0] && g1[1] == g2[1] {
		return 0
	}

	first := g1[0]
	second := g2[0]

	if first == second {
		return 0
	}

	if !(strings.HasPrefix(first, "GLIBC_") || first == "") ||
		!(strings.HasPrefix(second, "GLIBC_") || second == "") {
		return -1
	}

	if first == "" && second != "" {
		return 2
	} else if first != "" && second == "" {
		return 1
	}

	// Loose GLIBC_
	first = first[6:]
	second = second[6:]

	c := rpmcmp(first, second)
	if c == -1 {
		c = 2
	}

	return c
}

func rpmcmp(s1, s2 string) int {
	if s1 == s2 {
		return 0
	}

	// Split strings into number/alpha groups
	s := [2][]string{}
	for i, st := range []string{s1, s2} {
		buf := ""
		for _, c := range st {
			if buf == "" {
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
					buf = buf + string(c)
				} else {
					s[i] = append(s[i], buf)
				}
			} else if buf[0] >= '0' && buf[0] <= '9' {
				if c >= '0' && c <= '9' {
					buf = buf + string(c)
				} else if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
					s[i] = append(s[i], buf)
					buf = string(c)
				} else {
					s[i] = append(s[i], buf)
					buf = ""
				}
			} else if (buf[0] >= 'A' && buf[0] <= 'Z') || (buf[0] >= 'a' && buf[0] <= 'z') {
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
					buf = buf + string(c)
				} else if c >= '0' && c <= '9' {
					s[i] = append(s[i], buf)
					buf = string(c)
				} else {
					s[i] = append(s[i], buf)
					buf = ""
				}
			} else {
				s[i] = append(s[i], buf)
				buf = ""
			}
		}
		if buf != "" {
			s[i] = append(s[i], buf)
		}
	}

	min := 0
	if len(s[0]) < len(s[1]) {
		min = len(s[0])
	} else {
		min = len(s[1])
	}

	for i := 0; i < min; i++ {
		l, r := s[0][i], s[1][i]
		if l == r {
			continue
		}
		if l == "" {
			return 1
		}
		if r == "" {
			return -1
		}
		if l[0] >= '0' && l[0] <= '9' && r[0] >= '0' && r[0] <= '9' {
			li, _ := strconv.ParseUint(l, 10, 64)
			ri, _ := strconv.ParseUint(r, 10, 64)
			if li < ri {
				return -1
			} else if li > ri {
				return 1
			}
		} else if ((l[0] >= 'A' && l[0] <= 'Z') || (l[0] >= 'a' && l[0] <= 'z')) &&
			((r[0] >= 'A' && r[0] <= 'Z') || (r[0] >= 'a' && r[0] <= 'z')) {
			if l < r {
				return -1
			} else if l > r {
				return 1
			}
		} else if ((l[0] >= 'A' && l[0] <= 'Z') || (l[0] >= 'a' && l[0] <= 'z')) &&
			(r[0] >= '0' && r[0] <= '9') {
			return -1
		} else {
			return 1
		}
	}

	if len(s[0]) == len(s[1]) {
		return 0
	} else if len(s[0]) < len(s[1]) {
		return -1
	}

	return 1
}

func readParenthesis(s string) ([]string, bool) {
	var in bool
	var group string
	var ret []string
	for _, r := range s {
		if in {
			if r == '(' {
				return nil, false
			} else if r == ')' {
				ret = append(ret, group)
				group = ""
				in = false
			} else {
				group = group + string(r)
			}
		} else {
			if r != '(' {
				return nil, false
			}
			in = true
		}
	}

	return ret, !in
}

func getFlag(f int) (string, string) {
	var flag, pre string

	switch {
	case rpm.DepFlagLesserOrEqual == f&rpm.DepFlagLesserOrEqual:
		flag = "LE"
	case rpm.DepFlagGreaterOrEqual == f&rpm.DepFlagGreaterOrEqual:
		flag = "GE"
	case rpm.DepFlagLesser == f&rpm.DepFlagLesser:
		flag = "LT"
	case rpm.DepFlagGreater == f&rpm.DepFlagGreater:
		flag = "GT"
	case rpm.DepFlagEqual == f&rpm.DepFlagEqual:
		flag = "EQ"
	}

	if (rpm.DepFlagPrereq == f&rpm.DepFlagPrereq) ||
		(rpm.DepFlagScriptPre == f&rpm.DepFlagScriptPre) ||
		(rpm.DepFlagScriptPost == f&rpm.DepFlagScriptPost) {
		pre = "1"
	}

	return flag, pre
}

func bits(n int) string {
	var ret []string

	i := 0
	for i < 64 {
		if n&1 == 1 {
			ret = append(ret, fmt.Sprintf("%d", i))
		}
		n = n >> 1
		i++
	}

	return strings.Join(ret, ", ")
}

// getRPMFiles return a list with all files with suffix .rpm
// within the directory root
func getRPMFiles(baseDir string) ([]string, error) {
	var ls []string
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".rpm") {
			path, _ = strings.CutPrefix(path, baseDir)
			ls = append(ls, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ls, nil
}
