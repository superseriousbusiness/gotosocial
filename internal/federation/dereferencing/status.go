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
	"net/http"
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
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// statusFresh returns true if the given status is still
// considered "fresh" according to the desired freshness
// window (falls back to default status freshness if nil).
//
// Local statuses will always be considered fresh,
// because there's no remote state that may have changed.
//
// Return value of false indicates that the status
// is not fresh and should be refreshed from remote.
func statusFresh(
	status *gtsmodel.Status,
	window *FreshnessWindow,
) bool {
	// Take default if no
	// freshness window preferred.
	if window == nil {
		window = DefaultStatusFreshness
	}

	if status.IsLocal() {
		// Can't refresh
		// local statuses.
		return true
	}

	// Moment when the status is
	// considered stale according to
	// desired freshness window.
	staleAt := status.FetchedAt.Add(
		time.Duration(*window),
	)

	// It's still fresh if the time now
	// is not past the point of staleness.
	return !time.Now().After(staleAt)
}

// GetStatusByURI will attempt to fetch a status by its URI, first checking the database. In the case of a newly-met remote model, or a remote model whose 'last_fetched' date
// is beyond a certain interval, the status will be dereferenced. In the case of dereferencing, some low-priority status information may be enqueued for asynchronous fetching,
// e.g. dereferencing the status thread. An ActivityPub object indicates the status was dereferenced.
func (d *Dereferencer) GetStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error) {

	// Fetch and dereference / update status if necessary.
	status, statusable, isNew, err := d.getStatusByURI(ctx,
		requestUser,
		uri,
	)

	if err != nil {
		if status == nil {
			// err with no existing
			// status for fallback.
			return nil, nil, err
		}

		log.Errorf(ctx, "error updating status %s: %v", uri, err)

	} else if statusable != nil {

		// Deref parents + children.
		d.dereferenceThread(ctx,
			requestUser,
			uri,
			status,
			statusable,
			isNew,
		)
	}

	return status, statusable, nil
}

// getStatusByURI is a package internal form of .GetStatusByURI() that doesn't dereference thread on update, and may return an existing status with error on failed re-fetch.
func (d *Dereferencer) getStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, bool, error) {
	var (
		status *gtsmodel.Status
		uriStr = uri.String()
		err    error
	)

	// Search the database for existing by URI.
	status, err = d.state.DB.GetStatusByURI(

		// request a barebones object, it may be in the
		// db but with related models not yet dereferenced.
		gtscontext.SetBarebones(ctx),
		uriStr,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, nil, false, gtserror.Newf("error checking database for status %s by uri: %w", uriStr, err)
	}

	if status == nil {
		// Else, search database for existing by URL.
		status, err = d.state.DB.GetStatusByURL(
			gtscontext.SetBarebones(ctx),
			uriStr,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, nil, false, gtserror.Newf("error checking database for status %s by url: %w", uriStr, err)
		}
	}

	if status == nil {
		// Ensure not a failed search for a local
		// status, if so we know it doesn't exist.
		if uri.Host == config.GetHost() ||
			uri.Host == config.GetAccountDomain() {
			return nil, nil, false, gtserror.SetUnretrievable(err)
		}

		// Create and pass-through a new bare-bones model for deref.
		return d.enrichStatusSafely(ctx, requestUser, uri, &gtsmodel.Status{
			Local: util.Ptr(false),
			URI:   uriStr,
		}, nil)
	}

	if statusFresh(status, DefaultStatusFreshness) {
		// This is an existing status that is up-to-date,
		// before returning ensure it is fully populated.
		if err := d.state.DB.PopulateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error populating existing status: %v", err)
		}

		return status, nil, false, nil
	}

	// Try to deref and update existing.
	return d.enrichStatusSafely(ctx,
		requestUser,
		uri,
		status,
		nil,
	)
}

