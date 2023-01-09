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

package federation

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

const (
	nodeInfoVersion      = "2.0"
	nodeInfoSoftwareName = "gotosocial"
)

var (
	nodeInfoRel       = fmt.Sprintf("http://nodeinfo.diaspora.software/ns/schema/%s", nodeInfoVersion)
	nodeInfoProtocols = []string{"activitypub"}
)

func (p *processor) GetNodeInfoRel(ctx context.Context) (*apimodel.WellKnownResponse, gtserror.WithCode) {
	protocol := config.GetProtocol()
	host := config.GetHost()

	return &apimodel.WellKnownResponse{
		Links: []apimodel.Link{
			{
				Rel:  nodeInfoRel,
				Href: fmt.Sprintf("%s://%s/nodeinfo/%s", protocol, host, nodeInfoVersion),
			},
		},
	}, nil
}

func (p *processor) GetNodeInfo(ctx context.Context) (*apimodel.Nodeinfo, gtserror.WithCode) {
	openRegistration := config.GetAccountsRegistrationOpen()
	softwareVersion := config.GetSoftwareVersion()

	host := config.GetHost()
	userCount, err := p.db.CountInstanceUsers(ctx, host)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err, "Unable to query instance user count")
	}

	postCount, err := p.db.CountInstanceStatuses(ctx, host)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err, "Unable to query instance status count")
	}

	return &apimodel.Nodeinfo{
		Version: nodeInfoVersion,
		Software: apimodel.NodeInfoSoftware{
			Name:    nodeInfoSoftwareName,
			Version: softwareVersion,
		},
		Protocols: nodeInfoProtocols,
		Services: apimodel.NodeInfoServices{
			Inbound:  []string{},
			Outbound: []string{},
		},
		OpenRegistrations: openRegistration,
		Usage: apimodel.NodeInfoUsage{
			Users: apimodel.NodeInfoUsers{
				Total: userCount,
			},
			LocalPosts: postCount,
		},
		Metadata: make(map[string]interface{}),
	}, nil
}
