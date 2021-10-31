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

// SMTPConfig holds configuration for sending emails using the smtp protocol.
type SMTPConfig struct {
	// Host of the smtp server.
	Host string `yaml:"host"`
	// Port of the smtp server.
	Port int `yaml:"port"`
	// Username to use when authenticating with the smtp server.
	Username string `yaml:"username"`
	// Password to use when authenticating with the smtp server.
	Password string `yaml:"password"`
	// From address to use when sending emails.
	From string `yaml:"from"`
}
