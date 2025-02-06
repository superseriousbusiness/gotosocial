// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/spam"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/media/ffmpeg"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/observability"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	tlprocessor "github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/subscriptions"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/web"
	"github.com/superseriousbusiness/gotosocial/internal/webpush"
	"go.uber.org/automaxprocs/maxprocs"
)

// Maintenance starts and creates a GoToSocial server
// in maintenance mode (returns 503 for most requests).
var Maintenance action.GTSAction = func(ctx context.Context) error {
	route, err := router.New(ctx)
	if err != nil {
		return fmt.Errorf("error creating maintenance router: %w", err)
	}

	// Route maintenance handlers.
	maintenance := web.NewMaintenance()
	maintenance.Route(route)

	// Start the maintenance router.
	if err := route.Start(); err != nil {
		return fmt.Errorf("error starting maintenance router: %w", err)
	}

	// Catch shutdown signals from the OS.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs // block until signal received
	log.Infof(ctx, "received signal %s, shutting down", sig)

	if err := route.Stop(); err != nil {
		log.Errorf(ctx, "error stopping router: %v", err)
	}

	return nil
}

// Start creates and starts a gotosocial server
var Start action.GTSAction = func(ctx context.Context) error {
	// Set GOMAXPROCS / GOMEMLIMIT
	// to match container limits.
	setLimits(ctx)

	var (
		// Define necessary core variables
		// before anything so we can prepare
		// defer function for safe shutdown
		// depending on what services were
		// managed to be started.
		state   = new(state.State)
		route   *router.Router
		process *processing.Processor
	)

	defer func() {
		// Stop caches with
		// background tasks.
		state.Caches.Stop()

		if route != nil {
			// We reached a point where the API router
			// was created + setup. Ensure it gets stopped
			// first to stop processing new information.
			if err := route.Stop(); err != nil {
				log.Errorf(ctx, "error stopping router: %v", err)
			}
		}

		// Stop any currently running
		// worker processes / scheduled
		// tasks from being executed.
		state.Workers.Stop()

		if state.Timelines.Home != nil {
			// Home timeline mgr was setup, ensure it gets stopped.
			if err := state.Timelines.Home.Stop(); err != nil {
				log.Errorf(ctx, "error stopping home timeline: %v", err)
			}
		}

		if state.Timelines.List != nil {
			// List timeline mgr was setup, ensure it gets stopped.
			if err := state.Timelines.List.Stop(); err != nil {
				log.Errorf(ctx, "error stopping list timeline: %v", err)
			}
		}

		if process != nil {
			const timeout = time.Minute

			// Use a new timeout context to ensure
			// persisting queued tasks does not fail!
			// The main ctx is very likely canceled.
			ctx := context.WithoutCancel(ctx)
			ctx, cncl := context.WithTimeout(ctx, timeout)
			defer cncl()

			// Now that all the "moving" components have been stopped,
			// persist any remaining queued worker tasks to the database.
			if err := process.Admin().PersistWorkerQueues(ctx); err != nil {
				log.Errorf(ctx, "error persisting worker queues: %v", err)
			}
		}

		if state.DB != nil {
			// Lastly, if database service was started,
			// ensure it gets closed now all else stopped.
			if err := state.DB.Close(); err != nil {
				log.Errorf(ctx, "error stopping database: %v", err)
			}
		}

		// Finally reached end of shutdown.
		log.Info(ctx, "done! exiting...")
	}()

	// Create maintenance router.
	var err error
	route, err = router.New(ctx)
	if err != nil {
		return fmt.Errorf("error creating maintenance router: %w", err)
	}

	// Route maintenance handlers.
	maintenance := web.NewMaintenance()
	maintenance.Route(route)

	// Start the maintenance router to handle reqs
	// while the instance is starting up / migrating.
	if err := route.Start(); err != nil {
		return fmt.Errorf("error starting maintenance router: %w", err)
	}

	// Initialize tracing (noop if not enabled).
	if err := observability.InitializeTracing(); err != nil {
		return fmt.Errorf("error initializing tracing: %w", err)
	}

	// Initialize caches
	state.Caches.Init()
	state.Caches.Start()

	// Open connection to the database now caches started.
	dbService, err := bundb.NewBunDBService(ctx, state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set DB on state.
	state.DB = dbService

	// Set Actions on state, providing workers to
	// Actions as well for triggering side effects.
	state.AdminActions = admin.New(dbService, &state.Workers)

	// Ensure necessary database instance prerequisites exist.
	if err := dbService.CreateInstanceAccount(ctx); err != nil {
		return fmt.Errorf("error creating instance account: %s", err)
	}
	if err := dbService.CreateInstanceInstance(ctx); err != nil {
		return fmt.Errorf("error creating instance instance: %s", err)
	}
	if err := dbService.CreateInstanceApplication(ctx); err != nil {
		return fmt.Errorf("error creating instance application: %s", err)
	}

	// Get the instance account (we'll need this later).
	instanceAccount, err := dbService.GetInstanceAccount(ctx, "")
	if err != nil {
		return fmt.Errorf("error retrieving instance account: %w", err)
	}

	// Open the storage backend according to config.
	state.Storage, err = gtsstorage.AutoConfig()
	if err != nil {
		return fmt.Errorf("error opening storage backend: %w", err)
	}

	// Prepare wrapped httpclient with config.
	client := httpclient.New(httpclient.Config{
		AllowRanges:           config.MustParseIPPrefixes(config.GetHTTPClientAllowIPs()),
		BlockRanges:           config.MustParseIPPrefixes(config.GetHTTPClientBlockIPs()),
		Timeout:               config.GetHTTPClientTimeout(),
		TLSInsecureSkipVerify: config.GetHTTPClientTLSInsecureSkipVerify(),
	})

	// Compile WASM modules ahead of first use
	// to prevent unexpected initial slowdowns.
	//
	// Note that this can take a bit of memory
	// and processing so we perform this much
	// later after any database migrations.
	log.Info(ctx, "compiling WebAssembly")
	if err := compileWASM(ctx); err != nil {
		return err
	}

	// Build handlers used in later initializations.
	mediaManager := media.NewManager(state)
	oauthServer := oauth.New(ctx, dbService)
	typeConverter := typeutils.NewConverter(state)
	visFilter := visibility.NewFilter(state)
	intFilter := interaction.NewFilter(state)
	spamFilter := spam.NewFilter(state)
	federatingDB := federatingdb.New(state, typeConverter, visFilter, intFilter, spamFilter)
	transportController := transport.NewController(state, federatingDB, &federation.Clock{}, client)
	federator := federation.NewFederator(
		state,
		federatingDB,
		transportController,
		typeConverter,
		visFilter,
		intFilter,
		mediaManager,
	)

	// Decide whether to create a noop email
	// sender (won't send emails) or a real one.
	var emailSender email.Sender
	if smtpHost := config.GetSMTPHost(); smtpHost != "" {
		// Host is defined; create a proper sender.
		emailSender, err = email.NewSender()
		if err != nil {
			return fmt.Errorf("error creating email sender: %s", err)
		}
	} else {
		// No host is defined; create a noop sender.
		emailSender, err = email.NewNoopSender(nil)
		if err != nil {
			return fmt.Errorf("error creating noop email sender: %s", err)
		}
	}

	// Get or create a VAPID key pair.
	if _, err := dbService.GetVAPIDKeyPair(ctx); err != nil {
		return gtserror.Newf("error getting or creating VAPID key pair: %w", err)
	}

	// Create a Web Push notification sender.
	webPushSender := webpush.NewSender(client, state, typeConverter)

	// Initialize both home / list timelines.
	state.Timelines.Home = timeline.NewManager(
		tlprocessor.HomeTimelineGrab(state),
		tlprocessor.HomeTimelineFilter(state, visFilter),
		tlprocessor.HomeTimelineStatusPrepare(state, typeConverter),
		tlprocessor.SkipInsert(),
	)
	if err := state.Timelines.Home.Start(); err != nil {
		return fmt.Errorf("error starting home timeline: %s", err)
	}
	state.Timelines.List = timeline.NewManager(
		tlprocessor.ListTimelineGrab(state),
		tlprocessor.ListTimelineFilter(state, visFilter),
		tlprocessor.ListTimelineStatusPrepare(state, typeConverter),
		tlprocessor.SkipInsert(),
	)
	if err := state.Timelines.List.Start(); err != nil {
		return fmt.Errorf("error starting list timeline: %s", err)
	}

	// Start the job scheduler
	// (this is required for cleaner).
	state.Workers.StartScheduler()

	// Add a task to the scheduler to sweep caches.
	// Frequency = 1 * minute
	// Threshold = 60% capacity
	if !state.Workers.Scheduler.AddRecurring(
		"@cachesweep", // id
		time.Time{},   // start
		time.Minute,   // freq
		func(context.Context, time.Time) {
			state.Caches.Sweep(60)
		},
	) {
		return fmt.Errorf("error scheduling cache sweep: %w", err)
	}

	// Create background cleaner.
	cleaner := cleaner.New(state)

	// Create subscriptions fetcher.
	subscriptions := subscriptions.New(
		state,
		transportController,
		typeConverter,
	)

	// Create the processor using all the
	// other services we've created so far.
	process = processing.NewProcessor(
		cleaner,
		subscriptions,
		typeConverter,
		federator,
		oauthServer,
		mediaManager,
		state,
		emailSender,
		webPushSender,
		visFilter,
		intFilter,
	)

	// Schedule background cleaning tasks.
	if err := cleaner.ScheduleJobs(); err != nil {
		return fmt.Errorf("error scheduling cleaner jobs: %w", err)
	}

	// Schedule background subscriptions updating.
	if err := subscriptions.ScheduleJobs(); err != nil {
		return fmt.Errorf("error scheduling subscriptions jobs: %w", err)
	}

	// Initialize the specialized workers pools.
	state.Workers.Client.Init(messages.ClientMsgIndices())
	state.Workers.Federator.Init(messages.FederatorMsgIndices())
	state.Workers.Delivery.Init(client)
	state.Workers.Client.Process = process.Workers().ProcessFromClientAPI
	state.Workers.Federator.Process = process.Workers().ProcessFromFediAPI

	// Now start workers!
	state.Workers.Start()

	// Schedule notif tasks for all existing poll expiries.
	if err := process.Polls().ScheduleAll(ctx); err != nil {
		return fmt.Errorf("error scheduling poll expiries: %w", err)
	}

	// Initialize metrics.
	if err := observability.InitializeMetrics(state.DB); err != nil {
		return fmt.Errorf("error initializing metrics: %w", err)
	}

	// Run advanced migrations.
	if err := process.AdvancedMigrations().Migrate(ctx); err != nil {
		return err
	}

	/*
		HTTP router initialization
	*/

	// Close down the maintenance router.
	if err := route.Stop(); err != nil {
		return fmt.Errorf("error stopping maintenance router: %w", err)
	}

	// Instantiate the main router.
	route, err = router.New(ctx)
	if err != nil {
		return fmt.Errorf("error creating main router: %s", err)
	}

	// Start preparing global middleware
	// stack (used for every request).
	middlewares := make([]gin.HandlerFunc, 1)

	// RequestID middleware must run before tracing!
	middlewares[0] = middleware.AddRequestID(config.GetRequestIDHeader())

	// Add tracing middleware if enabled.
	if config.GetTracingEnabled() {
		middlewares = append(middlewares, observability.TracingMiddleware())
	}

	// Add metrics middleware if enabled.
	if config.GetMetricsEnabled() {
		middlewares = append(middlewares, observability.MetricsMiddleware())
	}

	middlewares = append(middlewares, []gin.HandlerFunc{
		// note: hooks adding ctx fields must be ABOVE
		// the logger, otherwise won't be accessible.
		middleware.Logger(config.GetLogClientIP()),
		middleware.HeaderFilter(state),
		middleware.UserAgent(),
		middleware.CORS(),
		middleware.ExtraHeaders(),
	}...)

	// Instantiate Content-Security-Policy
	// middleware, with extra URIs.
	cspExtraURIs := make([]string, 0)

	// Probe storage to check if extra URI is needed in CSP.
	// Error here means something is wrong with storage.
	storageCSPUri, err := state.Storage.ProbeCSPUri(ctx)
	if err != nil {
		return fmt.Errorf("error deriving Content-Security-Policy uri from storage: %w", err)
	}

	// storageCSPUri may be empty string if
	// not S3-backed storage; check for this.
	if storageCSPUri != "" {
		cspExtraURIs = append(cspExtraURIs, storageCSPUri)
	}

	// Add any extra CSP URIs from config.
	cspExtraURIs = append(cspExtraURIs, config.GetAdvancedCSPExtraURIs()...)

	// Add CSP to middlewares.
	middlewares = append(middlewares, middleware.ContentSecurityPolicy(cspExtraURIs...))

	// attach global middlewares which are used for every request
	route.AttachGlobalMiddleware(middlewares...)

	// attach global no route / 404 handler to the router
	route.AttachNoRouteHandler(func(c *gin.Context) {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), process.InstanceGetV1)
	})

	// build router modules
	var idp oidc.IDP
	if config.GetOIDCEnabled() {
		idp, err = oidc.NewIDP(ctx)
		if err != nil {
			return fmt.Errorf("error creating oidc idp: %w", err)
		}
	}

	routerSession, err := dbService.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving router session for session middleware: %w", err)
	}

	sessionName, err := middleware.SessionName()
	if err != nil {
		return fmt.Errorf("error generating session name for session middleware: %w", err)
	}

	var (
		authModule        = api.NewAuth(dbService, process, idp, routerSession, sessionName) // auth/oauth paths
		clientModule      = api.NewClient(state, process)                                    // api client endpoints
		metricsModule     = api.NewMetrics()                                                 // Metrics endpoints
		healthModule      = api.NewHealth(dbService.Ready)                                   // Health check endpoints
		fileserverModule  = api.NewFileserver(process)                                       // fileserver endpoints
		robotsModule      = api.NewRobots()                                                  // robots.txt endpoint
		wellKnownModule   = api.NewWellKnown(process)                                        // .well-known endpoints
		nodeInfoModule    = api.NewNodeInfo(process)                                         // nodeinfo endpoint
		activityPubModule = api.NewActivityPub(dbService, process)                           // ActivityPub endpoints
		webModule         = web.New(dbService, process)                                      // web pages + user profiles + settings panels etc
	)

	// Create per-route / per-grouping middlewares.
	// rate limiting
	rlLimit := config.GetAdvancedRateLimitRequests()
	clLimit := middleware.RateLimit(rlLimit, config.GetAdvancedRateLimitExceptionsParsed())        // client api
	s2sLimit := middleware.RateLimit(rlLimit, config.GetAdvancedRateLimitExceptionsParsed())       // server-to-server (AP)
	fsMainLimit := middleware.RateLimit(rlLimit, config.GetAdvancedRateLimitExceptionsParsed())    // fileserver / web templates
	fsEmojiLimit := middleware.RateLimit(rlLimit*2, config.GetAdvancedRateLimitExceptionsParsed()) // fileserver (emojis only, use high limit)

	// throttling
	cpuMultiplier := config.GetAdvancedThrottlingMultiplier()
	retryAfter := config.GetAdvancedThrottlingRetryAfter()
	clThrottle := middleware.Throttle(cpuMultiplier, retryAfter) // client api
	s2sThrottle := middleware.Throttle(cpuMultiplier, retryAfter)

	// server-to-server (AP)
	fsThrottle := middleware.Throttle(cpuMultiplier, retryAfter) // fileserver / web templates / emojis
	pkThrottle := middleware.Throttle(cpuMultiplier, retryAfter) // throttle public key endpoint separately

	// Robots http headers (x-robots-tag).
	//
	// robotsDisallowAll is used for client API + S2S endpoints
	// that definitely should never be indexed by crawlers.
	//
	// robotsDisallowAIOnly is used for utility endpoints,
	// fileserver, and for web endpoints that set their own
	// additional robots directives in HTML meta tags.
	//
	// Other endpoints like .well-known and nodeinfo handle
	// robots headers themselves based on configuration.
	robotsDisallowAll := middleware.RobotsHeaders("")
	robotsDisallowAIOnly := middleware.RobotsHeaders("aiOnly")

	// Gzip middleware is applied to all endpoints except
	// fileserver (compression too expensive for those),
	// health (which really doesn't need compression), and
	// metrics (which does its own compression handling that
	// is rather annoying to neatly override).
	gzip := middleware.Gzip()

	// these should be routed in order;
	// apply throttling *after* rate limiting
	authModule.Route(route, clLimit, clThrottle, robotsDisallowAll, gzip)
	clientModule.Route(route, clLimit, clThrottle, robotsDisallowAll, gzip)
	metricsModule.Route(route, clLimit, clThrottle, robotsDisallowAIOnly)
	healthModule.Route(route, clLimit, clThrottle, robotsDisallowAIOnly)
	fileserverModule.Route(route, fsMainLimit, fsThrottle, robotsDisallowAIOnly)
	fileserverModule.RouteEmojis(route, instanceAccount.ID, fsEmojiLimit, fsThrottle, robotsDisallowAIOnly)
	robotsModule.Route(route, fsMainLimit, fsThrottle, robotsDisallowAIOnly, gzip)
	wellKnownModule.Route(route, gzip, s2sLimit, s2sThrottle)
	nodeInfoModule.Route(route, s2sLimit, s2sThrottle, gzip)
	activityPubModule.Route(route, s2sLimit, s2sThrottle, robotsDisallowAll, gzip)
	activityPubModule.RoutePublicKey(route, s2sLimit, pkThrottle, robotsDisallowAll, gzip)
	webModule.Route(route, fsMainLimit, fsThrottle, robotsDisallowAIOnly, gzip)

	// Finally start the main http server!
	if err := route.Start(); err != nil {
		return fmt.Errorf("error starting router: %w", err)
	}

	// Fill worker queues from persisted task data in database.
	if err := process.Admin().FillWorkerQueues(ctx); err != nil {
		return fmt.Errorf("error filling worker queues: %w", err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs // block until signal received
	log.Infof(ctx, "received signal %s, shutting down", sig)

	return nil
}

func setLimits(ctx context.Context) {
	if _, err := maxprocs.Set(maxprocs.Logger(nil)); err != nil {
		log.Warnf(ctx, "could not set CPU limits from cgroup: %s", err)
	}

	if _, err := memlimit.SetGoMemLimitWithOpts(); err != nil {
		if !strings.Contains(err.Error(), "cgroup mountpoint does not exist") {
			log.Warnf(ctx, "could not set Memory limits from cgroup: %s", err)
		}
	}
}

func compileWASM(ctx context.Context) error {
	// Use admin-set ffmpeg pool size, and fall
	// back to GOMAXPROCS if number 0 or less.
	ffPoolSize := config.GetMediaFfmpegPoolSize()
	if ffPoolSize <= 0 {
		ffPoolSize = runtime.GOMAXPROCS(0)
	}

	if err := ffmpeg.InitFfmpeg(ctx, ffPoolSize); err != nil {
		return gtserror.Newf("error compiling ffmpeg: %w", err)
	}

	if err := ffmpeg.InitFfprobe(ctx, ffPoolSize); err != nil {
		return gtserror.Newf("error compiling ffprobe: %w", err)
	}

	return nil
}
