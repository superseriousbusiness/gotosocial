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

package typeutils

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// FollowRequestToFollow just converts a follow request
// into a follow, that's it! No bells and whistles.
func (c *Converter) FollowRequestToFollow(
	ctx context.Context,
	fr *gtsmodel.FollowRequest,
) *gtsmodel.Follow {
	return &gtsmodel.Follow{
		ID:              fr.ID,
		CreatedAt:       fr.CreatedAt,
		UpdatedAt:       fr.UpdatedAt,
		AccountID:       fr.AccountID,
		TargetAccountID: fr.TargetAccountID,
		ShowReblogs:     util.Ptr(*fr.ShowReblogs),
		URI:             fr.URI,
		Notify:          util.Ptr(*fr.Notify),
	}
}

// StatusToBoost wraps the target status into a
// boost wrapper status owned by the requester.
func (c *Converter) StatusToBoost(
	ctx context.Context,
	target *gtsmodel.Status,
	booster *gtsmodel.Account,
	applicationID string,
) (*gtsmodel.Status, error) {
	// The boost won't use the same IDs as the
	// target so we need to generate new ones.
	boostID := id.NewULID()
	accountURIs := uris.GenerateURIsForAccount(booster.Username)

	boost := &gtsmodel.Status{
		ID:  boostID,
		URI: accountURIs.StatusesURI + "/" + boostID,
		URL: accountURIs.StatusesURL + "/" + boostID,

		// Inherit some fields from the booster account.
		Local:                    util.Ptr(booster.IsLocal()),
		AccountID:                booster.ID,
		Account:                  booster,
		AccountURI:               booster.URI,
		CreatedWithApplicationID: applicationID,

		// Replies can be boosted, but
		// boosts are never replies.
		InReplyToID:        "",
		InReplyToAccountID: "",

		// These will all be wrapped in the
		// boosted status so set them empty.
		AttachmentIDs: []string{},
		TagIDs:        []string{},
		MentionIDs:    []string{},
		EmojiIDs:      []string{},

		// Remaining fields all
		// taken from boosted status.
		Content:             target.Content,
		ContentWarning:      target.ContentWarning,
		ActivityStreamsType: target.ActivityStreamsType,
		Sensitive:           util.Ptr(*target.Sensitive),
		Language:            target.Language,
		Text:                target.Text,
		BoostOfID:           target.ID,
		BoostOf:             target,
		BoostOfAccountID:    target.AccountID,
		BoostOfAccount:      target.Account,
		Visibility:          target.Visibility,
		Federated:           util.Ptr(*target.Federated),
	}

	return boost, nil
}
