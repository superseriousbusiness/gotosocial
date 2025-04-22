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

package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type profile struct {
	instance       *apimodel.InstanceV1
	account        *apimodel.WebAccount
	rssFeed        string
	robotsMeta     string
	pinnedStatuses []*apimodel.WebStatus
	statusResp     *apimodel.PageableResponse
	paging         bool
}

// prepareProfile does content type checks, fetches the
// targeted account from the db, and converts it to its
// web representation, along with other data needed to
// render the web view of the account.
func (m *Module) prepareProfile(c *gin.Context) *profile {
	ctx := c.Request.Context()

	// We'll need the instance later, and we can also use it
	// before then to make it easier to return a web error.
	instance, errWithCode := m.processor.InstanceGetV1(ctx)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return nil
	}

	// Return instance we already got from the db,
	// don't try to fetch it again when erroring.
	instanceGet := func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode) {
		return instance, nil
	}

	// Parse + normalize account username from the URL.
	requestedUsername, errWithCode := apiutil.ParseUsername(c.Param(apiutil.UsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return nil
	}
	requestedUsername = strings.ToLower(requestedUsername)

	// Check what type of content is being requested.
	// If we're getting an AP request on this endpoint
	// we should render the AP representation instead.
	contentType, err := apiutil.NegotiateAccept(c, apiutil.HTMLOrActivityPubHeaders...)
	if err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), instanceGet)
		return nil
	}

	if contentType == string(apiutil.AppActivityJSON) ||
		contentType == string(apiutil.AppActivityLDJSON) {
		// AP account representation has
		// been requested, return that.
		m.returnAPAccount(c, requestedUsername, contentType)
		return nil
	}

	// text/html has been requested.
	//
	// Proceed with getting the web
	// representation of the account.
	account, errWithCode := m.processor.Account().GetWeb(ctx, requestedUsername)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return nil
	}

	// If target account is suspended,
	// this page should not be visible.
	//
	// TODO: change this to 410?
	if account.Suspended {
		err := fmt.Errorf("target account %s is suspended", requestedUsername)
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotFound(err), instanceGet)
		return nil
	}

	// Only generate RSS link if
	// account has RSS enabled.
	var rssFeed string
	if account.EnableRSS {
		rssFeed = "/@" + account.Username + "/feed.rss"
	}

	// Only allow search robots
	// if account is discoverable.
	var robotsMeta string
	if account.Discoverable {
		robotsMeta = apiutil.RobotsDirectivesAllowSome
	}

	// Check if paging.
	maxStatusID := apiutil.ParseMaxID(c.Query(apiutil.MaxIDKey), "")
	paging := maxStatusID != ""

	// If not paging, load pinned statuses.
	var (
		mediaOnly      = account.WebLayout == "gallery"
		pinnedStatuses []*apimodel.WebStatus
	)
	if !paging {
		var errWithCode gtserror.WithCode
		pinnedStatuses, errWithCode = m.processor.Account().WebStatusesGetPinned(
			ctx,
			account.ID,
			mediaOnly,
		)
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, errWithCode, instanceGet)
			return nil
		}
	}

	// Limit varies depending on whether this is a gallery view or not.
	// If gallery view, we want a nice full screen of media, else we
	// don't want to overwhelm the viewer with a shitload of posts.
	var limit int
	if account.WebLayout == "gallery" {
		limit = 40
	} else {
		limit = 20
	}

	// Get statuses from maxStatusID onwards (or from top if empty string).
	statusResp, errWithCode := m.processor.Account().WebStatusesGet(
		ctx,
		account.ID,
		mediaOnly,
		limit,
		maxStatusID,
	)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return nil
	}

	return &profile{
		instance:       instance,
		account:        account,
		rssFeed:        rssFeed,
		robotsMeta:     robotsMeta,
		pinnedStatuses: pinnedStatuses,
		statusResp:     statusResp,
		paging:         paging,
	}
}

// profileGETHandler selects the appropriate rendering
// mode for the target account profile, and serves that.
func (m *Module) profileGETHandler(c *gin.Context) {
	p := m.prepareProfile(c)
	if p == nil {
		// Something went wrong,
		// error already written.
		return
	}

	// Choose desired web renderer for this acct.
	switch wrm := p.account.WebLayout; wrm {

	// El classico.
	case "", "microblog":
		m.profileMicroblog(c, p)

	// 'gram style media gallery.
	case "gallery":
		m.profileGallery(c, p)

	default:
		log.Panicf(
			c.Request.Context(),
			"unknown webrenderingmode %s", wrm,
		)
	}
}

