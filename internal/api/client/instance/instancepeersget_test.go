/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin/render"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InstancePeersGetTestSuite struct {
	InstanceStandardTestSuite
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParams() {
	recorder := httptest.NewRecorder()
	ctx, r := testrig.CreateGinTestContext(recorder, nil)
	r.HTMLRender = render.HTMLDebug{}

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  "example.org",
  "fossbros-anonymous.io"
]`, dst.String())
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParamsUnauthorized() {
	config.SetInstanceExposePeers(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusUnauthorized, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Unauthorized: peers open query requires an authenticated account/user"}`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParamsAuthorized() {
	config.SetInstanceExposePeers(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", true)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  "example.org",
  "fossbros-anonymous.io"
]`, dst.String())
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspended() {
	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  {
    "domain": "replyguys.com",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "reply-guying to tech posts"
  }
]`, dst.String())
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspendedUnauthorized() {
	config.SetInstanceExposeSuspended(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusUnauthorized, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Unauthorized: peers suspended query requires an authenticated account/user"}`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspendedAuthorized() {
	config.SetInstanceExposeSuspended(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", true)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  {
    "domain": "replyguys.com",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "reply-guying to tech posts"
  }
]`, dst.String())
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAll() {
	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended,open", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  {
    "domain": "example.org"
  },
  {
    "domain": "fossbros-anonymous.io"
  },
  {
    "domain": "replyguys.com",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "reply-guying to tech posts"
  }
]`, dst.String())
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAllWithObfuscated() {
	err := suite.db.Put(context.Background(), &gtsmodel.DomainBlock{
		ID:                 "01G633XTNK51GBADQZFZQDP6WR",
		CreatedAt:          testrig.TimeMustParse("2021-06-09T12:34:55+02:00"),
		UpdatedAt:          testrig.TimeMustParse("2021-06-09T12:34:55+02:00"),
		Domain:             "omg.just.the.worst.org.ever",
		CreatedByAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
		PublicComment:      "just absolutely the worst, wowza",
		Obfuscate:          testrig.TrueBool(),
	})
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended,open", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  {
    "domain": "example.org"
  },
  {
    "domain": "fossbros-anonymous.io"
  },
  {
    "domain": "o*g.*u**.t**.*or*t.*r**ev**",
    "suspended_at": "2021-06-09T10:34:55.000Z",
    "public_comment": "just absolutely the worst, wowza"
  },
  {
    "domain": "replyguys.com",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "reply-guying to tech posts"
  }
]`, dst.String())
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetFunkyParams() {
	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=aaaaaaaaaaaaaaaaa,open", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", true)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: filter aaaaaaaaaaaaaaaaa not recognized; accepted values are 'open', 'suspended'"}`, string(b))
}

func TestInstancePeersGetTestSuite(t *testing.T) {
	suite.Run(t, &InstancePeersGetTestSuite{})
}
