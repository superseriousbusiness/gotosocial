/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package blocks

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// BlocksGETHandler swagger:operation GET /api/v1/blocks blocksGet
//
// Get an array of accounts that requesting account has blocked.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/blocks?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/blocks?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
// ---
// tags:
// - blocks
//
// produces:
// - application/json
//
// parameters:
// - name: limit
//   type: integer
//   description: Number of blocks to return.
//   default: 20
//   in: query
// - name: max_id
//   type: string
//   description: |-
//     Return only blocks *OLDER* than the given max block ID.
//     The block with the specified ID will not be included in the response.
//   in: query
// - name: since_id
//   type: string
//   description: |-
//     Return only blocks *NEWER* than the given since block ID.
//     The block with the specified ID will not be included in the response.
//   in: query
//
// security:
// - OAuth2 Bearer:
//   - read:blocks
//
// responses:
//   '200':
//     headers:
//       Link:
//         type: string
//         description: Links to the next and previous queries.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/account"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) BlocksGETHandler(c *gin.Context) {
	l := logrus.WithField("func", "PublicTimelineGETHandler")

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("error authing: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	maxID := ""
	maxIDString := c.Query(MaxIDKey)
	if maxIDString != "" {
		maxID = maxIDString
	}

	sinceID := ""
	sinceIDString := c.Query(SinceIDKey)
	if sinceIDString != "" {
		sinceID = sinceIDString
	}

	limit := 20
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 64)
		if err != nil {
			l.Debugf("error parsing limit string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse limit query param"})
			return
		}
		limit = int(i)
	}

	resp, errWithCode := m.processor.BlocksGet(c.Request.Context(), authed, maxID, sinceID, limit)
	if errWithCode != nil {
		l.Debugf("error from processor BlocksGet: %s", errWithCode)
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Accounts)
}
