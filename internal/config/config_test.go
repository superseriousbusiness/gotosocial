/*
   GoToSocial
   Copyright (C) 2022 GoToSocial Authors admin@gotosocial.org

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

package config_test

import (
	"os"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"codeberg.org/gruf/go-kv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func expectedKV(kvs ...kv.Field) map[string]interface{} {
	ret := make(map[string]interface{}, len(kvs))
	for _, kv := range kvs {
		ret[kv.K] = kv.V
	}
	return ret
}

func expectedFile(t *testing.T, file string) map[string]interface{} {
	expectedConfig, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("error reading expected config from file %q: %v", file, err)
	}

	var ret map[string]interface{}
	if err := yaml.Unmarshal(expectedConfig, &ret); err != nil {
		t.Errorf("error parsing expected config from file %q: %v", file, err)
	}

	return ret
}

func TestCLIParsing(t *testing.T) {
	type testcase struct {
		cli      []string
		env      []string
		expected map[string]interface{}
	}

	defaults := config.Defaults.MarshalMap()

	testcases := map[string]testcase{
		"Make sure defaults are set correctly": {
			expected: defaults,
		},

		"Override db-address from default using cli flag": {
			cli: []string{
				"--db-address", "some.db.address",
			},
			expected: expectedKV(
				kv.Field{"db-address", "some.db.address"},
			),
		},

		"Override db-address from default using env var": {
			env: []string{
				"GTS_DB_ADDRESS=some.db.address",
			},
			expected: expectedKV(
				kv.Field{"db-address", "some.db.address"},
			),
		},

		"Override db-address from default using both env var and cli flag. The cli flag should take priority": {
			cli: []string{
				"--db-address", "some.db.address",
			},
			env: []string{
				"GTS_DB_ADDRESS=some.other.db.address",
			},
			expected: expectedKV(
				kv.Field{"db-address", "some.db.address"},
			),
		},

		"Loading a config file via env var": {
			env: []string{
				"GTS_CONFIG_PATH=testdata/test.yaml",
			},
			expected: expectedFile(t, "testdata/test.yaml"),
		},

		"Loading a config file via cli flag": {
			cli: []string{
				"--config-path", "testdata/test.yaml",
			},
			expected: expectedFile(t, "testdata/test.yaml"),
		},

		"Loading a config file and overriding one of the variables with a cli flag": {
			cli: []string{
				"--config-path", "testdata/test.yaml",
				"--account-domain", "my.test.domain",
			},
			// only checking our overridden one and one non-default from the config file here instead of including all of test.yaml
			expected: expectedKV(
				kv.Field{"account-domain", "my.test.domain"},
				kv.Field{"host", "gts.example.org"},
			),
		},

		"Loading a config file and overriding one of the variables with an env var": {
			cli: []string{
				"--config-path", "testdata/test.yaml",
			},
			env: []string{
				"GTS_ACCOUNT_DOMAIN=my.test.domain",
			},
			// only checking our overridden one and one non-default from the config file here instead of including all of test.yaml
			expected: expectedKV(
				kv.Field{"account-domain", "my.test.domain"},
				kv.Field{"host", "gts.example.org"},
			),
		},

		"Loading a config file and overriding one of the variables with both an env var and a cli flag. The cli flag should have priority": {
			cli: []string{
				"--config-path", "testdata/test.yaml",
				"--account-domain", "my.test.domain",
			},
			env: []string{
				"GTS_ACCOUNT_DOMAIN=my.wrong.test.domain",
			},
			// only checking our overridden one and one non-default from the config file here instead of including all of test.yaml
			expected: expectedKV(
				kv.Field{"account-domain", "my.test.domain"},
				kv.Field{"host", "gts.example.org"},
			),
		},

		"Loading a config file from json": {
			cli: []string{
				"--config-path", "testdata/test.json",
			},
			expected: expectedFile(t, "testdata/test.json"),
		},

		"Loading a partial config file. Default values should be used apart from those set in the config file": {
			cli: []string{
				"--config-path", "testdata/test2.yaml",
			},
			expected: expectedKV(
				kv.Field{"log-level", "trace"},
				kv.Field{"account-domain", "peepee.poopoo"},
				kv.Field{"application-name", "gotosocial"},
			),
		},

		"Loading nested config file. This should also work the same": {
			cli: []string{
				"--config-path", "testdata/test3.yaml",
			},
			expected: expectedKV(
				kv.Field{"advanced-scraper-deterrence-enabled", true},
				kv.Field{"advanced-rate-limit-requests", 5000},
			),
		},
	}

	for desc, data := range testcases {
		t.Run(desc, func(t *testing.T) {
			os.Clearenv()

			if data.env != nil {
				for _, s := range data.env {
					kv := strings.SplitN(s, "=", 2)
					os.Setenv(kv[0], kv[1])
				}
			}

			state := config.NewState()
			cmd := cobra.Command{}
			config.RegisterGlobalFlags(&cmd)

			if data.cli != nil {
				cmd.ParseFlags(data.cli)
			}

			state.BindFlags(&cmd)

			state.LoadConfigFile()

			state.Viper(func(v *viper.Viper) {
				for k, ev := range data.expected {
					assert.EqualValues(t, ev, v.Get(k))
				}
			})
		})
	}
}
