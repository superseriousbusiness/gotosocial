package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gotosocial/gotosocial/internal/api"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/gtsmodel"
	"github.com/gotosocial/oauth2/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type OauthTestSuite struct {
	suite.Suite
	tokenStore       oauth2.TokenStore
	clientStore      oauth2.ClientStore
	conn             *pg.DB
	testClientID     string
	testClientSecret string
	testClientDomain string
	testClientUserID string
	testUser         *gtsmodel.User
	config           *config.Config
}

const ()

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *OauthTestSuite) SetupSuite() {
	suite.testClientID = "test-client-id"
	suite.testClientSecret = "test-client-secret"
	suite.testClientDomain = "https://example.org"
	suite.testClientUserID = "test-client-user-id"
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)
	if err != nil {
		logrus.Panicf("error encrypting user pass: %s", err)
	}
	suite.testUser = &gtsmodel.User{
		EncryptedPassword: string(encryptedPassword),
		Email:             "user@example.org",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		AccountID:         "whatever",
	}
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *OauthTestSuite) SetupTest() {
	suite.conn = pg.Connect(&pg.Options{})
	if err := suite.conn.Ping(context.Background()); err != nil {
		logrus.Panicf("db connection error: %s", err)
	}

	models := []interface{}{
		&oauthClient{},
		&oauthToken{},
		&gtsmodel.User{},
	}

	for _, m := range models {
		if err := suite.conn.Model(m).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}

	suite.tokenStore = NewPGTokenStore(context.Background(), suite.conn, logrus.New())
	suite.clientStore = NewPGClientStore(suite.conn)

	if _, err := suite.conn.Model(suite.testUser).Insert(); err != nil {
		logrus.Panicf("could not insert test user into db: %s", err)
	}

}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *OauthTestSuite) TearDownTest() {
	models := []interface{}{
		&oauthClient{},
		&oauthToken{},
		&gtsmodel.User{},
	}
	for _, m := range models {
		if err := suite.conn.Model(m).DropTable(&orm.DropTableOptions{}); err != nil {
			logrus.Panicf("drop table error: %s", err)
		}
	}
	if err := suite.conn.Close(); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
	suite.conn = nil
}

func (suite *OauthTestSuite) TestAPIInitialize() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	r := api.New(suite.config, log)
	api := New(suite.tokenStore, suite.clientStore, suite.conn, log)
	api.AddRoutes(r)
	go r.Start()
	time.Sleep(30 * time.Second)
	// http://localhost:8080/oauth/authorize?client_id=whatever
}

func TestOauthTestSuite(t *testing.T) {
	suite.Run(t, new(OauthTestSuite))
}
