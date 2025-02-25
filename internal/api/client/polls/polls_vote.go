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
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// PollVotePOSTHandler swagger:operation POST /api/v1/polls/{id}/votes pollVote
//
// Vote with choices in the given poll.
//
//	---
//	tags:
//	- polls
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: Target poll ID.
//		in: path
//		required: true
//	-
//		name: choices
//		type: array
//		items:
//			type: integer
//		description: Poll choice indices on which to vote.
//		in: formData
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- write:statuses
//
//	responses:
//		'200':
//			description: "The updated poll with user vote choices."
//			schema:
//				"$ref": "#/definitions/poll"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'422':
//			description: unprocessable entity
//		'500':
//			description: internal server error
func (m *Module) PollVotePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		errWithCode := gtserror.NewErrorNotAcceptable(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	pollID, errWithCode := apiutil.ParseID(c.Param(IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	choices, err := bindChoices(c)
	if err != nil {
		errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	poll, errWithCode := m.processor.Polls().PollVote(
		c.Request.Context(),
		authed.Account,
		pollID,
		choices,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, poll)
}

func bindChoices(c *gin.Context) ([]int, error) {
	var form apimodel.PollVoteRequest
	if err := c.ShouldBind(&form); err != nil {
		return nil, err
	}

	if form.Choices != nil {
		// Easiest option: we parsed
		// from a form successfully.
		return form.Choices, nil
	}

	// More difficult option: we
	// parsed choices from json.
	//
	// Convert submitted choices
	// into the ints we need.
	choices := make([]int, 0, len(form.ChoicesI))
	for _, choiceI := range form.ChoicesI {
		switch i := choiceI.(type) {

		// JSON numbers normally
		// parse into float64.
		//
		// This is the most likely
		// option so try it first.
		case float64:
			choices = append(choices, int(i))

		// Fallback option for funky
		// clients (pinafore, semaphore).
		case string:
			choice, err := strconv.Atoi(i)
			if err != nil {
				return nil, err
			}

			choices = append(choices, choice)

		default:
			// Nothing else will do.
			return nil, fmt.Errorf("could not parse json poll choice %T to integer", choiceI)
		}
	}

	return choices, nil
}
