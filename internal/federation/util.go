package federation

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *federator) blockedDomain(host string) (bool, error) {
	b := &gtsmodel.DomainBlock{}
	err := f.db.GetWhere([]db.Where{{Key: "domain", Value: host, CaseInsensitive: true}}, b)
	if err == nil {
		// block exists
		return true, nil
	}

	if _, ok := err.(db.ErrNoEntries); ok {
		// there are no entries so there's no block
		return false, nil
	}

	// there's an actual error
	return false, err
}
