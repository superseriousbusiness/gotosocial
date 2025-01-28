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

package federatingdb

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Create adds a new entry to the database which must be able to be
// keyed by its id.
//
// Note that Activity values received from federated peers may also be
// created in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
//
// Under certain conditions and network activities, Create may be called
// multiple times for the same ActivityStreams object.
func (f *federatingDB) Create(ctx context.Context, asType vocab.Type) error {
	log.DebugKV(ctx, "create", serialize{asType})

	// Cache entry for this activity type's ID for later
	// checks in the Exist() function if we see it again.
	f.activityIDs.Set(ap.GetJSONLDId(asType).String(), struct{}{})

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if requesting.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	// Cast to the expected types we handle in this func.
	creatable, ok := asType.(vocab.ActivityStreamsCreate)
	if !ok {
		log.Debugf(ctx, "unhandled object type: %s", asType.GetTypeName())
		return nil
	}

	var errs gtserror.MultiError

	// Extract objects from create activity.
	objects := ap.ExtractObjects(creatable)

	// Extract PollOptionables (votes!) from objects slice.
	optionables, objects := ap.ExtractPollOptionables(objects)

	if len(optionables) > 0 {
		// Handle provided poll vote(s) creation, this can
		// be for single or multiple votes in the same poll.
		err := f.createPollOptionables(ctx,
			receiving,
			requesting,
			optionables,
		)
		if err != nil {
			errs.Appendf("error creating poll vote(s): %w", err)
		}
	}

	// Extract Statusables from objects slice (this must be
	// done AFTER extracting options due to how AS typing works).
	statusables, objects := ap.ExtractStatusables(objects)

	for _, statusable := range statusables {
		// Check if this is a forwarded object, i.e. did
		// the account making the request also create this?
		forwarded := !isSender(statusable, requesting)

		// Handle create event for this statusable.
		if err := f.createStatusable(ctx,
			receiving,
			requesting,
			statusable,
			forwarded,
		); err != nil {
			errs.Appendf("error creating statusable: %w", err)
		}
	}

	if len(objects) > 0 {
		// Log any unhandled objects after filtering for debug purposes.
		log.Debugf(ctx, "unhandled CREATE types: %v", typeNames(objects))
	}

	return errs.Combine()
}

// createPollOptionable handles a Create activity for a PollOptionable.
// This function doesn't handle database insertion, only validation checks
// before passing off to a worker for asynchronous processing.
func (f *federatingDB) createPollOptionables(
	ctx context.Context,
	receiver *gtsmodel.Account,
	requester *gtsmodel.Account,
	options []ap.PollOptionable,
) error {
	var (
		// the origin Status w/ Poll the vote
		// options are in. This gets set on first
		// iteration, relevant checks performed
		// then re-used in each further iteration.
		inReplyTo *gtsmodel.Status

		// the resulting slices of Poll.Option
		// choice indices passed into the new
		// created PollVote object.
		choices []int
	)

	for _, option := range options {
		// Extract the "inReplyTo" property.
		inReplyToURIs := ap.GetInReplyTo(option)
		if len(inReplyToURIs) != 1 {
			return gtserror.Newf("invalid inReplyTo property length: %d", len(inReplyToURIs))
		}

		// Stringify the inReplyTo URI.
		statusURI := inReplyToURIs[0].String()

		if inReplyTo == nil {
			var err error

			// This is the first object in the activity slice,
			// check database for the poll source status by URI.
			inReplyTo, err = f.state.DB.GetStatusByURI(ctx, statusURI)
			if err != nil {
				return gtserror.Newf("error getting poll source from database %s: %w", statusURI, err)
			}

			switch {
			// The origin status isn't a poll?
			case inReplyTo.PollID == "":
				return gtserror.Newf("poll vote in status %s without poll", statusURI)

			// We don't own the poll ...
			case !*inReplyTo.Local:
				return gtserror.Newf("poll vote in remote status %s", statusURI)
			}

			// Check whether user has already vote in this poll.
			// (we only check this for the first object, as multiple
			// may be sent in response to a multiple-choice poll).
			vote, err := f.state.DB.GetPollVoteBy(
				gtscontext.SetBarebones(ctx),
				inReplyTo.PollID,
				requester.ID,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return gtserror.Newf("error getting status %s poll votes from database: %w", statusURI, err)
			}

			if vote != nil {
				log.Warnf(ctx, "%s has already voted in poll %s", requester.URI, statusURI)
				return nil // this is a useful warning for admins to report to us from logs
			}
		}

		if statusURI != inReplyTo.URI {
			// All activity votes should be to the same poll per activity.
			return gtserror.New("votes to multiple polls in single activity")
		}

		// Extract the poll option name.
		name := ap.ExtractName(option)

		// Check that this is a valid option name.
		choice := inReplyTo.Poll.GetChoice(name)
		if choice == -1 {
			return gtserror.Newf("poll vote in status %s invalid: %s", statusURI, name)
		}

		// Append the option index to choices.
		choices = append(choices, choice)
	}

	// Enqueue message to the fedi API worker with poll vote(s).
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityQuestion,
		GTSModel: &gtsmodel.PollVote{
			ID:        id.NewULID(),
			Choices:   choices,
			AccountID: requester.ID,
			Account:   requester,
			PollID:    inReplyTo.PollID,
			Poll:      inReplyTo.Poll,
		},
		Receiving:  receiver,
		Requesting: requester,
	})

	return nil
}

