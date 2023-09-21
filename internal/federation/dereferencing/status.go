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

// GetStatus: implements Dereferencer{}.GetStatus().
func (d *deref) GetStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error) {
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
func (d *deref) getStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error) {
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
		return d.enrichStatus(ctx, requestUser, uri, &gtsmodel.Status{
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
	latest, apubStatus, err := d.enrichStatus(ctx,
		requestUser,
		uri,
		status,
		nil,
	)
	if err != nil {
		log.Errorf(ctx, "error enriching remote status: %v", err)

		// Update fetch-at to slow re-attempts.
		status.FetchedAt = time.Now()
		_ = d.state.DB.UpdateStatus(ctx, status, "fetched_at")

		// Fallback to existing.
		return status, nil, nil
	}

	return latest, apubStatus, nil
}

// RefreshStatus: implements Dereferencer{}.RefreshStatus().
func (d *deref) RefreshStatus(ctx context.Context, requestUser string, status *gtsmodel.Status, apubStatus ap.Statusable, force bool) (*gtsmodel.Status, ap.Statusable, error) {
	// Check whether needs update.
	if !force && statusUpToDate(status) {
		return status, nil, nil
	}

	// Parse the URI from status.
	uri, err := url.Parse(status.URI)
	if err != nil {
		return nil, nil, gtserror.Newf("invalid status uri %q: %w", status.URI, err)
	}

	// Try to update + deref existing status model.
	latest, apubStatus, err := d.enrichStatus(ctx,
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

// RefreshStatusAsync: implements Dereferencer{}.RefreshStatusAsync().
func (d *deref) RefreshStatusAsync(ctx context.Context, requestUser string, status *gtsmodel.Status, apubStatus ap.Statusable, force bool) {
	// Check whether needs update.
	if statusUpToDate(status) {
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
		latest, apubStatus, err := d.enrichStatus(ctx, requestUser, uri, status, apubStatus)
		if err != nil {
			log.Errorf(ctx, "error enriching remote status: %v", err)
			return
		}

		// This status was updated, re-dereference the whole thread.
		d.dereferenceThread(ctx, requestUser, uri, latest, apubStatus)
	})
}

// enrichStatus will enrich the given status, whether a new
// barebones model, or existing model from the database.
// It handles necessary dereferencing, database updates, etc.
func (d *deref) enrichStatus(
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

	// Ensure we have the author account of the status dereferenced (+ up-to-date).
	if author, _, err := d.getAccountByURI(ctx, requestUser, attributedTo); err != nil {
		if status.AccountID == "" {
			// Provided status account is nil, i.e. this is a new status / author, so a deref fail is unrecoverable.
			return nil, nil, gtserror.Newf("failed to dereference status author %s: %w", uri, err)
		}
	} else if status.AccountID != "" && status.AccountID != author.ID {
		// There already existed an account for this status author, but account ID changed. This shouldn't happen!
		log.Warnf(ctx, "status author account ID changed: old=%s new=%s", status.AccountID, author.ID)
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

	// Ensure the status' mentions are populated, and pass in existing to check for changes.
	if err := d.fetchStatusMentions(ctx, requestUser, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating mentions for status %s: %w", uri, err)
	}

	// Ensure the status' tags are populated.
	if err := d.fetchStatusTags(ctx, requestUser, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating tags for status %s: %w", uri, err)
	}

	// Ensure the status' media attachments are populated, passing in existing to check for changes.
	if err := d.fetchStatusAttachments(ctx, tsport, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating attachments for status %s: %w", uri, err)
	}

	// Ensure the status' emoji attachments are populated, passing in existing to check for changes.
	if err := d.fetchStatusEmojis(ctx, requestUser, status, latestStatus); err != nil {
		return nil, nil, gtserror.Newf("error populating emojis for status %s: %w", uri, err)
	}

	if status.CreatedAt.IsZero() {
		// CreatedAt will be zero if no local copy was
		// found in one of the GetStatusBy___() functions.
		//
		// This is new, put the status in the database.
		err := d.state.DB.PutStatus(ctx, latestStatus)

		if errors.Is(err, db.ErrAlreadyExists) {
			// TODO: replace this quick fix with per-URI deref locks.
			latestStatus, err = d.state.DB.GetStatusByURI(ctx, latestStatus.URI)
			return latestStatus, nil, err
		}

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

func (d *deref) fetchStatusMentions(ctx context.Context, requestUser string, existing, status *gtsmodel.Status) error {
	// Allocate new slice to take the yet-to-be created mention IDs.
	status.MentionIDs = make([]string, len(status.Mentions))

	for i := range status.Mentions {
		mention := status.Mentions[i]

		// Look for existing mention with target account URI first.
		existing, ok := existing.GetMentionByTargetURI(mention.TargetAccountURI)
		if ok && existing.ID != "" {
			status.Mentions[i] = existing
			status.MentionIDs[i] = existing.ID
			continue
		}

		// Ensure that mention account URI is parseable.
		accountURI, err := url.Parse(mention.TargetAccountURI)
		if err != nil {
			log.Errorf(ctx, "invalid account uri %q: %v", mention.TargetAccountURI, err)
			continue
		}

		// Ensure we have the account of the mention target dereferenced.
		mention.TargetAccount, _, err = d.getAccountByURI(ctx, requestUser, accountURI)
		if err != nil {
			log.Errorf(ctx, "failed to dereference account %s: %v", accountURI, err)
			continue
		}

		// Generate new ID according to status creation.
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

func (d *deref) fetchStatusTags(ctx context.Context, requestUser string, status *gtsmodel.Status) error {
	// Allocate new slice to take the yet-to-be determined tag IDs.
	status.TagIDs = make([]string, len(status.Tags))

	for i := range status.Tags {
		placeholder := status.Tags[i]

		// Look for existing tag with this name first.
		tag, err := d.state.DB.GetTagByName(ctx, placeholder.Name)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf(ctx, "db error getting tag %s: %v", tag.Name, err)
			continue
		}

		// No tag with this name yet, create it.
		if tag == nil {
			tag = &gtsmodel.Tag{
				ID:   id.NewULID(),
				Name: placeholder.Name,
			}

			if err := d.state.DB.PutTag(ctx, tag); err != nil {
				log.Errorf(ctx, "db error putting tag %s: %v", tag.Name, err)
				continue
			}
		}

		// Set the *new* tag and ID.
		status.Tags[i] = tag
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

func (d *deref) fetchStatusAttachments(ctx context.Context, tsport transport.Transport, existing, status *gtsmodel.Status) error {
	// Allocate new slice to take the yet-to-be fetched attachment IDs.
	status.AttachmentIDs = make([]string, len(status.Attachments))

	for i := range status.Attachments {
		placeholder := status.Attachments[i]

		// Look for existing media attachment with remoet URL first.
		existing, ok := existing.GetAttachmentByRemoteURL(placeholder.RemoteURL)
		if ok && existing.ID != "" && *existing.Cached {
			status.Attachments[i] = existing
			status.AttachmentIDs[i] = existing.ID
			continue
		}

		// Ensure a valid media attachment remote URL.
		remoteURL, err := url.Parse(placeholder.RemoteURL)
		if err != nil {
			log.Errorf(ctx, "invalid remote media url %q: %v", placeholder.RemoteURL, err)
			continue
		}

		// Start pre-processing remote media at remote URL.
		processing, err := d.mediaManager.PreProcessMedia(ctx, func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, remoteURL)
		}, status.AccountID, &media.AdditionalMediaInfo{
			StatusID:    &status.ID,
			RemoteURL:   &placeholder.RemoteURL,
			Description: &placeholder.Description,
			Blurhash:    &placeholder.Blurhash,
		})
		if err != nil {
			log.Errorf(ctx, "error processing attachment: %v", err)
			continue
		}

		// Force attachment loading *right now*.
		media, err := processing.LoadAttachment(ctx)
		if err != nil {
			log.Errorf(ctx, "error loading attachment: %v", err)
			continue
		}

		// Set the *new* attachment and ID.
		status.Attachments[i] = media
		status.AttachmentIDs[i] = media.ID
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

func (d *deref) fetchStatusEmojis(ctx context.Context, requestUser string, existing, status *gtsmodel.Status) error {
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
