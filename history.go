package createrepo

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"
)

// history represents the .history.xml content. Createrepo uses this
// content for cleaning up older versions.
type history struct {
	baseDir   string      `xml:"-"`
	XMLName   xml.Name    `xml:"history"`
	Revisions []*revision `xml:"revisions"`
}

// Append appends a repoMD to the history.
func (h *history) Append(r *repoMD) {
	e := &revision{
		Revision: r.Revision,
	}
	e.Data = append(e.Data, r.Data...)

	for _, ex := range h.Revisions {
		if ex.Revision == e.Revision {
			// Revision already exists
			return
		}
	}
	h.Revisions = append(h.Revisions, e)
}

// write writes the history to the named file, creating it if
// necessary.
func (h *history) write() error {
	b, err := h.XML()
	if err != nil {
		return err
	}
	if _, err := writeFile(h.baseDir+"/"+historyXML, b); err != nil {
		return err
	}
	return nil
}

// Clean cleans up revisions older than n seconds. If the current
// *repoMD is passed, it will be spared from the cleanup process.
func (h *history) Clean(seconds int64) (int, error) {
	var lastRevision *revision
	for _, r := range h.Revisions {
		if lastRevision == nil || r.Revision > lastRevision.Revision {
			lastRevision = r
		}
	}
	if lastRevision == nil {
		return 0, fmt.Errorf("no current revision found in history")
	}

	// Bless all data files for lastRevision. One or more data
	// sets might occur in a dirrerent Revision. The group dataset
	// might be the same etc.
	bless := make(map[string]bool)
	for _, data := range lastRevision.Data {
		bless[data.Location.Href] = true
	}

	now := time.Now().Unix()
	expunged := 0
	var newRevs []*revision
	for _, r := range h.Revisions {
		if r.Revision == lastRevision.Revision {
			newRevs = append(newRevs, r)
			continue
		}
		if r.Obsoleted == 0 {
			r.Obsoleted = now
		}

		if now >= r.Obsoleted+seconds {
			expunged++
			for _, data := range r.Data {
				if data.Location != nil && data.Location.Href != "" {
					if _, blessed := bless[data.Location.Href]; !blessed {
						os.Remove(h.baseDir + "/" + data.Location.Href)
					}
				}
			}
		} else {
			newRevs = append(newRevs, r)

		}
	}

	h.Revisions = newRevs
	if err := h.write(); err != nil {
		return 0, err
	}

	return expunged, nil
}

// revision represents a single repoMD in the .history.xml file.
type revision struct {
	Obsoleted int64   `xml:"obsoleted,omitempty"`
	Revision  float64 `xml:"revision"`
	Data      []*data `xml:"data"`
}

func (h *history) String() string {
	b, err := h.XML()
	if err != nil {
		return ""
	}

	return string(b)
}

// XML returns the history in XML format
func (h *history) XML() ([]byte, error) {
	b, err := xml.MarshalIndent(h, "", "  ")
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')

	return append([]byte(xmlHeader), b...), nil
}

// readhistory reads the .history.xml file and returns its content in
// a history struct. If the file doesn't exist, a nil pointer will be
// returned.
func readHistory(baseDir string) (*history, error) {
	content, err := os.ReadFile(baseDir + "/" + historyXML)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	history := &history{baseDir: baseDir}

	if err := xml.Unmarshal(content, history); err != nil {
		return nil, err
	}

	return history, nil
}

// newHistory returns a new *history, the named file will be used when
// writing history to disk.
func newHistory(baseDir string) *history {
	return &history{
		baseDir: baseDir,
	}
}
