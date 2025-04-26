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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
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
func (p *Processor) getTargetPoll(ctx context.Context, requester *gtsmodel.Account, targetID string) (*gtsmodel.Poll, gtserror.WithCode) {
	// Load the status the poll is attached to by the poll ID,
	// checking for visibility and ensuring it is up-to-date.
	status, errWithCode := p.c.GetVisibleTargetStatusBy(ctx,
		requester,
		func() (*gtsmodel.Status, error) {
			return p.state.DB.GetStatusByPollID(ctx, targetID)
		},
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Return most up-to-date
	// copy of the status poll.
	poll := status.Poll
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
