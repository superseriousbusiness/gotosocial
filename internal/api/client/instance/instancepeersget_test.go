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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

type InstancePeersGetTestSuite struct {
	InstanceStandardTestSuite
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParams() {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`[{"domain":"example.org"},{"domain":"fossbros-anonymous.io"}]`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParamsUnauthorized() {
	config.SetInstanceExposePeers(false)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

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
	ctx := suite.newContext(recorder, http.MethodGet, []byte{}, requestURI, "")

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`[{"domain":"example.org"},{"domain":"fossbros-anonymous.io"}]`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspended() {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`[{"domain":"replyguys.com","suspended_at":"2020-05-13T13:29:12.000Z"}]`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspendedUnauthorized() {
	config.SetInstanceExposeSuspended(false)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

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
	ctx := suite.newContext(recorder, http.MethodGet, []byte{}, requestURI, "")

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`[{"domain":"replyguys.com","suspended_at":"2020-05-13T13:29:12.000Z"}]`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAll() {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended,open", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

	suite.instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`[{"domain":"example.org"},{"domain":"fossbros-anonymous.io"},{"domain":"replyguys.com","suspended_at":"2020-05-13T13:29:12.000Z"}]`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetFunkyParams() {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=aaaaaaaaaaaaaaaaa,open", baseURI, instance.InstancePeersPath)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)

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
