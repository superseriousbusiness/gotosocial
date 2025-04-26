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

package polls_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/polls"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type PollCreateTestSuite struct {
	PollsStandardTestSuite
}

func (suite *PollCreateTestSuite) voteInPoll(
	pollID string,
	contentType string,
	body io.Reader,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Poll, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["admin_account"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["admin_account"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["admin_account"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodPost, config.GetProtocol()+"://"+config.GetHost()+"/api/"+polls.BasePath+"/"+pollID, body)
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Header.Set("content-type", contentType)

	ctx.AddParam("id", pollID)

	// trigger the handler
	suite.pollsModule.PollVotePOSTHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.Poll{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *PollCreateTestSuite) formVoteInPoll(
	pollID string,
	choices []int,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Poll, error) {
	choicesStrs := make([]string, 0, len(choices))
	for _, choice := range choices {
		choicesStrs = append(choicesStrs, strconv.Itoa(choice))
	}

	body, w, err := testrig.CreateMultipartFormData(nil, map[string][]string{
		"choices[]": choicesStrs,
	})

	if err != nil {
		suite.FailNow(err.Error())
	}

	b := body.Bytes()
	suite.T().Log(string(b))

	return suite.voteInPoll(
		pollID,
		w.FormDataContentType(),
		bytes.NewReader(b),
		expectedHTTPStatus,
		expectedBody,
	)
}

func (suite *PollCreateTestSuite) jsonVoteInPoll(
	pollID string,
	choices []interface{},
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Poll, error) {
	form := apimodel.PollVoteRequest{ChoicesI: choices}

	b, err := json.Marshal(&form)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.T().Log(string(b))

	return suite.voteInPoll(
		pollID,
		"application/json",
		bytes.NewReader(b),
		expectedHTTPStatus,
		expectedBody,
	)
}

func (suite *PollCreateTestSuite) TestPollVoteForm() {
	targetPoll := suite.testPolls["local_account_1_status_6_poll"]

	poll, err := suite.formVoteInPoll(targetPoll.ID, []int{2}, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(poll)
}

func (suite *PollCreateTestSuite) TestPollVoteJSONInt() {
	targetPoll := suite.testPolls["local_account_1_status_6_poll"]

	poll, err := suite.jsonVoteInPoll(targetPoll.ID, []interface{}{2}, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(poll)
}

func (suite *PollCreateTestSuite) TestPollVoteJSONStr() {
	targetPoll := suite.testPolls["local_account_1_status_6_poll"]

	poll, err := suite.jsonVoteInPoll(targetPoll.ID, []interface{}{"2"}, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(poll)
}

func TestPollCreateTestSuite(t *testing.T) {
	suite.Run(t, &PollCreateTestSuite{})
}
