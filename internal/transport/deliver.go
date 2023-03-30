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
	"errors"
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

		// mutex protects the error accumulator.
		mutex sync.Mutex

		// sender is a just a short-hand to sender worker.
		sender = &t.controller.state.Workers.Sender

		// Get current instance host info.
		domain = config.GetAccountDomain()
		host   = config.GetHost()
	)

	for _, to := range recipients {
		// Skip delivery to recipient if it is "us".
		if to.Host == host || to.Host == domain {
			continue
		}

		// Track sender.
		wait.Add(1)

		// Rescope to loop.
		recipient := to

		// Enqueue delivery for each recipient to send worker.
		if !sender.EnqueueCtx(ctx, func(ctx context.Context) {
			defer wait.Done()

			if err := t.deliver(ctx, b, recipient); err != nil {
				mutex.Lock() // safely append err to accumulator.
				errs.Appendf("error delivering to %s: %v", to, err)
				mutex.Unlock()
			}
		}) {
			// Enqueue failed, i.e. our ctx or worker ctx was canceled.
			return errors.New("context canceled during batch delivery")
		}
	}

	// Wait for finish.
	wait.Wait()

	// Return combined err.
	return errs.Combine()
}

func (t *transport) Deliver(ctx context.Context, b []byte, to *url.URL) (err error) {
	// sender is a just a short-hand to sender worker.
	sender := &t.controller.state.Workers.Sender

	// if 'to' host is our own, skip as we don't need to deliver to ourselves...
	if to.Host == config.GetHost() || to.Host == config.GetAccountDomain() {
		return nil
	}

	if !sender.EnqueueCtx(ctx, func(ctx context.Context) {
		// Deliver bytes to recipient.
		err = t.deliver(ctx, b, to)
	}) {
		// Pool/our ctx was canceled.
		err = context.Canceled
	}

	return err
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
