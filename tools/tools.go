//go:build tools

// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// tools exists to pull in command-line tools that we need to go get,
// and is behind a build tag that is otherwise unused and thus only visible
// to dependency management commands. See https://stackoverflow.com/a/54028731.
package tools

import (
	// Provides swagger command used by tests/swagger.sh
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
)
