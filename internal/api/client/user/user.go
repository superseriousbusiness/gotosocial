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

package user

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

const (
	BasePath               = "/v1/user"
	PasswordChangePath     = BasePath + "/password_change"
	EmailChangePath        = BasePath + "/email_change"
	TwoFactorPath          = BasePath + "/2fa"
	TwoFactorQRCodePngPath = TwoFactorPath + "/qr.png"
	TwoFactorQRCodeURIPath = TwoFactorPath + "/qruri"
	TwoFactorEnablePath    = TwoFactorPath + "/enable"
	TwoFactorDisablePath   = TwoFactorPath + "/disable"
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.UserGETHandler)
	attachHandler(http.MethodPost, PasswordChangePath, m.PasswordChangePOSTHandler)
	attachHandler(http.MethodPost, EmailChangePath, m.EmailChangePOSTHandler)
	attachHandler(http.MethodGet, TwoFactorQRCodePngPath, m.TwoFactorQRCodePngGETHandler)
	attachHandler(http.MethodGet, TwoFactorQRCodeURIPath, m.TwoFactorQRCodeURIGETHandler)
	attachHandler(http.MethodPost, TwoFactorEnablePath, m.TwoFactorEnablePOSTHandler)
	attachHandler(http.MethodPost, TwoFactorDisablePath, m.TwoFactorDisablePOSTHandler)
}
