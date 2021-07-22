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

// OIDCConfig contains configuration values for openID connect (oauth) authorization by an external service such as Dex.
type OIDCConfig struct {
	Enabled          bool     `yaml:"enabled"`
	IDPName          string   `yaml:"idpName"`
	SkipVerification bool     `yaml:"skipVerification"`
	Issuer           string   `yaml:"issuer"`
	ClientID         string   `yaml:"clientID"`
	ClientSecret     string   `yaml:"clientSecret"`
	Scopes           []string `yaml:"scopes"`
}
