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

package accounts_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AccountDeleteTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountDeleteTestSuite) TestAccountDeletePOSTHandler() {
	// set up the request
	// we're deleting zork
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"password": {"password"},
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, accounts.DeletePath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountDeletePOSTHandler(ctx)

	// 1. we should have Accepted because our request was valid
	suite.Equal(http.StatusAccepted, recorder.Code)
}

func (suite *AccountDeleteTestSuite) TestAccountDeletePOSTHandlerWrongPassword() {
	// set up the request
	// we're deleting zork
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"password": {"aaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, accounts.DeletePath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountDeletePOSTHandler(ctx)

	// 1. we should have Forbidden because we supplied the wrong password
	suite.Equal(http.StatusForbidden, recorder.Code)
}

func (suite *AccountDeleteTestSuite) TestAccountDeletePOSTHandlerNoPassword() {
	// set up the request
	// we're deleting zork
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, accounts.DeletePath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountDeletePOSTHandler(ctx)

	// 1. we should have StatusBadRequest because our request was invalid
	suite.Equal(http.StatusBadRequest, recorder.Code)
}

func TestAccountDeleteTestSuite(t *testing.T) {
	suite.Run(t, new(AccountDeleteTestSuite))
}
