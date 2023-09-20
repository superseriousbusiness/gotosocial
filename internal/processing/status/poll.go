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

package status

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (p *Processor) PollGet(ctx context.Context, requestingAccount *gtsmodel.Account, pollID string) (*apimodel.Poll, gtserror.WithCode) {
	// Get (+ check visibility of) requested poll with ID.
	poll, errWithCode := p.getTargetPoll(ctx, requestingAccount, pollID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Package poll API model response.
	return p.toAPIPoll(ctx, poll)
}

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
		if err := p.state.DB.DeletePollVotes(ctx, pollID, requestingAccount.ID); err != nil {
			log.Errorf(ctx, "error deleting poll %s vote(s) for %s: %v", poll.Status.URI, requestingAccount.URI, err)
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
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Package poll API model response.
	return p.toAPIPoll(ctx, poll)
}

// getTargetPoll fetches a target poll ID for requesting account, taking visibility of the poll's originating status into account.
func (p *Processor) getTargetPoll(ctx context.Context, requestingAccount *gtsmodel.Account, targetID string) (*gtsmodel.Poll, gtserror.WithCode) {
	// Load the requested poll with ID.
	// (barebones as we fetch status below)
	poll, err := p.state.DB.GetPollByID(
		gtscontext.SetBarebones(ctx),
		targetID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if poll == nil {
		// No poll could be found for given ID.
		err := errors.New("target poll not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Check that we can see + fetch the originating status for requesting account.
	status, errWithCode := p.c.GetVisibleTargetStatus(ctx, requestingAccount, poll.StatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Update poll status.
	poll.Status = status

	return poll, nil
}

func (p *Processor) toAPIPoll(ctx context.Context, poll *gtsmodel.Poll) (*apimodel.Poll, gtserror.WithCode) {
	// Preallocate a slice of frontend model poll options.
	options := make([]apimodel.PollOption, len(poll.Options))

	if *poll.HideCounts {
		// Simply set the poll option titles.
		for i, title := range poll.Options {
			options[i].Title = title
		}
	} else {
		// Set the poll option titles WITH vote counts.
		for i, title := range poll.Options {
			options[i].Title = title
			options[i].VotesCount = 999
		}
	}

	return nil, nil
}
