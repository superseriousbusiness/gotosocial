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

package dereferencing

import (
	"context"
	"errors"
	"io"
	"net/url"
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// statusUpToDate returns whether the given status model is both updateable
// (i.e. remote status) and whether it needs an update based on `fetched_at`.
func statusUpToDate(status *gtsmodel.Status) bool {
	if *status.Local {
		// Can't update local statuses.
		return true
	}

	// If this status was updated recently (last interval), we return as-is.
	if next := status.FetchedAt.Add(2 * time.Hour); time.Now().Before(next) {
		return true
	}

	return false
}

// GetStatusByURI will attempt to fetch a status by its URI, first checking the database. In the case of a newly-met remote model, or a remote model
// whose last_fetched date is beyond a certain interval, the status will be dereferenced. In the case of dereferencing, some low-priority status information
// may be enqueued for asynchronous fetching, e.g. dereferencing the remainder of the status thread. An ActivityPub object indicates the status was dereferenced.
func (d *Dereferencer) GetStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error) {
	// Fetch and dereference status if necessary.
	status, apubStatus, err := d.getStatusByURI(ctx,
		requestUser,
		uri,
	)
	if err != nil {
		return nil, nil, err
	}

	if apubStatus != nil {
		// This status was updated, enqueue re-dereferencing the whole thread.
		d.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
			d.dereferenceThread(ctx, requestUser, uri, status, apubStatus)
		})
	}

	return status, apubStatus, nil
}

// getStatusByURI is a package internal form of .GetStatusByURI() that doesn't bother dereferencing the whole thread on update.
func (d *Dereferencer) getStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error) {
	var (
		status *gtsmodel.Status
		uriStr = uri.String()
		err    error
	)

	// Search the database for existing status with URI.
	status, err = d.state.DB.GetStatusByURI(
		// request a barebones object, it may be in the
		// db but with related models not yet dereferenced.
		gtscontext.SetBarebones(ctx),
		uriStr,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, nil, gtserror.Newf("error checking database for status %s by uri: %w", uriStr, err)
	}

	if status == nil {
		// Else, search the database for existing by URL.
		status, err = d.state.DB.GetStatusByURL(
			gtscontext.SetBarebones(ctx),
			uriStr,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, nil, gtserror.Newf("error checking database for status %s by url: %w", uriStr, err)
		}
	}

	if status == nil {
		// Ensure that this isn't a search for a local status.
		if uri.Host == config.GetHost() || uri.Host == config.GetAccountDomain() {
			return nil, nil, gtserror.SetUnretrievable(err) // this will be db.ErrNoEntries
		}

		// Create and pass-through a new bare-bones model for deref.
		return d.enrichStatusSafely(ctx, requestUser, uri, &gtsmodel.Status{
			Local: func() *bool { var false bool; return &false }(),
			URI:   uriStr,
		}, nil)
	}

	// Check whether needs update.
	if statusUpToDate(status) {
		// This is existing up-to-date status, ensure it is populated.
		if err := d.state.DB.PopulateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error populating existing status: %v", err)
		}
		return status, nil, nil
	}

	// Try to update + deref existing status model.
	latest, apubStatus, err := d.enrichStatusSafely(ctx,
		requestUser,
		uri,
		status,
		nil,
	)
	if err != nil {
		log.Errorf(ctx, "error enriching remote status: %v", err)

		// Fallback to existing.
		return status, nil, nil
	}

	return latest, apubStatus, nil
}

// RefreshStatus updates the given status if remote and last_fetched is beyond fetch interval, or if force is set. An updated status model is returned,
// but in the case of dereferencing, some low-priority status information may be enqueued for asynchronous fetching, e.g. dereferencing the remainder of the
// status thread. An ActivityPub object indicates the status was dereferenced (i.e. updated).
func (d *Dereferencer) RefreshStatus(ctx context.Context, requestUser string, status *gtsmodel.Status, apubStatus ap.Statusable, force bool) (*gtsmodel.Status, ap.Statusable, error) {
	// Check whether needs update.
	if !force && statusUpToDate(status) {
		return status, nil, nil
	}

	// Parse the URI from status.
	uri, err := url.Parse(status.URI)
	if err != nil {
		return nil, nil, gtserror.Newf("invalid status uri %q: %w", status.URI, err)
	}

	// Try to update + deref the passed status model.
	latest, apubStatus, err := d.enrichStatusSafely(ctx,
		requestUser,
		uri,
		status,
		apubStatus,
	)
	if err != nil {
		return nil, nil, err
	}

	// This status was updated, enqueue re-dereferencing the whole thread.
	d.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
		d.dereferenceThread(ctx, requestUser, uri, latest, apubStatus)
	})

	return latest, apubStatus, nil
}

