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
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type singleDomainPermCreate func(
	context.Context,
	gtsmodel.DomainPermissionType, // block/allow
	*gtsmodel.Account, // admin account
	string, // domain
	bool, // obfuscate
	string, // publicComment
	string, // privateComment
	string, // subscriptionID
) (*apimodel.DomainPermission, string, gtserror.WithCode)

type multiDomainPermCreate func(
	context.Context,
	gtsmodel.DomainPermissionType, // block/allow
	*gtsmodel.Account, // admin account
	*multipart.FileHeader, // domains
) (*apimodel.MultiStatus, gtserror.WithCode)

// createDomainPemissions either creates a single domain
// permission entry (block/allow) or imports multiple domain
// permission entries (multiple blocks, multiple allows)
// using the given functions.
//
// Handling the creation of both types of permissions in
// one function in this way reduces code duplication.
func (m *Module) createDomainPermissions(
	c *gin.Context,
	permType gtsmodel.DomainPermissionType,
	single singleDomainPermCreate,
	multi multiDomainPermCreate,
) {
	// Scope differs based on permType.
	var requireScope apiutil.Scope
	if permType == gtsmodel.DomainPermissionBlock {
		requireScope = apiutil.ScopeAdminWriteDomainBlocks
	} else {
		requireScope = apiutil.ScopeAdminWriteDomainAllows
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		requireScope,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	importing, errWithCode := apiutil.ParseDomainPermissionImport(c.Query(apiutil.DomainPermissionImportKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse + validate form.
	form := new(apimodel.DomainPermissionRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	var err error
	if importing && form.Domains.Size == 0 {
		err = errors.New("import was specified but list of domains is empty")
	} else if !importing && form.Domain == "" {
		err = errors.New("no domain provided")
	}

	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if !importing {
		// Single domain permission creation.
		perm, _, errWithCode := single(
			c.Request.Context(),
			permType,
			authed.Account,
			form.Domain,
			util.PtrOrZero(form.Obfuscate),
			util.PtrOrZero(form.PublicComment),
			util.PtrOrZero(form.PrivateComment),
			"", // No sub ID for single perm creation.
		)

		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

		apiutil.JSON(c, http.StatusOK, perm)
		return
	}

	// We're importing multiple domain permissions,
	// so we're looking at a multi-status response.
	multiStatus, errWithCode := multi(
		c.Request.Context(),
		permType,
		authed.Account,
		form.Domains, // Pass the file through.
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// TODO: Return 207 and multiStatus data nicely
	//       when supported by the admin panel.
	if multiStatus.Metadata.Failure != 0 {
		failures := make(map[string]any, multiStatus.Metadata.Failure)
		for _, entry := range multiStatus.Data {
			failures[entry.Resource.(string)] = entry.Message
		}

		err := fmt.Errorf("one or more errors importing domain %ss: %+v", permType.String(), failures)
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Success, return slice of newly-created domain perms.
	domainPerms := make([]any, 0, multiStatus.Metadata.Success)
	for _, entry := range multiStatus.Data {
		domainPerms = append(domainPerms, entry.Resource)
	}

	apiutil.JSON(c, http.StatusOK, domainPerms)
}

func (m *Module) updateDomainPermission(
	c *gin.Context,
	permType gtsmodel.DomainPermissionType,
) {
	// Scope differs based on permType.
	var requireScope apiutil.Scope
	if permType == gtsmodel.DomainPermissionBlock {
		requireScope = apiutil.ScopeAdminWriteDomainBlocks
	} else {
		requireScope = apiutil.ScopeAdminWriteDomainAllows
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		requireScope,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	permID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse + validate form.
	form := new(apimodel.DomainPermissionRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if form.Obfuscate == nil &&
		form.PrivateComment == nil &&
		form.PublicComment == nil {
		const errText = "empty form submitted"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	perm, errWithCode := m.processor.Admin().DomainPermissionUpdate(
		c.Request.Context(),
		permType,
		permID,
		form.Obfuscate,
		form.PublicComment,
		form.PrivateComment,
		nil, // Can't update perm sub ID this way yet.
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, perm)
}

// deleteDomainPermission deletes a single domain permission (block or allow).
func (m *Module) deleteDomainPermission(
	c *gin.Context,
	permType gtsmodel.DomainPermissionType, // block/allow
) {
	// Scope differs based on permType.
	var requireScope apiutil.Scope
	if permType == gtsmodel.DomainPermissionBlock {
		requireScope = apiutil.ScopeAdminWriteDomainBlocks
	} else {
		requireScope = apiutil.ScopeAdminWriteDomainAllows
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		requireScope,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	domainPermID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	domainPerm, _, errWithCode := m.processor.Admin().DomainPermissionDelete(
		c.Request.Context(),
		permType,
		authed.Account,
		domainPermID,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, domainPerm)
}

// getDomainPermission gets a single domain permission (block or allow).
func (m *Module) getDomainPermission(
	c *gin.Context,
	permType gtsmodel.DomainPermissionType,
) {
	// Scope differs based on permType.
	var requireScope apiutil.Scope
	if permType == gtsmodel.DomainPermissionBlock {
		requireScope = apiutil.ScopeAdminReadDomainBlocks
	} else {
		requireScope = apiutil.ScopeAdminReadDomainAllows
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		requireScope,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	domainPermID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	export, errWithCode := apiutil.ParseDomainPermissionExport(c.Query(apiutil.DomainPermissionExportKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	domainPerm, errWithCode := m.processor.Admin().DomainPermissionGet(
		c.Request.Context(),
		permType,
		domainPermID,
		export,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, domainPerm)
}

// getDomainPermissions gets all domain permissions of the given type (block, allow).
func (m *Module) getDomainPermissions(
	c *gin.Context,
	permType gtsmodel.DomainPermissionType,
) {
	// Scope differs based on permType.
	var requireScope apiutil.Scope
	if permType == gtsmodel.DomainPermissionBlock {
		requireScope = apiutil.ScopeAdminReadDomainBlocks
	} else {
		requireScope = apiutil.ScopeAdminReadDomainAllows
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		requireScope,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	export, errWithCode := apiutil.ParseDomainPermissionExport(c.Query(apiutil.DomainPermissionExportKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	domainPerm, errWithCode := m.processor.Admin().DomainPermissionsGet(
		c.Request.Context(),
		permType,
		authed.Account,
		export,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, domainPerm)
}

// parseDomainPermissionType is a util function to parse i
// to a DomainPermissionType, or return a suitable error.
func parseDomainPermissionType(i string) (
	permType gtsmodel.DomainPermissionType,
	errWithCode gtserror.WithCode,
) {
	if i == "" {
		const errText = "permission_type not set, must be one of block or allow"
		errWithCode = gtserror.NewErrorBadRequest(errors.New(errText), errText)
		return
	}

	permType = gtsmodel.ParseDomainPermissionType(i)
	if permType == gtsmodel.DomainPermissionUnknown {
		var errText = fmt.Sprintf("permission_type %s not recognized, must be one of block or allow", i)
		errWithCode = gtserror.NewErrorBadRequest(errors.New(errText), errText)
	}

	return
}

// parseDomainPermSubContentType is a util function to parse i
// to a DomainPermSubContentType, or return a suitable error.
func parseDomainPermSubContentType(i string) (
	contentType gtsmodel.DomainPermSubContentType,
	errWithCode gtserror.WithCode,
) {
	if i == "" {
		const errText = "content_type not set, must be one of text/csv, text/plain or application/json"
		errWithCode = gtserror.NewErrorBadRequest(errors.New(errText), errText)
		return
	}

	contentType = gtsmodel.NewDomainPermSubContentType(i)
	if contentType == gtsmodel.DomainPermSubContentTypeUnknown {
		var errText = fmt.Sprintf("content_type %s not recognized, must be one of text/csv, text/plain or application/json", i)
		errWithCode = gtserror.NewErrorBadRequest(errors.New(errText), errText)
	}

	return
}
