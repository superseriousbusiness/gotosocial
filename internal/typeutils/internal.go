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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (c *converter) FollowRequestToFollow(ctx context.Context, f *gtsmodel.FollowRequest) *gtsmodel.Follow {
	showReblogs := *f.ShowReblogs
	notify := *f.Notify
	return &gtsmodel.Follow{
		ID:              f.ID,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
		AccountID:       f.AccountID,
		TargetAccountID: f.TargetAccountID,
		ShowReblogs:     &showReblogs,
		URI:             f.URI,
		Notify:          &notify,
	}
}

func (c *converter) StatusToBoost(ctx context.Context, s *gtsmodel.Status, boostingAccount *gtsmodel.Account) (*gtsmodel.Status, error) {
	// the wrapper won't use the same ID as the boosted status so we generate some new UUIDs
	accountURIs := uris.GenerateURIsForAccount(boostingAccount.Username)
	boostWrapperStatusID := id.NewULID()
	boostWrapperStatusURI := accountURIs.StatusesURI + "/" + boostWrapperStatusID
	boostWrapperStatusURL := accountURIs.StatusesURL + "/" + boostWrapperStatusID

	local := true
	if boostingAccount.Domain != "" {
		local = false
	}

	sensitive := *s.Sensitive
	federated := *s.Federated
	boostable := *s.Boostable
	replyable := *s.Replyable
	likeable := *s.Likeable

	boostWrapperStatus := &gtsmodel.Status{
		ID:  boostWrapperStatusID,
		URI: boostWrapperStatusURI,
		URL: boostWrapperStatusURL,

		// the boosted status is not created now, but the boost certainly is
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Local:      &local,
		AccountID:  boostingAccount.ID,
		AccountURI: boostingAccount.URI,

		// replies can be boosted, but boosts are never replies
		InReplyToID:        "",
		InReplyToAccountID: "",

		// these will all be wrapped in the boosted status so set them empty here
		AttachmentIDs: []string{},
		TagIDs:        []string{},
		MentionIDs:    []string{},
		EmojiIDs:      []string{},

		// the below fields will be taken from the target status
		Content:             s.Content,
		ContentWarning:      s.ContentWarning,
		ActivityStreamsType: s.ActivityStreamsType,
		Sensitive:           &sensitive,
		Language:            s.Language,
		Text:                s.Text,
		BoostOfID:           s.ID,
		BoostOfAccountID:    s.AccountID,
		Visibility:          s.Visibility,
		Federated:           &federated,
		Boostable:           &boostable,
		Replyable:           &replyable,
		Likeable:            &likeable,

		// attach these here for convenience -- the boosted status/account won't go in the DB
		// but they're needed in the processor and for the frontend. Since we have them, we can
		// attach them so we don't need to fetch them again later (save some DB calls)
		BoostOf: s,
	}

	return boostWrapperStatus, nil
}
