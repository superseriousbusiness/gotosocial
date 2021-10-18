package testrig

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/app"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/blocks"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/emoji"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/favourites"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/filter"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/followrequest"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/list"
	mediaModule "github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/notification"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/search"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/streaming"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/nodeinfo"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/webfinger"
	"github.com/superseriousbusiness/gotosocial/internal/api/security"
	"github.com/superseriousbusiness/gotosocial/internal/cliactions"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/web"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

// Start creates and starts a gotosocial testrig server
var Start cliactions.GTSAction = func(ctx context.Context, _ *config.Config) error {
	testrig.InitTestLog()

	c := testrig.NewTestConfig()
	dbService := testrig.NewTestDB()
	testrig.StandardDBSetup(dbService, nil)
	router := testrig.NewTestRouter(dbService)
	storageBackend := testrig.NewTestStorage()
	testrig.StandardStorageSetup(storageBackend, "./testrig/media")

	// build backend handlers
	oauthServer := testrig.NewTestOauthServer(dbService)
	transportController := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}), dbService)
	federator := testrig.NewTestFederator(dbService, transportController, storageBackend)

	emailSender := testrig.NewEmailSender("./web/template/", nil)

	processor := testrig.NewTestProcessor(dbService, storageBackend, federator, emailSender)
	if err := processor.Start(ctx); err != nil {
		return fmt.Errorf("error starting processor: %s", err)
	}

	idp, err := oidc.NewIDP(ctx, c)
	if err != nil {
		return fmt.Errorf("error creating oidc idp: %s", err)
	}

	// build client api modules
	authModule := auth.New(c, dbService, oauthServer, idp)
	accountModule := account.New(c, processor)
	instanceModule := instance.New(c, processor)
	appsModule := app.New(c, processor)
	followRequestsModule := followrequest.New(c, processor)
	webfingerModule := webfinger.New(c, processor)
	nodeInfoModule := nodeinfo.New(c, processor)
	webBaseModule := web.New(c, processor)
	usersModule := user.New(c, processor)
	timelineModule := timeline.New(c, processor)
	notificationModule := notification.New(c, processor)
	searchModule := search.New(c, processor)
	filtersModule := filter.New(c, processor)
	emojiModule := emoji.New(c, processor)
	listsModule := list.New(c, processor)
	mm := mediaModule.New(c, processor)
	fileServerModule := fileserver.New(c, processor)
	adminModule := admin.New(c, processor)
	statusModule := status.New(c, processor)
	securityModule := security.New(c, dbService)
	streamingModule := streaming.New(c, processor)
	favouritesModule := favourites.New(c, processor)
	blocksModule := blocks.New(c, processor)

	apis := []api.ClientModule{
		// modules with middleware go first
		securityModule,
		authModule,

		// now everything else
		webBaseModule,
		accountModule,
		instanceModule,
		appsModule,
		followRequestsModule,
		mm,
		fileServerModule,
		adminModule,
		statusModule,
		webfingerModule,
		nodeInfoModule,
		usersModule,
		timelineModule,
		notificationModule,
		searchModule,
		filtersModule,
		emojiModule,
		listsModule,
		streamingModule,
		favouritesModule,
		blocksModule,
	}

	for _, m := range apis {
		if err := m.Route(router); err != nil {
			return fmt.Errorf("routing error: %s", err)
		}
	}

	gts, err := gotosocial.NewServer(dbService, router, federator, c)
	if err != nil {
		return fmt.Errorf("error creating gotosocial service: %s", err)
	}

	if err := gts.Start(ctx); err != nil {
		return fmt.Errorf("error starting gotosocial service: %s", err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	logrus.Infof("received signal %s, shutting down", sig)

	testrig.StandardDBTeardown(dbService)
	testrig.StandardStorageTeardown(storageBackend)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	logrus.Info("done! exiting...")
	return nil
}
