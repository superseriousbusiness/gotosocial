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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *Processor) PollVote(ctx context.Context, requestingAccount *gtsmodel.Account, pollID string, choice int) (*apimodel.Poll, gtserror.WithCode) {
	// Get (+ check visibility of) requested poll with ID.
	poll, errWithCode := p.getTargetPoll(ctx, requestingAccount, pollID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if choice >= len(poll.Options) {
		// This is an invalid poll choice.
		err := errors.New("invalid poll choice")
		return nil, gtserror.NewErrorBadRequest(err)
	}

	if !*poll.Multiple {
		// This is a SINGLE choice poll, we need to delete any existing votes by account for poll.
		if err := p.state.DB.DeletePollVotesBy(ctx, pollID, requestingAccount.ID); err != nil {
			err := gtserror.Newf("error deleting poll vote: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Create new poll vote for choice.
	vote := &gtsmodel.PollVote{
		ID:        id.NewULID(),
		Choice:    choice,
		AccountID: requestingAccount.ID,
		Account:   requestingAccount,
		PollID:    pollID,
		Poll:      poll,
	}

	// Insert the new poll vote into the database.
	if err := p.state.DB.PutPollVote(ctx, vote); err != nil {
		err := gtserror.Newf("error inserting poll vote: %w", err)
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
