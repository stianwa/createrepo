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

	// Check to see if primary, filelists or group has changed
	// from the old version. If it hasn't, leave everything
	// unchanged.
	if oldRepoMD != nil {
		same := true
		primary := false
		filelists := false
		for _, d := range oldRepoMD.Data {
			switch d.Type {
			case "primary":
				same = same && d.sameChecksumAndExists(repoData.primary.OpenChecksum, r.baseDir)
				primary = true
			case "filelists":
				same = same && d.sameChecksumAndExists(repoData.fileLists.OpenChecksum, r.baseDir)
				filelists = true
			case "group":
				if repoData.comps != nil {
					same = same && d.sameChecksumAndExists(repoData.comps.OpenChecksum, r.baseDir)
				}
			}
		}
		if same && primary && filelists {
			// repo content is the same - do nothing
			return summary, nil
		}
	}

	repomd, err := repoData.writeData(r.baseDir, r.config.CompressAlgo)
	if err != nil {
		return nil, fmt.Errorf("write meta: %v", err)
	}

	if err := repomd.Write(); err != nil {
		return nil, err
	}

	hist, err := readHistory(r.baseDir)
	if err != nil {
		return nil, err
	}
	if hist == nil {
		hist = newHistory(r.baseDir)
	}

	hist.Append(repomd)
	if err := hist.write(); err != nil {
		return nil, err
	}

	if hist.Clean(r.config.ExpungeOldMetadata); err != nil {
		return summary, err
	}

	return summary, nil
}
