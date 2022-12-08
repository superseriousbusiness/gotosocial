package bookmarks

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

const (
	// LimitKey is for setting the return amount limit for eg., requesting an account's statuses
	LimitKey = "limit"

	// MaxIDKey is for specifying the maximum ID of the bookmark to retrieve.
	MaxIDKey = "max_id"
	// MinIDKey is for specifying the minimum ID of the bookmark to retrieve.
	MinIDKey = "min_id"
)

// BookmarksGETHandler swagger:operation GET /api/v1/bookmarks bookmarksGet
//
// Get an array of statuses bookmarked in the instance
//
//	---
//	tags:
//	- bookmarks
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read:bookmarks
//
//	responses:
//		'200':
//			description: Array of bookmarked statuses
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/status"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//		'401':
//			description: unauthorized
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) BookmarksGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	limit := 30
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 64)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		limit = int(i)
	}

	maxID := ""
	maxIDString := c.Query(MaxIDKey)
	if maxIDString != "" {
		maxID = maxIDString
	}

	minID := ""
	minIDString := c.Query(MinIDKey)
	if minIDString != "" {
		minID = minIDString
	}

	resp, errWithCode := m.processor.BookmarksGet(c.Request.Context(), authed, maxID, minID, limit)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Items)
}
