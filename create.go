package createrepo

import (
	"fmt"
	"os"
)

// Create creates or updates the epository.
func (r *Repo) Create() (*Summary, error) {
	if fi, err := os.Stat(r.baseDir + "/" + repoDataDir); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("%q exists, but is not a directory", r.baseDir+"/"+repoDataDir)
		}
	} else if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.Mkdir(r.baseDir+"/"+repoDataDir, 0777); err != nil {
			return nil, err
		}
	}

	if r.config.WriteConfig {
		if err := r.config.write(r.baseDir); err != nil {
			return nil, err
		}
	}

	repoData, err := r.getData()
	if err != nil {
		return nil, fmt.Errorf("rpm meta: %v", err)
	}

	oldRepoMD, err := r.readRepoMD()
	if err != nil {
		return nil, fmt.Errorf("repomd: %v", err)
	}

	summary := &Summary{Dir: r.baseDir, RPMs: len(repoData.primary.Packages)}

	hist, err := readHistory(r.baseDir)
	if err != nil {
		return nil, err
	}
	if hist == nil {
		hist = newHistory(r.baseDir)
	}

	// If not the same data content, create new
	if ! r.sameDataContent(oldRepoMD, repoData) {
		repomd, err := repoData.writeData(r.baseDir, r.config.CompressAlgo)
		if err != nil {
			return nil, fmt.Errorf("write meta: %v", err)
		}
		
		if err := repomd.Write(); err != nil {
			return nil, err
		}
		summary.Updated = true

		hist.Append(repomd)
		if err := hist.write(); err != nil {
			return nil, err
		}
	}
	
	expunged, err := hist.Clean(r.config.ExpungeOldMetadata)

	if err != nil {
		return summary, err
	}
	summary.Expunged = expunged
	
	return summary, nil
}

func (r *Repo) sameDataContent(old *repoMD, fresh *dataSet) bool {
	if old == nil || fresh == nil {
		return false
	}

	var primary, fileLists, comps *data
	for _, d := range old.Data {
		switch d.Type {
		case "primary":
			primary = d
		case "filelists":
			fileLists = d
		case "group":
			comps = d
		}
	}

	if fresh.primary == nil || primary == nil || fresh.fileLists == nil || fileLists == nil {
		return false
	}

	if ! (comps == nil && fresh.comps == nil || comps != nil && fresh.comps != nil) {
		return false
	}

	if ! (primary.sameChecksumAndExists(fresh.primary.OpenChecksum, r.baseDir) &&
		fileLists.sameChecksumAndExists(fresh.fileLists.OpenChecksum, r.baseDir)) {
		return false
	}
	
	if fresh.comps != nil && ! comps.sameChecksumAndExists(fresh.comps.OpenChecksum, r.baseDir) {
		return false
	}

	return true
}
