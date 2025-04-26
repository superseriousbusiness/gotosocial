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

package fedi

import (
	"context"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

const (
	hostMetaXMLNS                   = "http://docs.oasis-open.org/ns/xri/xrd-1.0"
	hostMetaRel                     = "lrdd"
	hostMetaType                    = "application/xrd+xml"
	hostMetaTemplate                = ".well-known/webfinger?resource={uri}"
	nodeInfoSoftwareName            = "gotosocial"
	nodeInfo20Rel                   = "http://nodeinfo.diaspora.software/ns/schema/2.0"
	nodeInfo21Rel                   = "http://nodeinfo.diaspora.software/ns/schema/2.1"
	nodeInfoRepo                    = "https://codeberg.org/superseriousbusiness/gotosocial"
	nodeInfoHomepage                = "https://docs.gotosocial.org"
	webfingerProfilePage            = "http://webfinger.net/rel/profile-page"
	webFingerProfilePageContentType = "text/html"
	webfingerSelf                   = "self"
	webFingerSelfContentType        = "application/activity+json"
	webfingerAccount                = "acct"
)

var (
	nodeInfoProtocols = []string{"activitypub"}
	nodeInfoInbound   = []string{}
	nodeInfoOutbound  = []string{}
	nodeInfoMetadata  = make(map[string]interface{})
)

// NodeInfoRelGet returns a well known response giving the path to node info.
func (p *Processor) NodeInfoRelGet(ctx context.Context) (*apimodel.WellKnownResponse, gtserror.WithCode) {
	protocol := config.GetProtocol()
	host := config.GetHost()

	return &apimodel.WellKnownResponse{
		Links: []apimodel.Link{
			{
				Rel:  nodeInfo20Rel,
				Href: fmt.Sprintf("%s://%s/nodeinfo/2.0", protocol, host),
			},
			{
				Rel:  nodeInfo21Rel,
				Href: fmt.Sprintf("%s://%s/nodeinfo/2.1", protocol, host),
			},
		},
	}, nil
}

// NodeInfoGet returns a node info struct in response to a 2.0 or 2.1 node info request.
func (p *Processor) NodeInfoGet(ctx context.Context, schemaVersion string) (*apimodel.Nodeinfo, gtserror.WithCode) {
	const ()

	var (
		userCount int
		postCount int
		mau       int
		err       error
	)

	switch config.GetInstanceStatsMode() {

	case config.InstanceStatsModeBaffle:
		// Use randomized stats.
		stats := p.converter.RandomStats()
		userCount = int(stats.TotalUsers)
		postCount = int(stats.Statuses)
		mau = int(stats.MonthlyActiveUsers)

	case config.InstanceStatsModeZero:
		// Use zeroed stats
		// (don't count anything).

	default:
		// Mode is either "serve" or "default".
		// Count actual stats.
		host := config.GetHost()

		userCount, err = p.state.DB.CountInstanceUsers(ctx, host)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		postCount, err = p.state.DB.CountInstanceStatuses(ctx, host)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	nodeInfo := &apimodel.Nodeinfo{
		Version: schemaVersion,
		Software: apimodel.NodeInfoSoftware{
			Name:    nodeInfoSoftwareName,
			Version: config.GetSoftwareVersion(),
		},
		Protocols: nodeInfoProtocols,
		Services: apimodel.NodeInfoServices{
			Inbound:  nodeInfoInbound,
			Outbound: nodeInfoOutbound,
		},
		OpenRegistrations: config.GetAccountsRegistrationOpen(),
		Usage: apimodel.NodeInfoUsage{
			Users: apimodel.NodeInfoUsers{
				Total:       userCount,
				ActiveMonth: mau,
			},
			LocalPosts: postCount,
		},
		Metadata: nodeInfoMetadata,
	}

	if schemaVersion == "2.1" {
		nodeInfo.Software.Repository = nodeInfoRepo
		nodeInfo.Software.Homepage = nodeInfoHomepage
	}

	return nodeInfo, nil
}

// HostMetaGet returns a host-meta struct in response to a host-meta request.
func (p *Processor) HostMetaGet() *apimodel.HostMeta {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return &apimodel.HostMeta{
		XMLNS: hostMetaXMLNS,
		Link: []apimodel.Link{
			{
				Rel:      hostMetaRel,
				Type:     hostMetaType,
				Template: fmt.Sprintf("%s://%s/%s", protocol, host, hostMetaTemplate),
			},
		},
	}
}

// WebfingerGet handles the GET for a webfinger resource. Most commonly, it will be used for returning account lookups.
func (p *Processor) WebfingerGet(ctx context.Context, requestedUsername string) (*apimodel.WellKnownResponse, gtserror.WithCode) {
	// Get the local account the request is referring to.
	requestedAccount, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	return &apimodel.WellKnownResponse{
		Subject: webfingerAccount + ":" + requestedAccount.Username + "@" + config.GetAccountDomain(),
		Aliases: []string{
			requestedAccount.URI,
			requestedAccount.URL,
		},
		Links: []apimodel.Link{
			{
				Rel:  webfingerProfilePage,
				Type: webFingerProfilePageContentType,
				Href: requestedAccount.URL,
			},
			{
				Rel:  webfingerSelf,
				Type: webFingerSelfContentType,
				Href: requestedAccount.URI,
			},
		},
	}, nil
}
