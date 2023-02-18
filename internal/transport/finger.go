/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
)

func (t *transport) Finger(ctx context.Context, targetUsername string, targetDomain string) ([]byte, error) {
	// Prepare URL string
	urlStr := "https://" +
		targetDomain +
		"/.well-known/webfinger?resource=acct:" +
		targetUsername + "@" + targetDomain

	// Generate new GET request from URL string
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", string(apiutil.AppJSON))
	req.Header.Add("Accept", "application/jrd+json")
	req.Header.Set("Host", req.URL.Host)

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	// Check for an expected status code
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request to %s failed: %s", urlStr, rsp.Status)
	}

	return io.ReadAll(rsp.Body)
}
