package status

import (
	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// Processor wraps a bunch of functions for processing statuses.
type Processor interface {
	// Create processes the given form to create a new status, returning the api model representation of that status if it's OK.
	Create(account *gtsmodel.Account, application *gtsmodel.Application, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, gtserror.WithCode)
	// Delete processes the delete of a given status, returning the deleted status if the delete goes through.
	Delete(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Fave processes the faving of a given status, returning the updated status if the fave goes through.
	Fave(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Boost processes the boost/reblog of a given status, returning the newly-created boost if all is well.
	Boost(account *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// BoostedBy returns a slice of accounts that have boosted the given status, filtered according to privacy settings.
	BoostedBy(account *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode)
	// FavedBy returns a slice of accounts that have liked the given status, filtered according to privacy settings.
	FavedBy(account *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode)
	// Get gets the given status, taking account of privacy settings and blocks etc.
	Get(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Unfave processes the unfaving of a given status, returning the updated status if the fave goes through.
	Unfave(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Context returns the context (previous and following posts) from the given status ID
	Context(account *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode)
}

type processor struct {
	tc            typeutils.TypeConverter
	config        *config.Config
	db            db.DB
	filter        visibility.Filter
	fromClientAPI chan gtsmodel.FromClientAPI
	log           *logrus.Logger
}

// New returns a new status processor.
func New(db db.DB, tc typeutils.TypeConverter, config *config.Config, fromClientAPI chan gtsmodel.FromClientAPI, log *logrus.Logger) Processor {
	return &processor{
		tc:            tc,
		config:        config,
		db:            db,
		filter:        visibility.NewFilter(db, log),
		fromClientAPI: fromClientAPI,
		log:           log,
	}
}
