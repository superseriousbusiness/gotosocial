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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *Processor) PollVote(ctx context.Context, requestingAccount *gtsmodel.Account, pollID string, choices []int) (*apimodel.Poll, gtserror.WithCode) {
	// Get (+ check visibility of) requested poll with ID.
	poll, errWithCode := p.getTargetPoll(ctx, requestingAccount, pollID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if !*poll.Multiple && len(choices) > 1 {
		// Multiple given for single-choice poll.
		const text = "poll only accepts single choice"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	for _, choice := range choices {
		if choice >= len(poll.Options) {
			// This is an invalid poll choice (index out of range).
			const text = "poll choice out of range"
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
	switch err := p.state.DB.PutPollVotes(ctx, votes...); {

	case err == nil:
		// no issue.

	case errors.Is(err, db.ErrAlreadyExists):
		// Users cannot vote multiple *times* (not choices).
		const text = "already voted in poll"
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)

	default:
		// Any other irrecoverable database error.
		err := gtserror.Newf("error inserting poll votes: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert the poll to API model view for the requesting account.
	apiPoll, err := p.converter.PollToAPIPoll(ctx, requestingAccount, poll)
	if err != nil {
		err := gtserror.Newf("error converting to api model: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiPoll, nil
}