// RefreshStatusAsync enqueues the given status for an asychronous update fetching, if last_fetched is beyond fetch interval, or if force is set.
// This is a more optimized form of manually enqueueing .UpdateStatus() to the federation worker, since it only enqueues update if necessary.
func (d *Dereferencer) RefreshStatusAsync(ctx context.Context, requestUser string, status *gtsmodel.Status, apubStatus ap.Statusable, force bool) {
	// Check whether needs update.
	if !force && statusUpToDate(status) {
		return
	}

	// Parse the URI from status.
	uri, err := url.Parse(status.URI)
	if err != nil {
		log.Errorf(ctx, "invalid status uri %q: %v", status.URI, err)
		return
	}

	// Enqueue a worker function to re-fetch this status async.
	d.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
		latest, apubStatus, err := d.enrichStatusSafely(ctx, requestUser, uri, status, apubStatus)
		if err != nil {
			log.Errorf(ctx, "error enriching remote status: %v", err)
			return
		}

		if apubStatus != nil {
			// This status was updated, re-dereference the whole thread.
			d.dereferenceThread(ctx, requestUser, uri, latest, apubStatus)
		}
	})
}

// enrichStatusSafely wraps enrichStatus() to perform
// it within the State{}.FedLocks mutexmap, which protects
// dereferencing actions with per-URI mutex locks.
func (d *Dereferencer) enrichStatusSafely(
	ctx context.Context,
	requestUser string,
	uri *url.URL,
	status *gtsmodel.Status,
	apubStatus ap.Statusable,
) (*gtsmodel.Status, ap.Statusable, error) {
	uriStr := status.URI

	if status.ID != "" {
		// This is an existing status, first try to populate it. This
		// is required by the checks below for existing tags, media etc.
		if err := d.state.DB.PopulateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error populating existing status %s: %v", uriStr, err)
		}
	}

	// Acquire per-URI deref lock, wraping unlock
	// to safely defer in case of panic, while still
	// performing more granular unlocks when needed.
	unlock := d.state.FedLocks.Lock(uriStr)
	unlock = doOnce(unlock)
	defer unlock()

	// Perform status enrichment with passed vars.
	latest, apubStatus, err := d.enrichStatus(ctx,
		requestUser,
		uri,
		status,
		apubStatus,
	)

	if gtserror.StatusCode(err) >= 400 {
		// Update fetch-at to slow re-attempts.
		status.FetchedAt = time.Now()
		_ = d.state.DB.UpdateStatus(ctx, status, "fetched_at")
	}

	// Unlock now
	// we're done.
	unlock()

	if errors.Is(err, db.ErrAlreadyExists) {
		// Ensure AP model isn't set,
		// otherwise this indicates WE
		// enriched the status.
		apubStatus = nil

		// DATA RACE! We likely lost out to another goroutine
		// in a call to db.Put(Status). Look again in DB by URI.
		latest, err = d.state.DB.GetStatusByURI(ctx, status.URI)
		if err != nil {
			err = gtserror.Newf("error getting status %s from database after race: %w", uriStr, err)
		}
	}

	return latest, apubStatus, err
}

