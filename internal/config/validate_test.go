/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

	config.SetHost("")

	err := config.Validate()
	suite.EqualError(err, "host must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateAccountDomainOK1() {
	testrig.InitTestConfig()

	err := config.Validate()
	suite.NoError(err)

	suite.Equal(config.GetHost(), config.GetAccountDomain())
}

func (suite *ConfigValidateTestSuite) TestValidateAccountDomainOK2() {
	testrig.InitTestConfig()

	config.SetAccountDomain("localhost:8080")

	err := config.Validate()
	suite.NoError(err)
}

func (suite *ConfigValidateTestSuite) TestValidateAccountDomainOK3() {
	testrig.InitTestConfig()

	config.SetHost("gts.example.org")
	config.SetAccountDomain("example.org")

	err := config.Validate()
	suite.NoError(err)
}

func (suite *ConfigValidateTestSuite) TestValidateAccountDomainNotSubdomain1() {
	testrig.InitTestConfig()

	config.SetHost("gts.example.org")
	config.SetAccountDomain("example.com")

	err := config.Validate()
	suite.EqualError(err, "host was gts.example.org and account-domain was example.com, but gts.example.org is not a valid subdomain of example.com")
}

func (suite *ConfigValidateTestSuite) TestValidateAccountDomainNotSubdomain2() {
	testrig.InitTestConfig()

	config.SetHost("example.org")
	config.SetAccountDomain("gts.example.org")

	err := config.Validate()
	suite.EqualError(err, "host was example.org and account-domain was gts.example.org, but example.org is not a valid subdomain of gts.example.org")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigNoProtocol() {
	testrig.InitTestConfig()

	config.SetProtocol("")

	err := config.Validate()
	suite.EqualError(err, "protocol must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigNoWebAssetBaseDir() {
	testrig.InitTestConfig()

	config.SetWebAssetBaseDir("")

	err := config.Validate()
	suite.EqualError(err, "web-asset-base-dir must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigNoProtocolOrHost() {
	testrig.InitTestConfig()

	config.SetHost("")
	config.SetProtocol("")

	err := config.Validate()
	suite.EqualError(err, "host must be set; protocol must be set")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigBadProtocol() {
	testrig.InitTestConfig()

	config.SetProtocol("foo")

	err := config.Validate()
	suite.EqualError(err, "protocol must be set to either http or https, provided value was foo")
}

func (suite *ConfigValidateTestSuite) TestValidateConfigBadProtocolNoHost() {
	testrig.InitTestConfig()

	config.SetHost("")
	config.SetProtocol("foo")

	err := config.Validate()
	suite.EqualError(err, "host must be set; protocol must be set to either http or https, provided value was foo")
}

func TestConfigValidateTestSuite(t *testing.T) {
	suite.Run(t, &ConfigValidateTestSuite{})
}
