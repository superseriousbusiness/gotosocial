package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/auth"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-contrib/sessions"
	"github.com/stretchr/testify/suite"
)

type AuthAuthorizeTestSuite struct {
	AuthStandardTestSuite
}

type authorizeHandlerTestCase struct {
	description            string
	mutateUserAccount      func(*gtsmodel.User, *gtsmodel.Account) []string
	expectedStatusCode     int
	expectedLocationHeader string
}

func (suite *AuthAuthorizeTestSuite) TestAccountAuthorizeHandler() {
	tests := []authorizeHandlerTestCase{
		{
			description: "user has their email unconfirmed",
			mutateUserAccount: func(user *gtsmodel.User, account *gtsmodel.Account) []string {
				user.ConfirmedAt = time.Time{}
				return []string{"confirmed_at"}
			},
			expectedStatusCode:     http.StatusSeeOther,
			expectedLocationHeader: "/auth" + auth.AuthCheckYourEmailPath,
		},
		{
			description: "user has their email confirmed but is not approved",
			mutateUserAccount: func(user *gtsmodel.User, account *gtsmodel.Account) []string {
				user.ConfirmedAt = time.Now()
				user.Email = user.UnconfirmedEmail
				return []string{"confirmed_at", "email"}
			},
			expectedStatusCode:     http.StatusSeeOther,
			expectedLocationHeader: "/auth" + auth.AuthWaitForApprovalPath,
		},
		{
			description: "user has their email confirmed and is approved, but User entity has been disabled",
			mutateUserAccount: func(user *gtsmodel.User, account *gtsmodel.Account) []string {
				user.ConfirmedAt = time.Now()
				user.Email = user.UnconfirmedEmail
				user.Approved = util.Ptr(true)
				user.Disabled = util.Ptr(true)
				return []string{"confirmed_at", "email", "approved", "disabled"}
			},
			expectedStatusCode:     http.StatusSeeOther,
			expectedLocationHeader: "/auth" + auth.AuthAccountDisabledPath,
		},
		{
			description: "user has their email confirmed and is approved, but Account entity has been suspended",
			mutateUserAccount: func(user *gtsmodel.User, account *gtsmodel.Account) []string {
				user.ConfirmedAt = time.Now()
				user.Email = user.UnconfirmedEmail
				user.Approved = util.Ptr(true)
				user.Disabled = util.Ptr(false)
				account.SuspendedAt = time.Now()
				return []string{"confirmed_at", "email", "approved", "disabled"}
			},
			expectedStatusCode:     http.StatusSeeOther,
			expectedLocationHeader: "/auth" + auth.AuthAccountDisabledPath,
		},
	}

	doTest := func(testCase authorizeHandlerTestCase) {
		ctx, recorder := suite.newContext(http.MethodGet, auth.OauthAuthorizePath, nil, "")

		user := &gtsmodel.User{}
		account := &gtsmodel.Account{}

		*user = *suite.testUsers["unconfirmed_account"]
		*account = *suite.testAccounts["unconfirmed_account"]

		testSession := sessions.Default(ctx)
		testSession.Set(sessionUserID, user.ID)
		testSession.Set(sessionClientID, suite.testApplications["application_1"].ClientID)
		if err := testSession.Save(); err != nil {
			panic(fmt.Errorf("failed on case %s: %w", testCase.description, err))
		}

		columns := testCase.mutateUserAccount(user, account)

		testCase.description = fmt.Sprintf("%s, %t, %s", user.Email, *user.Disabled, account.SuspendedAt)

		err := suite.db.UpdateUser(context.Background(), user, columns...)
		suite.NoError(err)
		err = suite.db.UpdateAccount(context.Background(), account)
		suite.NoError(err)

		// call the handler
		suite.authModule.AuthorizeGETHandler(ctx)

		// 1. we should have a redirect
		suite.Equal(testCase.expectedStatusCode, recorder.Code, fmt.Sprintf("failed on case: %s", testCase.description))

		// 2. we should have a redirect to the check your email path, as this user has not confirmed their email yet.
		suite.Equal(testCase.expectedLocationHeader, recorder.Header().Get("Location"), fmt.Sprintf("failed on case: %s", testCase.description))
	}

	for _, testCase := range tests {
		doTest(testCase)
	}
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AuthAuthorizeTestSuite))
}
