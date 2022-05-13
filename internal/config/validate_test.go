/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ConfigValidateTestSuite struct {
	suite.Suite
}

func (suite *ConfigValidateTestSuite) TestValidateConfigOK() {
	testrig.InitTestConfig()

	err := config.Validate()
	suite.NoError(err)
}

func (suite *ConfigValidateTestSuite) TestValidateConfigNoHost() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.Host, "")

	err := config.Validate()
	suite.EqualError(err, "host must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigNoProtocol() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.Protocol, "")

	err := config.Validate()
	suite.EqualError(err, "protocol must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigNoProtocolOrHost() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.Host, "")
	viper.Set(config.Keys.Protocol, "")

	err := config.Validate()
	suite.EqualError(err, "host must be set; protocol must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigBadProtocol() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.Protocol, "foo")

	err := config.Validate()
	suite.EqualError(err, "protocol must be set to either http or https, provided value was foo")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigBadProtocolNoHost() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.Host, "")
	viper.Set(config.Keys.Protocol, "foo")

	err := config.Validate()
	suite.EqualError(err, "host must be set; protocol must be set to either http or https, provided value was foo")
}

func TestConfigValidateTestSuite(t *testing.T) {
	suite.Run(t, &ConfigValidateTestSuite{})
}
