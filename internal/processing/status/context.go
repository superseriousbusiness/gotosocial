package status

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Context(account *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	return &apimodel.Context{}, nil
}
