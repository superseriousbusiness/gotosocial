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

package util

import (
	"fmt"
	"strconv"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

const (
	LimitKey = "limit"
	LocalKey = "local"
)

func ParseLimit(limit string, defaultLimit int) (int, gtserror.WithCode) {
	if limit == "" {
		return defaultLimit, nil
	}

	i, err := strconv.Atoi(limit)
	if err != nil {
		err := fmt.Errorf("error parsing %s: %w", LimitKey, err)
		return 0, gtserror.NewErrorBadRequest(err, err.Error())
	}

	return i, nil
}

func ParseLocal(local string, defaultLocal bool) (bool, gtserror.WithCode) {
	if local == "" {
		return defaultLocal, nil
	}

	i, err := strconv.ParseBool(local)
	if err != nil {
		err := fmt.Errorf("error parsing %s: %w", LocalKey, err)
		return false, gtserror.NewErrorBadRequest(err, err.Error())
	}

	return i, nil
}
