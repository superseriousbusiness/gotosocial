package admin

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) DomainBlocksGet(account *gtsmodel.Account, export bool) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlocks := []*gtsmodel.DomainBlock{}

	if err := p.db.GetAll(&domainBlocks); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// something has gone really wrong
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	mastoDomainBlocks := []*apimodel.DomainBlock{}
	for _, b := range domainBlocks {
		mastoDomainBlock, err := p.tc.DomainBlockToMasto(b, export)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		mastoDomainBlocks = append(mastoDomainBlocks, mastoDomainBlock)
	}

	return mastoDomainBlocks, nil
}
