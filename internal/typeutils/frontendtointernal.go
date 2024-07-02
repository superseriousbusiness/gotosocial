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

package typeutils

import (
	"fmt"
	"net/url"
	"slices"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func APIVisToVis(m apimodel.Visibility) gtsmodel.Visibility {
	switch m {
	case apimodel.VisibilityPublic:
		return gtsmodel.VisibilityPublic
	case apimodel.VisibilityUnlisted:
		return gtsmodel.VisibilityUnlocked
	case apimodel.VisibilityPrivate:
		return gtsmodel.VisibilityFollowersOnly
	case apimodel.VisibilityMutualsOnly:
		return gtsmodel.VisibilityMutualsOnly
	case apimodel.VisibilityDirect:
		return gtsmodel.VisibilityDirect
	}
	return ""
}

func APIMarkerNameToMarkerName(m apimodel.MarkerName) gtsmodel.MarkerName {
	switch m {
	case apimodel.MarkerNameHome:
		return gtsmodel.MarkerNameHome
	case apimodel.MarkerNameNotifications:
		return gtsmodel.MarkerNameNotifications
	}
	return ""
}

func APIFilterActionToFilterAction(m apimodel.FilterAction) gtsmodel.FilterAction {
	switch m {
	case apimodel.FilterActionWarn:
		return gtsmodel.FilterActionWarn
	case apimodel.FilterActionHide:
		return gtsmodel.FilterActionHide
	}
	return gtsmodel.FilterActionNone
}

func APIPolicyURIToPolicyURI(u apimodel.PolicyURI) (gtsmodel.PolicyURI, error) {
	switch u {
	case apimodel.PolicyURIPublic:
		return gtsmodel.PolicyURIPublic, nil

	case apimodel.PolicyURIFollowers:
		return gtsmodel.PolicyURIFollowers, nil

	case apimodel.PolicyURIFollowing:
		return gtsmodel.PolicyURIFollowing, nil

	case apimodel.PolicyURIMutuals:
		return gtsmodel.PolicyURIMutuals, nil

	case apimodel.PolicyURIMentioned:
		return gtsmodel.PolicyURIMentioned, nil

	case apimodel.PolicyURIAuthor:
		return gtsmodel.PolicyURISelf, nil

	case apimodel.PolicyURIMe:
		err := fmt.Errorf("policyURI %s has no corresponding internal model", apimodel.PolicyURIMe)
		return "", err

	default:
		// Parse URI to ensure it's a
		// url with a valid protocol.
		url, err := url.Parse(string(u))
		if err != nil {
			return "", err
		}

		if url.Host != "http" && url.Host != "https" {
			err := fmt.Errorf("non-URN policyURIs must have protocol 'http' or 'https' (%s)", u)
			return "", err
		}

		return gtsmodel.PolicyURI(u), nil
	}
}

func APIInteractionPolicyToInteractionPolicy(
	p *apimodel.InteractionPolicy,
	v apimodel.Visibility,
) (*gtsmodel.InteractionPolicy, error) {
	visibility := APIVisToVis(v)

	convertURIs := func(apiURIs []apimodel.PolicyURI) (gtsmodel.PolicyURIs, error) {
		policyURIs := gtsmodel.PolicyURIs{}
		for _, apiURI := range apiURIs {
			uri, err := APIPolicyURIToPolicyURI(apiURI)
			if err != nil {
				return nil, err
			}

			if !uri.FeasibleForVisibility(visibility) {
				err := fmt.Errorf("policyURI %s is not feasible for visibility %s", apiURI, v)
				return nil, err
			}

			policyURIs = append(policyURIs, uri)
		}
		return policyURIs, nil
	}

	canLikeAlways, err := convertURIs(p.CanFavourite.Always)
	if err != nil {
		err := fmt.Errorf("error converting %s.can_favourite.always: %w", v, err)
		return nil, err
	}

	canLikeWithApproval, err := convertURIs(p.CanFavourite.WithApproval)
	if err != nil {
		err := fmt.Errorf("error converting %s.can_favourite.with_approval: %w", v, err)
		return nil, err
	}

	canReplyAlways, err := convertURIs(p.CanReply.Always)
	if err != nil {
		err := fmt.Errorf("error converting %s.can_reply.always: %w", v, err)
		return nil, err
	}

	canReplyWithApproval, err := convertURIs(p.CanReply.WithApproval)
	if err != nil {
		err := fmt.Errorf("error converting %s.can_reply.with_approval: %w", v, err)
		return nil, err
	}

	canAnnounceAlways, err := convertURIs(p.CanReblog.Always)
	if err != nil {
		err := fmt.Errorf("error converting %s.can_reblog.always: %w", v, err)
		return nil, err
	}

	canAnnounceWithApproval, err := convertURIs(p.CanReblog.WithApproval)
	if err != nil {
		err := fmt.Errorf("error converting %s.can_reblog.with_approval: %w", v, err)
		return nil, err
	}

	// Normalize URIs.
	//
	// 1. Ensure canLikeAlways, canReplyAlways,
	//    and canAnnounceAlways include self
	//    (either explicitly or within public).
	for _, uris := range []*gtsmodel.PolicyURIs{
		&canLikeAlways,
		&canReplyAlways,
		&canAnnounceAlways,
	} {
		if !slices.ContainsFunc(
			*uris,
			func(uri gtsmodel.PolicyURI) bool {
				return uri == gtsmodel.PolicyURIPublic ||
					uri == gtsmodel.PolicyURISelf
			},
		) {
			*uris = *util.Ptr(append(
				*uris,
				gtsmodel.PolicyURISelf,
			))
		}
	}

	// 2. Ensure canReplyAlways includes mentioned
	//    accounts (either explicitly or within public).
	if !slices.ContainsFunc(
		canReplyAlways,
		func(uri gtsmodel.PolicyURI) bool {
			return uri == gtsmodel.PolicyURIPublic ||
				uri == gtsmodel.PolicyURIMentioned
		},
	) {
		canReplyAlways = append(
			canReplyAlways,
			gtsmodel.PolicyURIMentioned,
		)
	}

	return &gtsmodel.InteractionPolicy{
		CanLike: gtsmodel.PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: gtsmodel.PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: gtsmodel.PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}, nil
}