// createStatusable handles a Create activity for a Statusable.
// This function won't insert anything in the database yet,
// but will pass the Statusable (if appropriate) through to
// the processor for further asynchronous processing.
func (f *federatingDB) createStatusable(
	ctx context.Context,
	receiver *gtsmodel.Account,
	requester *gtsmodel.Account,
	statusable ap.Statusable,
	forwarded bool,
) error {
	// Check whether this status is both
	// relevant, and doesn't look like spam.
	err := f.spamFilter.StatusableOK(ctx,
		receiver,
		requester,
		statusable,
	)

	switch {
	case err == nil:
		// No problem!

	case gtserror.IsNotRelevant(err):
		// This case is quite common if a remote (Mastodon)
		// instance forwards a message to us which is a reply
		// from someone else to a status we've also replied to.
		//
		// It does this to try to ensure thread completion, but
		// we have our own thread fetching mechanism anyway.
		log.Debugf(ctx, "status %s is not relevant to receiver (%v); dropping it",
			ap.GetJSONLDId(statusable), err,
		)
		return nil

	case gtserror.IsSpam(err):
		// Log this at a higher level so admins can
		// gauge how much spam is being sent to them.
		//
		// TODO: add Prometheus metrics for this.
		log.Infof(ctx, "status %s looked like spam (%v); dropping it",
			ap.GetJSONLDId(statusable), err,
		)
		return nil

	default:
		// A real error has occurred.
		return gtserror.Newf("error checking relevancy/spam: %w", err)
	}

	// If we do have a forward, we should ignore the content
	// and instead deref based on the URI of the statusable.
	//
	// In other words, don't automatically trust whoever sent
	// this status to us, but fetch the authentic article from
	// the server it originated from.
	if forwarded {

		// Pass the statusable URI (APIri) into the processor
		// worker and do the rest of the processing asynchronously.
		f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			APIRI:          ap.GetJSONLDId(statusable),
			APObject:       nil,
			GTSModel:       nil,
			Receiving:      receiver,
			Requesting:     requester,
		})
		return nil
	}

	// Do the rest of the processing asynchronously. The processor
	// will handle inserting/updating + further dereferencing the status.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		APIRI:          nil,
		GTSModel:       nil,
		APObject:       statusable,
		Receiving:      receiver,
		Requesting:     requester,
	})

	return nil
}
