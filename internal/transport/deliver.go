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
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	if to.Host == viper.GetString(config.Keys.Host) || to.Host == viper.GetString(config.Keys.AccountDomain) {
		return nil
	}

	logrus.Debugf("Deliver: posting as %s to %s", t.pubKeyID, to.String())
	return t.sigTransport.Deliver(ctx, b, to)
}
