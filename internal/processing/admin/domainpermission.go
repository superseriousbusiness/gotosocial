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
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// DomainPermissionCreate creates an instance-level permission
// targeting the given domain, and then processes any side
// effects of the permission creation.
//
// If the same permission type already exists for the domain,
// side effects will be retried.
//
// Return values for this function are the new or existing
// domain permission, the ID of the admin action resulting
// from this call, and/or an error if something goes wrong.
func (p *Processor) DomainPermissionCreate(
	ctx context.Context,
	permissionType gtsmodel.DomainPermissionType,
	adminAcct *gtsmodel.Account,
	domain string,
	obfuscate bool,
	publicComment string,
	privateComment string,
	subscriptionID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	switch permissionType {

	// Explicitly block a domain.
	case gtsmodel.DomainPermissionBlock:
		return p.createDomainBlock(
			ctx,
			adminAcct,
			domain,
			obfuscate,
			publicComment,
			privateComment,
			subscriptionID,
		)

	// Explicitly allow a domain.
	case gtsmodel.DomainPermissionAllow:
		return p.createDomainAllow(
			ctx,
			adminAcct,
			domain,
			obfuscate,
			publicComment,
			privateComment,
			subscriptionID,
		)

	// Weeping, roaring, red-faced.
	default:
		err := gtserror.Newf("unrecognized permission type %d", permissionType)
		return nil, "", gtserror.NewErrorInternalError(err)
	}
}

// DomainPermissionUpdate updates a domain permission
// of the given permissionType, with the given ID.
func (p *Processor) DomainPermissionUpdate(
	ctx context.Context,
	permissionType gtsmodel.DomainPermissionType,
	permID string,
	obfuscate *bool,
	publicComment *string,
	privateComment *string,
	subscriptionID *string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	switch permissionType {

	// Explicitly block a domain.
	case gtsmodel.DomainPermissionBlock:
		return p.updateDomainBlock(
			ctx,
			permID,
			obfuscate,
			publicComment,
			privateComment,
			subscriptionID,
		)

	// Explicitly allow a domain.
	case gtsmodel.DomainPermissionAllow:
		return p.updateDomainAllow(
			ctx,
			permID,
			obfuscate,
			publicComment,
			privateComment,
			subscriptionID,
		)

	// 🎵 Why don't we all strap bombs to our chests,
	// and ride our bikes to the next G7 picnic?
	// Seems easier with every clock-tick. 🎵
	default:
		err := gtserror.Newf("unrecognized permission type %d", permissionType)
		return nil, gtserror.NewErrorInternalError(err)
	}
}

// DomainPermissionDelete removes one domain block with the given ID,
// and processes side effects of removing the block asynchronously.
//
// Return values for this function are the deleted domain block, the ID of the admin
// action resulting from this call, and/or an error if something goes wrong.
func (p *Processor) DomainPermissionDelete(
	ctx context.Context,
	permissionType gtsmodel.DomainPermissionType,
	adminAcct *gtsmodel.Account,
	domainBlockID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	switch permissionType {

	// Delete explicit domain block.
	case gtsmodel.DomainPermissionBlock:
		return p.deleteDomainBlock(
			ctx,
			adminAcct,
			domainBlockID,
		)

	// Delete explicit domain allow.
	case gtsmodel.DomainPermissionAllow:
		return p.deleteDomainAllow(
			ctx,
			adminAcct,
			domainBlockID,
		)

	// You do the hokey-cokey and you turn
	// around, that's what it's all about.
	default:
		err := gtserror.Newf("unrecognized permission type %d", permissionType)
		return nil, "", gtserror.NewErrorInternalError(err)
	}
}

