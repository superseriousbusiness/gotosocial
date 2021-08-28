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

package typeutils

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) ASRepresentationToAccount(ctx context.Context, accountable ap.Accountable, update bool) (*gtsmodel.Account, error) {
	// first check if we actually already know this account
	uriProp := accountable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found on person, or id was not an iri")
	}
	uri := uriProp.GetIRI()

	if !update {
		acct, err := c.db.GetAccountByURI(ctx, uri.String())
		if err == nil {
			// we already know this account so we can skip generating it
			return acct, nil
		}
		if err != db.ErrNoEntries {
			// we don't know the account and there's been a real error
			return nil, fmt.Errorf("error getting account with uri %s from the database: %s", uri.String(), err)
		}
	}

	// we don't know the account, or we're being told to update it, so we need to generate it from the person -- at least we already have the URI!
	acct := &gtsmodel.Account{}
	acct.URI = uri.String()

	// Username aka preferredUsername
	// We need this one so bail if it's not set.
	username, err := ap.ExtractPreferredUsername(accountable)
	if err != nil {
		return nil, fmt.Errorf("couldn't extract username: %s", err)
	}
	acct.Username = username

	// Domain
	acct.Domain = uri.Host

	// avatar aka icon
	// if this one isn't extractable in a format we recognise we'll just skip it
	if avatarURL, err := ap.ExtractIconURL(accountable); err == nil {
		acct.AvatarRemoteURL = avatarURL.String()
	}

	// header aka image
	// if this one isn't extractable in a format we recognise we'll just skip it
	if headerURL, err := ap.ExtractImageURL(accountable); err == nil {
		acct.HeaderRemoteURL = headerURL.String()
	}

	// display name aka name
	// we default to the username, but take the more nuanced name property if it exists
	acct.DisplayName = username
	if displayName, err := ap.ExtractName(accountable); err == nil {
		acct.DisplayName = displayName
	}

	// TODO: fields aka attachment array

	// note aka summary
	note, err := ap.ExtractSummary(accountable)
	if err == nil && note != "" {
		acct.Note = note
	}

	// check for bot and actor type
	switch accountable.GetTypeName() {
	case gtsmodel.ActivityStreamsPerson, gtsmodel.ActivityStreamsGroup, gtsmodel.ActivityStreamsOrganization:
		// people, groups, and organizations aren't bots
		acct.Bot = false
		// apps and services are
	case gtsmodel.ActivityStreamsApplication, gtsmodel.ActivityStreamsService:
		acct.Bot = true
	default:
		// we don't know what this is!
		return nil, fmt.Errorf("type name %s not recognised or not convertible to gtsmodel.ActivityStreamsActor", accountable.GetTypeName())
	}
	acct.ActorType = accountable.GetTypeName()

	// locked aka manuallyApprovesFollowers
	acct.Locked = true // assume locked by default
	maf := accountable.GetActivityStreamsManuallyApprovesFollowers()
	if maf != nil && maf.IsXMLSchemaBoolean() {
		acct.Locked = maf.Get()
	}

	// discoverable
	// default to false -- take custom value if it's set though
	acct.Discoverable = false
	discoverable, err := ap.ExtractDiscoverable(accountable)
	if err == nil {
		acct.Discoverable = discoverable
	}

	// url property
	url, err := ap.ExtractURL(accountable)
	if err == nil {
		// take the URL if we can find it
		acct.URL = url.String()
	} else {
		// otherwise just take the account URI as the URL
		acct.URL = uri.String()
	}

	// InboxURI
	if accountable.GetActivityStreamsInbox() != nil && accountable.GetActivityStreamsInbox().GetIRI() != nil {
		acct.InboxURI = accountable.GetActivityStreamsInbox().GetIRI().String()
	}

	// OutboxURI
	if accountable.GetActivityStreamsOutbox() != nil && accountable.GetActivityStreamsOutbox().GetIRI() != nil {
		acct.OutboxURI = accountable.GetActivityStreamsOutbox().GetIRI().String()
	}

	// FollowingURI
	if accountable.GetActivityStreamsFollowing() != nil && accountable.GetActivityStreamsFollowing().GetIRI() != nil {
		acct.FollowingURI = accountable.GetActivityStreamsFollowing().GetIRI().String()
	}

	// FollowersURI
	if accountable.GetActivityStreamsFollowers() != nil && accountable.GetActivityStreamsFollowers().GetIRI() != nil {
		acct.FollowersURI = accountable.GetActivityStreamsFollowers().GetIRI().String()
	}

	// FeaturedURI
	if accountable.GetTootFeatured() != nil && accountable.GetTootFeatured().GetIRI() != nil {
		acct.FeaturedCollectionURI = accountable.GetTootFeatured().GetIRI().String()
	}

	// TODO: FeaturedTagsURI

	// TODO: alsoKnownAs

	// publicKey
	pkey, pkeyURL, err := ap.ExtractPublicKeyForOwner(accountable, uri)
	if err != nil {
		return nil, fmt.Errorf("couldn't get public key for person %s: %s", uri.String(), err)
	}
	acct.PublicKey = pkey
	acct.PublicKeyURI = pkeyURL.String()

	return acct, nil
}

