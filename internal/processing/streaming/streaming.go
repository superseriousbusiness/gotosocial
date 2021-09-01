package streaming

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// Processor wraps a bunch of functions for processing streaming.
type Processor interface {
	// AuthorizeStreamingRequest returns an oauth2 token info in response to an access token query from the streaming API
	AuthorizeStreamingRequest(ctx context.Context, accessToken string) (*gtsmodel.Account, error)
	// OpenStreamForAccount returns a new Stream for the given account, which will contain a channel for passing messages back to the caller.
	OpenStreamForAccount(ctx context.Context, account *gtsmodel.Account, streamType string) (*stream.Stream, gtserror.WithCode)
	// StreamStatusToAccount streams the given status to any open, appropriate streams belonging to the given account.
	StreamStatusToAccount(s *apimodel.Status, account *gtsmodel.Account) error
	// StreamNotificationToAccount streams the given notification to any open, appropriate streams belonging to the given account.
	StreamNotificationToAccount(n *apimodel.Notification, account *gtsmodel.Account) error
	// StreamDelete streams the delete of the given statusID to *ALL* open streams.
	StreamDelete(statusID string) error
}

type processor struct {
	tc          typeutils.TypeConverter
	config      *config.Config
	db          db.DB
	filter      visibility.Filter
	log         *logrus.Logger
	oauthServer oauth.Server
	streamMap   *sync.Map
}

// New returns a new status processor.
func New(db db.DB, tc typeutils.TypeConverter, oauthServer oauth.Server, config *config.Config, log *logrus.Logger) Processor {
	return &processor{
		tc:          tc,
		config:      config,
		db:          db,
		filter:      visibility.NewFilter(db, log),
		log:         log,
		oauthServer: oauthServer,
		streamMap:   &sync.Map{},
	}
}
