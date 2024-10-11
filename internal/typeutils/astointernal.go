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
	"cmp"
	"context"
	"errors"
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

// ASRepresentationToAccount converts a remote account / person
// / application representation into a gts model account.
//
// If accountDomain is provided then this value will be
// used as the account's Domain, else the AP ID host.
//
// If accountUsername is provided then this is used as
// a fallback when no preferredUsername is provided. Else
// a lack of username will result in error return.
func (c *Converter) ASRepresentationToAccount(
	ctx context.Context,
	accountable ap.Accountable,
	accountDomain string,
	accountUsername string,
) (
	*gtsmodel.Account,
	error,
) {
	var err error

	// Extract URI from accountable
	uriObj := ap.GetJSONLDId(accountable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	// Create DB account with URI
	var acct gtsmodel.Account
	acct.URI = uri

	// Check whether account is a usable actor type.
	switch acct.ActorType = accountable.GetTypeName(); acct.ActorType {

	// people, groups, and organizations aren't bots
	case ap.ActorPerson, ap.ActorGroup, ap.ActorOrganization:
		acct.Bot = util.Ptr(false)

	// apps and services are
	case ap.ActorApplication, ap.ActorService:
		acct.Bot = util.Ptr(true)

	// we don't know what this is!
	default:
		err := gtserror.Newf("unusable actor type for %s", uri)
		return nil, gtserror.SetMalformed(err)
	}

	// Set account username.
	acct.Username = cmp.Or(

		// Prefer the AP model provided username.
		ap.ExtractPreferredUsername(accountable),

		// Fallback username.
		accountUsername,
	)
	if acct.Username == "" {
		err := gtserror.Newf("missing username for %s", uri)
		return nil, gtserror.SetMalformed(err)
	}

	// Extract published time if possible.
	//
	// This denotes original creation time
	// of the account on the remote instance.
	//
	// Not every implementation uses this property;
	// so don't bother warning if we can't find it.
	if pub := ap.GetPublished(accountable); !pub.IsZero() {
		acct.CreatedAt = pub
		acct.UpdatedAt = pub
	}

	// Extract a preferred name (display name), fallback to username.
	if displayName := ap.ExtractName(accountable); displayName != "" {
		acct.DisplayName = displayName
	} else {
		acct.DisplayName = acct.Username
	}

	// Check for separaate account
	// domain to the instance hostname.
	if accountDomain != "" {
		acct.Domain = accountDomain
	} else {
		acct.Domain = uriObj.Host
	}

	// avatar aka icon
	// if this one isn't extractable in a format we recognise we'll just skip it
	avatarURL, err := ap.ExtractIconURI(accountable)
	if err == nil {
		acct.AvatarRemoteURL = avatarURL.String()
	}

	// header aka image
	// if this one isn't extractable in a format we recognise we'll just skip it
	headerURL, err := ap.ExtractImageURI(accountable)
	if err == nil {
		acct.HeaderRemoteURL = headerURL.String()
	}

	// account emojis (used in bio, display name, fields)
	acct.Emojis, err = ap.ExtractEmojis(accountable)
	if err != nil {
		log.Warnf(ctx, "error(s) extracting account emojis for %s: %v", uri, err)
	}

	// Extract account attachments (key-value fields).
	acct.Fields = ap.ExtractFields(accountable)

	// Extract account note (bio / summary).
	acct.Note = ap.ExtractSummary(accountable)

	// Assume not memorial (todo)
	acct.Memorial = util.Ptr(false)

	// Extract 'manuallyApprovesFollowers' aka locked account (default = true).
	manuallyApprovesFollowers := ap.GetManuallyApprovesFollowers(accountable)
	acct.Locked = &manuallyApprovesFollowers

	// Extract account discoverability (default = false).
	discoverable := ap.GetDiscoverable(accountable)
	acct.Discoverable = &discoverable

	// Extract the URL property.
	urls := ap.GetURL(accountable)
	if len(urls) == 0 {
		// just use account uri string
		acct.URL = uri
	} else {
		// else use provided URL string
		acct.URL = urls[0].String()
	}

	// Extract the inbox IRI property.
	inboxIRI := ap.GetInbox(accountable)
	if inboxIRI != nil {
		acct.InboxURI = inboxIRI.String()
	}

	// Extract the outbox IRI property.
	outboxIRI := ap.GetOutbox(accountable)
	if outboxIRI != nil {
		acct.OutboxURI = outboxIRI.String()
	}

	// Extract a SharedInboxURI, but only trust if equal to / subdomain of account's domain.
	if sharedInboxURI := ap.ExtractSharedInbox(accountable); // nocollapse
	sharedInboxURI != nil && dns.CompareDomainName(acct.Domain, sharedInboxURI.Host) >= 2 {
		sharedInbox := sharedInboxURI.String()
		acct.SharedInboxURI = &sharedInbox
	}

	// Extract the following IRI property.
	followingURI := ap.GetFollowing(accountable)
	if followingURI != nil {
		acct.FollowingURI = followingURI.String()
	}

	// Extract the following IRI property.
	followersURI := ap.GetFollowers(accountable)
	if followersURI != nil {
		acct.FollowersURI = followersURI.String()
	}

	// Extract a FeaturedURI, but only trust if equal to / subdomain of account's domain.
	if featuredURI := ap.GetFeatured(accountable); // nocollapse
	featuredURI != nil && dns.CompareDomainName(acct.Domain, featuredURI.Host) >= 2 {
		acct.FeaturedCollectionURI = featuredURI.String()
	}

	// TODO: FeaturedTagsURI

	// Moved and AlsoKnownAsURIs,
	// needed for account migrations.
	movedToURI := ap.GetMovedTo(accountable)
	if movedToURI != nil {
		acct.MovedToURI = movedToURI.String()
	}

	alsoKnownAsURIs := ap.GetAlsoKnownAs(accountable)
	for i, uri := range alsoKnownAsURIs {
		// Don't store more than
		// 20 AKA URIs for remotes,
		// to prevent people playing
		// silly buggers.
		if i >= 20 {
			break
		}

		acct.AlsoKnownAsURIs = append(acct.AlsoKnownAsURIs, uri.String())
	}

	// Extract account public key and verify ownership to account.
	pkey, pkeyURL, pkeyOwnerID, err := ap.ExtractPubKeyFromActor(accountable)
	if err != nil {
		err := gtserror.Newf("error extracting public key for %s: %w", uri, err)
		return nil, gtserror.SetMalformed(err)
	} else if pkeyOwnerID.String() != acct.URI {
		err := gtserror.Newf("public key not owned by account %s", uri)
		return nil, gtserror.SetMalformed(err)
	}

	acct.PublicKey = pkey
	acct.PublicKeyURI = pkeyURL.String()

	return &acct, nil
}

// ASStatus converts a remote activitystreams 'status' representation into a gts model status.
func (c *Converter) ASStatusToStatus(ctx context.Context, statusable ap.Statusable) (*gtsmodel.Status, error) {
	var err error

	// Extract URI from statusable
	uriObj := ap.GetJSONLDId(statusable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	// Create DB status with URI
	var status gtsmodel.Status
	status.URI = uri

	// status.URL
	//
	// Web URL of this status (optional).
	if statusURL, err := ap.ExtractURL(statusable); err == nil {
		status.URL = statusURL.String()
	} else {
		status.URL = status.URI // Fall back to the URI.
	}

	// status.Content
	// status.Language
	//
	// Many implementations set both content
	// and contentMap; we can use these to
	// infer the language of the status.
	status.Content, status.Language = ContentToContentLanguage(
		ctx,
		ap.ExtractContent(statusable),
	)

	// status.Attachments
	//
	// Media attachments for later dereferencing.
	status.Attachments, err = ap.ExtractAttachments(statusable)
	if err != nil {
		log.Warnf(ctx, "error(s) extracting attachments for %s: %v", uri, err)
	}

	// status.Poll
	//
	// Attached poll information (the statusable will actually
	// be a Pollable, as a Question is a subset of our Status).
	if pollable, ok := ap.ToPollable(statusable); ok {
		status.Poll, err = ap.ExtractPoll(pollable)
		if err != nil {
			log.Warnf(ctx, "error(s) extracting poll for %s: %v", uri, err)
		}
	}

	// status.Hashtags
	//
	// Hashtags for later dereferencing.
	if hashtags, err := ap.ExtractHashtags(statusable); err != nil {
		log.Warnf(ctx, "error extracting hashtags for %s: %v", uri, err)
	} else {
		status.Tags = hashtags
	}

	// status.Emojis
	//
	// Custom emojis for later dereferencing.
	if emojis, err := ap.ExtractEmojis(statusable); err != nil {
		log.Warnf(ctx, "error extracting emojis for %s: %v", uri, err)
	} else {
		status.Emojis = emojis
	}

	// status.Mentions
	//
	// Mentions of other accounts for later dereferencing.
	if mentions, err := ap.ExtractMentions(statusable); err != nil {
		log.Warnf(ctx, "error extracting mentions for %s: %v", uri, err)
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
	// Extract published time for the status,
	// zero-time will fall back to db defaults.
	if pub := ap.GetPublished(statusable); !pub.IsZero() {
		status.CreatedAt = pub
		status.UpdatedAt = pub
	} else {
		log.Warnf(ctx, "unusable published property on %s", uri)
	}

	// status.AccountURI
	// status.AccountID
	// status.Account
	//
	// Account that created the status. Assume we have
	// this in the db by the time this function is called,
	// error if we don't.
	status.Account, err = c.getASAttributedToAccount(ctx,
		status.URI,
		statusable,
	)
	if err != nil {
		return nil, err
	}

	// Set the related status<->account fields.
	status.AccountURI = status.Account.URI
	status.AccountID = status.Account.ID

	// status.InReplyToURI
	// status.InReplyToID
	// status.InReplyTo
	// status.InReplyToAccountID
	// status.InReplyToAccount
	//
	// Status that this status replies to, if applicable.
	// If we don't have this status in the database, we
	// just set the URI and assume we can deref it later.
	inReplyTo := ap.GetInReplyTo(statusable)
	if len(inReplyTo) > 0 {

		// Extract the URI from inReplyTo slice.
		inReplyToURI := inReplyTo[0].String()
		status.InReplyToURI = inReplyToURI

		// Check if we already have the replied-to status.
		inReplyTo, err := c.state.DB.GetStatusByURI(ctx, inReplyToURI)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting reply %s from db: %w", inReplyToURI, err)
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

	// Calculate intended visibility of the status.
	status.Visibility, err = ap.ExtractVisibility(
		statusable,
		status.Account.FollowersURI,
	)
	if err != nil {
		err := gtserror.Newf("error extracting status visibility for %s: %w", uri, err)
		return nil, gtserror.SetMalformed(err)
	}

	// Status was sent to us or dereffed
	// by us so it must be federated.
	status.Federated = util.Ptr(true)

	// Derive interaction policy for this status.
	status.InteractionPolicy = ap.ExtractInteractionPolicy(
		statusable,
		status.Account,
	)

	// Set approvedByURI if present,
	// for later dereferencing.
	approvedByURI := ap.GetApprovedBy(statusable)
	if approvedByURI != nil {
		status.ApprovedByURI = approvedByURI.String()
	}

	// Assume not pending approval; this may
	// change when permissivity is checked.
	status.PendingApproval = util.Ptr(false)

	// status.Sensitive
	sensitive := ap.ExtractSensitive(statusable)
	status.Sensitive = &sensitive

	// ActivityStreamsType
	status.ActivityStreamsType = statusable.GetTypeName()

	return &status, nil
}

// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow request.
func (c *Converter) ASFollowToFollowRequest(ctx context.Context, followable ap.Followable) (*gtsmodel.FollowRequest, error) {
	uriObj := ap.GetJSONLDId(followable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	origin, err := c.getASActorAccount(ctx, uri, followable)
	if err != nil {
		return nil, err
	}

	target, err := c.getASObjectAccount(ctx, uri, followable)
	if err != nil {
		return nil, err
	}

	followRequest := &gtsmodel.FollowRequest{
		URI:             uri,
		AccountID:       origin.ID,
		TargetAccountID: target.ID,
	}

	return followRequest, nil
}

// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow.
func (c *Converter) ASFollowToFollow(ctx context.Context, followable ap.Followable) (*gtsmodel.Follow, error) {
	uriObj := ap.GetJSONLDId(followable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	origin, err := c.getASActorAccount(ctx, uri, followable)
	if err != nil {
		return nil, err
	}

	target, err := c.getASObjectAccount(ctx, uri, followable)
	if err != nil {
		return nil, err
	}

	follow := &gtsmodel.Follow{
		URI:             uri,
		AccountID:       origin.ID,
		TargetAccountID: target.ID,
	}

	return follow, nil
}

// ASLikeToFave converts a remote activitystreams 'like' representation into a gts model status fave.
func (c *Converter) ASLikeToFave(ctx context.Context, likeable ap.Likeable) (*gtsmodel.StatusFave, error) {
	uriObj := ap.GetJSONLDId(likeable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	origin, err := c.getASActorAccount(ctx, uri, likeable)
	if err != nil {
		return nil, err
	}

	target, err := c.getASObjectStatus(ctx, uri, likeable)
	if err != nil {
		return nil, err
	}

	return &gtsmodel.StatusFave{
		AccountID:       origin.ID,
		Account:         origin,
		TargetAccountID: target.AccountID,
		TargetAccount:   target.Account,
		StatusID:        target.ID,
		Status:          target,
		URI:             uri,

		// Assume not pending approval; this may
		// change when permissivity is checked.
		PendingApproval: util.Ptr(false),
	}, nil
}

// ASBlockToBlock converts a remote activity streams 'block' representation into a gts model block.
func (c *Converter) ASBlockToBlock(ctx context.Context, blockable ap.Blockable) (*gtsmodel.Block, error) {
	uriObj := ap.GetJSONLDId(blockable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	origin, err := c.getASActorAccount(ctx, uri, blockable)
	if err != nil {
		return nil, err
	}

	target, err := c.getASObjectAccount(ctx, uri, blockable)
	if err != nil {
		return nil, err
	}

	return &gtsmodel.Block{
		AccountID:       origin.ID,
		Account:         origin,
		TargetAccountID: target.ID,
		TargetAccount:   target,
		URI:             uri,
	}, nil
}

// ASAnnounceToStatus converts an activitystreams 'announce' into a boost
// wrapper status. The returned bool indicates whether this boost is new
// (true) or not. If new, callers should use `status.BoostOfURI` to see the
// status being boosted, and do dereferencing on it as appropriate. If not
// new, then the boost has already been fully processed and can be ignored.
func (c *Converter) ASAnnounceToStatus(
	ctx context.Context,
	announceable ap.Announceable,
) (*gtsmodel.Status, bool, error) {
	// Default assume
	// we already have.
	isNew := false

	// Extract uri ID from announceable.
	uriObj := ap.GetJSONLDId(announceable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, isNew, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	// Check if we already have this boost in the database.
	boost, err := c.state.DB.GetStatusByURI(ctx, uri)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error trying to get status with uri %s: %w", uri, err)
		return nil, isNew, err
	}

	if boost != nil {
		// We already have this boost,
		// no need to proceed further.
		return boost, isNew, nil
	}

	// Create boost with URI
	boost = new(gtsmodel.Status)
	boost.URI = uri
	isNew = true

	// Get the URI of the boosted status.
	boostOf := ap.GetObjectIRIs(announceable)
	if len(boostOf) == 0 {
		err := gtserror.Newf("unusable object property iri for %s", uri)
		return nil, isNew, gtserror.SetMalformed(err)
	}

	// Set the URI of the boosted status on
	// the boost, for later dereferencing.
	boost.BoostOfURI = boostOf[0].String()

	// Extract published time for the boost,
	// zero-time will fall back to db defaults.
	if pub := ap.GetPublished(announceable); !pub.IsZero() {
		boost.CreatedAt = pub
		boost.UpdatedAt = pub
	} else {
		log.Warnf(ctx, "unusable published property on %s", uri)
	}

	// Extract and load the boost actor account,
	// (this MUST already be in database by now).
	boost.Account, err = c.getASActorAccount(ctx,
		uri,
		announceable,
	)
	if err != nil {
		return nil, isNew, err
	}

	// Set the related status<->account fields.
	boost.AccountURI = boost.Account.URI
	boost.AccountID = boost.Account.ID

	// Calculate intended visibility of the boost.
	boost.Visibility, err = ap.ExtractVisibility(
		announceable,
		boost.Account.FollowersURI,
	)
	if err != nil {
		err := gtserror.Newf("error extracting status visibility for %s: %w", uri, err)
		return nil, isNew, gtserror.SetMalformed(err)
	}

	// Below IDs will all be included in the
	// boosted status, so set them empty here.
	boost.AttachmentIDs = make([]string, 0)
	boost.TagIDs = make([]string, 0)
	boost.MentionIDs = make([]string, 0)
	boost.EmojiIDs = make([]string, 0)

	// Assume not pending approval; this may
	// change when permissivity is checked.
	boost.PendingApproval = util.Ptr(false)

	// Remaining fields on the boost will be
	// taken from the target status; it's not
	// our job to do all that dereferencing here.
	return boost, isNew, nil
}

// ASFlagToReport converts a remote activitystreams 'flag' representation into a gts model report.
func (c *Converter) ASFlagToReport(ctx context.Context, flaggable ap.Flaggable) (*gtsmodel.Report, error) {
	uriObj := ap.GetJSONLDId(flaggable)
	if uriObj == nil {
		err := gtserror.New("unusable iri property")
		return nil, gtserror.SetMalformed(err)
	}

	// Stringify uri obj.
	uri := uriObj.String()

	// Extract the origin (actor) account for report.
	origin, err := c.getASActorAccount(ctx, uri, flaggable)
	if err != nil {
		return nil, err
	}

	var (
		// Gathered from objects
		// (+ content for misskey).
		statusURIs   []*url.URL
		targetAccURI *url.URL

		// Get current hostname.
		host = config.GetHost()
	)

	// Get the content of the report.
	// For Mastodon, this will just be a string, or nothing.
	// In Misskey's case, it may also contain the URLs of
	// one or more reported statuses, so extract these too.
	content := ap.ExtractContent(flaggable).Content
	statusURIs = misskeyReportInlineURLs(content)

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
	objects := ap.GetObjectIRIs(flaggable)
	if len(objects) == 0 {
		err := gtserror.Newf("unusable object property iris for %s", uri)
		return nil, gtserror.SetMalformed(err)
	}

	for _, object := range objects {
		switch {
		case object.Host != host:
			// object doesn't belong
			// to us, just ignore it
			continue

		case uris.IsUserPath(object):
			if targetAccURI != nil {
				err := gtserror.Newf("multiple target account uris for %s", uri)
				return nil, gtserror.SetMalformed(err)
			}
			targetAccURI = object

		case uris.IsStatusesPath(object):
			statusURIs = append(statusURIs, object)
		}
	}

	// Ensure we have a target.
	if targetAccURI == nil {
		err := gtserror.Newf("missing target account uri for %s", uri)
		return nil, gtserror.SetMalformed(err)
	}

	// Fetch target account from the database by its URI.
	targetAcc, err := c.state.DB.GetAccountByURI(ctx, targetAccURI.String())
	if err != nil {
		return nil, gtserror.Newf("error getting target account %s from database: %w", targetAccURI, err)
	}

	var (
		// Preallocate expected status + IDs slice lengths.
		statusIDs = make([]string, 0, len(statusURIs))
		statuses  = make([]*gtsmodel.Status, 0, len(statusURIs))
	)

	for _, statusURI := range statusURIs {
		// Rescope as just the URI string.
		statusURI := statusURI.String()

		// Try getting status by URI from database.
		status, err := c.state.DB.GetStatusByURI(ctx, statusURI)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting target status %s from database: %w", statusURI, err)
			return nil, err
		}

		if status == nil {
			// Status was not found, try again with URL.
			status, err = c.state.DB.GetStatusByURL(ctx, statusURI)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				err := gtserror.Newf("error getting target status %s from database: %w", statusURI, err)
				return nil, err
			}

			if status == nil {
				log.Warnf(ctx, "missing target status %s for %s", statusURI, uri)
				continue
			}
		}

		if status.AccountID != targetAcc.ID {
			// status doesn't belong
			// to target, ignore it.
			continue
		}

		// Append the discovered status to slices.
		statusIDs = append(statusIDs, status.ID)
		statuses = append(statuses, status)
	}

	// id etc should be handled the caller,
	// so just return what we got
	return &gtsmodel.Report{
		URI:             uri,
		AccountID:       origin.ID,
		Account:         origin,
		TargetAccountID: targetAcc.ID,
		TargetAccount:   targetAcc,
		Comment:         content,
		StatusIDs:       statusIDs,
		Statuses:        statuses,
	}, nil
}

func (c *Converter) getASActorAccount(ctx context.Context, id string, with ap.WithActor) (*gtsmodel.Account, error) {
	// Get actor IRIs from type.
	actor := ap.GetActorIRIs(with)
	if len(actor) == 0 {
		err := gtserror.Newf("unusable actor property iri for %s", id)
		return nil, gtserror.SetMalformed(err)
	}

	// Check for account in database with provided actor URI.
	account, err := c.state.DB.GetAccountByURI(ctx, actor[0].String())
	if err != nil {
		return nil, gtserror.Newf("error getting actor account from database: %w", err)
	}

	return account, nil
}

func (c *Converter) getASAttributedToAccount(ctx context.Context, id string, with ap.WithAttributedTo) (*gtsmodel.Account, error) {
	// Get attribTo IRIs from type.
	attribTo := ap.GetAttributedTo(with)
	if len(attribTo) == 0 {
		err := gtserror.Newf("unusable attributedTo property iri for %s", id)
		return nil, gtserror.SetMalformed(err)
	}

	// Check for account in database with provided attributedTo URI.
	account, err := c.state.DB.GetAccountByURI(ctx, attribTo[0].String())
	if err != nil {
		return nil, gtserror.Newf("error getting actor account from database: %w", err)
	}

	return account, nil
}

func (c *Converter) getASObjectAccount(ctx context.Context, id string, with ap.WithObject) (*gtsmodel.Account, error) {
	// Get object IRIs from type.
	object := ap.GetObjectIRIs(with)
	if len(object) == 0 {
		err := gtserror.Newf("unusable object property iri for %s", id)
		return nil, gtserror.SetMalformed(err)
	}

	// Check for account in database with provided object URI.
	account, err := c.state.DB.GetAccountByURI(ctx, object[0].String())
	if err != nil {
		return nil, gtserror.Newf("error getting object account from database: %w", err)
	}

	return account, nil
}

func (c *Converter) getASObjectStatus(ctx context.Context, id string, with ap.WithObject) (*gtsmodel.Status, error) {
	// Get object IRIs from type.
	object := ap.GetObjectIRIs(with)
	if len(object) == 0 {
		err := gtserror.Newf("unusable object property iri for %s", id)
		return nil, gtserror.SetMalformed(err)
	}

	// Check for status in database with provided object URI.
	status, err := c.state.DB.GetStatusByURI(ctx, object[0].String())
	if err != nil {
		return nil, gtserror.Newf("error getting object status from database: %w", err)
	}

	return status, nil
}
