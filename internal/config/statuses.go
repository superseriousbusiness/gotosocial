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

// StatusesConfig pertains to posting/deleting/interacting with statuses
type StatusesConfig struct {
	// Maximum amount of characters allowed in a status, excluding CW
	MaxChars int `yaml:"max_chars"`
	// Maximum amount of characters allowed in a content-warning/spoiler field
	CWMaxChars int `yaml:"cw_max_chars"`
	// Maximum number of options allowed in a poll
	PollMaxOptions int `yaml:"poll_max_options"`
	// Maximum characters allowed per poll option
	PollOptionMaxChars int `yaml:"poll_option_max_chars"`
	// Maximum amount of media files allowed to be attached to one status
	MaxMediaFiles int `yaml:"max_media_files"`
}
