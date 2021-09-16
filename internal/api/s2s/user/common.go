/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package user

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ActivityPubAcceptHeaders represents the Accept headers mentioned here:
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubAcceptHeaders = []string{
	`application/activity+json`,
	`application/ld+json; profile="https://www.w3.org/ns/activitystreams"`,
}

// populateContext transfers the signature verifier and signature from the gin context to the request context
func populateContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	verifier, signed := c.Get(string(util.APRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, util.APRequestingPublicKeyVerifier, verifier)
	}

	signature, signed := c.Get(string(util.APRequestingPublicKeySignature))
	if signed {
		ctx = context.WithValue(ctx, util.APRequestingPublicKeySignature, signature)
	}

	return ctx
}

func negotiateFormat(c *gin.Context) (string, error) {
	format := c.NegotiateFormat(ActivityPubAcceptHeaders...)
	if format == "" {
		return "", fmt.Errorf("no format can be offered for Accept headers %s", c.Request.Header.Get("Accept"))
	}
	return format, nil
}
