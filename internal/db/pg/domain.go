/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package pg

import (
	"context"
	"net/url"

	"github.com/go-pg/pg/v10"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type domainDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (d *domainDB) IsDomainBlocked(domain string) (bool, db.Error) {
	if domain == "" {
		return false, nil
	}

	blocked, err := d.conn.
		Model(&gtsmodel.DomainBlock{}).
		Where("LOWER(domain) = LOWER(?)", domain).
		Exists()

	err = processErrorResponse(err)

	return blocked, err
}

func (d *domainDB) AreDomainsBlocked(domains []string) (bool, db.Error) {
	// filter out any doubles
	uniqueDomains := util.UniqueStrings(domains)

	for _, domain := range uniqueDomains {
		if blocked, err := d.IsDomainBlocked(domain); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}

	// no blocks found
	return false, nil
}

func (d *domainDB) IsURIBlocked(uri *url.URL) (bool, db.Error) {
	domain := uri.Hostname()
	return d.IsDomainBlocked(domain)
}

func (d *domainDB) AreURIsBlocked(uris []*url.URL) (bool, db.Error) {
	domains := []string{}
	for _, uri := range uris {
		domains = append(domains, uri.Hostname())
	}

	return d.AreDomainsBlocked(domains)
}
