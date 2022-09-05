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

package emoji

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// EmojiGetHandler 
func (m *Module) EmojiGetHandler(c *gin.Context) {
	// usernames on our instance are always lowercase
	requestedEmojiID := strings.ToUpper(c.Param(EmojiIDKey))
	if requestedEmojiID == "" {
		err := errors.New("no emoji id specified in request")
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	format, err := api.NegotiateAccept(c, api.ActivityPubAcceptHeaders...)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	ctx := c.Request.Context()
	verifier, signed := c.Get(string(ap.ContextRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeyVerifier, verifier)
	}

	signature, signed := c.Get(string(ap.ContextRequestingPublicKeySignature))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeySignature, signature)
	}

	resp, errWithCode := m.processor.GetFediEmoji(ctx, requestedEmojiID, c.Request.URL)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	b, err := json.Marshal(resp)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	c.Data(http.StatusOK, format, b)
}
