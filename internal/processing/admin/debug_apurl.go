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

package admin

import (
	"context"
	"io"
	"net/http"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// DebugAPUrl performs a GET to the given url, using the
// signature of the given admin account. The GET will
// have Accept set to the ActivityPub content types.
//
// Only urls with schema http or https are allowed.
//
// Calls to blocked domains are not allowed, not only
// because it's unfair to call them when they can't
// call us, but because it probably won't work anyway
// if they try to dereference the calling account.
//
// Errors returned from this function should be fairly
// verbose, to help with debugging.
func (p *Processor) DebugAPUrl(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	urlStr string,
) (*apimodel.DebugAPUrlResponse, gtserror.WithCode) {
	// Validate URL.
	if urlStr == "" {
		err := gtserror.New("empty URL")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		err := gtserror.Newf("invalid URL: %w", err)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	if url == nil || (url.Scheme != "http" && url.Scheme != "https") {
		err = gtserror.New("invalid URL scheme, acceptable schemes are http or https")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure URL not blocked.
	blocked, err := p.state.DB.IsDomainBlocked(ctx, url.Host)
	if err != nil {
		err = gtserror.Newf("db error checking for domain block: %w", err)
		return nil, gtserror.NewErrorInternalError(err, err.Error())
	}

	if blocked {
		err = gtserror.Newf("target domain %s is blocked", url.Host)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// All looks fine. Prepare the transport and (signed) GET request.
	tsport, err := p.transport.NewTransportForUsername(ctx, adminAcct.Username)
	if err != nil {
		err = gtserror.Newf("error creating transport: %w", err)
		return nil, gtserror.NewErrorInternalError(err, err.Error())
	}

	req, err := http.NewRequestWithContext(
		// Caller will want a snappy
		// response so don't retry.
		gtscontext.SetFastFail(ctx),
		http.MethodGet, urlStr, nil,
	)
	if err != nil {
		err = gtserror.Newf("error creating request: %w", err)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	req.Header.Add("Accept", string(apiutil.AppActivityLDJSON)+","+string(apiutil.AppActivityJSON))
	req.Header.Add("Accept-Charset", "utf-8")

	// Perform the HTTP request,
	// and return everything.
	rsp, err := tsport.GET(req)
	if err != nil {
		err = gtserror.Newf("error doing dereference: %w", err)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}
	defer rsp.Body.Close()

	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		err := gtserror.Newf("error reading response body bytes: %w", err)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	debugResponse := &apimodel.DebugAPUrlResponse{
		RequestURL:      urlStr,
		RequestHeaders:  req.Header,
		ResponseHeaders: rsp.Header,
		ResponseCode:    rsp.StatusCode,
		ResponseBody:    string(b),
	}

	return debugResponse, nil
}
