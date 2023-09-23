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
	"fmt"
	"net/url"

	"github.com/miekg/dns"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ASRepresentationToAccount converts a remote account/person/application representation into a gts model account.
//
// If accountDomain is provided then this value will be used as the account's Domain, else the AP ID host.
func (c *Converter) ASRepresentationToAccount(ctx context.Context, accountable ap.Accountable, accountDomain string) (*gtsmodel.Account, error) {
	// first check if we actually already know this account
	uriProp := accountable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found on person, or id was not an iri")
	}
	uri := uriProp.GetIRI()

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
	if accountDomain != "" {
		acct.Domain = accountDomain
	} else {
		acct.Domain = uri.Host
	}

	// avatar aka icon
	// if this one isn't extractable in a format we recognise we'll just skip it
	if avatarURL, err := ap.ExtractIconURI(accountable); err == nil {
		acct.AvatarRemoteURL = avatarURL.String()
	}

	// header aka image
	// if this one isn't extractable in a format we recognise we'll just skip it
	if headerURL, err := ap.ExtractImageURI(accountable); err == nil {
		acct.HeaderRemoteURL = headerURL.String()
	}

	// display name aka name
	// we default to the username, but take the more nuanced name property if it exists
	if displayName := ap.ExtractName(accountable); displayName != "" {
		acct.DisplayName = displayName
	} else {
		acct.DisplayName = username
	}

	// account emojis (used in bio, display name, fields)
	if emojis, err := ap.ExtractEmojis(accountable); err != nil {
		log.Infof(nil, "error extracting account emojis: %s", err)
	} else {
		acct.Emojis = emojis
	}

	// fields aka attachment array
	acct.Fields = ap.ExtractFields(accountable)

	// note aka summary
	acct.Note = ap.ExtractSummary(accountable)

	// check for bot and actor type
	switch accountable.GetTypeName() {
	case ap.ActorPerson, ap.ActorGroup, ap.ActorOrganization:
		// people, groups, and organizations aren't bots
		bot := false
		acct.Bot = &bot
		// apps and services are
	case ap.ActorApplication, ap.ActorService:
		bot := true
		acct.Bot = &bot
	default:
		// we don't know what this is!
		return nil, fmt.Errorf("type name %s not recognised or not convertible to ap.ActivityStreamsActor", accountable.GetTypeName())
	}
	acct.ActorType = accountable.GetTypeName()

	// assume not memorial (todo)
	memorial := false
	acct.Memorial = &memorial

	// assume not sensitive (todo)
	sensitive := false
	acct.Sensitive = &sensitive

	// assume not hide collections (todo)
	hideCollections := false
	acct.HideCollections = &hideCollections

	// locked aka manuallyApprovesFollowers
	locked := true
	acct.Locked = &locked // assume locked by default
	maf := accountable.GetActivityStreamsManuallyApprovesFollowers()
	if maf != nil && maf.IsXMLSchemaBoolean() {
		locked = maf.Get()
		acct.Locked = &locked
	}

	// discoverable
	// default to false -- take custom value if it's set though
	discoverable := false
	acct.Discoverable = &discoverable
	d, err := ap.ExtractDiscoverable(accountable)
	if err == nil {
		acct.Discoverable = &d
	}

	// assume not rss feed
	enableRSS := false
	acct.EnableRSS = &enableRSS

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

	// SharedInboxURI:
	// only trust shared inbox if it has at least two domains,
	// from the right, in common with the domain of the account
	if sharedInboxURI := ap.ExtractSharedInbox(accountable); // nocollapse
	sharedInboxURI != nil && dns.CompareDomainName(acct.Domain, sharedInboxURI.Host) >= 2 {
		sharedInbox := sharedInboxURI.String()
		acct.SharedInboxURI = &sharedInbox
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

	// FeaturedURI aka pinned collection:
	// Only trust featured URI if it has at least two domains,
	// from the right, in common with the domain of the account
	if featured := accountable.GetTootFeatured(); featured != nil && featured.IsIRI() {
		if featuredURI := featured.GetIRI(); // nocollapse
		featuredURI != nil && dns.CompareDomainName(acct.Domain, featuredURI.Host) >= 2 {
			acct.FeaturedCollectionURI = featuredURI.String()
		}
	}

	// TODO: FeaturedTagsURI

	// TODO: alsoKnownAs

	// publicKey
	pkey, pkeyURL, pkeyOwnerID, err := ap.ExtractPublicKey(accountable)
	if err != nil {
		return nil, fmt.Errorf("couldn't get public key for person %s: %s", uri.String(), err)
	}

	if pkeyOwnerID.String() != acct.URI {
		return nil, fmt.Errorf("public key %s was owned by %s and not by %s", pkeyURL, pkeyOwnerID, acct.URI)
	}

	acct.PublicKey = pkey
	acct.PublicKeyURI = pkeyURL.String()

	return acct, nil
}

