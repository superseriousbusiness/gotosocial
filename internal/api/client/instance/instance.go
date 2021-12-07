package instance

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// InstanceInformationPath is for serving instance info requests
	InstanceInformationPath = "api/v1/instance"
)

// Module implements the ClientModule interface
type Module struct {
	processor processing.Processor
}

// New returns a new instance information module
func New(processor processing.Processor) api.ClientModule {
	return &Module{
		processor: processor,
	}
}

// Route satisfies the ClientModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, InstanceInformationPath, m.InstanceInformationGETHandler)
	s.AttachHandler(http.MethodPatch, InstanceInformationPath, m.InstanceUpdatePATCHHandler)
	return nil
}
