package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StatusRepliesGETHandler swagger:operation GET /users/{username}/statuses/{status}/replies s2sRepliesGet
//
// Get the replies collection for a status.
//
// Note that the response will be a Collection with a page as `first`, as shown below, if `page` is `false`.
//
// If `page` is `true`, then the response will be a single `CollectionPage` without the wrapping `Collection`.
//
// HTTP signature is required on the request.
//
// ---
// tags:
// - s2s/federation
//
// produces:
// - application/activity+json
//
// parameters:
// - name: username
//   type: string
//   description: Username of the account.
//   in: path
//   required: true
// - name: status
//   type: string
//   description: ID of the status.
//   in: path
//   required: true
// - name: page
//   type: boolean
//   description: Return response as a CollectionPage.
//   in: query
//   default: false
// - name: only_other_accounts
//   type: boolean
//   description: Return replies only from accounts other than the status owner.
//   in: query
//   default: false
// - name: min_id
//   type: string
//   description: Minimum ID of the next status, used for paging.
//   in: query
//
// responses:
//   '200':
//      in: body
//      schema:
//        "$ref": "#/definitions/swaggerStatusRepliesCollection"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '404':
//      description: not found
func (m *Module) StatusRepliesGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func": "StatusRepliesGETHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	requestedStatusID := c.Param(StatusIDKey)
	if requestedStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id specified in request"})
		return
	}

	page := false
	pageString := c.Query(PageKey)
	if pageString != "" {
		i, err := strconv.ParseBool(pageString)
		if err != nil {
			l.Debugf("error parsing page string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse page query param"})
			return
		}
		page = i
	}

	onlyOtherAccounts := false
	onlyOtherAccountsString := c.Query(OnlyOtherAccountsKey)
	if onlyOtherAccountsString != "" {
		i, err := strconv.ParseBool(onlyOtherAccountsString)
		if err != nil {
			l.Debugf("error parsing only_other_accounts string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse only_other_accounts query param"})
			return
		}
		onlyOtherAccounts = i
	}

	minID := ""
	minIDString := c.Query(MinIDKey)
	if minIDString != "" {
		minID = minIDString
	}

	format, err := negotiateFormat(c)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": fmt.Sprintf("could not negotiate format with given Accept header(s): %s", err)})
		return
	}
	l.Tracef("negotiated format: %s", format)

	ctx := populateContext(c)

	replies, errWithCode := m.processor.GetFediStatusReplies(ctx, requestedUsername, requestedStatusID, page, onlyOtherAccounts, minID, c.Request.URL)
	if err != nil {
		l.Info(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(replies)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, format, b)
}

// SwaggerStatusRepliesCollection represents a response to GET /users/{username}/statuses/{status}/replies.
// swagger:model swaggerStatusRepliesCollection
type SwaggerStatusRepliesCollection struct {
	// ActivityStreams context.
	// example: https://www.w3.org/ns/activitystreams
	Context string `json:"@context"`
	// ActivityStreams ID.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies
	ID string `json:"id"`
	// ActivityStreams type.
	// example: Collection
	Type string `json:"type"`
	// ActivityStreams first property.
	First SwaggerStatusRepliesCollectionPage `json:"first"`
}

// SwaggerStatusRepliesCollectionPage represents one page of a collection.
// swagger:model swaggerStatusRepliesCollectionPage
type SwaggerStatusRepliesCollectionPage struct {
	// ActivityStreams ID.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies?page=true
	ID string `json:"id"`
	// ActivityStreams type.
	// example: CollectionPage
	Type string `json:"type"`
	// Link to the next page.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies?only_other_accounts=true&page=true
	Next string `json:"next"`
	// Collection this page belongs to.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies
	PartOf string `json:"partOf"`
	// Items on this page.
	// example: ["https://example.org/users/some_other_user/statuses/086417595981111564", "https://another.example.com/users/another_user/statuses/01FCN8XDV3YG7B4R42QA6YQZ9R"]
	Items []string `json:"items"`
}
