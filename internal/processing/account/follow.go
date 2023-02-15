package account

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// AccountFollowCreate handles a follow request to an account, either remote or local.
func (p *AccountProcessor) AccountFollowCreate(ctx context.Context, requestingAccount *gtsmodel.Account, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode) {
	// if there's a block between the accounts we shouldn't create the request ofc
	if blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, form.ID, true); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts"))
	}

	// make sure the target account actually exists in our db
	targetAcct, err := p.db.GetAccountByID(ctx, form.ID)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("accountfollowcreate: account %s not found in the db: %s", form.ID, err))
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	// check if a follow exists already
	if follows, err := p.db.IsFollowing(ctx, requestingAccount, targetAcct); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error checking follow in db: %s", err))
	} else if follows {
		// already follows so just return the relationship
		return p.AccountRelationshipGet(ctx, requestingAccount, form.ID)
	}

	// check if a follow request exists already
	if followRequested, err := p.db.IsFollowRequested(ctx, requestingAccount, targetAcct); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error checking follow request in db: %s", err))
	} else if followRequested {
		// already follow requested so just return the relationship
		return p.AccountRelationshipGet(ctx, requestingAccount, form.ID)
	}

	// check for attempt to follow self
	if requestingAccount.ID == targetAcct.ID {
		return nil, gtserror.NewErrorNotAcceptable(fmt.Errorf("accountfollowcreate: account %s cannot follow itself", requestingAccount.ID))
	}

	// make the follow request
	newFollowID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	showReblogs := true
	notify := false
	fr := &gtsmodel.FollowRequest{
		ID:              newFollowID,
		AccountID:       requestingAccount.ID,
		TargetAccountID: form.ID,
		ShowReblogs:     &showReblogs,
		URI:             uris.GenerateURIForFollow(requestingAccount.Username, newFollowID),
		Notify:          &notify,
	}
	if form.Reblogs != nil {
		fr.ShowReblogs = form.Reblogs
	}
	if form.Notify != nil {
		fr.Notify = form.Notify
	}

	// whack it in the database
	if err := p.db.Put(ctx, fr); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error creating follow request in db: %s", err))
	}

	// if it's a local account that's not locked we can just straight up accept the follow request
	if !*targetAcct.Locked && targetAcct.Domain == "" {
		if _, err := p.db.AcceptFollowRequest(ctx, requestingAccount.ID, form.ID); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error accepting folow request for local unlocked account: %s", err))
		}
		// return the new relationship
		return p.AccountRelationshipGet(ctx, requestingAccount, form.ID)
	}

	// otherwise we leave the follow request as it is and we handle the rest of the process asynchronously
	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityCreate,
		GTSModel:       fr,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetAcct,
	})

	// return whatever relationship results from this
	return p.AccountRelationshipGet(ctx, requestingAccount, form.ID)
}

// AccountFollowRemove handles the removal of a follow/follow request to an account, either remote or local.
func (p *AccountProcessor) AccountFollowRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	// if there's a block between the accounts we shouldn't do anything
	blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, targetAccountID, true)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("AccountFollowRemove: block exists between accounts"))
	}

	// make sure the target account actually exists in our db
	targetAcct, err := p.db.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("AccountFollowRemove: account %s not found in the db: %s", targetAccountID, err))
		}
	}

	// check if a follow request exists, and remove it if it does (storing the URI for later)
	var frChanged bool
	var frURI string
	fr := &gtsmodel.FollowRequest{}
	if err := p.db.GetWhere(ctx, []db.Where{
		{Key: "account_id", Value: requestingAccount.ID},
		{Key: "target_account_id", Value: targetAccountID},
	}, fr); err == nil {
		frURI = fr.URI
		if err := p.db.DeleteByID(ctx, fr.ID, fr); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("AccountFollowRemove: error removing follow request from db: %s", err))
		}
		frChanged = true
	}

	// now do the same thing for any existing follow
	var fChanged bool
	var fURI string
	f := &gtsmodel.Follow{}
	if err := p.db.GetWhere(ctx, []db.Where{
		{Key: "account_id", Value: requestingAccount.ID},
		{Key: "target_account_id", Value: targetAccountID},
	}, f); err == nil {
		fURI = f.URI
		if err := p.db.DeleteByID(ctx, f.ID, f); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("AccountFollowRemove: error removing follow from db: %s", err))
		}
		fChanged = true
	}

	// follow request status changed so send the UNDO activity to the channel for async processing
	if frChanged {
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccountID,
				URI:             frURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAcct,
		})
	}

	// follow status changed so send the UNDO activity to the channel for async processing
	if fChanged {
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccountID,
				URI:             fURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAcct,
		})
	}

	// return whatever relationship results from all this
	return p.AccountRelationshipGet(ctx, requestingAccount, targetAccountID)
}
