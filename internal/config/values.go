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

package config

// Values contains contains the type of each configuration value.
type Values struct {
	LogLevel        string
	LogDbQueries    bool
	ApplicationName string
	ConfigPath      string
	Host            string
	AccountDomain   string
	Protocol        string
	BindAddress     string
	Port            int
	TrustedProxies  []string
	SoftwareVersion string

	DbType      string
	DbAddress   string
	DbPort      int
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DbTLSMode   string
	DbTLSCACert string

	WebTemplateBaseDir string
	WebAssetBaseDir    string

	AccountsRegistrationOpen bool
	AccountsApprovalRequired bool
	AccountsReasonRequired   bool

	MediaImageMaxSize        int
	MediaVideoMaxSize        int
	MediaDescriptionMinChars int
	MediaDescriptionMaxChars int
	MediaRemoteCacheDays     int

	StorageBackend       string
	StorageLocalBasePath string

	StatusesMaxChars           int
	StatusesCWMaxChars         int
	StatusesPollMaxOptions     int
	StatusesPollOptionMaxChars int
	StatusesMediaMaxFiles      int

	LetsEncryptEnabled      bool
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         int

	OIDCEnabled          bool
	OIDCIdpName          string
	OIDCSkipVerification bool
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           []string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	SyslogEnabled  bool
	SyslogProtocol string
	SyslogAddress  string

	AdminAccountUsername string
	AdminAccountEmail    string
	AdminAccountPassword string
	AdminTransPath       string
}
