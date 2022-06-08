package favourites

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// FavouritesGETHandler handles GETting favourites.
func (m *Module) FavouritesGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
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

	limit := 20
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

	resp, errWithCode := m.processor.FavedTimelineGet(c.Request.Context(), authed, maxID, minID, limit)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Items)
}
