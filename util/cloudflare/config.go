package cloudflare

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/frankbraun/codechain/util/file"
)

// ConfigFilename defines the default filename for Cloudflare API Config files.
const ConfigFilename = "cloudflare.json"

// A Config for the Cloudflare API.
type Config struct {
	APIKey string // API key generated on the "My Account" page
	Email  string // Email address associated with Cloudflare account
}

// ReadConfig from filename.
func ReadConfig(filename string) (*Config, error) {
	var c Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Write config to filename.
func (c *Config) Write(filename string) error {
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("configuration file '%s' exists already", filename)
	}
	jsn, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, jsn, 0644)
}
