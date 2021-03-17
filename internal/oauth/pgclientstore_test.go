package oauth

import (
	"context"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gotosocial/oauth2/v4/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type PgClientStoreTestSuite struct {
	suite.Suite
	conn             *pg.DB
	testClientID     string
	testClientSecret string
	testClientDomain string
	testClientUserID string
}

const ()

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *PgClientStoreTestSuite) SetupSuite() {
	suite.testClientID = "test-client-id"
	suite.testClientSecret = "test-client-secret"
	suite.testClientDomain = "https://example.org"
	suite.testClientUserID = "test-client-user-id"
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *PgClientStoreTestSuite) SetupTest() {
	suite.conn = pg.Connect(&pg.Options{})
	if err := suite.conn.Ping(context.Background()); err != nil {
		logrus.Panicf("db connection error: %s", err)
	}
	if err := suite.conn.Model(&oauthClient{}).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	}); err != nil {
		logrus.Panicf("db connection error: %s", err)
	}
}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *PgClientStoreTestSuite) TearDownTest() {
	if err := suite.conn.Model(&oauthClient{}).DropTable(&orm.DropTableOptions{}); err != nil {
		logrus.Panicf("drop table error: %s", err)
	}
	if err := suite.conn.Close(); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
	suite.conn = nil
}

func (suite *PgClientStoreTestSuite) TestClientStoreSetAndGet() {
	// set a new client in the store
	cs := NewPGClientStore(suite.conn)
	if err := cs.Set(context.Background(), suite.testClientID, models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID)); err != nil {
		suite.FailNow(err.Error())
	}

	// fetch that client from the store
	client, err := cs.GetByID(context.Background(), suite.testClientID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// check that the values are the same
	suite.NotNil(client)
	suite.EqualValues(models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID), client)
}

func (suite *PgClientStoreTestSuite) TestClientSetAndDelete() {
	// set a new client in the store
	cs := NewPGClientStore(suite.conn)
	if err := cs.Set(context.Background(), suite.testClientID, models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID)); err != nil {
		suite.FailNow(err.Error())
	}

	// fetch the client from the store
	client, err := cs.GetByID(context.Background(), suite.testClientID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// check that the values are the same
	suite.NotNil(client)
	suite.EqualValues(models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID), client)
	if err := cs.Delete(context.Background(), suite.testClientID); err != nil {
		suite.FailNow(err.Error())
	}

	// try to get the deleted client; we should get an error
	deletedClient, err := cs.GetByID(context.Background(), suite.testClientID)
	suite.Assert().Nil(deletedClient)
	suite.Assert().NotNil(err)
}

func TestPgClientStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PgClientStoreTestSuite))
}
