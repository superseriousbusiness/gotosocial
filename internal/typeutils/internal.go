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
	"errors"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
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

		// Boosts are not considered sensitive even if their target is.
		Sensitive: util.Ptr(false),

		// Remaining fields all
		// taken from boosted status.
		ActivityStreamsType: target.ActivityStreamsType,
		BoostOfID:           target.ID,
		BoostOf:             target,
		BoostOfAccountID:    target.AccountID,
		BoostOfAccount:      target.Account,
		Visibility:          target.Visibility,
		Federated:           util.Ptr(*target.Federated),
	}

	return boost, nil
}

func StatusToInteractionRequest(status *gtsmodel.Status) *gtsmodel.InteractionRequest {
	reqID := id.NewULIDFromTime(status.CreatedAt)

	var (
		targetID        string
		target          *gtsmodel.Status
		targetAccountID string
		targetAccount   *gtsmodel.Account
		interactionType gtsmodel.InteractionType
		reply           *gtsmodel.Status
		announce        *gtsmodel.Status
	)

	if status.InReplyToID != "" {
		// It's a reply.
		targetID = status.InReplyToID
		target = status.InReplyTo
		targetAccountID = status.InReplyToAccountID
		targetAccount = status.InReplyToAccount
		interactionType = gtsmodel.InteractionReply
		reply = status
	} else {
		// It's a boost.
		targetID = status.BoostOfID
		target = status.BoostOf
		targetAccountID = status.BoostOfAccountID
		targetAccount = status.BoostOfAccount
		interactionType = gtsmodel.InteractionAnnounce
		announce = status
	}

	return &gtsmodel.InteractionRequest{
		ID:                   reqID,
		CreatedAt:            status.CreatedAt,
		StatusID:             targetID,
		Status:               target,
		TargetAccountID:      targetAccountID,
		TargetAccount:        targetAccount,
		InteractingAccountID: status.AccountID,
		InteractingAccount:   status.Account,
		InteractionURI:       status.URI,
		InteractionType:      interactionType,
		Reply:                reply,
		Announce:             announce,
	}
}

func StatusFaveToInteractionRequest(fave *gtsmodel.StatusFave) *gtsmodel.InteractionRequest {
	reqID := id.NewULIDFromTime(fave.CreatedAt)

	return &gtsmodel.InteractionRequest{
		ID:                   reqID,
		CreatedAt:            fave.CreatedAt,
		StatusID:             fave.StatusID,
		Status:               fave.Status,
		TargetAccountID:      fave.TargetAccountID,
		TargetAccount:        fave.TargetAccount,
		InteractingAccountID: fave.AccountID,
		InteractingAccount:   fave.Account,
		InteractionURI:       fave.URI,
		InteractionType:      gtsmodel.InteractionLike,
		Like:                 fave,
	}
}

func (c *Converter) StatusToSinBinStatus(
	ctx context.Context,
	status *gtsmodel.Status,
) (*gtsmodel.SinBinStatus, error) {
	// Populate status first so we have
	// polls, mentions etc to copy over.
	//
	// ErrNoEntries is fine, we'll do our best.
	err := c.state.DB.PopulateStatus(ctx, status)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("db error populating status: %w", err)
	}

	// Get domain of this status,
	// empty for our own domain.
	var domain string
	if status.Account != nil {
		domain = status.Account.Domain
	} else {
		uri, err := url.Parse(status.URI)
		if err != nil {
			return nil, gtserror.Newf("error parsing status URI: %w", err)
		}

		host := uri.Host
		if host != config.GetAccountDomain() &&
			host != config.GetHost() {
			domain = host
		}
	}

	// Extract just the image URLs from attachments.
	attachLinks := make([]string, len(status.Attachments))
	for i, attach := range status.Attachments {
		if attach.IsLocal() {
			attachLinks[i] = attach.URL
		} else {
			attachLinks[i] = attach.RemoteURL
		}
	}

	// Extract just the target account URIs from mentions.
	mentionTargetURIs := make([]string, 0, len(status.Mentions))
	for _, mention := range status.Mentions {
		if err := c.state.DB.PopulateMention(ctx, mention); err != nil {
			log.Errorf(ctx, "error populating mention: %v", err)
			continue
		}

		mentionTargetURIs = append(mentionTargetURIs, mention.TargetAccount.URI)
	}

	// Extract just the image URLs from emojis.
	emojiLinks := make([]string, len(status.Emojis))
	for i, emoji := range status.Emojis {
		if emoji.IsLocal() {
			emojiLinks[i] = emoji.ImageURL
		} else {
			emojiLinks[i] = emoji.ImageRemoteURL
		}
	}

	// Extract just the poll option strings.
	var pollOptions []string
	if status.Poll != nil {
		pollOptions = status.Poll.Options
	}

	return &gtsmodel.SinBinStatus{
		ID:                  status.ID, // Reuse the status ID.
		URI:                 status.URI,
		URL:                 status.URL,
		Domain:              domain,
		AccountURI:          status.AccountURI,
		InReplyToURI:        status.InReplyToURI,
		Content:             status.Content,
		AttachmentLinks:     attachLinks,
		MentionTargetURIs:   mentionTargetURIs,
		EmojiLinks:          emojiLinks,
		PollOptions:         pollOptions,
		ContentWarning:      status.ContentWarning,
		Visibility:          status.Visibility,
		Sensitive:           status.Sensitive,
		Language:            status.Language,
		ActivityStreamsType: status.ActivityStreamsType,
	}, nil
}
