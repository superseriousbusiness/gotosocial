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
	"errors"
	"fmt"
	"strconv"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ParseFocus parses a media attachment focus parameters from incoming API string.
func ParseFocus(focus string) (focusx, focusy float32, errWithCode gtserror.WithCode) {
	if focus == "" {
		return
	}
	spl := strings.Split(focus, ",")
	if len(spl) != 2 {
		const text = "missing comma separator"
		errWithCode = gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
		return
	}
	xStr := spl[0]
	yStr := spl[1]
	fx, err := strconv.ParseFloat(xStr, 32)
	if err != nil || fx > 1 || fx < -1 {
		text := fmt.Sprintf("invalid x focus: %s", xStr)
		errWithCode = gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
		return
	}
	fy, err := strconv.ParseFloat(yStr, 32)
	if err != nil || fy > 1 || fy < -1 {
		text := fmt.Sprintf("invalid y focus: %s", xStr)
		errWithCode = gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
		return
	}
	focusx = float32(fx)
	focusy = float32(fy)
	return
}

// ParseDuration parses the given raw interface belonging
// the given fieldName as an integer duration.
func ParseDuration(rawI any, fieldName string) (*int, error) {
	var (
		asInteger int
		err       error
	)

	switch raw := rawI.(type) {
	case float64:
		// Submitted as JSON number
		// (casts to float64 by default).
		asInteger = int(raw)

	case string:
		// Submitted as JSON string or form field.
		asInteger, err = strconv.Atoi(raw)
		if err != nil {
			err = fmt.Errorf(
				"could not parse %s value %s as integer: %w",
				fieldName, raw, err,
			)
		}

	default:
		// Submitted as god-knows-what.
		err = fmt.Errorf(
			"could not parse %s type %T as integer",
			fieldName, rawI,
		)
	}

	if err != nil {
		return nil, err
	}

	return &asInteger, nil
}

// ParseNullableDuration is like ParseDuration, but
// for JSON values that may have been sent as `null`.
//
// IsSpecified should be checked and "true" on the
// given nullable before calling this function.
func ParseNullableDuration(
	nullable apimodel.Nullable[any],
	fieldName string,
) (*int, error) {
	if nullable.IsNull() {
		// Was specified as `null`,
		// return pointer to zero value.
		return util.Ptr(0), nil
	}

	rawI, err := nullable.Get()
	if err != nil {
		return nil, err
	}

	return ParseDuration(rawI, fieldName)
}
