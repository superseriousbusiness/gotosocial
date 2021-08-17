package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Instance interface {
	// GetUserCountForInstance returns the number of known accounts registered with the given domain.
	GetUserCountForInstance(domain string) (int, error)

	// GetStatusCountForInstance returns the number of known statuses posted from the given domain.
	GetStatusCountForInstance(domain string) (int, error)

	// GetDomainCountForInstance returns the number of known instances known that the given domain federates with.
	GetDomainCountForInstance(domain string) (int, error)

	// GetAccountsForInstance returns a slice of accounts from the given instance, arranged by ID.
	GetAccountsForInstance(domain string, maxID string, limit int) ([]*gtsmodel.Account, error)
}
