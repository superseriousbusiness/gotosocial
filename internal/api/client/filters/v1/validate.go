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

package v1

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func validateNormalizeCreateUpdateFilter(form *model.FilterCreateUpdateRequestV1) error {
	if err := validate.FilterKeyword(form.Phrase); err != nil {
		return err
	}
	// For filter v1 forwards compatibility, the phrase is used as the title of a v2 filter, so it must pass that as well.
	if err := validate.FilterTitle(form.Phrase); err != nil {
		return err
	}
	if err := validate.FilterContexts(form.Context); err != nil {
		return err
	}

	// Apply defaults for missing fields.
	form.WholeWord = util.Ptr(util.PtrOrValue(form.WholeWord, false))
	form.Irreversible = util.Ptr(util.PtrOrValue(form.Irreversible, false))

	if *form.Irreversible {
		return errors.New("irreversible aka server-side drop filters are not supported yet")
	}

	// Normalize filter expiry if necessary.
	// If we parsed this as JSON, expires_in
	// may be either a float64 or a string.
	if ei := form.ExpiresInI; ei != nil {
		switch e := ei.(type) {
		case float64:
			form.ExpiresIn = util.Ptr(int(e))

		case string:
			expiresIn, err := strconv.Atoi(e)
			if err != nil {
				return fmt.Errorf("could not parse expires_in value %s as integer: %w", e, err)
			}

			form.ExpiresIn = &expiresIn

		default:
			return fmt.Errorf("could not parse expires_in type %T as integer", ei)
		}
	}

	return nil
}
