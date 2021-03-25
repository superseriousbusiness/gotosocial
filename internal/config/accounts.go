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

// AccountsConfig contains configuration to do with creating accounts, new registrations, and defaults.
type AccountsConfig struct {
	// Do we want people to be able to just submit sign up requests, or do we want invite only?
	OpenRegistration bool `yaml:"openRegistration"`
	// Do sign up requests require approval from an admin/moderator?
	RequireApproval bool `yaml:"requireApproval"`
	// Do we require a reason for a sign up or is an empty string OK?
	ReasonRequired bool `yaml:"reasonRequired"`
}
