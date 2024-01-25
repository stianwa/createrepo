// Package createrepo provides methods for creating and maintaining an RPM repository.
package createrepo

import (
	"fmt"
	"os"
	"path"
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
}

func (s *Summary) String() string {
	if s.Updated {
		return fmt.Sprintf("%s: %d rpms, metadata updated", s.Dir, s.RPMs)
	}
	return fmt.Sprintf("%s: %d rpms, metadata not changed", s.Dir, s.RPMs)
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
	baseDir := path.Clean(dir)

	if baseDir == "" {
		return nil, fmt.Errorf("dir is missing")
	}

	fi, err := os.Stat(baseDir)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", baseDir)
	}

	if config == nil {
		c, err := readConfig(baseDir)
		if err != nil {
			return nil, err
		}
		if c != nil {
			config = c
		} else {
			config = &Config{WriteConfig: true, ExpungeOldMetadata: 172800}
		}
	}

	switch config.CompressAlgo {
	case "xz", "gz": // Supported
	case "":
		config.CompressAlgo = "xz" // Default
	default:
		return nil, fmt.Errorf("bad compression algorithm: %s", config.CompressAlgo)
	}

	return &Repo{baseDir: baseDir, config: config}, nil
}