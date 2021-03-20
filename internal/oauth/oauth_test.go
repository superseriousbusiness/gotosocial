package oauth

import (
	"context"
	"fmt"
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
	tokenStore  oauth2.TokenStore
	clientStore oauth2.ClientStore
	conn        *pg.DB
	testAccount *gtsmodel.Account
	testUser    *gtsmodel.User
	testClient  *oauthClient
	config      *config.Config
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *OauthTestSuite) SetupSuite() {
	c := config.Empty()
	// we're running on localhost without https so set the protocol to http
	c.Protocol = "http"
	// just for testing
	c.Host = "localhost:8080"
	// because go tests are run within the test package directory, we need to fiddle with the templateconfig
	// basedir in a way that we wouldn't normally have to do when running the binary, in order to make
	// the templates actually load
	c.TemplateConfig.BaseDir = "../../web/template/"
	suite.config = c

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		logrus.Panicf("error encrypting user pass: %s", err)
	}

	suite.testAccount = &gtsmodel.Account{}
	suite.testUser = &gtsmodel.User{
		EncryptedPassword: string(encryptedPassword),
		Email:             "user@example.org",
		AccountID:         "some-account-id-it-doesn't-matter-really-since-this-user-doesn't-actually-have-an-account!",
	}
	suite.testClient = &oauthClient{
		ID:     "a-known-client-id",
		Secret: "some-secret",
		Domain: fmt.Sprintf("%s://%s", c.Protocol, c.Host),
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
		&gtsmodel.Account{},
		&gtsmodel.Application{},
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

	if _, err := suite.conn.Model(suite.testClient).Insert(); err != nil {
		logrus.Panicf("could not insert test client into db: %s", err)
	}

}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *OauthTestSuite) TearDownTest() {
	models := []interface{}{
		&oauthClient{},
		&oauthToken{},
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Application{},
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
	log.SetLevel(logrus.TraceLevel)

	r := api.New(suite.config, log)
	api := New(suite.tokenStore, suite.clientStore, suite.conn, log)
	if err := api.Route(r); err != nil {
		suite.FailNow(fmt.Sprintf("error initializing api: %s", err))
	}
	go r.Start()
	time.Sleep(60 * time.Second)
	// http://localhost:8080/oauth/authorize?client_id=a-known-client-id&response_type=code&redirect_uri=http://localhost:8080
	// curl -v -F client_id=a-known-client-id -F client_secret=some-secret -F redirect_uri=http://localhost:8080 -F code=[ INSERT CODE HERE ] -F grant_type=authorization_code localhost:8080/oauth/token
	// curl -v -H "Authorization: bearer [INSERT TOKEN HERE]" http://localhost:8080
}

func TestOauthTestSuite(t *testing.T) {
	suite.Run(t, new(OauthTestSuite))
}
