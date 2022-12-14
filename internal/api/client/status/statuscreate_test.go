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

package status_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusCreateTestSuite struct {
	StatusStandardTestSuite
}

const (
	statusWithLinksAndTags = "#test alright, should be able to post #links with fragments in them now, let's see........\n\nhttps://docs.gotosocial.org/en/latest/user_guide/posts/#links\n\n#gotosocial\n\n(tobi remember to pull the docker image challenge)"
	statusMarkdown         = "# Title\n\n## Smaller title\n\nThis is a post written in [markdown](https://www.markdownguide.org/)\n\n<img src=\"https://d33wubrfki0l68.cloudfront.net/f1f475a6fda1c2c4be4cac04033db5c3293032b4/513a4/assets/images/markdown-mark-white.svg\"/>"
	statusMarkdownExpected = "<h1>Title</h1><h2>Smaller title</h2><p>This is a post written in <a href=\"https://www.markdownguide.org/\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">markdown</a></p><img src=\"https://d33wubrfki0l68.cloudfront.net/f1f475a6fda1c2c4be4cac04033db5c3293032b4/513a4/assets/images/markdown-mark-white.svg\" crossorigin=\"anonymous\">"
)

// Post a new status with some custom visibility settings
func (suite *StatusCreateTestSuite) TestPostNewStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status":       {"this is a brand new status! #helloworld"},
		"spoiler_text": {"hello hello"},
		"sensitive":    {"true"},
		"visibility":   {string(model.VisibilityMutualsOnly)},
		"likeable":     {"false"},
		"replyable":    {"false"},
		"federated":    {"false"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.Equal("hello hello", statusReply.SpoilerText)
	suite.Equal("<p>this is a brand new status! <a href=\"http://localhost:8080/tags/helloworld\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>helloworld</span></a></p>", statusReply.Content)
	suite.True(statusReply.Sensitive)
	suite.Equal(model.VisibilityPrivate, statusReply.Visibility) // even though we set this status to mutuals only, it should serialize to private, because the mastodon api has no idea about mutuals_only
	suite.Len(statusReply.Tags, 1)
	suite.Equal(model.Tag{
		Name: "helloworld",
		URL:  "http://localhost:8080/tags/helloworld",
	}, statusReply.Tags[0])

	gtsTag := &gtsmodel.Tag{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "name", Value: "helloworld"}}, gtsTag)
	suite.NoError(err)
	suite.Equal(statusReply.Account.ID, gtsTag.FirstSeenFromAccountID)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusMarkdown() {
	// set default post language of account 1 to markdown
	testAccount := suite.testAccounts["local_account_1"]
	testAccount.StatusFormat = "markdown"
	a := testAccount

	err := suite.db.UpdateAccount(context.Background(), a)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(a.StatusFormat, "markdown")

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, a)

	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil)
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status":     {statusMarkdown},
		"visibility": {string(model.VisibilityPublic)},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.Equal(statusMarkdownExpected, statusReply.Content)
}

// mention an account that is not yet known to the instance -- it should be looked up and put in the db
func (suite *StatusCreateTestSuite) TestMentionUnknownAccount() {
	// first remove remote account 1 from the database so it gets looked up again
	remoteAccount := suite.testAccounts["remote_account_1"]
	err := suite.db.DeleteAccount(context.Background(), remoteAccount.ID)
	suite.NoError(err)

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status":     {"hello @brand_new_person@unknown-instance.com"},
		"visibility": {string(model.VisibilityPublic)},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	// if the status is properly formatted, that means the account has been put in the db
	suite.Equal(`<p>hello <span class="h-card"><a href="https://unknown-instance.com/@brand_new_person" class="u-url mention" rel="nofollow noreferrer noopener" target="_blank">@<span>brand_new_person</span></a></span></p>`, statusReply.Content)
	suite.Equal(model.VisibilityPublic, statusReply.Visibility)
}

