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

package transport

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

func (t *transport) BatchDeliver(ctx context.Context, b []byte, recipients []*url.URL) error {
	// concurrently deliver to recipients; for each delivery, buffer the error if it fails
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(recipients))
	for _, recipient := range recipients {
		wg.Add(1)
		go func(r *url.URL) {
			defer wg.Done()
			if err := t.Deliver(ctx, b, r); err != nil {
				errCh <- err
			}
		}(recipient)
	}

	// wait until all deliveries have succeeded or failed
	wg.Wait()

	// receive any buffered errors
	errs := make([]string, 0, len(recipients))
outer:
	for {
		select {
		case e := <-errCh:
			errs = append(errs, e.Error())
		default:
			break outer
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("BatchDeliver: at least one failure: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (t *transport) Deliver(ctx context.Context, b []byte, to *url.URL) error {
	// if the 'to' host is our own, just skip this delivery since we by definition already have the message!
	if to.Host == config.GetHost() || to.Host == config.GetAccountDomain() {
		return nil
	}

	urlStr := to.String()

	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", string(apiutil.AppActivityLDJSON))
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Set("Host", to.Host)

	resp, err := t.POST(req, b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK &&
		code != http.StatusCreated && code != http.StatusAccepted {
		return fmt.Errorf("POST request to %s failed (%d): %s", urlStr, resp.StatusCode, resp.Status)
	}

	return nil
}
