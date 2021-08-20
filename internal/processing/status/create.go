package status

import (
	"fmt"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) Create(account *gtsmodel.Account, application *gtsmodel.Application, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, gtserror.WithCode) {
	uris := util.GenerateURIsForAccount(account.Username, p.config.Protocol, p.config.Host)
	thisStatusID, err := id.NewULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	thisStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, thisStatusID)
	thisStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, thisStatusID)

	newStatus := &gtsmodel.Status{
		ID:                       thisStatusID,
		URI:                      thisStatusURI,
		URL:                      thisStatusURL,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		Local:                    true,
		AccountID:                account.ID,
		AccountURI:               account.URI,
		ContentWarning:           text.RemoveHTML(form.SpoilerText),
		ActivityStreamsType:      gtsmodel.ActivityStreamsNote,
		Sensitive:                form.Sensitive,
		Language:                 form.Language,
		CreatedWithApplicationID: application.ID,
		Text:                     form.Status,
	}

	if err := p.ProcessReplyToID(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessMediaIDs(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessVisibility(form, account.Privacy, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessLanguage(form, account.Language, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessMentions(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessTags(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessEmojis(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessContent(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// put the new status in the database
	if err := p.db.PutStatus(newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it back to the processor for async processing
	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsNote,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		GTSModel:       newStatus,
		OriginAccount:  account,
	}

	// return the frontend representation of the new status to the submitter
	mastoStatus, err := p.tc.StatusToMasto(newStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", newStatus.ID, err))
	}

	return mastoStatus, nil
}