func (c *converter) ASStatusToStatus(ctx context.Context, statusable ap.Statusable) (*gtsmodel.Status, error) {
	status := &gtsmodel.Status{}

	// uri at which this status is reachable
	uriProp := statusable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found, or id was not an iri")
	}
	status.URI = uriProp.GetIRI().String()

	l := c.log.WithField("statusURI", status.URI)

	// web url for viewing this status
	if statusURL, err := ap.ExtractURL(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status URL: %s", err)
	} else {
		status.URL = statusURL.String()
	}

	// the html-formatted content of this status
	if content, err := ap.ExtractContent(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status content: %s", err)
	} else {
		status.Content = content
	}

	// attachments to dereference and fetch later on (we don't do that here)
	if attachments, err := ap.ExtractAttachments(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status attachments: %s", err)
	} else {
		status.Attachments = attachments
	}

	// hashtags to dereference later on
	if hashtags, err := ap.ExtractHashtags(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status hashtags: %s", err)
	} else {
		status.Tags = hashtags
	}

	// emojis to dereference and fetch later on
	if emojis, err := ap.ExtractEmojis(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status emojis: %s", err)
	} else {
		status.Emojis = emojis
	}

	// mentions to dereference later on
	if mentions, err := ap.ExtractMentions(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status mentions: %s", err)
	} else {
		status.Mentions = mentions
	}

	// cw string for this status
	if cw, err := ap.ExtractSummary(statusable); err != nil {
		l.Infof("ASStatusToStatus: error extracting status summary: %s", err)
	} else {
		status.ContentWarning = cw
	}

	// when was this status created?
	published, err := ap.ExtractPublished(statusable)
	if err != nil {
		l.Infof("ASStatusToStatus: error extracting status published: %s", err)
	} else {
		status.CreatedAt = published
		status.UpdatedAt = published
	}

	// which account posted this status?
	// if we don't know the account yet we can dereference it later
	attributedTo, err := ap.ExtractAttributedTo(statusable)
	if err != nil {
		return nil, errors.New("attributedTo was empty")
	}
	status.AccountURI = attributedTo.String()

	statusOwner, err := c.db.GetAccountByURI(ctx, attributedTo.String())
	if err != nil {
		return nil, fmt.Errorf("couldn't get status owner from db: %s", err)
	}
	status.AccountID = statusOwner.ID
	status.AccountURI = statusOwner.URI
	status.Account = statusOwner

	// check if there's a post that this is a reply to
	inReplyToURI := ap.ExtractInReplyToURI(statusable)
	if inReplyToURI != nil {
		// something is set so we can at least set this field on the
		// status and dereference using this later if we need to
		status.InReplyToURI = inReplyToURI.String()

		// now we can check if we have the replied-to status in our db already
		if inReplyToStatus, err := c.db.GetStatusByURI(ctx, inReplyToURI.String()); err == nil {
			// we have the status in our database already
			// so we can set these fields here and now...
			status.InReplyToID = inReplyToStatus.ID
			status.InReplyToAccountID = inReplyToStatus.AccountID
			status.InReplyTo = inReplyToStatus
			if status.InReplyToAccount == nil {
				if inReplyToAccount, err := c.db.GetAccountByID(ctx, inReplyToStatus.AccountID); err == nil {
					status.InReplyToAccount = inReplyToAccount
				}
			}
		}
	}

	// visibility entry for this status
	var visibility gtsmodel.Visibility

	to, err := ap.ExtractTos(statusable)
	if err != nil {
		return nil, fmt.Errorf("error extracting TO values: %s", err)
	}

	cc, err := ap.ExtractCCs(statusable)
	if err != nil {
		return nil, fmt.Errorf("error extracting CC values: %s", err)
	}

	if len(to) == 0 && len(cc) == 0 {
		return nil, errors.New("message wasn't TO or CC anyone")
	}

	// for visibility derivation, we start by assuming most restrictive, and work our way to least restrictive

	// if it's a DM then it's addressed to SPECIFIC ACCOUNTS and not followers or public
	if len(to) != 0 && len(cc) == 0 {
		visibility = gtsmodel.VisibilityDirect
	}

	// if it's just got followers in TO and it's not also CC'ed to public, it's followers only
	if isFollowers(to, statusOwner.FollowersURI) {
		visibility = gtsmodel.VisibilityFollowersOnly
	}

	// if it's CC'ed to public, it's public or unlocked
	// mentioned SPECIFIC ACCOUNTS also get added to CC'es if it's not a direct message
	if isPublic(cc) {
		visibility = gtsmodel.VisibilityUnlocked
	}
	if isPublic(to) {
		visibility = gtsmodel.VisibilityPublic
	}

	// we should have a visibility by now
	if visibility == "" {
		return nil, errors.New("couldn't derive visibility")
	}
	status.Visibility = visibility

	// advanced visibility for this status
	// TODO: a lot of work to be done here -- a new type needs to be created for this in go-fed/activity using ASTOOL

	// sensitive
	// TODO: this is a bool

	// language
	// we might be able to extract this from the contentMap field

	// ActivityStreamsType
	status.ActivityStreamsType = statusable.GetTypeName()

	return status, nil
}

