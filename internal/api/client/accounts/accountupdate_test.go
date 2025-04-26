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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AccountUpdateTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountUpdateTestSuite) updateAccountFromForm(data map[string][]string, expectedHTTPStatus int, expectedBody string) (*apimodel.Account, error) {
	form := url.Values{}
	for key, val := range data {
		if form.Has(key) {
			form[key] = append(form[key], val...)
		} else {
			form[key] = val
		}
	}
	return suite.updateAccount([]byte(form.Encode()), "application/x-www-form-urlencoded", expectedHTTPStatus, expectedBody)
}

func (suite *AccountUpdateTestSuite) updateAccountFromFormData(data map[string][]string, expectedHTTPStatus int, expectedBody string) (*apimodel.Account, error) {
	requestBody, w, err := testrig.CreateMultipartFormData(nil, data)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return suite.updateAccount(requestBody.Bytes(), w.FormDataContentType(), expectedHTTPStatus, expectedBody)
}

func (suite *AccountUpdateTestSuite) updateAccountFromFormDataWithFile(fieldName string, filePath string, data map[string][]string, expectedHTTPStatus int, expectedBody string) (*apimodel.Account, error) {
	requestBody, w, err := testrig.CreateMultipartFormData(testrig.FileToDataF(fieldName, filePath), data)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return suite.updateAccount(requestBody.Bytes(), w.FormDataContentType(), expectedHTTPStatus, expectedBody)
}

func (suite *AccountUpdateTestSuite) updateAccountFromJSON(data string, expectedHTTPStatus int, expectedBody string) (*apimodel.Account, error) {
	return suite.updateAccount([]byte(data), "application/json", expectedHTTPStatus, expectedBody)
}

