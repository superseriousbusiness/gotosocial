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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

const (
	hostMetaXMLNS                   = "http://docs.oasis-open.org/ns/xri/xrd-1.0"
	hostMetaRel                     = "lrdd"
	hostMetaType                    = "application/xrd+xml"
	hostMetaTemplate                = ".well-known/webfinger?resource={uri}"
	nodeInfoVersion                 = "2.0"
	nodeInfoSoftwareName            = "gotosocial"
	nodeInfoRel                     = "http://nodeinfo.diaspora.software/ns/schema/" + nodeInfoVersion
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
				Rel:  nodeInfoRel,
				Href: fmt.Sprintf("%s://%s/nodeinfo/%s", protocol, host, nodeInfoVersion),
			},
		},
	}, nil
}

// NodeInfoGet returns a node info struct in response to a node info request.
func (p *Processor) NodeInfoGet(ctx context.Context) (*apimodel.Nodeinfo, gtserror.WithCode) {
	host := config.GetHost()

	userCount, err := p.state.DB.CountInstanceUsers(ctx, host)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	postCount, err := p.state.DB.CountInstanceStatuses(ctx, host)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apimodel.Nodeinfo{
		Version: nodeInfoVersion,
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
				Total: userCount,
			},
			LocalPosts: postCount,
		},
		Metadata: nodeInfoMetadata,
	}, nil
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
