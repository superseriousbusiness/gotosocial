package admin

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) DomainBlockDelete(account *gtsmodel.Account, id string) (*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlock := &gtsmodel.DomainBlock{}

	if err := p.db.GetByID(id, domainBlock); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// something has gone really wrong
			return nil, gtserror.NewErrorInternalError(err)
		}
		// there are no entries for this ID
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no entry for ID %s", id))
	}

	// prepare the domain block to return
	mastoDomainBlock, err := p.tc.DomainBlockToMasto(domainBlock, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// delete the domain block
	if err := p.db.DeleteByID(id, domainBlock); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return mastoDomainBlock, nil
}
