package fedi_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type CollectionsTestSuite struct {
	FediTestSuite
}

func (suite *CollectionsTestSuite) TestGetFollowing() {
	testAccount := suite.testAccounts["local_account_1"]

	ctx := createTestContext(testAccount, testAccount)

	data, errWithCode := suite.fedi.FollowingGet(ctx, testAccount.Username, nil)
	suite.NoError(errWithCode)

	b, err := json.MarshalIndent(data, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "totalItems": 2,
  "type": "OrderedCollection",
  "id": "http://localhost:8080/users/zork"
}`, string(b))
}

func TestCollectionsTestSuite(t *testing.T) {
	suite.Run(t, &CollectionsTestSuite{})
}

func createTestContext(receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) context.Context {
	ctx := context.Background()
	ctx = gtscontext.SetReceivingAccount(ctx, receivingAccount)
	ctx = gtscontext.SetRequestingAccount(ctx, requestingAccount)
	return ctx
}