func (c *converter) ASFollowToFollowRequest(ctx context.Context, followable ap.Followable) (*gtsmodel.FollowRequest, error) {

	idProp := followable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on follow, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActor(followable)
	if err != nil {
		return nil, errors.New("error extracting actor property from follow")
	}
	originAccount, err := c.db.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObject(followable)
	if err != nil {
		return nil, errors.New("error extracting object property from follow")
	}
	targetAccount, err := c.db.GetAccountByURI(ctx, target.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	followRequest := &gtsmodel.FollowRequest{
		URI:             uri,
		AccountID:       originAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	return followRequest, nil
}

func (c *converter) ASFollowToFollow(ctx context.Context, followable ap.Followable) (*gtsmodel.Follow, error) {
	idProp := followable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on follow, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActor(followable)
	if err != nil {
		return nil, errors.New("error extracting actor property from follow")
	}
	originAccount, err := c.db.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObject(followable)
	if err != nil {
		return nil, errors.New("error extracting object property from follow")
	}
	targetAccount, err := c.db.GetAccountByURI(ctx, target.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	follow := &gtsmodel.Follow{
		URI:             uri,
		AccountID:       originAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	return follow, nil
}

func (c *converter) ASLikeToFave(ctx context.Context, likeable ap.Likeable) (*gtsmodel.StatusFave, error) {
	idProp := likeable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on like, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActor(likeable)
	if err != nil {
		return nil, errors.New("error extracting actor property from like")
	}
	originAccount, err := c.db.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObject(likeable)
	if err != nil {
		return nil, errors.New("error extracting object property from like")
	}

	targetStatus, err := c.db.GetStatusByURI(ctx, target.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting status with uri %s from the database: %s", target.String(), err)
	}

	var targetAccount *gtsmodel.Account
	if targetStatus.Account != nil {
		targetAccount = targetStatus.Account
	} else {
		a, err := c.db.GetAccountByID(ctx, targetStatus.AccountID)
		if err != nil {
			return nil, fmt.Errorf("error extracting account with id %s from the database: %s", targetStatus.AccountID, err)
		}
		targetAccount = a
	}

	return &gtsmodel.StatusFave{
		AccountID:       originAccount.ID,
		Account:         originAccount,
		TargetAccountID: targetAccount.ID,
		TargetAccount:   targetAccount,
		StatusID:        targetStatus.ID,
		Status:          targetStatus,
		URI:             uri,
	}, nil
}

func (c *converter) ASBlockToBlock(ctx context.Context, blockable ap.Blockable) (*gtsmodel.Block, error) {
	idProp := blockable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("ASBlockToBlock: no id property set on block, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActor(blockable)
	if err != nil {
		return nil, errors.New("ASBlockToBlock: error extracting actor property from block")
	}
	originAccount, err := c.db.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObject(blockable)
	if err != nil {
		return nil, errors.New("ASBlockToBlock: error extracting object property from block")
	}

	targetAccount, err := c.db.GetAccountByURI(ctx, target.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	return &gtsmodel.Block{
		AccountID:       originAccount.ID,
		Account:         originAccount,
		TargetAccountID: targetAccount.ID,
		TargetAccount:   targetAccount,
		URI:             uri,
	}, nil
}

func (c *converter) ASAnnounceToStatus(ctx context.Context, announceable ap.Announceable) (*gtsmodel.Status, bool, error) {
	status := &gtsmodel.Status{}
	isNew := true

	// check if we already have the boost in the database
	idProp := announceable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, isNew, errors.New("no id property set on announce, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	if status, err := c.db.GetStatusByURI(ctx, uri); err == nil {
		// we already have it, great, just return it as-is :)
		isNew = false
		return status, isNew, nil
	}
	status.URI = uri

	// get the URI of the announced/boosted status
	boostedStatusURI, err := ap.ExtractObject(announceable)
	if err != nil {
		return nil, isNew, fmt.Errorf("ASAnnounceToStatus: error getting object from announce: %s", err)
	}

	// set the URI on the new status for dereferencing later
	status.BoostOf = &gtsmodel.Status{
		URI: boostedStatusURI.String(),
	}

	// get the published time for the announce
	published, err := ap.ExtractPublished(announceable)
	if err != nil {
		return nil, isNew, fmt.Errorf("ASAnnounceToStatus: error extracting published time: %s", err)
	}
	status.CreatedAt = published
	status.UpdatedAt = published

	// get the actor's IRI (ie., the person who boosted the status)
	actor, err := ap.ExtractActor(announceable)
	if err != nil {
		return nil, isNew, fmt.Errorf("ASAnnounceToStatus: error extracting actor: %s", err)
	}

	// get the boosting account based on the URI
	// this should have been dereferenced already before we hit this point so we can confidently error out if we don't have it
	boostingAccount, err := c.db.GetAccountByURI(ctx, actor.String())
	if err != nil {
		return nil, isNew, fmt.Errorf("ASAnnounceToStatus: error in db fetching account with uri %s: %s", actor.String(), err)
	}
	status.AccountID = boostingAccount.ID
	status.AccountURI = boostingAccount.URI
	status.Account = boostingAccount

	// these will all be wrapped in the boosted status so set them empty here
	status.AttachmentIDs = []string{}
	status.TagIDs = []string{}
	status.MentionIDs = []string{}
	status.EmojiIDs = []string{}

	// parse the visibility from the To and CC entries
	var visibility gtsmodel.Visibility

	to, err := ap.ExtractTos(announceable)
	if err != nil {
		return nil, isNew, fmt.Errorf("error extracting TO values: %s", err)
	}

	cc, err := ap.ExtractCCs(announceable)
	if err != nil {
		return nil, isNew, fmt.Errorf("error extracting CC values: %s", err)
	}

	if len(to) == 0 && len(cc) == 0 {
		return nil, isNew, errors.New("message wasn't TO or CC anyone")
	}

	// if it's CC'ed to public, it's public or unlocked
	if isPublic(cc) {
		visibility = gtsmodel.VisibilityUnlocked
	}
	if isPublic(to) {
		visibility = gtsmodel.VisibilityPublic
	}

	// we should have a visibility by now
	if visibility == "" {
		return nil, isNew, errors.New("couldn't derive visibility")
	}
	status.Visibility = visibility

	// the rest of the fields will be taken from the target status, but it's not our job to do the dereferencing here
	return status, isNew, nil
}

func isPublic(tos []*url.URL) bool {
	for _, entry := range tos {
		if strings.EqualFold(entry.String(), "https://www.w3.org/ns/activitystreams#Public") {
			return true
		}
	}
	return false
}

func isFollowers(ccs []*url.URL, followersURI string) bool {
	for _, entry := range ccs {
		if strings.EqualFold(entry.String(), followersURI) {
			return true
		}
	}
	return false
}