func (suite *StatusCreateTestSuite) TestPostAnotherNewStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status": {statusWithLinksAndTags},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.Equal("<p><a href=\"http://localhost:8080/tags/test\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>test</span></a> alright, should be able to post <a href=\"http://localhost:8080/tags/links\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>links</span></a> with fragments in them now, let&#39;s see........<br/><br/><a href=\"https://docs.gotosocial.org/en/latest/user_guide/posts/#links\" rel=\"noopener nofollow noreferrer\" target=\"_blank\">docs.gotosocial.org/en/latest/user_guide/posts/#links</a><br/><br/><a href=\"http://localhost:8080/tags/gotosocial\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>gotosocial</span></a><br/><br/>(tobi remember to pull the docker image challenge)</p>", statusReply.Content)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusWithEmoji() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status": {"here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: \n here's an emoji that isn't in the db: :test_emoji: "},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.Equal("", statusReply.SpoilerText)
	suite.Equal("<p>here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: <br/> here&#39;s an emoji that isn&#39;t in the db: :test_emoji:</p>", statusReply.Content)

	suite.Len(statusReply.Emojis, 1)
	apiEmoji := statusReply.Emojis[0]
	gtsEmoji := testrig.NewTestEmojis()["rainbow"]

	suite.Equal(gtsEmoji.Shortcode, apiEmoji.Shortcode)
	suite.Equal(gtsEmoji.ImageURL, apiEmoji.URL)
	suite.Equal(gtsEmoji.ImageStaticURL, apiEmoji.StaticURL)
}

// Try to reply to a status that doesn't exist
func (suite *StatusCreateTestSuite) TestReplyToNonexistentStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status":         {"this is a reply to a status that doesn't exist"},
		"spoiler_text":   {"don't open cuz it won't work"},
		"in_reply_to_id": {"3759e7ef-8ee1-4c0c-86f6-8b70b9ad3d50"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Equal(`{"error":"Bad Request: status with id 3759e7ef-8ee1-4c0c-86f6-8b70b9ad3d50 not replyable because it doesn't exist"}`, string(b))
}

// Post a reply to the status of a local user that allows replies.
func (suite *StatusCreateTestSuite) TestReplyToLocalStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status":         {fmt.Sprintf("hello @%s this reply should work!", testrig.NewTestAccounts()["local_account_2"].Username)},
		"in_reply_to_id": {testrig.NewTestStatuses()["local_account_2_status_1"].ID},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.Equal("", statusReply.SpoilerText)
	suite.Equal(fmt.Sprintf("<p>hello <span class=\"h-card\"><a href=\"http://localhost:8080/@%s\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>%s</span></a></span> this reply should work!</p>", testrig.NewTestAccounts()["local_account_2"].Username, testrig.NewTestAccounts()["local_account_2"].Username), statusReply.Content)
	suite.False(statusReply.Sensitive)
	suite.Equal(model.VisibilityPublic, statusReply.Visibility)
	suite.Equal(testrig.NewTestStatuses()["local_account_2_status_1"].ID, *statusReply.InReplyToID)
	suite.Equal(testrig.NewTestAccounts()["local_account_2"].ID, *statusReply.InReplyToAccountID)
	suite.Len(statusReply.Mentions, 1)
}

// Take a media file which is currently not associated with a status, and attach it to a new status.
func (suite *StatusCreateTestSuite) TestAttachNewMediaSuccess() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	attachment := suite.testAttachments["local_account_1_unattached_1"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"status":      {"here's an image attachment"},
		"media_ids[]": {attachment.ID},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusResponse := &model.Status{}
	err = json.Unmarshal(b, statusResponse)
	suite.NoError(err)

	suite.Equal("", statusResponse.SpoilerText)
	suite.Equal("<p>here&#39;s an image attachment</p>", statusResponse.Content)
	suite.False(statusResponse.Sensitive)
	suite.Equal(model.VisibilityPublic, statusResponse.Visibility)

	// there should be one media attachment
	suite.Len(statusResponse.MediaAttachments, 1)

	// get the updated media attachment from the database
	gtsAttachment, err := suite.db.GetAttachmentByID(context.Background(), statusResponse.MediaAttachments[0].ID)
	suite.NoError(err)

	// convert it to a api attachment
	gtsAttachmentAsapi, err := suite.tc.AttachmentToAPIAttachment(context.Background(), gtsAttachment)
	suite.NoError(err)

	// compare it with what we have now
	suite.EqualValues(statusResponse.MediaAttachments[0], gtsAttachmentAsapi)

	// the status id of the attachment should now be set to the id of the status we just created
	suite.Equal(statusResponse.ID, gtsAttachment.StatusID)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}
