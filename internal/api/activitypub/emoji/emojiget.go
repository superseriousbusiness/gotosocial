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

package emoji

import (
	"errors"
	"net/http"
	"strings"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

func (m *Module) EmojiGetHandler(c *gin.Context) {
	requestedEmojiID := strings.ToUpper(c.Param(EmojiIDKey))
	if requestedEmojiID == "" {
		err := errors.New("no emoji id specified in request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	contentType, err := apiutil.NegotiateAccept(c, apiutil.ActivityPubHeaders...)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Fedi().EmojiGet(c.Request.Context(), requestedEmojiID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Encode JSON HTTP response.
	apiutil.EncodeJSONResponse(
		c.Writer,
		c.Request,
		http.StatusOK,
		contentType,
		resp,
	)
}