func (suite *AccountUpdateTestSuite) updateAccount(
	bodyBytes []byte,
	contentType string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Account, error) {
	// Initialize http test context.
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdatePath, contentType)

	// Trigger the handler.
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// Read the result.
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// Check expected code + body.
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// If we got an expected body, return early.
	if expectedBody != "" && string(b) != expectedBody {
		errs.Appendf("expected %s got %s", expectedBody, string(b))
	}

	if err := errs.Combine(); err != nil {
		return nil, fmt.Errorf("%v (body %s)", err, string(b))
	}

	// Return account response.
	resp := &apimodel.Account{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountBasicForm() {
	data := map[string][]string{
		"note":                        {"this is my new bio read it and weep"},
		"fields_attributes[0][name]":  {"pronouns"},
		"fields_attributes[0][value]": {"they/them"},
		"fields_attributes[1][name]":  {"Website"},
		"fields_attributes[1][value]": {"https://example.com"},
	}

	apimodelAccount, err := suite.updateAccountFromForm(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
	suite.Equal("this is my new bio read it and weep", apimodelAccount.Source.Note)

	if l := len(apimodelAccount.Fields); l != 2 {
		suite.FailNow("", "expected %d fields, got %d", 2, l)
	}
	suite.Equal(`pronouns`, apimodelAccount.Fields[0].Name)
	suite.Equal(`they/them`, apimodelAccount.Fields[0].Value)
	suite.Equal(`Website`, apimodelAccount.Fields[1].Name)
	suite.Equal(`<a href="https://example.com" rel="nofollow noreferrer noopener" target="_blank">https://example.com</a>`, apimodelAccount.Fields[1].Value)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountBasicFormData() {
	data := map[string][]string{
		"note":                        {"this is my new bio read it and weep"},
		"fields_attributes[0][name]":  {"pronouns"},
		"fields_attributes[0][value]": {"they/them"},
		"fields_attributes[1][name]":  {"Website"},
		"fields_attributes[1][value]": {"https://example.com"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
	suite.Equal("this is my new bio read it and weep", apimodelAccount.Source.Note)

	if l := len(apimodelAccount.Fields); l != 2 {
		suite.FailNow("", "expected %d fields, got %d", 2, l)
	}
	suite.Equal(`pronouns`, apimodelAccount.Fields[0].Name)
	suite.Equal(`they/them`, apimodelAccount.Fields[0].Value)
	suite.Equal(`Website`, apimodelAccount.Fields[1].Name)
	suite.Equal(`<a href="https://example.com" rel="nofollow noreferrer noopener" target="_blank">https://example.com</a>`, apimodelAccount.Fields[1].Value)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountBasicJSON() {
	data := `
{
  "note": "this is my new bio read it and weep",
  "fields_attributes": {
    "0": {
      "name": "pronouns",
      "value": "they/them"
    },
    "1": {
      "name": "Website",
      "value": "https://example.com"
    }
  }
}
`

	apimodelAccount, err := suite.updateAccountFromJSON(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
	suite.Equal("this is my new bio read it and weep", apimodelAccount.Source.Note)

	if l := len(apimodelAccount.Fields); l != 2 {
		suite.FailNow("", "expected %d fields, got %d", 2, l)
	}
	suite.Equal(`pronouns`, apimodelAccount.Fields[0].Name)
	suite.Equal(`they/them`, apimodelAccount.Fields[0].Value)
	suite.Equal(`Website`, apimodelAccount.Fields[1].Name)
	suite.Equal(`<a href="https://example.com" rel="nofollow noreferrer noopener" target="_blank">https://example.com</a>`, apimodelAccount.Fields[1].Value)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountLockForm() {
	data := map[string][]string{
		"locked": {"true"},
	}

	apimodelAccount, err := suite.updateAccountFromForm(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountLockFormData() {
	data := map[string][]string{
		"locked": {"true"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountLockJSON() {
	data := `
{
  "locked": true
}`

	apimodelAccount, err := suite.updateAccountFromJSON(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountUnlockForm() {
	data := map[string][]string{
		"locked": {"false"},
	}

	apimodelAccount, err := suite.updateAccountFromForm(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountUnlockFormData() {
	data := map[string][]string{
		"locked": {"false"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountUnlockJSON() {
	data := `
{
  "locked": false
}`

	apimodelAccount, err := suite.updateAccountFromJSON(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountCache() {
	// Get the account first to make sure it's in the database
	// cache. When the account is updated via the PATCH handler,
	// it should invalidate the cache and return the new version.
	if _, err := suite.db.GetAccountByID(context.Background(), suite.testAccounts["local_account_1"].ID); err != nil {
		suite.FailNow(err.Error())
	}

	data := map[string][]string{
		"note": {"this is my new bio read it and weep"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountDiscoverableForm() {
	data := map[string][]string{
		"discoverable": {"false"},
	}

	apimodelAccount, err := suite.updateAccountFromForm(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(apimodelAccount.Discoverable)

	// Check the account in the database too.
	dbZork, err := suite.db.GetAccountByID(context.Background(), apimodelAccount.ID)
	suite.NoError(err)
	suite.False(*dbZork.Discoverable)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountDiscoverableFormData() {
	data := map[string][]string{
		"discoverable": {"false"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(apimodelAccount.Discoverable)

	// Check the account in the database too.
	dbZork, err := suite.db.GetAccountByID(context.Background(), apimodelAccount.ID)
	suite.NoError(err)
	suite.False(*dbZork.Discoverable)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountDiscoverableJSON() {
	data := `
{
  "discoverable": false
}`

	apimodelAccount, err := suite.updateAccountFromJSON(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(apimodelAccount.Discoverable)

	// Check the account in the database too.
	dbZork, err := suite.db.GetAccountByID(context.Background(), apimodelAccount.ID)
	suite.NoError(err)
	suite.False(*dbZork.Discoverable)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountWithImageFormData() {
	data := map[string][]string{
		"display_name": {"updated zork display name!!!"},
		"note":         {""},
		"locked":       {"true"},
	}

	apimodelAccount, err := suite.updateAccountFromFormDataWithFile("header", "../../../../testrig/media/test-jpeg.jpg", data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(data["display_name"][0], apimodelAccount.DisplayName)
	suite.True(apimodelAccount.Locked)
	suite.Empty(apimodelAccount.Note)
	suite.Empty(apimodelAccount.Source.Note)
	suite.NotEmpty(apimodelAccount.Header)
	suite.NotEmpty(apimodelAccount.HeaderStatic)

	// Can't predict IDs generated for new media
	// so just ensure it's different than before.
	suite.NotEqual("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg", apimodelAccount.Header)
	suite.NotEqual("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp", apimodelAccount.HeaderStatic)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountEmptyForm() {
	data := make(map[string][]string)

	_, err := suite.updateAccountFromForm(data, http.StatusBadRequest, `{"error":"Bad Request: empty form submitted"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountEmptyFormData() {
	data := make(map[string][]string)

	_, err := suite.updateAccountFromFormData(data, http.StatusBadRequest, `{"error":"Bad Request: empty form submitted"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountSourceForm() {
	data := map[string][]string{
		"source[privacy]":   {string(apimodel.VisibilityPrivate)},
		"source[language]":  {"de"},
		"source[sensitive]": {"true"},
		"locked":            {"true"},
	}

	apimodelAccount, err := suite.updateAccountFromForm(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(data["source[language]"][0], apimodelAccount.Source.Language)
	suite.EqualValues(apimodel.VisibilityPrivate, apimodelAccount.Source.Privacy)
	suite.True(apimodelAccount.Source.Sensitive)
	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountSourceFormData() {
	data := map[string][]string{
		"source[privacy]":   {string(apimodel.VisibilityPrivate)},
		"source[language]":  {"de"},
		"source[sensitive]": {"true"},
		"locked":            {"true"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(data["source[language]"][0], apimodelAccount.Source.Language)
	suite.EqualValues(apimodel.VisibilityPrivate, apimodelAccount.Source.Privacy)
	suite.True(apimodelAccount.Source.Sensitive)
	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountSourceJSON() {
	data := `
{
  "source": {
    "privacy": "private",
    "language": "de",
    "sensitive": true
  },
  "locked": true
}
`

	apimodelAccount, err := suite.updateAccountFromJSON(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("de", apimodelAccount.Source.Language)
	suite.EqualValues(apimodel.VisibilityPrivate, apimodelAccount.Source.Privacy)
	suite.True(apimodelAccount.Source.Sensitive)
	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestUpdateAccountSourceBadContentTypeFormData() {
	data := map[string][]string{
		"source[status_content_type]": {"text/markdown"},
	}

	apimodelAccount, err := suite.updateAccountFromFormData(data, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(data["source[status_content_type]"][0], apimodelAccount.Source.StatusContentType)

	// Check the account in the database too.
	dbAccount, err := suite.db.GetAccountByID(context.Background(), suite.testAccounts["local_account_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(data["source[status_content_type]"][0], dbAccount.Settings.StatusContentType)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerUpdateStatusContentTypeBad() {
	data := map[string][]string{
		"source[status_content_type]": {"peepeepoopoo"},
	}

	_, err := suite.updateAccountFromFormData(data, http.StatusBadRequest, `{"error":"Bad Request: status content type 'peepeepoopoo' was not recognized, valid options are 'text/plain', 'text/markdown'"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
