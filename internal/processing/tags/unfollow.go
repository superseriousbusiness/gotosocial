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

package tags

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Unfollow unfollows the tag with the given name as the given account.
// If there is no tag with that name, it creates a tag.
func (p *Processor) Unfollow(
	ctx context.Context,
	account *gtsmodel.Account,
	name string,
) (*apimodel.Tag, gtserror.WithCode) {
	// Try to get an existing tag with that name.
	tag, err := p.state.DB.GetTagByName(ctx, name)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("DB error getting tag with name %s: %w", name, err),
		)
	}
	if tag == nil {
		return nil, gtserror.NewErrorNotFound(
			gtserror.Newf("couldn't find tag with name %s: %w", name, err),
		)
	}

	// Unfollow the tag.
	if err := p.state.DB.DeleteFollowedTag(ctx, account.ID, tag.ID); err != nil {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("DB error unfollowing tag %s: %w", tag.ID, err),
		)
	}

	return p.apiTag(ctx, tag, false)
}
