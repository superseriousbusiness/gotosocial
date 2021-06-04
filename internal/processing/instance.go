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

package processing

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) InstanceGet(domain string) (*apimodel.Instance, gtserror.WithCode) {
	i := &gtsmodel.Instance{}
	if err := p.db.GetWhere([]db.Where{{Key: "domain", Value: domain}}, i); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance %s: %s", p.config.Host, err))
	}

	ai, err := p.tc.InstanceToMasto(i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}