// enrichStatus will enrich the given status, whether a new
// barebones model, or existing model from the database.
// It handles necessary dereferencing, database updates, etc.
func (d *Dereferencer) enrichStatus(
	ctx context.Context,
	requestUser string,
	uri *url.URL,
	status *gtsmodel.Status,
	apubStatus ap.Statusable,
) (*gtsmodel.Status, ap.Statusable, error) {
	// Pre-fetch a transport for requesting username, used by later dereferencing.
	tsport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
	if err != nil {
		return nil, nil, gtserror.Newf("couldn't create transport: %w", err)
	}

	// Check whether this account URI is a blocked domain / subdomain.
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, uri.Host); err != nil {
		return nil, nil, gtserror.Newf("error checking blocked domain: %w", err)
	} else if blocked {
		err = gtserror.Newf("%s is blocked", uri.Host)
		return nil, nil, gtserror.SetUnretrievable(err)
	}

	if apubStatus == nil {
		// Dereference latest version of the status.
		b, err := tsport.Dereference(ctx, uri)
		if err != nil {
			err := gtserror.Newf("error deferencing %s: %w", uri, err)
			return nil, nil, gtserror.SetUnretrievable(err)
		}

		// Attempt to resolve ActivityPub status from data.
		apubStatus, err = ap.ResolveStatusable(ctx, b)
		if err != nil {
			return nil, nil, gtserror.Newf("error resolving statusable from data for account %s: %w", uri, err)
		}
	}

	// Get the attributed-to account in order to fetch profile.
	attributedTo, err := ap.ExtractAttributedToURI(apubStatus)
	if err != nil {
		return nil, nil, gtserror.New("attributedTo was empty")
	}

	// Ensure we have the author account of the status dereferenced (+ up-to-date). If this is a new status
	// (i.e. status.AccountID == "") then any error here is irrecoverable. AccountID must ALWAYS be set.
	if _, _, err := d.getAccountByURI(ctx, requestUser, attributedTo); err != nil && status.AccountID == "" {
		return nil, nil, gtserror.Newf("failed to dereference status author %s: %w", uri, err)
	}

	// ActivityPub model was recently dereferenced, so assume that passed status
	// may contain out-of-date information, convert AP model to our GTS model.
	latestStatus, err := d.converter.ASStatusToStatus(ctx, apubStatus)
	if err != nil {
		return nil, nil, gtserror.Newf("error converting statusable to gts model for status %s: %w", uri, err)
	}

	// Use existing status ID.
	latestStatus.ID = status.ID

	if latestStatus.ID == "" {
		// Generate new status ID from the provided creation date.
		latestStatus.ID, err = id.NewULIDFromTime(latestStatus.CreatedAt)
		if err != nil {
			return nil, nil, gtserror.Newf("invalid created at date: %w", err)
		}
	}

	// Carry-over values and set fetch time.
	latestStatus.FetchedAt = time.Now()
	latestStatus.Local = status.Local

	// Ensure the status' poll remains consistent, else reset the poll.
	if err := d.fetchStatusPoll(ctx, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating poll for status %s: %w", uri, err)
	}

	// Ensure the status' mentions are populated, and pass in existing to check for changes.
	if err := d.fetchStatusMentions(ctx, requestUser, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating mentions for status %s: %w", uri, err)
	}

	// Now that we know who this status replies to (handled by ASStatusToStatus)
	// and who it mentions, we can add a ThreadID to it if necessary.
	if err := d.threadStatus(ctx, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error checking / creating threadID for status %s: %w", uri, err)
	}

	// Ensure the status' tags are populated, (changes are expected / okay).
	if err := d.fetchStatusTags(ctx, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating tags for status %s: %w", uri, err)
	}

	// Ensure the status' media attachments are populated, passing in existing to check for changes.
	if err := d.fetchStatusAttachments(ctx, tsport, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating attachments for status %s: %w", uri, err)
	}

	// Ensure the status' emoji attachments are populated, (changes are expected / okay).
	if err := d.fetchStatusEmojis(ctx, requestUser, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating emojis for status %s: %w", uri, err)
	}

	if status.CreatedAt.IsZero() {
		// CreatedAt will be zero if no local copy was
		// found in one of the GetStatusBy___() functions.
		//
		// This is new, put the status in the database.
		err := d.state.DB.PutStatus(ctx, latestStatus)
		if err != nil {
			return nil, nil, gtserror.Newf("error putting in database: %w", err)
		}
	} else {
		// This is an existing status, update the model in the database.
		if err := d.state.DB.UpdateStatus(ctx, latestStatus); err != nil {
			return nil, nil, gtserror.Newf("error updating database: %w", err)
		}
	}

	return latestStatus, apubStatus, nil
}

