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

package importdata

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
)

const (
	BasePath = "/v1/import"
)

var types = []string{
	"following",
	"blocks",
	"mutes",
}

var modes = []string{
	"merge",
	"overwrite",
}

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodPost, BasePath, m.ImportPOSTHandler)
}

// ImportPOSTHandler swagger:operation POST /api/v1/import importData
//
// Upload some CSV-formatted data to your account.
//
// This can be used to migrate data from a Mastodon-compatible CSV file to a GoToSocial account.
//
// Uploaded data will be processed asynchronously, and not all entries may be processed depending
// on domain blocks, user-level blocks, network availability of referenced accounts and statuses, etc.
//
//	---
//	tags:
//	- import-export
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: data
//		in: formData
//		description: The CSV data file to upload.
//		type: file
//		required: true
//	-
//		name: type
//		in: formData
//		description: >-
//			Type of entries contained in the data file:
//
//			- `following` - accounts to follow.
//			- `blocks` - accounts to block.
//			- `mutes` - accounts to mute.
//
//		type: string
//		required: true
//	-
//		name: mode
//		in: formData
//		description: >-
//			Mode to use when creating entries from the data file:
//
//			- `merge` to merge entries in file with existing entries.
//			- `overwrite` to replace existing entries with entries in file.
//		type: string
//		default: merge
//
//	security:
//	- OAuth2 Bearer:
//		- write
//
//	responses:
//		'202':
//			description: Upload accepted.
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) ImportPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWrite,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
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

	form := &apimodel.ImportRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if form.Data == nil {
		const text = "no data file provided"
		err := errors.New(text)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, text), m.processor.InstanceGetV1)
		return
	}

	if form.Type == "" {
		const text = "no type provided"
		err := errors.New(text)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, text), m.processor.InstanceGetV1)
		return
	}

	form.Type = strings.ToLower(form.Type)
	if !slices.Contains(types, form.Type) {
		text := fmt.Sprintf("type %s not recognized, valid types are: %+v", form.Type, types)
		err := errors.New(text)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, text), m.processor.InstanceGetV1)
		return
	}

	if form.Mode != "" {
		form.Mode = strings.ToLower(form.Mode)
		if !slices.Contains(modes, form.Mode) {
			text := fmt.Sprintf("mode %s not recognized, valid modes are: %+v", form.Mode, modes)
			err := errors.New(text)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, text), m.processor.InstanceGetV1)
			return
		}
	}
	overwrite := form.Mode == "overwrite"

	// Trigger the import.
	errWithCode = m.processor.Account().ImportData(
		c.Request.Context(),
		authed.Account,
		form.Data,
		form.Type,
		overwrite,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusAccepted, gin.H{"status": "accepted"})
}
