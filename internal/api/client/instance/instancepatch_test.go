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

package instance_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InstancePatchTestSuite struct {
	InstanceStandardTestSuite
}

func (suite *InstancePatchTestSuite) TestInstancePatch1() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"title":            "Example Instance",
			"contact_username": "admin",
			"contact_email":    "someone@example.org",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","account_domain":"localhost:8080","title":"Example Instance","description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","short_description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","email":"someone@example.org","version":"0.0.0-testrig","registrations":true,"approval_required":true,"invites_enabled":false,"configuration":{"statuses":{"max_characters":5000,"max_media_attachments":6,"characters_reserved_per_url":25},"media_attachments":{"supported_mime_types":["image/jpeg","image/gif","image/png","image/webp","video/mp4"],"image_size_limit":10485760,"image_matrix_limit":16777216,"video_size_limit":41943040,"video_frame_rate_limit":60,"video_matrix_limit":16777216},"polls":{"max_options":6,"max_characters_per_option":50,"min_expiration":300,"max_expiration":2629746},"accounts":{"allow_custom_css":true},"emojis":{"emoji_size_limit":51200}},"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":2,"status_count":16,"user_count":4},"thumbnail":"http://localhost:8080/assets/logo.png","contact_account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"admin"},"max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch2() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"title": "<p>Geoff's Instance</p>",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","account_domain":"localhost:8080","title":"Geoff's Instance","description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","short_description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","email":"admin@example.org","version":"0.0.0-testrig","registrations":true,"approval_required":true,"invites_enabled":false,"configuration":{"statuses":{"max_characters":5000,"max_media_attachments":6,"characters_reserved_per_url":25},"media_attachments":{"supported_mime_types":["image/jpeg","image/gif","image/png","image/webp","video/mp4"],"image_size_limit":10485760,"image_matrix_limit":16777216,"video_size_limit":41943040,"video_frame_rate_limit":60,"video_matrix_limit":16777216},"polls":{"max_options":6,"max_characters_per_option":50,"min_expiration":300,"max_expiration":2629746},"accounts":{"allow_custom_css":true},"emojis":{"emoji_size_limit":51200}},"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":2,"status_count":16,"user_count":4},"thumbnail":"http://localhost:8080/assets/logo.png","contact_account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"admin"},"max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch3() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"short_description": "<p>This is some html, which is <em>allowed</em> in short descriptions.</p>",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","account_domain":"localhost:8080","title":"GoToSocial Testrig Instance","description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","short_description":"\u003cp\u003eThis is some html, which is \u003cem\u003eallowed\u003c/em\u003e in short descriptions.\u003c/p\u003e","email":"admin@example.org","version":"0.0.0-testrig","registrations":true,"approval_required":true,"invites_enabled":false,"configuration":{"statuses":{"max_characters":5000,"max_media_attachments":6,"characters_reserved_per_url":25},"media_attachments":{"supported_mime_types":["image/jpeg","image/gif","image/png","image/webp","video/mp4"],"image_size_limit":10485760,"image_matrix_limit":16777216,"video_size_limit":41943040,"video_frame_rate_limit":60,"video_matrix_limit":16777216},"polls":{"max_options":6,"max_characters_per_option":50,"min_expiration":300,"max_expiration":2629746},"accounts":{"allow_custom_css":true},"emojis":{"emoji_size_limit":51200}},"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":2,"status_count":16,"user_count":4},"thumbnail":"http://localhost:8080/assets/logo.png","contact_account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"admin"},"max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch4() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: empty form submitted"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch5() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"short_description": "<p>This is some html, which is <em>allowed</em> in short descriptions.</p>",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	suite.Equal(http.StatusForbidden, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Forbidden: user is not an admin so cannot update instance settings"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch6() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"contact_email": "",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","account_domain":"localhost:8080","title":"GoToSocial Testrig Instance","description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","short_description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","email":"","version":"0.0.0-testrig","registrations":true,"approval_required":true,"invites_enabled":false,"configuration":{"statuses":{"max_characters":5000,"max_media_attachments":6,"characters_reserved_per_url":25},"media_attachments":{"supported_mime_types":["image/jpeg","image/gif","image/png","image/webp","video/mp4"],"image_size_limit":10485760,"image_matrix_limit":16777216,"video_size_limit":41943040,"video_frame_rate_limit":60,"video_matrix_limit":16777216},"polls":{"max_options":6,"max_characters_per_option":50,"min_expiration":300,"max_expiration":2629746},"accounts":{"allow_custom_css":true},"emojis":{"emoji_size_limit":51200}},"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":2,"status_count":16,"user_count":4},"thumbnail":"http://localhost:8080/assets/logo.png","contact_account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"admin"},"max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch7() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"contact_email": "not.an.email.address",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: mail: missing '@' or angle-addr"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch8() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"thumbnail", "../../../../testrig/media/peglin.gif",
		map[string]string{
			"thumbnail_description": "A bouncing little green peglin.",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPath, bodyBytes, w.FormDataContentType(), true)

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	instanceAccount, err := suite.db.GetInstanceAccount(context.Background(), "")
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(instanceAccount.AvatarMediaAttachmentID)

	expectedInstanceResponse := fmt.Sprintf(`{"uri":"http://localhost:8080","account_domain":"localhost:8080","title":"GoToSocial Testrig Instance","description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","short_description":"\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e","email":"admin@example.org","version":"0.0.0-testrig","registrations":true,"approval_required":true,"invites_enabled":false,"configuration":{"statuses":{"max_characters":5000,"max_media_attachments":6,"characters_reserved_per_url":25},"media_attachments":{"supported_mime_types":["image/jpeg","image/gif","image/png","image/webp","video/mp4"],"image_size_limit":10485760,"image_matrix_limit":16777216,"video_size_limit":41943040,"video_frame_rate_limit":60,"video_matrix_limit":16777216},"polls":{"max_options":6,"max_characters_per_option":50,"min_expiration":300,"max_expiration":2629746},"accounts":{"allow_custom_css":true},"emojis":{"emoji_size_limit":51200}},"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":2,"status_count":16,"user_count":4},"thumbnail":"http://localhost:8080/fileserver/%s/attachment/original/%s.gif","thumbnail_type":"image/gif","thumbnail_description":"A bouncing little green peglin.","contact_account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"admin"},"max_toot_chars":5000}`, instanceAccount.ID, instanceAccount.AvatarMediaAttachmentID)
	suite.Equal(expectedInstanceResponse, string(b))
}

func TestInstancePatchTestSuite(t *testing.T) {
	suite.Run(t, &InstancePatchTestSuite{})
}
