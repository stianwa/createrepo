package createrepo

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// Config represents a configuration for repo.
type Config struct {
	// CompressAlgo specifies which compression algorithm to be
	// used for compressing the meta files. Supported algorithms
	// are: xz (default) and gz.
	CompressAlgo string `yaml:"compressAlgo,omitempty"`

	// CompsFile specifies a path to a comps group (yumgroup)
	// file, if used.
	CompsFile string `yaml:"compsFile,omitempty"`

	// ExpungeOldMetadata specifies the time in seconds when old
	// metadata should be deleted from disk and history. The
	// default is 172800 (48 hours).
	ExpungeOldMetadata int64 `yaml:"expungeOldMetadata"`

	// WriteConfig writes this Config to disk.
	WriteConfig bool `yaml:"-"`
}

// readConfig reads a configuration from file. It is ok if the file
// does not exists. In that case both the Config and error will be
// returned as nil.
func readConfig(baseDir string) (*Config, error) {
	content, err := os.ReadFile(baseDir + "/" + configYAML)
	if err != nil {
		return nil, err
	}

	c := &Config{}

	decoder := yaml.NewDecoder(bytes.NewReader(content))

	// make sure we don't have invalid fields in configuration
	decoder.KnownFields(true)

	// parse configuration data
	err = decoder.Decode(c)
	if err != nil {
		return nil, fmt.Errorf("config %q: %v", baseDir+"/"+configYAML, err)
	}

	return c, nil
}

// write writes the configuration to disk.
func (c *Config) write(baseDir string) error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if _, err := writeFile(baseDir+"/"+configYAML, content); err != nil {
		return err
	}

	return nil
}
