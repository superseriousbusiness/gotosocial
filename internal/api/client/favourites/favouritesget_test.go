package favourites_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/favourites"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FavouritesTestSuite struct {
	FavouritesStandardTestSuite
}

func (suite *FavouritesTestSuite) TestGetFavourites() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_2"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s?limit=80", favourites.BasePath), nil)
	ctx.Request.Header.Set("accept", "application/json")

	suite.favModule.FavouritesGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	favs := []model.Status{}
	err = json.Unmarshal(b, &favs)

	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), favs, 4)
	assert.Equal(suite.T(), "01F8MHCP5P2NWYQ416SBA0XSEV", favs[0].ID)
	assert.Equal(suite.T(), "01F8MH75CBF9JFX4ZAD54N0W0R", favs[len(favs)-1].ID)
}

func TestStatusGetTestSuite(t *testing.T) {
	suite.Run(t, new(FavouritesTestSuite))
}
