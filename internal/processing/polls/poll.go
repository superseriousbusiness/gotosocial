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
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	// common processor logic
	c *common.Processor

	state     *state.State
	converter *typeutils.Converter
}

func New(common *common.Processor, state *state.State, converter *typeutils.Converter) Processor {
	return Processor{
		c:         common,
		state:     state,
		converter: converter,
	}
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
		const text = "target poll not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
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

// toAPIPoll converrts a given Poll to frontend API model, returning an appropriate error with HTTP code on failure.
func (p *Processor) toAPIPoll(ctx context.Context, requester *gtsmodel.Account, poll *gtsmodel.Poll) (*apimodel.Poll, gtserror.WithCode) {
	apiPoll, err := p.converter.PollToAPIPoll(ctx, requester, poll)
	if err != nil {
		err := gtserror.Newf("error converting to api model: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return apiPoll, nil
}