func (c *Converter) extractAttachments(i ap.WithAttachment) []*gtsmodel.MediaAttachment {
	attachmentProp := i.GetActivityStreamsAttachment()
	if attachmentProp == nil {
		return nil
	}

	attachments := make([]*gtsmodel.MediaAttachment, 0, attachmentProp.Len())

	for iter := attachmentProp.Begin(); iter != attachmentProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}

		attachmentable, ok := t.(ap.Attachmentable)
		if !ok {
			log.Error(nil, "ap attachment was not attachmentable")
			continue
		}

		attachment, err := ap.ExtractAttachment(attachmentable)
		if err != nil {
			log.Errorf(nil, "error extracting attachment: %s", err)
			continue
		}

		attachments = append(attachments, attachment)
	}

	return attachments
}

// ASStatus converts a remote activitystreams 'status' representation into a gts model status.
func (c *Converter) ASStatusToStatus(ctx context.Context, statusable ap.Statusable) (*gtsmodel.Status, error) {
	status := new(gtsmodel.Status)

	// status.URI
	//
	// ActivityPub ID/URI of this status.
	idProp := statusable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, gtserror.New("no id property found, or id was not an iri")
	}
	status.URI = idProp.GetIRI().String()

	l := log.WithContext(ctx).
		WithField("statusURI", status.URI)

	// status.URL
	//
	// Web URL of this status (optional).
	if statusURL, err := ap.ExtractURL(statusable); err == nil {
		status.URL = statusURL.String()
	} else {
		status.URL = status.URI // Fall back to the URI.
	}

	// status.Content
	//
	// The (html-formatted) content of this status.
	status.Content = ap.ExtractContent(statusable)

	// status.Attachments
	//
	// Media attachments for later dereferencing.
	status.Attachments = c.extractAttachments(statusable)

	// status.Hashtags
	//
	// Hashtags for later dereferencing.
	if hashtags, err := ap.ExtractHashtags(statusable); err != nil {
		l.Warnf("error extracting hashtags: %v", err)
	} else {
		status.Tags = hashtags
	}

	// status.Emojis
	//
	// Custom emojis for later dereferencing.
	if emojis, err := ap.ExtractEmojis(statusable); err != nil {
		l.Warnf("error extracting emojis: %v", err)
	} else {
		status.Emojis = emojis
	}

	// status.Mentions
	//
	// Mentions of other accounts for later dereferencing.
	if mentions, err := ap.ExtractMentions(statusable); err != nil {
		l.Warnf("error extracting mentions: %v", err)
	} else {
		status.Mentions = mentions
	}

	// status.ContentWarning
	//
	// Topic or content warning for this status;
	// prefer Summary, fall back to Name.
	if summary := ap.ExtractSummary(statusable); summary != "" {
		status.ContentWarning = summary
	} else {
		status.ContentWarning = ap.ExtractName(statusable)
	}

	// status.Published
	//
	// Publication time of this status. Thanks to
	// db defaults, will fall back to now if not set.
	published, err := ap.ExtractPublished(statusable)
	if err != nil {
		l.Warnf("error extracting published: %v", err)
	} else {
		status.CreatedAt = published
		status.UpdatedAt = published
	}

	// status.AccountURI
	// status.AccountID
	// status.Account
	//
	// Account that created the status. Assume we have
	// this in the db by the time this function is called,
	// error if we don't.
	attributedTo, err := ap.ExtractAttributedToURI(statusable)
	if err != nil {
		return nil, gtserror.Newf("%w", err)
	}
	accountURI := attributedTo.String()

	account, err := c.state.DB.GetAccountByURI(ctx, accountURI)
	if err != nil {
		err = gtserror.Newf("db error getting status author account %s: %w", accountURI, err)
		return nil, err
	}
	status.AccountURI = accountURI
	status.AccountID = account.ID
	status.Account = account

	// status.InReplyToURI
	// status.InReplyToID
	// status.InReplyTo
	// status.InReplyToAccountID
	// status.InReplyToAccount
	//
	// Status that this status replies to, if applicable.
	// If we don't have this status in the database, we
	// just set the URI and assume we can deref it later.
	if uri := ap.ExtractInReplyToURI(statusable); uri != nil {
		inReplyToURI := uri.String()
		status.InReplyToURI = inReplyToURI

		// Check if we already have the replied-to status.
		inReplyTo, err := c.state.DB.GetStatusByURI(ctx, inReplyToURI)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			// Real database error.
			err = gtserror.Newf("db error getting replied-to status %s: %w", inReplyToURI, err)
			return nil, err
		}

		if inReplyTo != nil {
			// We have it in the DB! Set
			// appropriate fields here and now.
			status.InReplyToID = inReplyTo.ID
			status.InReplyTo = inReplyTo
			status.InReplyToAccountID = inReplyTo.AccountID
			status.InReplyToAccount = inReplyTo.Account
		}
	}

	// status.Visibility
	visibility, err := ap.ExtractVisibility(
		statusable,
		status.Account.FollowersURI,
	)
	if err != nil {
		err = gtserror.Newf("error extracting visibility: %w", err)
		return nil, err
	}
	status.Visibility = visibility

	// Advanced visibility toggles for this status.
	//
	// TODO: a lot of work to be done here -- a new type
	// needs to be created for this in go-fed/activity.
	// Until this is implemented, assume all true.
	status.Federated = util.Ptr(true)
	status.Boostable = util.Ptr(true)
	status.Replyable = util.Ptr(true)
	status.Likeable = util.Ptr(true)

	// status.Sensitive
	status.Sensitive = func() *bool {
		s := ap.ExtractSensitive(statusable)
		return &s
	}()

	// language
	// TODO: we might be able to extract this from the contentMap field

	// ActivityStreamsType
	status.ActivityStreamsType = statusable.GetTypeName()

	return status, nil
}

// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow request.
func (c *Converter) ASFollowToFollowRequest(ctx context.Context, followable ap.Followable) (*gtsmodel.FollowRequest, error) {
	idProp := followable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on follow, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActorURI(followable)
	if err != nil {
		return nil, errors.New("error extracting actor property from follow")
	}
	originAccount, err := c.state.DB.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObjectURI(followable)
	if err != nil {
		return nil, errors.New("error extracting object property from follow")
	}
	targetAccount, err := c.state.DB.GetAccountByURI(ctx, target.String())
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

// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow.
func (c *Converter) ASFollowToFollow(ctx context.Context, followable ap.Followable) (*gtsmodel.Follow, error) {
	idProp := followable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on follow, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActorURI(followable)
	if err != nil {
		return nil, errors.New("error extracting actor property from follow")
	}
	originAccount, err := c.state.DB.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObjectURI(followable)
	if err != nil {
		return nil, errors.New("error extracting object property from follow")
	}
	targetAccount, err := c.state.DB.GetAccountByURI(ctx, target.String())
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

// ASLikeToFave converts a remote activitystreams 'like' representation into a gts model status fave.
func (c *Converter) ASLikeToFave(ctx context.Context, likeable ap.Likeable) (*gtsmodel.StatusFave, error) {
	idProp := likeable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on like, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActorURI(likeable)
	if err != nil {
		return nil, errors.New("error extracting actor property from like")
	}
	originAccount, err := c.state.DB.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObjectURI(likeable)
	if err != nil {
		return nil, errors.New("error extracting object property from like")
	}

	targetStatus, err := c.state.DB.GetStatusByURI(ctx, target.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting status with uri %s from the database: %s", target.String(), err)
	}

	var targetAccount *gtsmodel.Account
	if targetStatus.Account != nil {
		targetAccount = targetStatus.Account
	} else {
		a, err := c.state.DB.GetAccountByID(ctx, targetStatus.AccountID)
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

// ASBlockToBlock converts a remote activity streams 'block' representation into a gts model block.
func (c *Converter) ASBlockToBlock(ctx context.Context, blockable ap.Blockable) (*gtsmodel.Block, error) {
	idProp := blockable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("ASBlockToBlock: no id property set on block, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := ap.ExtractActorURI(blockable)
	if err != nil {
		return nil, errors.New("ASBlockToBlock: error extracting actor property from block")
	}
	originAccount, err := c.state.DB.GetAccountByURI(ctx, origin.String())
	if err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := ap.ExtractObjectURI(blockable)
	if err != nil {
		return nil, errors.New("ASBlockToBlock: error extracting object property from block")
	}

	targetAccount, err := c.state.DB.GetAccountByURI(ctx, target.String())
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

// ASAnnounceToStatus converts an activitystreams 'announce' into a status.
//
// The returned bool indicates whether this status is new (true) or not new (false).
//
// In other words, if the status is already in the database with the ID set on the announceable, then that will be returned,
// the returned bool will be false, and no further processing is necessary. If the returned bool is true, indicating
// that this is a new announce, then further processing will be necessary, because the returned status will be bareboned and
// require further dereferencing.
//
// This is useful when multiple users on an instance might receive the same boost, and we only want to process the boost once.
//
// NOTE -- this is different from one status being boosted multiple times! In this case, new boosts should indeed be created.
//
// Implementation note: this function creates and returns a boost WRAPPER
// status which references the boosted status in its BoostOf field. No
// dereferencing is done on the boosted status by this function. Callers
// should look at `status.BoostOf` to see the status being boosted, and do
// dereferencing on it as appropriate.
//
// The returned boolean indicates whether or not the boost has already been
// seen before by this instance. If it was, then status.BoostOf should be a
// fully filled-out status. If not, then only status.BoostOf.URI will be set.
func (c *Converter) ASAnnounceToStatus(ctx context.Context, announceable ap.Announceable) (*gtsmodel.Status, bool, error) {
	// Ensure item has an ID URI set.
	_, statusURIStr, err := getURI(announceable)
	if err != nil {
		err = gtserror.Newf("error extracting URI: %w", err)
		return nil, false, err
	}

	var (
		status *gtsmodel.Status
		isNew  bool
	)

	// Check if we already have this boost in the database.
	status, err = c.state.DB.GetStatusByURI(ctx, statusURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real database error.
		err = gtserror.Newf("db error trying to get status with uri %s: %w", statusURIStr, err)
		return nil, isNew, err
	}

	if status != nil {
		// We already have this status,
		// no need to proceed further.
		return status, isNew, nil
	}

	// If we reach here, we're dealing
	// with a boost we haven't seen before.
	isNew = true

	// Start assembling the new status
	// (we already know the URI).
	status = new(gtsmodel.Status)
	status.URI = statusURIStr

	// Get the URI of the boosted status.
	boostOfURI, err := ap.ExtractObjectURI(announceable)
	if err != nil {
		err = gtserror.Newf("error extracting Object: %w", err)
		return nil, isNew, err
	}

	// Set the URI of the boosted status on
	// the new status, for later dereferencing.
	boostOf := &gtsmodel.Status{
		URI: boostOfURI.String(),
	}
	status.BoostOf = boostOf

	// Extract published time for the boost.
	published, err := ap.ExtractPublished(announceable)
	if err != nil {
		err = gtserror.Newf("error extracting published: %w", err)
		return nil, isNew, err
	}
	status.CreatedAt = published
	status.UpdatedAt = published

	// Extract URI of the boosting account.
	accountURI, err := ap.ExtractActorURI(announceable)
	if err != nil {
		err = gtserror.Newf("error extracting Actor: %w", err)
		return nil, isNew, err
	}
	accountURIStr := accountURI.String()

	// Try to get the boosting account based on the URI.
	// This should have been dereferenced already before
	// we hit this point so we can confidently error out
	// if we don't have it.
	account, err := c.state.DB.GetAccountByURI(ctx, accountURIStr)
	if err != nil {
		err = gtserror.Newf("db error trying to get account with uri %s: %w", accountURIStr, err)
		return nil, isNew, err
	}
	status.AccountID = account.ID
	status.AccountURI = account.URI
	status.Account = account

	// Calculate intended visibility of the boost.
	visibility, err := ap.ExtractVisibility(announceable, account.FollowersURI)
	if err != nil {
		err = gtserror.Newf("error extracting visibility: %w", err)
		return nil, isNew, err
	}
	status.Visibility = visibility

	// Below IDs will all be included in the
	// boosted status, so set them empty here.
	status.AttachmentIDs = make([]string, 0)
	status.TagIDs = make([]string, 0)
	status.MentionIDs = make([]string, 0)
	status.EmojiIDs = make([]string, 0)

	// Remaining fields on the boost status will be taken
	// from the boosted status; it's not our job to do all
	// that dereferencing here.
	return status, isNew, nil
}

// ASFlagToReport converts a remote activitystreams 'flag' representation into a gts model report.
func (c *Converter) ASFlagToReport(ctx context.Context, flaggable ap.Flaggable) (*gtsmodel.Report, error) {
	// Extract flag uri.
	idProp := flaggable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("ASFlagToReport: no id property set on flaggable, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	// Extract account that created the flag / report.
	// This will usually be an instance actor.
	actor, err := ap.ExtractActorURI(flaggable)
	if err != nil {
		return nil, fmt.Errorf("ASFlagToReport: error extracting actor: %w", err)
	}
	account, err := c.state.DB.GetAccountByURI(ctx, actor.String())
	if err != nil {
		return nil, fmt.Errorf("ASFlagToReport: error in db fetching account with uri %s: %w", actor.String(), err)
	}

	// Get the content of the report.
	// For Mastodon, this will just be a string, or nothing.
	// In Misskey's case, it may also contain the URLs of
	// one or more reported statuses, so extract these too.
	content := ap.ExtractContent(flaggable)
	statusURIs := []*url.URL{}
	inlineURLs := misskeyReportInlineURLs(content)
	statusURIs = append(statusURIs, inlineURLs...)

	// Extract account and statuses targeted by the flag / report.
	//
	// Incoming flags from mastodon usually have a target account uri as
	// first entry in objects, followed by URIs of one or more statuses.
	// Misskey on the other hand will just contain the target account uri.
	// We shouldn't assume the order of the objects will correspond to this,
	// but we can check that he objects slice contains just one account, and
	// maybe some statuses.
	//
	// Throw away anything that's not relevant to us.
	objects, err := ap.ExtractObjectURIs(flaggable)
	if err != nil {
		return nil, fmt.Errorf("ASFlagToReport: error extracting objects: %w", err)
	}
	if len(objects) == 0 {
		return nil, errors.New("ASFlagToReport: flaggable objects empty, can't create report")
	}

	var targetAccountURI *url.URL
	for _, object := range objects {
		switch {
		case object.Host != config.GetHost():
			// object doesn't belong to us, just ignore it
			continue
		case uris.IsUserPath(object):
			if targetAccountURI != nil {
				return nil, errors.New("ASFlagToReport: flaggable objects contained more than one target account uri")
			}
			targetAccountURI = object
		case uris.IsStatusesPath(object):
			statusURIs = append(statusURIs, object)
		}
	}

	// Make sure we actually have a target account now.
	if targetAccountURI == nil {
		return nil, errors.New("ASFlagToReport: flaggable objects contained no recognizable target account uri")
	}
	targetAccount, err := c.state.DB.GetAccountByURI(ctx, targetAccountURI.String())
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, fmt.Errorf("ASFlagToReport: account with uri %s could not be found in the db", targetAccountURI.String())
		}
		return nil, fmt.Errorf("ASFlagToReport: db error getting account with uri %s: %w", targetAccountURI.String(), err)
	}

	// If we got some status URIs, try to get them from the db now
	var (
		statusIDs = make([]string, 0, len(statusURIs))
		statuses  = make([]*gtsmodel.Status, 0, len(statusURIs))
	)
	for _, statusURI := range statusURIs {
		statusURIString := statusURI.String()

		// try getting this status by URI first, then URL
		status, err := c.state.DB.GetStatusByURI(ctx, statusURIString)
		if err != nil {
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, fmt.Errorf("ASFlagToReport: db error getting status with uri %s: %w", statusURIString, err)
			}

			status, err = c.state.DB.GetStatusByURL(ctx, statusURIString)
			if err != nil {
				if !errors.Is(err, db.ErrNoEntries) {
					return nil, fmt.Errorf("ASFlagToReport: db error getting status with url %s: %w", statusURIString, err)
				}

				log.Warnf(nil, "reported status %s could not be found in the db, skipping it", statusURIString)
				continue
			}
		}

		if status.AccountID != targetAccount.ID {
			// status doesn't belong to this account, ignore it
			continue
		}

		statusIDs = append(statusIDs, status.ID)
		statuses = append(statuses, status)
	}

	// id etc should be handled the caller, so just return what we got
	return &gtsmodel.Report{
		URI:             uri,
		AccountID:       account.ID,
		Account:         account,
		TargetAccountID: targetAccount.ID,
		TargetAccount:   targetAccount,
		Comment:         content,
		StatusIDs:       statusIDs,
		Statuses:        statuses,
	}, nil
}