// DomainPermissionsImport handles the import of multiple
// domain permissions, by calling the DomainPermissionCreate
// function for each domain in the provided file. Will return
// a slice of processed domain permissions.
//
// In the case of total failure, a gtserror.WithCode will be
// returned so that the caller can respond appropriately. In
// the case of partial or total success, a MultiStatus model
// will be returned, which contains information about success
// + failure count, so that the caller can retry any failures
// as they wish.
func (p *Processor) DomainPermissionsImport(
	ctx context.Context,
	permissionType gtsmodel.DomainPermissionType,
	account *gtsmodel.Account,
	domainsF *multipart.FileHeader,
) (*apimodel.MultiStatus, gtserror.WithCode) {
	// Ensure known permission type.
	if permissionType != gtsmodel.DomainPermissionBlock &&
		permissionType != gtsmodel.DomainPermissionAllow {
		err := gtserror.Newf("unrecognized permission type %d", permissionType)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Open the provided file.
	file, err := domainsF.Open()
	if err != nil {
		err = gtserror.Newf("error opening attachment: %w", err)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
	defer file.Close()

	// Parse file as slice of domain permissions.
	apiDomainPerms := make([]*apimodel.DomainPermission, 0)
	if err := json.NewDecoder(file).Decode(&apiDomainPerms); err != nil {
		err = gtserror.Newf("error parsing attachment as domain permissions: %w", err)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	count := len(apiDomainPerms)
	if count == 0 {
		err = gtserror.New("error importing domain permissions: 0 entries provided")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Try to process each domain permission, differentiating
	// between successes and errors so that the caller can
	// try failed imports again if desired.
	multiStatusEntries := make([]apimodel.MultiStatusEntry, 0, count)
	for _, apiDomainPerm := range apiDomainPerms {
		multiStatusEntries = append(
			multiStatusEntries,
			p.importOrUpdateDomainPerm(
				ctx,
				permissionType,
				account,
				apiDomainPerm,
			),
		)
	}

	return apimodel.NewMultiStatus(multiStatusEntries), nil
}

func (p *Processor) importOrUpdateDomainPerm(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
	account *gtsmodel.Account,
	apiDomainPerm *apimodel.DomainPermission,
) apimodel.MultiStatusEntry {
	var (
		domain         = apiDomainPerm.Domain.Domain
		obfuscate      = apiDomainPerm.Obfuscate
		publicComment  = cmp.Or(apiDomainPerm.PublicComment, apiDomainPerm.Comment)
		privateComment = apiDomainPerm.PrivateComment
		subscriptionID = "" // No sub ID for imports.
	)

	// Check if this domain
	// perm already exists.
	var (
		domainPerm gtsmodel.DomainPermission
		err        error
	)
	if permType == gtsmodel.DomainPermissionBlock {
		domainPerm, err = p.state.DB.GetDomainBlock(ctx, domain)
	} else {
		domainPerm, err = p.state.DB.GetDomainAllow(ctx, domain)
	}

	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real db error.
		return apimodel.MultiStatusEntry{
			Resource: domain,
			Message:  "db error checking for existence of domain permission",
			Status:   http.StatusInternalServerError,
		}
	}

	var errWithCode gtserror.WithCode
	if !util.IsNil(domainPerm) {
		// Permission already exists, update it.
		apiDomainPerm, errWithCode = p.DomainPermissionUpdate(
			ctx,
			permType,
			domainPerm.GetID(),
			obfuscate,
			publicComment,
			privateComment,
			nil,
		)
	} else {
		// Permission didn't exist yet, create it.
		apiDomainPerm, _, errWithCode = p.DomainPermissionCreate(
			ctx,
			permType,
			account,
			domain,
			util.PtrOrZero(obfuscate),
			util.PtrOrZero(publicComment),
			util.PtrOrZero(privateComment),
			subscriptionID,
		)
	}

	if errWithCode != nil {
		return apimodel.MultiStatusEntry{
			Resource: domain,
			Message:  errWithCode.Safe(),
			Status:   errWithCode.Code(),
		}
	}

	return apimodel.MultiStatusEntry{
		Resource: apiDomainPerm,
		Message:  http.StatusText(http.StatusOK),
		Status:   http.StatusOK,
	}
}

// DomainPermissionsGet returns all existing domain
// permissions of the requested type. If export is
// true, the format will be suitable for writing out
// to an export.
func (p *Processor) DomainPermissionsGet(
	ctx context.Context,
	permissionType gtsmodel.DomainPermissionType,
	account *gtsmodel.Account,
	export bool,
) ([]*apimodel.DomainPermission, gtserror.WithCode) {
	var (
		domainPerms []gtsmodel.DomainPermission
		err         error
	)

	switch permissionType {
	case gtsmodel.DomainPermissionBlock:
		var blocks []*gtsmodel.DomainBlock

		blocks, err = p.state.DB.GetDomainBlocks(ctx)
		if err != nil {
			break
		}

		for _, block := range blocks {
			domainPerms = append(domainPerms, block)
		}

	case gtsmodel.DomainPermissionAllow:
		var allows []*gtsmodel.DomainAllow

		allows, err = p.state.DB.GetDomainAllows(ctx)
		if err != nil {
			break
		}

		for _, allow := range allows {
			domainPerms = append(domainPerms, allow)
		}

	default:
		err = errors.New("unrecognized permission type")
	}

	if err != nil {
		err := gtserror.Newf("error getting %ss: %w", permissionType.String(), err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiDomainPerms := make([]*apimodel.DomainPermission, len(domainPerms))
	for i, domainPerm := range domainPerms {
		apiDomainBlock, errWithCode := p.apiDomainPerm(ctx, domainPerm, export)
		if errWithCode != nil {
			return nil, errWithCode
		}

		apiDomainPerms[i] = apiDomainBlock
	}

	return apiDomainPerms, nil
}

// DomainPermissionGet returns one domain
// permission with the given id and type.
//
// If export is true, the format will be
// suitable for writing out to an export.
func (p *Processor) DomainPermissionGet(
	ctx context.Context,
	permissionType gtsmodel.DomainPermissionType,
	id string,
	export bool,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	var (
		domainPerm gtsmodel.DomainPermission
		err        error
	)

	switch permissionType {
	case gtsmodel.DomainPermissionBlock:
		domainPerm, err = p.state.DB.GetDomainBlockByID(ctx, id)
	case gtsmodel.DomainPermissionAllow:
		domainPerm, err = p.state.DB.GetDomainAllowByID(ctx, id)
	default:
		err = gtserror.New("unrecognized permission type")
	}

	if err != nil && errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf(
			"db error getting domain %s with id %s: %w",
			permissionType.String(), id, err,
		)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if util.IsNil(domainPerm) {
		errText := fmt.Sprintf(
			"no domain %s exists with id %s",
			permissionType.String(), id,
		)
		return nil, gtserror.NewErrorNotFound(errors.New(errText), errText)
	}

	return p.apiDomainPerm(ctx, domainPerm, export)
}
