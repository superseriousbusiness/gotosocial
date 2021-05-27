package message

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) NotificationsGet(authed *oauth.Auth, limit int, maxID string) ([]*apimodel.Notification, ErrorWithCode) {
	notifs, err := p.db.GetNotificationsForAccount(authed.Account.ID, limit, maxID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	mastoNotifs := []*apimodel.Notification{}
	for _, n := range notifs {
		mastoNotif, err := p.tc.NotificationToMasto(n)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		mastoNotifs = append(mastoNotifs, mastoNotif)
	}

	return mastoNotifs, nil
}
