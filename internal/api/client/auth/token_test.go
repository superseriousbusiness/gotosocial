package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
)

type TokenTestSuite struct {
	AuthStandardTestSuite
}

func (suite *TokenTestSuite) TestBindTokenRequestJson() {
	// create the recorder and gin test context
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	b, err := json.Marshal(&auth.TokenRequestForm{
		ClientID:     "whatever",
		ClientSecret: "some-secret",
		Code:         "don't tell anyone!",
		GrantType:    "authorization_code",
		RedirectURI:  "localhost",
		Scope:        "read write",
	})
	if err != nil {
		panic(err)
	}

	// set up the request
	c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(b)) // the endpoint we're hitting
	c.Request.Header = http.Header{
		"Content-Type": []string{"application/json"},
	}
	c.Request.Form = url.Values{}

	err = suite.authModule.BindTokenRequestForm(c)
	suite.NoError(err)

	suite.Equal("whatever", c.Request.Form.Get("client_id"))
	suite.Equal("some-secret", c.Request.Form.Get("client_secret"))
	suite.Equal("don't tell anyone!", c.Request.Form.Get("code"))
	suite.Equal("authorization_code", c.Request.Form.Get("grant_type"))
	suite.Equal("localhost", c.Request.Form.Get("redirect_uri"))
	suite.Equal("read write", c.Request.Form.Get("scope"))
}

func (suite *TokenTestSuite) TestBindTokenRequestJsonScopeArray() {
	// create the recorder and gin test context
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	b, err := json.Marshal(&auth.TokenRequestForm{
		ClientID:     "whatever",
		ClientSecret: "some-secret",
		Code:         "don't tell anyone!",
		GrantType:    "authorization_code",
		RedirectURI:  "localhost",
		Scope:        []string{"read", "write"},
	})
	if err != nil {
		panic(err)
	}

	// set up the request
	c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewBuffer(b)) // the endpoint we're hitting
	c.Request.Header = http.Header{
		"Content-Type": []string{"application/json"},
	}
	c.Request.Form = url.Values{}

	err = suite.authModule.BindTokenRequestForm(c)
	suite.NoError(err)

	suite.Equal("whatever", c.Request.Form.Get("client_id"))
	suite.Equal("some-secret", c.Request.Form.Get("client_secret"))
	suite.Equal("don't tell anyone!", c.Request.Form.Get("code"))
	suite.Equal("authorization_code", c.Request.Form.Get("grant_type"))
	suite.Equal("localhost", c.Request.Form.Get("redirect_uri"))
	suite.Equal("read write", c.Request.Form.Get("scope"))
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, new(TokenTestSuite))
}
