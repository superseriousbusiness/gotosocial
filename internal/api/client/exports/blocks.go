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

package exports

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// ExportBlocksGETHandler swagger:operation GET /api/v1/exports/blocks.csv exportBlocks
//
// Export a CSV file of accounts that you block.
//
//	---
//	tags:
//	- import-export
//
//	produces:
//	- text/csv
//
//	security:
//	- OAuth2 Bearer:
//		- read:blocks
//
//	responses:
//		'200':
//			name: accounts
//			description: CSV file of accounts that you block.
//		'401':
//			description: unauthorized
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) ExportBlocksGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadBlocks,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.CSVHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	records, errWithCode := m.processor.Account().ExportBlocks(
		c.Request.Context(),
		authed.Account,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.EncodeCSVResponse(c.Writer, c.Request, http.StatusOK, records)
}
