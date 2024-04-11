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
	"net"
	"slices"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) AccountsGet(
	ctx context.Context,
	request *apimodel.AdminGetAccountsRequest,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Validate "origin".
	if v := request.Origin; v != "" {
		valid := []string{"local", "remote"}
		if !slices.Contains(valid, v) {
			err := fmt.Errorf("origin %s not recognized; valid choices are %+v", v, valid)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	// Validate "status".
	if v := request.Status; v != "" {
		valid := []string{"active", "pending", "disabled", "silenced", "suspended"}
		if !slices.Contains(valid, v) {
			err := fmt.Errorf("status %s not recognized; valid choices are %+v", v, valid)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	// Validate "permissions".
	if v := request.Permissions; v != "" {
		valid := []string{"staff"}
		if !slices.Contains(valid, v) {
			err := fmt.Errorf("permissions %s not recognized; valid choices are %+v", v, valid)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	var ip net.IP
	if v := request.IP; v != "" {
		ip = net.ParseIP(v)
		if ip == nil {
			err := fmt.Errorf("ip %s not a valid IP address", v)
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
		request.MaxID,
		request.SinceID,
		request.MinID,
		request.Limit,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting accounts: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(accounts)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	nextMax := accounts[count-1].ID
	prevMin := accounts[0].ID

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
		return packageAccountsV1(items, nextMax, prevMin, request)

	case 2:
		return packageAccountsV2(items, nextMax, prevMin, request)

	default:
		log.Panic(ctx, "api version was neither 1 nor 2")
		return nil, nil
	}
}

func packageAccountsV1(
	items []interface{},
	nextMax string,
	prevMin string,
	request *apimodel.AdminGetAccountsRequest,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	extraQueryParams := []string{}

	// Translate origin to v1.
	if v := request.Origin; v != "" {
		var k string

		if v == "local" {
			k = apiutil.LocalKey
		} else {
			k = apiutil.AdminRemoteKey
		}

		extraQueryParams = append(
			extraQueryParams,
			k+"=true",
		)
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

		extraQueryParams = append(
			extraQueryParams,
			k+"=true",
		)
	}

	if v := request.Username; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.UsernameKey+"="+v,
		)
	}

	if v := request.DisplayName; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminDisplayNameKey+"="+v,
		)
	}

	if v := request.ByDomain; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminByDomainKey+"="+v,
		)
	}

	if v := request.Email; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminEmailKey+"="+v,
		)
	}

	if v := request.IP; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminIPKey+"="+v,
		)
	}

	// Translate permissions to v1.
	if v := request.Permissions; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminStaffKey+"=true",
		)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:            items,
		Path:             "/api/v1/admin/accounts",
		NextMaxIDValue:   nextMax,
		PrevMinIDValue:   prevMin,
		Limit:            request.Limit,
		ExtraQueryParams: extraQueryParams,
	})
}

func packageAccountsV2(
	items []interface{},
	nextMax string,
	prevMin string,
	request *apimodel.AdminGetAccountsRequest,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	extraQueryParams := []string{}

	if v := request.Origin; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminOriginKey+"="+v,
		)
	}

	if v := request.Status; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminStatusKey+"="+v,
		)
	}

	if v := request.Permissions; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminPermissionsKey+"="+v,
		)
	}

	if v := request.InvitedBy; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminInvitedByKey+"="+v,
		)
	}

	if v := request.Username; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.UsernameKey+"="+v,
		)
	}

	if v := request.DisplayName; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminDisplayNameKey+"="+v,
		)
	}

	if v := request.ByDomain; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminByDomainKey+"="+v,
		)
	}

	if v := request.Email; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminEmailKey+"="+v,
		)
	}

	if v := request.IP; v != "" {
		extraQueryParams = append(
			extraQueryParams,
			apiutil.AdminIPKey+"="+v,
		)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:            items,
		Path:             "/api/v2/admin/accounts",
		NextMaxIDValue:   nextMax,
		PrevMinIDValue:   prevMin,
		Limit:            request.Limit,
		ExtraQueryParams: extraQueryParams,
	})
}