// RefreshStatus is functionally equivalent to GetStatusByURI(), except that it requires a pre
// populated status model (with AT LEAST uri set), and ALL thread dereferencing is asynchronous.
func (d *Dereferencer) RefreshStatus(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
	statusable ap.Statusable,
	window *FreshnessWindow,
) (*gtsmodel.Status, ap.Statusable, error) {
	// If no incoming data is provided,
	// check whether status needs update.
	if statusable == nil &&
		statusFresh(status, window) {
		return status, nil, nil
	}

	// Parse the URI from status.
	uri, err := url.Parse(status.URI)
	if err != nil {
		return nil, nil, gtserror.Newf("invalid status uri %q: %w", status.URI, err)
	}

	// Try to update + dereference the passed status model.
	latest, statusable, isNew, err := d.enrichStatusSafely(ctx,
		requestUser,
		uri,
		status,
		statusable,
	)

	if statusable != nil {
		// Deref parents + children.
		d.dereferenceThread(ctx,
			requestUser,
			uri,
			latest,
			statusable,
			isNew,
		)
	}

	return latest, statusable, err
}

// RefreshStatusAsync is functionally equivalent to RefreshStatus(), except that ALL
// dereferencing is queued for asynchronous processing, (both thread AND status).
func (d *Dereferencer) RefreshStatusAsync(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
	statusable ap.Statusable,
	window *FreshnessWindow,
) {
	// If no incoming data is provided,
	// check whether status needs update.
	if statusable == nil &&
		statusFresh(status, window) {
		return
	}

	// Parse the URI from status.
	uri, err := url.Parse(status.URI)
	if err != nil {
		log.Errorf(ctx, "invalid status uri %q: %v", status.URI, err)
		return
	}

	// Enqueue a worker function to re-fetch this status entirely async.
	d.state.Workers.Dereference.Queue.Push(func(ctx context.Context) {
		latest, statusable, _, err := d.enrichStatusSafely(ctx,
			requestUser,
			uri,
			status,
			statusable,
		)
		if err != nil {
			log.Errorf(ctx, "error enriching remote status: %v", err)
			return
		}
		if statusable != nil {
			if err := d.DereferenceStatusAncestors(ctx, requestUser, latest); err != nil {
				log.Error(ctx, err)
			}
			if err := d.DereferenceStatusDescendants(ctx, requestUser, uri, statusable); err != nil {
				log.Error(ctx, err)
			}
		}
	})
}

