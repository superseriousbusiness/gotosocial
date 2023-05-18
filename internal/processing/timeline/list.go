package timeline

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) ListTimelineGet(ctx context.Context, authed *oauth.Auth, listID string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Ensure list exists + is owned by this account.
	list, err := p.state.DB.GetListByID(ctx, listID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list.AccountID != authed.Account.ID {
		err = fmt.Errorf("list with id %s does not belong to account %s", list.ID, authed.Account.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statuses, err := p.state.DB.GetListTimeline(ctx, listID, maxID, sinceID, minID, limit)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("ListTimelineGet: db error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items          = make([]interface{}, 0, count)
		nextMaxIDValue string
		prevMinIDValue string
	)

	for i, s := range statuses {
		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		if i == count-1 {
			nextMaxIDValue = s.ID
		}

		if i == 0 {
			prevMinIDValue = s.ID
		}

		// Since list timeline is essentially a slice of
		// a home timeline, and uses much the same rules,
		// we can reuse the HomeTimelineable function.
		timelineable, err := p.filter.StatusHomeTimelineable(ctx, authed.Account, s)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because of an error checking StatusPublicTimelineable: %s", s.ID, err)
			continue
		}

		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		items = append(items, apiStatus)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/timelines/list/" + listID,
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
