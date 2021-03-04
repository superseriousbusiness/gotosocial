/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"fmt"
	"os"

	"github.com/gotosocial/gotosocial/internal/db"
	"gopkg.in/yaml.v2"
)

// Config contains all the configuration needed to run gotosocial
type Config struct {
	LogLevel        string     `yaml:"logLevel"`
	ApplicationName string     `yaml:"applicationName,omitempty"`
	DBConfig        *db.Config `yaml:"db,omitempty"`
}

// New returns a new config, or an error if something goes amiss.
// The path parameter is optional, for loading a configuration json from the given path.
func New(path string) (*Config, error) {
	config := &Config{}
	if path != "" {
		var err error
		if config, err = loadFromFile(path); err != nil {
			return nil, fmt.Errorf("error creating config: %s", err)
		}
	}

	return config, nil
}

// loadFromFile takes a path to a yaml file and attempts to load a Config object from it
func loadFromFile(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file at path %s: %s", path, err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("could not unmarshal file at path %s: %s", path, err)
	}

	return config, nil
}

// ParseFlags sets flags on the config using the provided Flags object
func (c *Config) ParseFlags(f Flags) {

}

// Flags is a wrapper for any type that can store keyed flags and give them back
type Flags interface {
	String(k string) string
	Int(k string) int
}
