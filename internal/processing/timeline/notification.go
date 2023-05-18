package timeline

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) NotificationsGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, excludeTypes []string) (*apimodel.PageableResponse, gtserror.WithCode) {
	notifs, err := p.state.DB.GetAccountNotifications(ctx, authed.Account.ID, maxID, sinceID, minID, limit, excludeTypes)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("NotificationsGet: db error getting notifications: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(notifs)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items          = make([]interface{}, 0, count)
		nextMaxIDValue string
		prevMinIDValue string
	)

	for i, n := range notifs {
		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		if i == count-1 {
			nextMaxIDValue = n.ID
		}

		if i == 0 {
			prevMinIDValue = n.ID
		}

		// Ensure this notification should be shown to requester.
		if n.OriginAccount != nil {
			// Account is set, ensure it's visible to notif target.
			visible, err := p.filter.AccountVisible(ctx, authed.Account, n.OriginAccount)
			if err != nil {
				log.Debugf(ctx, "skipping notification %s because of an error checking notification visibility: %s", n.ID, err)
				continue
			}

			if !visible {
				continue
			}
		}

		if n.Status != nil {
			// Status is set, ensure it's visible to notif target.
			visible, err := p.filter.StatusVisible(ctx, authed.Account, n.Status)
			if err != nil {
				log.Debugf(ctx, "skipping notification %s because of an error checking notification visibility: %s", n.ID, err)
				continue
			}

			if !visible {
				continue
			}
		}

		item, err := p.tc.NotificationToAPINotification(ctx, n)
		if err != nil {
			log.Debugf(ctx, "skipping notification %s because it couldn't be converted to its api representation: %s", n.ID, err)
			continue
		}

		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/notifications",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

func (p *Processor) NotificationGet(ctx context.Context, account *gtsmodel.Account, targetNotifID string) (*apimodel.Notification, gtserror.WithCode) {
	notif, err := p.state.DB.GetNotificationByID(ctx, targetNotifID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}

		// Real error.
		return nil, gtserror.NewErrorInternalError(err)
	}

	if notifTargetAccountID := notif.TargetAccountID; notifTargetAccountID != account.ID {
		err = fmt.Errorf("account %s does not have permission to view notification belong to account %s", account.ID, notifTargetAccountID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}

		// Real error.
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiNotif, nil
}

func (p *Processor) NotificationsClear(ctx context.Context, authed *oauth.Auth) gtserror.WithCode {
	// Delete all notifications of all types that target the authorized account.
	if err := p.state.DB.DeleteNotifications(ctx, nil, authed.Account.ID, ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
