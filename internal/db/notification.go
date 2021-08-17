package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Notification interface {
	// GetNotificationsForAccount returns a list of notifications that pertain to the given accountID.
	GetNotificationsForAccount(accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, error)
}
