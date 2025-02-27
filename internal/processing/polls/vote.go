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

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *Processor) PollVote(ctx context.Context, requester *gtsmodel.Account, pollID string, choices []int) (*apimodel.Poll, gtserror.WithCode) {
	// Get (+ check visibility of) requested poll with ID.
	poll, errWithCode := p.getTargetPoll(ctx, requester, pollID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	switch {
	// Poll author isn't allowed to vote in their own poll.
	case requester.ID == poll.Status.AccountID:
		const text = "you can't vote in your own poll"
		return nil, gtserror.NewErrorUnprocessableEntity(errors.New(text), text)

	// Poll has already closed, no more voting!
	case poll.Closed():
		const text = "poll already closed"
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

	// Wrap the choices in a PollVote model.
	vote := &gtsmodel.PollVote{
		ID:        id.NewULID(),
		Choices:   choices,
		AccountID: requester.ID,
		Account:   requester,
		PollID:    pollID,
		Poll:      poll,
	}

	// Insert the new poll votes into the database.
	err := p.state.DB.PutPollVote(ctx, vote)
	switch {

	case err == nil:
		// no issue.

	case errors.Is(err, db.ErrAlreadyExists):
		// Users cannot vote multiple *times* (not choices).
		const text = "you have already voted in poll"
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)

	default:
		// Any other irrecoverable database error.
		err := gtserror.Newf("error inserting poll vote: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Before enqueuing it, increment the poll
	// vote counts on the copy attached to the
	// PollVote (that we also later return).
	poll.IncrementVotes(choices, true)

	// Enqueue worker task to handle side-effects of user poll vote(s).
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityQuestion,
		GTSModel:       vote, // the vote choices
		Origin:         requester,
	})

	// Return converted API model poll.
	return p.toAPIPoll(ctx, requester, poll)
}
