package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"codeberg.org/gruf/go-store/kv"
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
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	timelineprocessing "github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/web"
)

// Start creates and starts a gotosocial server
var Start cliactions.GTSAction = func(ctx context.Context, c *config.Config) error {
	dbService, err := bundb.NewBunDBService(ctx, c)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	if err := dbService.CreateInstanceAccount(ctx); err != nil {
		return fmt.Errorf("error creating instance account: %s", err)
	}

	if err := dbService.CreateInstanceInstance(ctx); err != nil {
		return fmt.Errorf("error creating instance instance: %s", err)
	}

	federatingDB := federatingdb.New(dbService, c)

	router, err := router.New(ctx, c, dbService)
	if err != nil {
		return fmt.Errorf("error creating router: %s", err)
	}

	// build converters and util
	typeConverter := typeutils.NewConverter(c, dbService)
	timelineManager := timelineprocessing.NewManager(dbService, typeConverter, c)

	// Open the storage backend
	storage, err := kv.OpenFile(c.StorageConfig.BasePath, nil)
	if err != nil {
		return fmt.Errorf("error creating storage backend: %s", err)
	}

	// build backend handlers
	mediaHandler := media.New(c, dbService, storage)
	oauthServer := oauth.New(ctx, dbService)
	transportController := transport.NewController(c, dbService, &federation.Clock{}, http.DefaultClient)
	federator := federation.NewFederator(dbService, federatingDB, transportController, c, typeConverter, mediaHandler)

	// decide whether to create a noop email sender (won't send emails) or a real one
	var emailSender email.Sender
	if c.SMTPConfig.Host != "" {
		// host is defined so create a proper sender
		emailSender, err = email.NewSender(c)
		if err != nil {
			return fmt.Errorf("error creating email sender: %s", err)
		}
	} else {
		// no host is defined so create a noop sender
		emailSender, err = email.NewNoopSender(c.TemplateConfig.BaseDir, nil)
		if err != nil {
			return fmt.Errorf("error creating noop email sender: %s", err)
		}
	}

	// create and start the message processor using the other services we've created so far
	processor := processing.NewProcessor(c, typeConverter, federator, oauthServer, mediaHandler, storage, timelineManager, dbService, emailSender)
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

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	logrus.Info("done! exiting...")
	return nil
}
