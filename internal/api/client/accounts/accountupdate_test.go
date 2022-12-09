/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package accounts_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/accounts"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountUpdateTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandler() {
	// set up the request
	// we're updating the note of zork
	newBio := "this is my new bio read it and weep"
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"note": newBio,
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
	suite.Equal(newBio, apimodelAccount.Source.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerUnlockLock() {
	// set up the first request
	requestBody1, w1, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"locked": "false",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes1 := requestBody1.Bytes()
	recorder1 := httptest.NewRecorder()
	ctx1 := suite.newContext(recorder1, http.MethodPatch, bodyBytes1, accounts.UpdateCredentialsPath, w1.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx1)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder1.Code)

	// 2. we should have no error message in the result body
	result1 := recorder1.Result()
	defer result1.Body.Close()

	// check the response
	b1, err := ioutil.ReadAll(result1.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount1 := &apimodel.Account{}
	err = json.Unmarshal(b1, apimodelAccount1)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.False(apimodelAccount1.Locked)

	// set up the first request
	requestBody2, w2, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"locked": "true",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes2 := requestBody2.Bytes()
	recorder2 := httptest.NewRecorder()
	ctx2 := suite.newContext(recorder2, http.MethodPatch, bodyBytes2, accounts.UpdateCredentialsPath, w2.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx2)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder1.Code)

	// 2. we should have no error message in the result body
	result2 := recorder2.Result()
	defer result2.Body.Close()

	// check the response
	b2, err := ioutil.ReadAll(result2.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount2 := &apimodel.Account{}
	err = json.Unmarshal(b2, apimodelAccount2)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.True(apimodelAccount2.Locked)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerGetAccountFirst() {
	// get the account first to make sure it's in the database cache -- when the account is updated via
	// the PATCH handler, it should invalidate the cache and not return the old version
	_, err := suite.db.GetAccountByID(context.Background(), suite.testAccounts["local_account_1"].ID)
	suite.NoError(err)

	// set up the request
	// we're updating the note of zork
	newBio := "this is my new bio read it and weep"
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"note": newBio,
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
	suite.Equal(newBio, apimodelAccount.Source.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerTwoFields() {
	// set up the request
	// we're updating the note of zork, and setting locked to true
	newBio := "this is my new bio read it and weep :rainbow:"
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"note":   newBio,
			"locked": "true",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("<p>this is my new bio read it and weep :rainbow:</p>", apimodelAccount.Note)
	suite.Equal(newBio, apimodelAccount.Source.Note)
	suite.True(apimodelAccount.Locked)
	suite.NotEmpty(apimodelAccount.Emojis)
	suite.Equal(apimodelAccount.Emojis[0].Shortcode, "rainbow")

	// check the account in the database
	dbZork, err := suite.db.GetAccountByID(context.Background(), apimodelAccount.ID)
	suite.NoError(err)
	suite.Equal(newBio, dbZork.NoteRaw)
	suite.Equal("<p>this is my new bio read it and weep :rainbow:</p>", dbZork.Note)
	suite.True(*dbZork.Locked)
	suite.NotEmpty(dbZork.EmojiIDs)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerWithMedia() {
	// set up the request
	// we're updating the header image, the display name, and the locked status of zork
	// we're removing the note/bio
	requestBody, w, err := testrig.CreateMultipartFormData(
		"header", "../../../../testrig/media/test-jpeg.jpg",
		map[string]string{
			"display_name": "updated zork display name!!!",
			"note":         "",
			"locked":       "true",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("updated zork display name!!!", apimodelAccount.DisplayName)
	suite.True(apimodelAccount.Locked)
	suite.Empty(apimodelAccount.Note)
	suite.Empty(apimodelAccount.Source.Note)

	// header values...
	// should be set
	suite.NotEmpty(apimodelAccount.Header)
	suite.NotEmpty(apimodelAccount.HeaderStatic)

	// should be different from the values set before
	suite.NotEqual("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg", apimodelAccount.Header)
	suite.NotEqual("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg", apimodelAccount.HeaderStatic)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerEmptyForm() {
	// set up the request
	bodyBytes := []byte{}
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, "")

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Equal(`{"error":"Bad Request: empty form submitted"}`, string(b))
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerUpdateSource() {
	// set up the request
	// we're updating the language of zork
	newLanguage := "de"
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"source[privacy]":   string(apimodel.VisibilityPrivate),
			"source[language]":  "de",
			"source[sensitive]": "true",
			"locked":            "true",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal(newLanguage, apimodelAccount.Source.Language)
	suite.EqualValues(apimodel.VisibilityPrivate, apimodelAccount.Source.Privacy)
	suite.True(apimodelAccount.Source.Sensitive)
	suite.True(apimodelAccount.Locked)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerUpdateStatusFormatOK() {
	// set up the request
	// we're updating the language of zork
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"source[status_format]": "markdown",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("markdown", apimodelAccount.Source.StatusFormat)

	dbAccount, err := suite.db.GetAccountByID(context.Background(), suite.testAccounts["local_account_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(dbAccount.StatusFormat, "markdown")
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerUpdateStatusFormatBad() {
	// set up the request
	// we're updating the language of zork
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"source[status_format]": "peepeepoopoo",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, accounts.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountsModule.AccountUpdateCredentialsPATCHHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: status format 'peepeepoopoo' was not recognized, valid options are 'plain', 'markdown'"}`, string(b))
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
