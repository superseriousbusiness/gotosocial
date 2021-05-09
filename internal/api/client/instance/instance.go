package instance

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/message"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// InstanceInformationPath
	InstanceInformationPath = "api/v1/instance"
)

// Module implements the ClientModule interface
type Module struct {
	config    *config.Config
	processor message.Processor
	log       *logrus.Logger
}

// New returns a new instance information module
func New(config *config.Config, processor message.Processor, log *logrus.Logger) api.ClientModule {
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