// populateMentionTarget tries to populate the given
// mention with the correct TargetAccount and (if not
// yet set) TargetAccountURI, returning the populated
// mention.
//
// Will check on the existing status if the mention
// is already there and populated; if so, existing
// mention will be returned along with `true`.
//
// Otherwise, this function will try to parse first
// the Href of the mention, and then the namestring,
// to see who it targets, and go fetch that account.
func (d *Dereferencer) populateMentionTarget(
	ctx context.Context,
	mention *gtsmodel.Mention,
	requestUser string,
	existing, status *gtsmodel.Status,
) (
	*gtsmodel.Mention,
	bool, // True if mention already exists in the DB.
	error,
) {
	// Mentions can be created using Name or Href.
	// Prefer Href (TargetAccountURI), fall back to Name.
	if mention.TargetAccountURI != "" {
		// Look for existing mention with this URI.
		// If we already have it we can return early.
		existingMention, ok := existing.GetMentionByTargetURI(mention.TargetAccountURI)
		if ok && existingMention.ID != "" {
			return existingMention, true, nil
		}

		// Ensure that mention account URI is parseable.
		accountURI, err := url.Parse(mention.TargetAccountURI)
		if err != nil {
			err = gtserror.Newf("invalid account uri %q: %w", mention.TargetAccountURI, err)
			return nil, false, err
		}

		// Ensure we have the account of the mention target dereferenced.
		mention.TargetAccount, _, err = d.getAccountByURI(ctx, requestUser, accountURI)
		if err != nil {
			err = gtserror.Newf("failed to dereference account %s: %w", accountURI, err)
			return nil, false, err
		}
	} else {
		// Href wasn't set. Find the target account using namestring.
		username, domain, err := util.ExtractNamestringParts(mention.NameString)
		if err != nil {
			err = gtserror.Newf("failed to parse namestring %s: %w", mention.NameString, err)
			return nil, false, err
		}

		mention.TargetAccount, _, err = d.getAccountByUsernameDomain(ctx, requestUser, username, domain)
		if err != nil {
			err = gtserror.Newf("failed to dereference account %s: %w", mention.NameString, err)
			return nil, false, err
		}

		// Look for existing mention with this URI.
		mention.TargetAccountURI = mention.TargetAccount.URI
		existingMention, ok := existing.GetMentionByTargetURI(mention.TargetAccountURI)
		if ok && existingMention.ID != "" {
			return existingMention, true, nil
		}
	}

	// At this point, mention.TargetAccountURI
	// and mention.TargetAccount must be set.
	return mention, false, nil
}

func (d *Dereferencer) fetchStatusMentions(ctx context.Context, requestUser string, existing, status *gtsmodel.Status) error {
	// Allocate new slice to take the yet-to-be created mention IDs.
	status.MentionIDs = make([]string, len(status.Mentions))

	for i := range status.Mentions {
		var (
			mention       = status.Mentions[i]
			alreadyExists bool
			err           error
		)

		mention, alreadyExists, err = d.populateMentionTarget(
			ctx,
			mention,
			requestUser,
			existing,
			status,
		)
		if err != nil {
			log.Errorf(ctx, "failed to derive mention: %v", err)
			continue
		}

		if alreadyExists {
			// This mention was already attached
			// to the status, use it and continue.
			status.Mentions[i] = mention
			status.MentionIDs[i] = mention.ID
			continue
		}

		// This mention didn't exist yet.
		// Generate new ID according to status creation.
		// TODO: update this to use "edited_at" when we add
		//       support for edited status revision history.
		mention.ID, err = id.NewULIDFromTime(status.CreatedAt)
		if err != nil {
			log.Errorf(ctx, "invalid created at date: %v", err)
			mention.ID = id.NewULID() // just use "now"
		}

		// Set known further mention details.
		mention.CreatedAt = status.CreatedAt
		mention.UpdatedAt = status.UpdatedAt
		mention.OriginAccount = status.Account
		mention.OriginAccountID = status.AccountID
		mention.OriginAccountURI = status.AccountURI
		mention.TargetAccountID = mention.TargetAccount.ID
		mention.TargetAccountURI = mention.TargetAccount.URI
		mention.TargetAccountURL = mention.TargetAccount.URL
		mention.StatusID = status.ID
		mention.Status = status

		// Place the new mention into the database.
		if err := d.state.DB.PutMention(ctx, mention); err != nil {
			return gtserror.Newf("error putting mention in database: %w", err)
		}

		// Set the *new* mention and ID.
		status.Mentions[i] = mention
		status.MentionIDs[i] = mention.ID
	}

	for i := 0; i < len(status.MentionIDs); {
		if status.MentionIDs[i] == "" {
			// This is a failed mention population, likely due
			// to invalid incoming data / now-deleted accounts.
			copy(status.Mentions[i:], status.Mentions[i+1:])
			copy(status.MentionIDs[i:], status.MentionIDs[i+1:])
			status.Mentions = status.Mentions[:len(status.Mentions)-1]
			status.MentionIDs = status.MentionIDs[:len(status.MentionIDs)-1]
			continue
		}
		i++
	}

	return nil
}

