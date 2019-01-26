package dyn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
)

// ConfigFilename defines the default filename for Dyn Managed DNS API Config
// files.
const ConfigFilename = "dyn.json"

// A Config for the Dyn Managed DNS API.
type Config struct {
	CustomerName string // customer_name
	UserName     string // user_name
	Password     string // password
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
	if err := ioutil.WriteFile(filename, jsn, 0644); err != nil {
		return err
	}
	return nil
}

// Check that the config is valid.
func (c *Config) Check(zone, fqdn string) error {
	s, err := NewWithConfig(c)
	if err != nil {
		return err
	}
	defer s.Close()
	// create TXT record, but do not commit it
	if err := s.TXTCreate(zone, fqdn, "test", ssot.TTL); err != nil {
		return err
	}
	return nil
}
