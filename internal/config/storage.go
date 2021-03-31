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

// StorageConfig contains configuration for storage and serving of media files and attachments
type StorageConfig struct {
	// Type of storage backend to use: currently only 'local' is supported.
	// TODO: add S3 support here.
	Backend string `yaml:"backend"`

	// The base path for storing things. Should be an already-existing directory.
	BasePath string `yaml:"basePath"`

	// Protocol to use when *serving* media files from storage
	ServeProtocol string `yaml:"serveProtocol"`
	// Host to use when *serving* media files from storage
	ServeHost string `yaml:"serveHost"`
	// Base path to use when *serving* media files from storage
	ServeBasePath string `yaml:"serveBasePath"`
}
