// Package createrepo provides methods for creating and maintaining an RPM repository.
package createrepo

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// repoDataDir is the name of the directory for storing metadata
	repoDataDir = "repodata"

	// repoMDXML is the name of the main repodata file
	repoMDXML = repoDataDir + "/repomd.xml"

	// historyXML is the name of the file storing historic
	// repomd.xml entries for clean ups.
	historyXML = repoDataDir + "/.history.xml"

	// configYAML is the name of the configuration file. If the
	// file doesn't exists, a new will be created.
	configYAML = repoDataDir + "/.config.yaml"
)

// Summary represents the Create summary.
type Summary struct {
	Dir     string
	RPMs    int
	Updated bool
	Expunged int
}

func (s *Summary) String() string {
	return fmt.Sprintf("repo:%s rpms:%d expunged:%d repomd:%t", s.Dir, s.RPMs, s.Expunged, s.Updated)
}

// Repo represents the repo handler.
type Repo struct {
	baseDir string
	config  *Config
}

// NewRepo returns a new repo handler. The directory is mandatory, and
// must exists. If *Config is nil, the config file in the repodata
// will be read. If the file doesn't exist, a new default Config will
// be created and saved to disk.
func NewRepo(dir string, config *Config) (*Repo, error) {
	if dir == "" {
		return nil, fmt.Errorf("dir is missing")
	}

	baseDir := filepath.Clean(dir)

	fi, err := os.Stat(baseDir)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", baseDir)
	}

	if config == nil {
		c, err := readConfig(baseDir)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if c != nil {
			config = c
		} else {
			config = &Config{WriteConfig: true}
		}
	}
	if config.WriteConfig {
		if config.ExpungeOldMetadata == 0 {
			config.ExpungeOldMetadata = 172800
		}
		if config.CompsFile != "" {
			a, err := filepath.Abs(config.CompsFile)
			if err != nil {
				return nil, err
			}
			config.CompsFile = a
		}
	}

	switch config.CompressAlgo {
	case "xz", "gz": // Supported
	case "":
		config.CompressAlgo = "xz" // Default
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", config.CompressAlgo)
	}

	return &Repo{baseDir: baseDir, config: config}, nil
}
