package user_test

import (
	"git.iim.gay/grufwub/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/api/security"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// nolint
type UserStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config         *config.Config
	db             db.DB
	log            *logrus.Logger
	tc             typeutils.TypeConverter
	federator      federation.Federator
	processor      processing.Processor
	storage        *kv.KVStore
	securityModule *security.Module

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status

	// module being tested
	userModule *user.Module
}
