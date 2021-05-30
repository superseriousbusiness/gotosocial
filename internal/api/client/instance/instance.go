package instance

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// InstanceInformationPath is for serving instance info requests
	InstanceInformationPath = "api/v1/instance"
)

// Module implements the ClientModule interface
type Module struct {
	config    *config.Config
	processor processing.Processor
	log       *logrus.Logger
}

// New returns a new instance information module
func New(config *config.Config, processor processing.Processor, log *logrus.Logger) api.ClientModule {
	return &Module{
		config:    config,
		processor: processor,
		log:       log,
	}
}

// Route satisfies the ClientModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, InstanceInformationPath, m.InstanceInformationGETHandler)
	return nil
}
