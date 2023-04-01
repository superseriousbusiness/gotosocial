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

package transport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"codeberg.org/gruf/go-byteutil"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (t *transport) BatchDeliver(ctx context.Context, b []byte, recipients []*url.URL) error {
	var (
		// errs accumulates errors received during
		// attempted delivery by deliverer routines.
		errs gtserror.MultiError

		// wait blocks until all sender
		// routines have returned.
		wait sync.WaitGroup

		// mutex protects 'recipients' and
		// 'errs' for concurrent access.
		mutex sync.Mutex

		// Get current instance host info.
		domain = config.GetAccountDomain()
		host   = config.GetHost()
	)

	// Block on expect no. senders.
	wait.Add(t.controller.senders)

	for i := 0; i < t.controller.senders; i++ {
		go func() {
			// Mark returned.
			defer wait.Done()

			for {
				// Acquire lock.
				mutex.Lock()

				if len(recipients) == 0 {
					// Reached end.
					return
				}

				// Pop next recipient.
				i := len(recipients) - 1
				to := recipients[i]
				recipients = recipients[:i]

				// Done with lock.
				mutex.Unlock()

				// Skip delivery to recipient if it is "us".
				if to.Host == host || to.Host == domain {
					continue
				}

				// Attempt to deliver data to recipient.
				if err := t.deliver(ctx, b, to); err != nil {
					mutex.Lock() // safely append err to accumulator.
					errs.Appendf("error delivering to %s: %v", to, err)
					mutex.Unlock()
				}
			}
		}()
	}

	// Wait for finish.
	wait.Wait()

	// Return combined err.
	return errs.Combine()
}

func (t *transport) Deliver(ctx context.Context, b []byte, to *url.URL) error {
	// if 'to' host is our own, skip as we don't need to deliver to ourselves...
	if to.Host == config.GetHost() || to.Host == config.GetAccountDomain() {
		return nil
	}

	// Deliver data to recipient.
	return t.deliver(ctx, b, to)
}

func (t *transport) deliver(ctx context.Context, b []byte, to *url.URL) error {
	url := to.String()

	// Use rewindable bytes reader for body.
	var body byteutil.ReadNopCloser
	body.Reset(b)

	req, err := http.NewRequestWithContext(ctx, "POST", url, &body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", string(apiutil.AppActivityLDJSON))
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Set("Host", to.Host)

	rsp, err := t.POST(req, b)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if code := rsp.StatusCode; code != http.StatusOK &&
		code != http.StatusCreated && code != http.StatusAccepted {
		err := fmt.Errorf("POST request to %s failed: %s", url, rsp.Status)
		return gtserror.WithStatusCode(err, rsp.StatusCode)
	}

	return nil
}