func (d *Dereferencer) threadStatus(ctx context.Context, status *gtsmodel.Status) error {
	if status.InReplyTo != nil {
		if parentThreadID := status.InReplyTo.ThreadID; parentThreadID != "" {
			// Simplest case: parent status
			// is threaded, so inherit threadID.
			status.ThreadID = parentThreadID
			return nil
		}
	}

	// Parent wasn't threaded. If this
	// status mentions a local account,
	// we should thread it so that local
	// account can mute it if they want.
	mentionsLocal := slices.ContainsFunc(
		status.Mentions,
		func(m *gtsmodel.Mention) bool {
			// If TargetAccount couldn't
			// be deref'd, we know it's not
			// a local account, so only
			// check for non-nil accounts.
			return m.TargetAccount != nil &&
				m.TargetAccount.IsLocal()
		},
	)

	if !mentionsLocal {
		// Status doesn't mention a
		// local account, so we don't
		// need to thread it.
		return nil
	}

	// Status mentions a local account.
	// Create a new thread and assign
	// it to the status.
	threadID := id.NewULID()

	if err := d.state.DB.PutThread(
		ctx,
		&gtsmodel.Thread{
			ID: threadID,
		},
	); err != nil {
		return gtserror.Newf("error inserting new thread in db: %w", err)
	}

	status.ThreadID = threadID
	return nil
}

func (d *Dereferencer) fetchStatusTags(ctx context.Context, existing, status *gtsmodel.Status) error {
	// Allocate new slice to take the yet-to-be determined tag IDs.
	status.TagIDs = make([]string, len(status.Tags))

	for i := range status.Tags {
		tag := status.Tags[i]

		// Look for tag in existing status with name.
		existing, ok := existing.GetTagByName(tag.Name)
		if ok && existing.ID != "" {
			status.Tags[i] = existing
			status.TagIDs[i] = existing.ID
			continue
		}

		// Look for existing tag with name in the database.
		existing, err := d.state.DB.GetTagByName(ctx, tag.Name)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return gtserror.Newf("db error getting tag %s: %w", tag.Name, err)
		} else if existing != nil {
			status.Tags[i] = existing
			status.TagIDs[i] = existing.ID
			continue
		}

		// Create new ID for tag.
		tag.ID = id.NewULID()

		// Insert this tag with new name into the database.
		if err := d.state.DB.PutTag(ctx, tag); err != nil {
			log.Errorf(ctx, "db error putting tag %s: %v", tag.Name, err)
			continue
		}

		// Set new tag ID in slice.
		status.TagIDs[i] = tag.ID
	}

	// Remove any tag we couldn't get or create.
	for i := 0; i < len(status.TagIDs); {
		if status.TagIDs[i] == "" {
			// This is a failed tag population, likely due
			// to some database peculiarity / race condition.
			copy(status.Tags[i:], status.Tags[i+1:])
			copy(status.TagIDs[i:], status.TagIDs[i+1:])
			status.Tags = status.Tags[:len(status.Tags)-1]
			status.TagIDs = status.TagIDs[:len(status.TagIDs)-1]
			continue
		}
		i++
	}

	return nil
}

