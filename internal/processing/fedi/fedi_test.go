package fedi_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing/fedi"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FediTestSuite struct {
	suite.Suite
	state state.State
	fedi  *fedi.Processor

	testAccounts map[string]*gtsmodel.Account
}

func (suite *FediTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.Caches.Init()
	_ = testrig.NewTestDB(&suite.state)
	testrig.StandardDBSetup(suite.state.DB, nil)

	suite.testAccounts = testrig.NewTestAccounts()

	fedi := fedi.New(
		&suite.state,
		testrig.NewTestTypeConverter(suite.state.DB),
		testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, nil), testrig.NewTestMediaManager(&suite.state)),
		visibility.NewFilter(&suite.state),
	)
	suite.fedi = &fedi
}

func (suite *FediTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
}
