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

package admin

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"slices"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

var (
	accountsValidOrigins     = []string{"local", "remote"}
	accountsValidStatuses    = []string{"active", "pending", "disabled", "silenced", "suspended"}
	accountsValidPermissions = []string{"staff"}
)

func (p *Processor) AccountsGet(
	ctx context.Context,
	request *apimodel.AdminGetAccountsRequest,
	page *paging.Page,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	// Validate "origin".
	if v := request.Origin; v != "" {
		if !slices.Contains(accountsValidOrigins, v) {
			err := fmt.Errorf(
				"origin %s not recognized; valid choices are %+v",
				v, accountsValidOrigins,
			)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	// Validate "status".
	if v := request.Status; v != "" {
		if !slices.Contains(accountsValidStatuses, v) {
			err := fmt.Errorf(
				"status %s not recognized; valid choices are %+v",
				v, accountsValidStatuses,
			)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	// Validate "permissions".
	if v := request.Permissions; v != "" {
		if !slices.Contains(accountsValidPermissions, v) {
			err := fmt.Errorf(
				"permissions %s not recognized; valid choices are %+v",
				v, accountsValidPermissions,
			)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	// Validate/parse IP.
	var ip netip.Addr
	if v := request.IP; v != "" {
		var err error
		ip, err = netip.ParseAddr(request.IP)
		if err != nil {
			err := fmt.Errorf("invalid ip provided: %w", err)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	// Get accounts with the given params.
	accounts, err := p.state.DB.GetAccounts(
		ctx,
		request.Origin,
		request.Status,
		func() bool { return request.Permissions == "staff" }(),
		request.InvitedBy,
		request.Username,
		request.DisplayName,
		request.ByDomain,
		request.Email,
		ip,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting accounts: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(accounts)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		loAcct = accounts[count-1]
		hiAcct = accounts[0]
		lo     = loAcct.Domain + "/@" + loAcct.Username
		hi     = hiAcct.Domain + "/@" + hiAcct.Username
	)

	items := make([]interface{}, 0, count)
	for _, account := range accounts {
		apiAccount, err := p.converter.AccountToAdminAPIAccount(ctx, account)
		if err != nil {
			log.Errorf(ctx, "error converting to api account: %v", err)
			continue
		}
		items = append(items, apiAccount)
	}

	// Return packaging + paging appropriate for
	// the API version used to call this function.
	switch request.APIVersion {
	case 1:
		return packageAccountsV1(items, lo, hi, request, page)

	case 2:
		return packageAccountsV2(items, lo, hi, request, page)

	default:
		log.Panic(ctx, "api version was neither 1 nor 2")
		return nil, nil
	}
}

func packageAccountsV1(
	items []interface{},
	loID, hiID string,
	request *apimodel.AdminGetAccountsRequest,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	queryParams := make(url.Values, 8)

	// Translate origin to v1.
	if v := request.Origin; v != "" {
		var k string

		if v == "local" {
			k = apiutil.LocalKey
		} else {
			k = apiutil.AdminRemoteKey
		}

		queryParams.Add(k, "true")
	}

	// Translate status to v1.
	if v := request.Status; v != "" {
		var k string

		switch v {
		case "active":
			k = apiutil.AdminActiveKey
		case "pending":
			k = apiutil.AdminPendingKey
		case "disabled":
			k = apiutil.AdminDisabledKey
		case "silenced":
			k = apiutil.AdminSilencedKey
		case "suspended":
			k = apiutil.AdminSuspendedKey
		}

		queryParams.Add(k, "true")
	}

	if v := request.Username; v != "" {
		queryParams.Add(apiutil.UsernameKey, v)
	}

	if v := request.DisplayName; v != "" {
		queryParams.Add(apiutil.AdminDisplayNameKey, v)
	}

	if v := request.ByDomain; v != "" {
		queryParams.Add(apiutil.AdminByDomainKey, v)
	}

	if v := request.Email; v != "" {
		queryParams.Add(apiutil.AdminEmailKey, v)
	}

	if v := request.IP; v != "" {
		queryParams.Add(apiutil.AdminIPKey, v)
	}

	// Translate permissions to v1.
	if v := request.Permissions; v != "" {
		queryParams.Add(apiutil.AdminStaffKey, v)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/accounts",
		Next:  page.Next(loID, hiID),
		Prev:  page.Prev(loID, hiID),
		Query: queryParams,
	}), nil
}

func packageAccountsV2(
	items []interface{},
	loID, hiID string,
	request *apimodel.AdminGetAccountsRequest,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	queryParams := make(url.Values, 9)

	if v := request.Origin; v != "" {
		queryParams.Add(apiutil.AdminOriginKey, v)
	}

	if v := request.Status; v != "" {
		queryParams.Add(apiutil.AdminStatusKey, v)
	}

	if v := request.Permissions; v != "" {
		queryParams.Add(apiutil.AdminPermissionsKey, v)
	}

	if v := request.InvitedBy; v != "" {
		queryParams.Add(apiutil.AdminInvitedByKey, v)
	}

	if v := request.Username; v != "" {
		queryParams.Add(apiutil.UsernameKey, v)
	}

	if v := request.DisplayName; v != "" {
		queryParams.Add(apiutil.AdminDisplayNameKey, v)
	}

	if v := request.ByDomain; v != "" {
		queryParams.Add(apiutil.AdminByDomainKey, v)
	}

	if v := request.Email; v != "" {
		queryParams.Add(apiutil.AdminEmailKey, v)
	}

	if v := request.IP; v != "" {
		queryParams.Add(apiutil.AdminIPKey, v)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v2/admin/accounts",
		Next:  page.Next(loID, hiID),
		Prev:  page.Prev(loID, hiID),
		Query: queryParams,
	}), nil
}
