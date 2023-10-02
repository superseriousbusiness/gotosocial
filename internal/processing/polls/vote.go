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

package polls

import (
	"context"
	"errors"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *Processor) PollVote(ctx context.Context, requestingAccount *gtsmodel.Account, pollID string, choices []int) (*apimodel.Poll, gtserror.WithCode) {
	// Get (+ check visibility of) requested poll with ID.
	poll, errWithCode := p.getTargetPoll(ctx, requestingAccount, pollID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	switch {
	// Poll author isn't allowed to vote in their own poll.
	case requestingAccount.ID == poll.Status.AccountID:
		const text = "you can't vote in your own poll"
		return nil, gtserror.NewErrorUnprocessableEntity(errors.New(text), text)

	// Poll has already expired, no more voting!
	case time.Now().After(poll.ExpiresAt):
		const text = "poll has already expired"
		return nil, gtserror.NewErrorUnprocessableEntity(errors.New(text), text)

	// No choices given, or multiple given for single-choice poll.
	case len(choices) == 0 || (!*poll.Multiple && len(choices) > 1):
		const text = "invalid number of choices for poll"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	for _, choice := range choices {
		if choice < 0 || choice >= len(poll.Options) {
			// This is an invalid choice (index out of range).
			const text = "invalid option index for poll"
			return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
		}
	}

	// Convert choices to a slice of DB model poll votes.
	votes := make([]*gtsmodel.PollVote, len(choices))
	for i, choice := range choices {
		votes[i] = &gtsmodel.PollVote{
			ID:        id.NewULID(),
			Choice:    choice,
			AccountID: requestingAccount.ID,
			Account:   requestingAccount,
			PollID:    pollID,
			Poll:      poll,
		}
	}

	// Insert the new poll votes into the database.
	err := p.state.DB.PutPollVotes(ctx, votes...)
	switch {

	case err == nil:
		// no issue.

	case errors.Is(err, db.ErrAlreadyExists):
		// Users cannot vote multiple *times* (not choices).
		const text = "you have already voted in poll"
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)

	default:
		// Any other irrecoverable database error.
		err := gtserror.Newf("error inserting poll votes: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Enqueue worker task to handle side-effects of user poll vote(s)
	// (note this is handled as a NOTE UPDATE activity due to how polls
	// are represents in ActivityPub, i.e. as an alternative Note type).
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       poll.Status,
		OriginAccount:  requestingAccount,
	})

	// Convert the poll to API model view for the requesting account.
	apiPoll, err := p.converter.PollToAPIPoll(ctx, requestingAccount, poll)
	if err != nil {
		err := gtserror.Newf("error converting to api model: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiPoll, nil
}