func (d *Dereferencer) fetchStatusPoll(ctx context.Context, existing, status *gtsmodel.Status) error {
	if existing.Poll != nil {
		// Check poll is up-to-date.
		if status.Poll != nil &&
			slices.Equal(existing.Poll.Options, status.Poll.Options) &&
			existing.Poll.ExpiresAt.Equal(status.Poll.ExpiresAt) {
			return nil // latest and existing are equal.
		}

		// Poll has changed from existing to latest. Delete existing!
		if err := d.state.DB.DeletePollByID(ctx, existing.Poll.ID); err != nil {
			return gtserror.Newf("error deleting existing poll from database: %w", err)
		}

		// Delete any poll votes pointing to the existing poll ID.
		if err := d.state.DB.DeletePollVotes(ctx, existing.Poll.ID); err != nil {
			return gtserror.Newf("error deleting existing votes from database: %w", err)
		}

		// Cancel any scheduled expiry task for existing poll.
		_ = d.state.Workers.Scheduler.Cancel(existing.Poll.ID)

		if status.Poll == nil {
			// Old poll deleted,
			// all done here.
			return nil
		}
	}

	if status.Poll == nil {
		// no new poll to create.
		return nil
	}

	var err error

	// Generate new ID for poll from the status CreatedAt.
	// TODO: update this to use "edited_at" when we add
	//       support for edited status revision history.
	status.Poll.ID, err = id.NewULIDFromTime(status.CreatedAt)
	if err != nil {
		log.Errorf(ctx, "invalid created at date: %v", err)
		status.Poll.ID = id.NewULID() // just use "now"
	}

	// Update the status<->poll links.
	status.PollID = status.Poll.ID
	status.Poll.StatusID = status.ID
	status.Poll.Status = status

	// Insert this latest poll into the database.
	err = d.state.DB.PutPoll(ctx, status.Poll)
	if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
		return gtserror.Newf("error putting in database: %w", err)
	}

	if err != nil /* i.e. db.ErrAlreadyExists */ {
		// TODO: replace this quick fix with per-URI deref locks.
		status.Poll, err = d.state.DB.GetPollByStatusID(ctx, status.ID)
		if err != nil {
			return gtserror.Newf("error getting from database after race: %w", err)
		}
	}

	return nil
}

func (d *Dereferencer) fetchStatusAttachments(ctx context.Context, tsport transport.Transport, existing, status *gtsmodel.Status) error {
	// Allocate new slice to take the yet-to-be fetched attachment IDs.
	status.AttachmentIDs = make([]string, len(status.Attachments))

	for i := range status.Attachments {
		attachment := status.Attachments[i]

		// Look for existing media attachment with remoet URL first.
		existing, ok := existing.GetAttachmentByRemoteURL(attachment.RemoteURL)
		if ok && existing.ID != "" && *existing.Cached {
			status.Attachments[i] = existing
			status.AttachmentIDs[i] = existing.ID
			continue
		}

		// Ensure a valid media attachment remote URL.
		remoteURL, err := url.Parse(attachment.RemoteURL)
		if err != nil {
			log.Errorf(ctx, "invalid remote media url %q: %v", attachment.RemoteURL, err)
			continue
		}

		// Start pre-processing remote media at remote URL.
		processing, err := d.mediaManager.PreProcessMedia(ctx, func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, remoteURL)
		}, status.AccountID, &media.AdditionalMediaInfo{
			StatusID:    &status.ID,
			RemoteURL:   &attachment.RemoteURL,
			Description: &attachment.Description,
			Blurhash:    &attachment.Blurhash,
		})
		if err != nil {
			log.Errorf(ctx, "error processing attachment: %v", err)
			continue
		}

		// Force attachment loading *right now*.
		attachment, err = processing.LoadAttachment(ctx)
		if err != nil {
			log.Errorf(ctx, "error loading attachment: %v", err)
			continue
		}

		// Set the *new* attachment and ID.
		status.Attachments[i] = attachment
		status.AttachmentIDs[i] = attachment.ID
	}

	for i := 0; i < len(status.AttachmentIDs); {
		if status.AttachmentIDs[i] == "" {
			// This is a failed attachment population, this may
			// be due to us not currently supporting a media type.
			copy(status.Attachments[i:], status.Attachments[i+1:])
			copy(status.AttachmentIDs[i:], status.AttachmentIDs[i+1:])
			status.Attachments = status.Attachments[:len(status.Attachments)-1]
			status.AttachmentIDs = status.AttachmentIDs[:len(status.AttachmentIDs)-1]
			continue
		}
		i++
	}

	return nil
}

func (d *Dereferencer) fetchStatusEmojis(ctx context.Context, requestUser string, status *gtsmodel.Status) error {
	// Fetch the full-fleshed-out emoji objects for our status.
	emojis, err := d.populateEmojis(ctx, status.Emojis, requestUser)
	if err != nil {
		return gtserror.Newf("failed to populate emojis: %w", err)
	}

	// Iterate over and get their IDs.
	emojiIDs := make([]string, 0, len(emojis))
	for _, e := range emojis {
		emojiIDs = append(emojiIDs, e.ID)
	}

	// Set known emoji details.
	status.Emojis = emojis
	status.EmojiIDs = emojiIDs

	return nil
}