// profileMicroblog serves the profile
// in classic GtS "microblog" view.
func (m *Module) profileMicroblog(c *gin.Context, p *profile) {
	// Prepare stylesheets for profile.
	stylesheets := make([]string, 0, 7)

	// Basic profile stylesheets.
	stylesheets = append(
		stylesheets,
		[]string{
			cssFA,
			cssStatus,
			cssThread,
			cssProfile,
		}...,
	)

	// User-selected theme if set.
	if theme := p.account.Theme; theme != "" {
		stylesheets = append(
			stylesheets,
			themesPathPrefix+"/"+theme,
		)
	}

	// Custom CSS for this user last in cascade.
	stylesheets = append(
		stylesheets,
		"/@"+p.account.Username+"/custom.css",
	)

	page := apiutil.WebPage{
		Template:    "profile.tmpl",
		Instance:    p.instance,
		OGMeta:      apiutil.OGBase(p.instance).WithAccount(p.account),
		Stylesheets: stylesheets,
		Javascript: []apiutil.JavascriptEntry{
			{
				Src:   jsFrontend,
				Async: true,
				Defer: true,
			},
			{
				Bottom: true,
				Src:    jsFrontendPrerender,
			},
		},
		Extra: map[string]any{
			"account":          p.account,
			"rssFeed":          p.rssFeed,
			"robotsMeta":       p.robotsMeta,
			"statuses":         p.statusResp.Items,
			"statuses_next":    p.statusResp.NextLink,
			"pinned_statuses":  p.pinnedStatuses,
			"show_back_to_top": p.paging,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

// profileMicroblog serves the profile
// in media-only 'gram-style gallery view.
func (m *Module) profileGallery(c *gin.Context, p *profile) {
	// Get just attachments from pinned,
	// making a rough guess for slice size.
	pinnedGalleryItems := make([]*apimodel.WebAttachment, 0, len(p.pinnedStatuses)*4)
	for _, status := range p.pinnedStatuses {
		pinnedGalleryItems = append(pinnedGalleryItems, status.MediaAttachments...)
	}

	// Get just attachments from statuses,
	// making a rough guess for slice size.
	galleryItems := make([]*apimodel.WebAttachment, 0, len(p.statusResp.Items)*4)
	for _, statusI := range p.statusResp.Items {
		status := statusI.(*apimodel.WebStatus)
		galleryItems = append(galleryItems, status.MediaAttachments...)
	}

	// Prepare stylesheets for profile.
	stylesheets := make([]string, 0, 4)

	// Profile gallery stylesheets.
	stylesheets = append(
		stylesheets,
		[]string{
			cssFA,
			cssProfileGallery,
		}...)

	// User-selected theme if set.
	if theme := p.account.Theme; theme != "" {
		stylesheets = append(
			stylesheets,
			themesPathPrefix+"/"+theme,
		)
	}

	// Custom CSS for this
	// user last in cascade.
	stylesheets = append(
		stylesheets,
		"/@"+p.account.Username+"/custom.css",
	)

	page := apiutil.WebPage{
		Template:    "profile-gallery.tmpl",
		Instance:    p.instance,
		OGMeta:      apiutil.OGBase(p.instance).WithAccount(p.account),
		Stylesheets: stylesheets,
		Javascript: []apiutil.JavascriptEntry{
			{
				Src:   jsFrontend,
				Async: true,
				Defer: true,
			},
			{
				Bottom: true,
				Src:    jsFrontendPrerender,
			},
		},
		Extra: map[string]any{
			"account":            p.account,
			"rssFeed":            p.rssFeed,
			"robotsMeta":         p.robotsMeta,
			"pinnedGalleryItems": pinnedGalleryItems,
			"galleryItems":       galleryItems,
			"statuses":           p.statusResp.Items,
			"statuses_next":      p.statusResp.NextLink,
			"pinned_statuses":    p.pinnedStatuses,
			"show_back_to_top":   p.paging,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

// returnAPAccount returns an ActivityPub representation of
// target account. It will do http signature authentication.
func (m *Module) returnAPAccount(
	c *gin.Context,
	targetUsername string,
	contentType string,
) {
	user, errWithCode := m.processor.Fedi().UserGet(c.Request.Context(), targetUsername, c.Request.URL)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSONType(c, http.StatusOK, contentType, user)
}
