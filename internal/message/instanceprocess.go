package message

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) InstanceGet(domain string) (*apimodel.Instance, ErrorWithCode) {
	i := &gtsmodel.Instance{}
	if err := p.db.GetWhere("domain", domain, i); err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("db error fetching instance %s: %s", p.config.Host, err))
	}

	ai, err := p.tc.InstanceToMasto(i)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return  ai, nil
}
