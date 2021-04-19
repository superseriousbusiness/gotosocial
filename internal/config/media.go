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

// MediaConfig contains configuration for receiving and parsing media files and attachments
type MediaConfig struct {
	// Max size of uploaded images in bytes
	MaxImageSize int `yaml:"maxImageSize"`
	// Max size of uploaded video in bytes
	MaxVideoSize int `yaml:"maxVideoSize"`
	// Minimum amount of chars required in an image description
	MinDescriptionChars int `yaml:"minDescriptionChars"`
	// Max amount of chars allowed in an image description
	MaxDescriptionChars int `yaml:"maxDescriptionChars"`
}
