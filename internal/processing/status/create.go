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

	// check if replyToID is ok
	if err := p.ProcessReplyToID(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// check if mediaIDs are ok
	if err := p.ProcessMediaIDs(form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// check if visibility settings are ok
	if err := p.ProcessVisibility(form, account.Privacy, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// handle language settings
	if err := p.ProcessLanguage(form, account.Language, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// handle mentions
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

	// put the new status in the database, generating an ID for it in the process
	if err := p.db.Put(newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// change the status ID of the media attachments to the new status
	for _, a := range newStatus.GTSMediaAttachments {
		a.StatusID = newStatus.ID
		a.UpdatedAt = time.Now()
		if err := p.db.UpdateByID(a.ID, a); err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
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