// enrichStatusSafely wraps enrichStatus() to perform it within
// a State{}.FedLocks mutexmap, which protects it within per-URI
// mutex locks. This also handles necessary delete of now-deleted
// statuses, and updating fetched_at on returned HTTP errors.
func (d *Dereferencer) enrichStatusSafely(
	ctx context.Context,
	requestUser string,
	uri *url.URL,
	status *gtsmodel.Status,
	statusable ap.Statusable,
) (*gtsmodel.Status, ap.Statusable, bool, error) {
	uriStr := status.URI

	var isNew bool

	// Check if this is a new status (to us).
	if isNew = (status.ID == ""); !isNew {

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
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Perform status enrichment with passed vars.
	latest, statusable, err := d.enrichStatus(ctx,
		requestUser,
		uri,
		status,
		statusable,
	)

	// Check for a returned HTTP code via error.
	switch code := gtserror.StatusCode(err); {

	// Gone (410) definitely indicates deletion.
	// Remove status if it was an existing one.
	case code == http.StatusGone && !isNew:
		if err := d.state.DB.DeleteStatusByID(ctx, status.ID); err != nil {
			log.Error(ctx, "error deleting gone status %s: %v", uriStr, err)
		}

		// Don't return any status.
		return nil, nil, false, err

	// Any other HTTP error mesg
	// code, with existing status.
	case code >= 400 && !isNew:

		// Update fetched_at to slow re-attempts
		// but don't return early. We can still
		// return the model we had stored already.
		status.FetchedAt = time.Now()
		if err := d.state.DB.UpdateStatus(ctx, status, "fetched_at"); err != nil {
			log.Error(ctx, "error updating %s fetched_at: %v", uriStr, err)
		}

		// See below.
		fallthrough

	// In case of error with an existing
	// status in the database, return error
	// but still return existing status.
	case err != nil && !isNew:
		latest = status
		statusable = nil
	}

	// Unlock now
	// we're done.
	unlock()

	if errors.Is(err, db.ErrAlreadyExists) {
		// We leave 'isNew' set so that caller
		// still dereferences parents, otherwise
		// the version we pass back may not have
		// these attached as inReplyTos yet (since
		// those happen OUTSIDE federator lock).
		//
		// TODO: performance-wise, this won't be
		// great. should improve this if we can!

		// DATA RACE! We likely lost out to another goroutine
		// in a call to db.Put(Status). Look again in DB by URI.
		latest, err = d.state.DB.GetStatusByURI(ctx, status.URI)
		if err != nil {
			err = gtserror.Newf("error getting status %s from database after race: %w", uriStr, err)
		}
	}

	return latest, statusable, isNew, err
}

// enrichStatus will enrich the given status, whether a new
// barebones model, or existing model from the database.
// It handles necessary dereferencing, database updates, etc.
func (d *Dereferencer) enrichStatus(
	ctx context.Context,
	requestUser string,
	uri *url.URL,
	status *gtsmodel.Status,
	statusable ap.Statusable,
) (
	*gtsmodel.Status,
	ap.Statusable,
	error,
) {
	// Pre-fetch a transport for requesting username, used by later dereferencing.
	tsport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
	if err != nil {
		return nil, nil, gtserror.Newf("couldn't create transport: %w", err)
	}

	// Check whether this account URI is a blocked domain / subdomain.
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, uri.Host); err != nil {
		return nil, nil, gtserror.Newf("error checking blocked domain: %w", err)
	} else if blocked {
		err := gtserror.Newf("%s is blocked", uri.Host)
		return nil, nil, gtserror.SetUnretrievable(err)
	}

	if statusable == nil {
		// Dereference latest version of the status.
		rsp, err := tsport.Dereference(ctx, uri)
		if err != nil {
			err := gtserror.Newf("error dereferencing %s: %w", uri, err)
			return nil, nil, gtserror.SetUnretrievable(err)
		}

		// Attempt to resolve ActivityPub status from response.
		statusable, err = ap.ResolveStatusable(ctx, rsp.Body)

		// Tidy up now done.
		_ = rsp.Body.Close()

		if err != nil {
			// ResolveStatusable will set gtserror.WrongType
			// on the returned error, so we don't need to do it here.
			err := gtserror.Newf("error resolving statusable %s: %w", uri, err)
			return nil, nil, err
		}

		// Check whether input URI and final returned URI
		// have changed (i.e. we followed some redirects).
		if finalURIStr := rsp.Request.URL.String(); //
		finalURIStr != uri.String() {

			// NOTE: this URI check + database call is performed
			// AFTER reading and closing response body, for performance.
			//
			// Check whether we have this status stored under *final* URI.
			alreadyStatus, err := d.state.DB.GetStatusByURI(ctx, finalURIStr)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return nil, nil, gtserror.Newf("db error getting status after redirects: %w", err)
			}

			if alreadyStatus != nil {
				// We had this status stored
				// under discovered final URI.
				//
				// Proceed with this status.
				status = alreadyStatus
			}

			// Update the input URI to
			// the final determined URI
			// for later URI checks.
			uri = rsp.Request.URL
		}
	}

	// Get the attributed-to account in order to fetch profile.
	attributedTo, err := ap.ExtractAttributedToURI(statusable)
	if err != nil {
		return nil, nil, gtserror.New("attributedTo was empty")
	}

	// Ensure we have the author account of the status dereferenced (+ up-to-date). If this is a new status
	// (i.e. status.AccountID == "") then any error here is irrecoverable. status.AccountID must ALWAYS be set.
	if _, _, err := d.getAccountByURI(ctx, requestUser, attributedTo); err != nil && status.AccountID == "" {

		// Note that we specifically DO NOT wrap the error, instead collapsing it as string.
		// Errors fetching an account do not necessarily relate to dereferencing the status.
		return nil, nil, gtserror.Newf("failed to dereference status author %s: %v", uri, err)
	}

	// ActivityPub model was recently dereferenced, so assume passed status
	// may contain out-of-date information. Convert AP model to our GTS model.
	latestStatus, err := d.converter.ASStatusToStatus(ctx, statusable)
	if err != nil {
		return nil, nil, gtserror.Newf("error converting statusable to gts model for status %s: %w", uri, err)
	}

	// Ensure final status isn't attempting
	// to claim being authored by local user.
	if latestStatus.Account.IsLocal() {
		return nil, nil, gtserror.Newf(
			"dereferenced status %s claiming to be local",
			latestStatus.URI,
		)
	}

	// Ensure the final parsed status URI or URL matches
	// the input URI we fetched (or received) it as.
	matches, err := util.URIMatches(
		uri,
		append(
			ap.GetURL(statusable),      // status URL(s)
			ap.GetJSONLDId(statusable), // status URI
		)...,
	)
	if err != nil {
		return nil, nil, gtserror.Newf(
			"error checking dereferenced status uri %s: %w",
			latestStatus.URI, err,
		)
	}

	if !matches {
		return nil, nil, gtserror.Newf(
			"dereferenced status uri %s does not match %s",
			latestStatus.URI, uri.String(),
		)
	}

	var isNew bool

	// Based on the original provided
	// status model, determine whether
	// this is a new insert / update.
	if isNew = (status.ID == ""); isNew {

		// Generate new status ID from the provided creation date.
		latestStatus.ID, err = id.NewULIDFromTime(latestStatus.CreatedAt)
		if err != nil {
			log.Errorf(ctx, "invalid created at date (falling back to 'now'): %v", err)
			latestStatus.ID = id.NewULID() // just use "now"
		}
	} else {

		// Reuse existing status ID.
		latestStatus.ID = status.ID
	}

	// Carry-over values and set fetch time.
	latestStatus.UpdatedAt = status.UpdatedAt
	latestStatus.FetchedAt = time.Now()
	latestStatus.Local = status.Local

	// Carry-over approvals. Remote instances might not yet
	// serve statuses with the `approved_by` field, but we
	// might have marked a status as pre-approved on our side
	// based on the author's inclusion in a followers/following
	// collection. By carrying over previously-set values we
	// can avoid marking such statuses as "pending" again.
	//
	// If a remote has in the meantime retracted its approval,
	// the next call to 'isPermittedStatus' will catch that.
	if latestStatus.ApprovedByURI == "" && status.ApprovedByURI != "" {
		latestStatus.ApprovedByURI = status.ApprovedByURI
	}

	// Check if this is a permitted status we should accept.
	// Function also sets "PendingApproval" bool as necessary.
	permit, err := d.isPermittedStatus(ctx, requestUser, status, latestStatus)
	if err != nil {
		return nil, nil, gtserror.Newf("error checking permissibility for status %s: %w", uri, err)
	}

	if !permit {
		// Return a checkable error type that can be ignored.
		err := gtserror.Newf("dropping unpermitted status: %s", uri)
		return nil, nil, gtserror.SetNotPermitted(err)
	}

	// Ensure the status' mentions are populated, and pass in existing to check for changes.
	if err := d.fetchStatusMentions(ctx, requestUser, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating mentions for status %s: %w", uri, err)
	}

	// Ensure the status' poll remains consistent, else reset the poll.
	if err := d.fetchStatusPoll(ctx, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating poll for status %s: %w", uri, err)
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
	if err := d.fetchStatusAttachments(ctx, requestUser, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating attachments for status %s: %w", uri, err)
	}

	// Ensure the status' emoji attachments are populated, passing in existing to check for changes.
	if err := d.fetchStatusEmojis(ctx, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating emojis for status %s: %w", uri, err)
	}

	if isNew {
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

	return latestStatus, statusable, nil
}

func (d *Dereferencer) fetchStatusMentions(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) error {
	// Allocate new slice to take the yet-to-be created mention IDs.
	status.MentionIDs = make([]string, len(status.Mentions))

	for i := range status.Mentions {
		var (
			mention       = status.Mentions[i]
			alreadyExists bool
			err           error
		)

		// Search existing status for a mention already stored,
		// else ensure new mention's target account is populated.
		mention, alreadyExists, err = d.getPopulatedMention(ctx,
			requestUser,
			existing,
			mention,
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
			log.Errorf(ctx, "invalid created at date (falling back to 'now'): %v", err)
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

func (d *Dereferencer) fetchStatusTags(
	ctx context.Context,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) error {
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

func (d *Dereferencer) fetchStatusPoll(
	ctx context.Context,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) error {
	var (
		// insertStatusPoll generates ID and inserts the poll attached to status into the database.
		insertStatusPoll = func(ctx context.Context, status *gtsmodel.Status) error {
			var err error

			// Generate new ID for poll from the status CreatedAt.
			// TODO: update this to use "edited_at" when we add
			//       support for edited status revision history.
			status.Poll.ID, err = id.NewULIDFromTime(status.CreatedAt)
			if err != nil {
				log.Errorf(ctx, "invalid created at date (falling back to 'now'): %v", err)
				status.Poll.ID = id.NewULID() // just use "now"
			}

			// Update the status<->poll links.
			status.PollID = status.Poll.ID
			status.Poll.StatusID = status.ID
			status.Poll.Status = status

			// Insert this latest poll into the database.
			err = d.state.DB.PutPoll(ctx, status.Poll)
			if err != nil {
				return gtserror.Newf("error putting in database: %w", err)
			}

			return nil
		}

		// deleteStatusPoll deletes the poll with ID, and all attached votes, from the database.
		deleteStatusPoll = func(ctx context.Context, pollID string) error {
			if err := d.state.DB.DeletePollByID(ctx, pollID); err != nil {
				return gtserror.Newf("error deleting existing poll from database: %w", err)
			}
			return nil
		}
	)

	switch {
	case existing.Poll == nil && status.Poll == nil:
		// no poll before or after, nothing to do.
		return nil

	case existing.Poll == nil && status.Poll != nil:
		// no previous poll, insert new poll!
		return insertStatusPoll(ctx, status)

	case status.Poll == nil:
		// existing poll has been deleted, remove this.
		return deleteStatusPoll(ctx, existing.PollID)

	case pollChanged(existing.Poll, status.Poll):
		// poll has changed since original, delete and reinsert new.
		if err := deleteStatusPoll(ctx, existing.PollID); err != nil {
			return err
		}
		return insertStatusPoll(ctx, status)

	case pollUpdated(existing.Poll, status.Poll):
		// Since we last saw it, the poll has updated!
		// Whether that be stats, or close time.
		poll := existing.Poll
		poll.Closing = pollJustClosed(existing.Poll, status.Poll)
		poll.ClosedAt = status.Poll.ClosedAt
		poll.Voters = status.Poll.Voters
		poll.Votes = status.Poll.Votes

		// Update poll model in the database (specifically only the possible changed columns).
		if err := d.state.DB.UpdatePoll(ctx, poll, "closed_at", "voters", "votes"); err != nil {
			return gtserror.Newf("error updating poll: %w", err)
		}

		// Update poll on status.
		status.PollID = poll.ID
		status.Poll = poll
		return nil

	default:
		// latest and existing
		// polls are up to date.
		poll := existing.Poll
		status.PollID = poll.ID
		status.Poll = poll
		return nil
	}
}

func (d *Dereferencer) fetchStatusAttachments(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) error {
	// Allocate new slice to take the yet-to-be fetched attachment IDs.
	status.AttachmentIDs = make([]string, len(status.Attachments))

	for i := range status.Attachments {
		placeholder := status.Attachments[i]

		// Look for existing media attachment with remote URL first.
		existing, ok := existing.GetAttachmentByRemoteURL(placeholder.RemoteURL)
		if ok && existing.ID != "" {

			// Ensure the existing media attachment is up-to-date and cached.
			existing, err := d.updateAttachment(ctx, requestUser, existing, placeholder)
			if err != nil {
				log.Errorf(ctx, "error updating existing attachment: %v", err)

				// specifically do NOT continue here,
				// we already have a model, we don't
				// want to drop it from the status, just
				// log that an update for it failed.
			}

			// Set the existing attachment.
			status.Attachments[i] = existing
			status.AttachmentIDs[i] = existing.ID
			continue
		}

		// Load this new media attachment.
		attachment, err := d.GetMedia(
			ctx,
			requestUser,
			status.AccountID,
			placeholder.RemoteURL,
			media.AdditionalMediaInfo{
				StatusID:    &status.ID,
				RemoteURL:   &placeholder.RemoteURL,
				Description: &placeholder.Description,
				Blurhash:    &placeholder.Blurhash,
			},
		)
		if err != nil {
			if attachment == nil {
				log.Errorf(ctx, "error loading attachment %s: %v", placeholder.RemoteURL, err)
				continue
			}

			// non-fatal error occurred during loading, still use it.
			log.Warnf(ctx, "partially loaded attachment: %v", err)
		}

		// Set the *new* attachment and ID.
		status.Attachments[i] = attachment
		status.AttachmentIDs[i] = attachment.ID
	}

	for i := 0; i < len(status.AttachmentIDs); {
		if status.AttachmentIDs[i] == "" {
			// Remove totally failed attachment populations
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

func (d *Dereferencer) fetchStatusEmojis(
	ctx context.Context,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) error {
	// Fetch the updated emojis for our status.
	emojis, changed, err := d.fetchEmojis(ctx,
		existing.Emojis,
		status.Emojis,
	)
	if err != nil {
		return gtserror.Newf("error fetching emojis: %w", err)
	}

	if !changed {
		// Use existing status emoji objects.
		status.EmojiIDs = existing.EmojiIDs
		status.Emojis = existing.Emojis
		return nil
	}

	// Set latest emojis.
	status.Emojis = emojis

	// Iterate over and set changed emoji IDs.
	status.EmojiIDs = make([]string, len(emojis))
	for i, emoji := range emojis {
		status.EmojiIDs[i] = emoji.ID
	}

	return nil
}

// getPopulatedMention tries to populate the given
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
func (d *Dereferencer) getPopulatedMention(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.Status,
	mention *gtsmodel.Mention,
) (
	*gtsmodel.Mention,
	bool, // True if mention already exists in the DB.
	error,
) {
	// Mentions can be created using Name or Href.
	// Prefer Href (TargetAccountURI), fall back to Name.
	if mention.TargetAccountURI != "" {

		// Look for existing mention with target account's URI, if so use this.
		existingMention, ok := existing.GetMentionByTargetURI(mention.TargetAccountURI)
		if ok && existingMention.ID != "" {
			return existingMention, true, nil
		}

		// Ensure that mention account URI is parseable.
		accountURI, err := url.Parse(mention.TargetAccountURI)
		if err != nil {
			err := gtserror.Newf("invalid account uri %q: %w", mention.TargetAccountURI, err)
			return nil, false, err
		}

		// Ensure we have  account of the mention target dereferenced.
		mention.TargetAccount, _, err = d.getAccountByURI(ctx,
			requestUser,
			accountURI,
		)
		if err != nil {
			err := gtserror.Newf("failed to dereference account %s: %w", accountURI, err)
			return nil, false, err
		}
	} else {

		// Href wasn't set, extract the username and domain parts from namestring.
		username, domain, err := util.ExtractNamestringParts(mention.NameString)
		if err != nil {
			err := gtserror.Newf("failed to parse namestring %s: %w", mention.NameString, err)
			return nil, false, err
		}

		// Look for existing mention with username domain target, if so use this.
		existingMention, ok := existing.GetMentionByUsernameDomain(username, domain)
		if ok && existingMention.ID != "" {
			return existingMention, true, nil
		}

		// Ensure we have the account of the mention target dereferenced.
		mention.TargetAccount, _, err = d.getAccountByUsernameDomain(ctx,
			requestUser,
			username,
			domain,
		)
		if err != nil {
			err := gtserror.Newf("failed to dereference account %s: %w", mention.NameString, err)
			return nil, false, err
		}

		// Look for existing mention with target account's URI, if so use this.
		existingMention, ok = existing.GetMentionByTargetURI(mention.TargetAccountURI)
		if ok && existingMention.ID != "" {
			return existingMention, true, nil
		}
	}

	// At this point, mention.TargetAccountURI
	// and mention.TargetAccount must be set.
	return mention, false, nil
}
