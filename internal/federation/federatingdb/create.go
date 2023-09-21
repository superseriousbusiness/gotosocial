// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package federatingdb

import (
	"context"
	"errors"
	"fmt"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Create adds a new entry to the database which must be able to be
// keyed by its id.
//
// Note that Activity values received from federated peers may also be
// created in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
//
// Under certain conditions and network activities, Create may be called
// multiple times for the same ActivityStreams object.
func (f *federatingDB) Create(ctx context.Context, asType vocab.Type) error {
	if log.Level() >= level.TRACE {
		i, err := marshalItem(asType)
		if err != nil {
			return err
		}

		log.
			WithContext(ctx).
			WithField("create", i).
			Trace("entering Create")
	}

	receivingAccount, requestingAccount, internal := extractFromCtx(ctx)
	if internal {
		return nil // Already processed.
	}

	switch asType.GetTypeName() {
	case ap.ActivityBlock:
		// BLOCK SOMETHING
		return f.activityBlock(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityCreate:
		// CREATE SOMETHING
		return f.activityCreate(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityFollow:
		// FOLLOW SOMETHING
		return f.activityFollow(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityLike:
		// LIKE SOMETHING
		return f.activityLike(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityFlag:
		// FLAG / REPORT SOMETHING
		return f.activityFlag(ctx, asType, receivingAccount, requestingAccount)
	}
	return nil
}

/*
	BLOCK HANDLERS
*/

func (f *federatingDB) activityBlock(ctx context.Context, asType vocab.Type, receiving *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	blockable, ok := asType.(vocab.ActivityStreamsBlock)
	if !ok {
		return errors.New("activityBlock: could not convert type to block")
	}

	block, err := f.converter.ASBlockToBlock(ctx, blockable)
	if err != nil {
		return fmt.Errorf("activityBlock: could not convert Block to gts model block")
	}

	block.ID = id.NewULID()

	if err := f.state.DB.PutBlock(ctx, block); err != nil {
		return fmt.Errorf("activityBlock: database error inserting block: %s", err)
	}

	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ActivityBlock,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         block,
		ReceivingAccount: receiving,
	})
	return nil
}

/*
	CREATE HANDLERS
*/

// activityCreate handles asType Create by checking
// the Object entries of the Create and calling other
// handlers as appropriate.
func (f *federatingDB) activityCreate(
	ctx context.Context,
	asType vocab.Type,
	receivingAccount *gtsmodel.Account,
	requestingAccount *gtsmodel.Account,
) error {
	create, ok := asType.(vocab.ActivityStreamsCreate)
	if !ok {
		return gtserror.Newf("could not convert asType %T to ActivityStreamsCreate", asType)
	}

	// Create must have an Object.
	objectProp := create.GetActivityStreamsObject()
	if objectProp == nil {
		return gtserror.New("create had no Object")
	}

	// Iterate through the Object property and process FIRST provided statusable.
	// todo: https://github.com/superseriousbusiness/gotosocial/issues/1905
	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		object := iter.GetType()
		if object == nil {
			// Can't do Create with Object that's just a URI.
			// Warn log this because it's an AP error.
			log.Warn(ctx, "object entry was not a type: %[1]T%[1]+v", iter)
			continue
		}

		// Ensure given object type is a statusable.
		statusable, ok := object.(ap.Statusable)
		if !ok {
			// Can't (currently) Create anything other than a Statusable. ([1] is a format arg index)
			log.Debugf(ctx, "object entry type (currently) unsupported: %[1]T%[1]+v", object)
			continue
		}

		// Handle creation of statusable.
		return f.createStatusable(ctx,
			statusable,
			receivingAccount,
			requestingAccount,
		)
	}

	return nil
}

// createStatusable handles a Create activity for a Statusable.
// This function won't insert anything in the database yet,
// but will pass the Statusable (if appropriate) through to
// the processor for further asynchronous processing.
func (f *federatingDB) createStatusable(
	ctx context.Context,
	statusable ap.Statusable,
	receivingAccount *gtsmodel.Account,
	requestingAccount *gtsmodel.Account,
) error {
	// Statusable must have an attributedTo.
	attrToProp := statusable.GetActivityStreamsAttributedTo()
	if attrToProp == nil {
		return gtserror.Newf("statusable had no attributedTo")
	}

	// Statusable must have an ID.
	idProp := statusable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return gtserror.Newf("statusable had no id, or id was not a URI")
	}

	statusableURI := idProp.GetIRI()

	// Check if we have a forward. In other words, was the
	// statusable posted to our inbox by at least one actor
	// who actually created it, or are they forwarding it?
	forward := true
	for iter := attrToProp.Begin(); iter != attrToProp.End(); iter = iter.Next() {
		actorURI, err := pub.ToId(iter)
		if err != nil {
			return gtserror.Newf("error extracting id from attributedTo entry: %w", err)
		}

		if requestingAccount.URI == actorURI.String() {
			// The actor who posted this statusable to our inbox is
			// (one of) its creator(s), so this is not a forward.
			forward = false
			break
		}
	}

	// Check if we already have a status entry
	// for this statusable, based on the ID/URI.
	statusableURIStr := statusableURI.String()
	status, err := f.state.DB.GetStatusByURI(ctx, statusableURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error checking existence of status %s: %w", statusableURIStr, err)
	}

	if status != nil {
		// We already had this status in the db, no need for further action.
		log.Trace(ctx, "status already exists: %s", statusableURIStr)
		return nil
	}

	// If we do have a forward, we should ignore the content
	// and instead deref based on the URI of the statusable.
	//
	// In other words, don't automatically trust whoever sent
	// this status to us, but fetch the authentic article from
	// the server it originated from.
	if forward {
		// Pass the statusable URI (APIri) into the processor worker
		// and do the rest of the processing asynchronously.
		f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
			APObjectType:     ap.ObjectNote,
			APActivityType:   ap.ActivityCreate,
			APIri:            statusableURI,
			APObjectModel:    nil,
			GTSModel:         nil,
			ReceivingAccount: receivingAccount,
		})
		return nil
	}

	// This is a non-forwarded status we can trust the requester on,
	// convert this provided statusable data to a useable gtsmodel status.
	status, err = f.converter.ASStatusToStatus(ctx, statusable)
	if err != nil {
		return gtserror.Newf("error converting statusable to status: %w", err)
	}

	// Check whether we should accept this new status.
	accept, err := f.shouldAcceptStatusable(ctx,
		receivingAccount,
		requestingAccount,
		status,
	)
	if err != nil {
		return gtserror.Newf("error checking status acceptibility: %w", err)
	}

	if !accept {
		// This is a status sent with no relation to receiver, i.e.
		// - receiving account does not follow requesting account
		// - received status does not mention receiving account
		//
		// We just pretend that all is fine (dog with cuppa, flames everywhere)
		log.Trace(ctx, "status failed acceptability check")
		return nil
	}

	// ID the new status based on the time it was created.
	status.ID, err = id.NewULIDFromTime(status.CreatedAt)
	if err != nil {
		return err
	}

	// Put this newly parsed status in the database.
	if err := f.state.DB.PutStatus(ctx, status); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// The status already exists in the database, which
			// means we've already processed it and some race
			// condition means we didn't catch it yet. We can
			// just return nil here and be done with it.
			return nil
		}
		return gtserror.Newf("db error inserting status: %w", err)
	}

	// Do the rest of the processing asynchronously. The processor
	// will handle inserting/updating + further dereferencing the status.
	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ObjectNote,
		APActivityType:   ap.ActivityCreate,
		APIri:            nil,
		APObjectModel:    statusable,
		GTSModel:         status,
		ReceivingAccount: receivingAccount,
	})

	return nil
}

func (f *federatingDB) shouldAcceptStatusable(ctx context.Context, receiver *gtsmodel.Account, requester *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	// Check whether status mentions the receiver,
	// this is the quickest check so perform it first.
	for _, mention := range status.Mentions {
		if mention.TargetAccountURI == receiver.URI {
			return true, nil
		}
	}

	// Check whether receiving account follows the requesting account.
	follows, err := f.state.DB.IsFollowing(ctx, receiver.ID, requester.ID)
	if err != nil {
		return false, gtserror.Newf("error checking follow status: %w", err)
	}

	// Status will only be acceptable
	// if receiver follows requester.
	return follows, nil
}

/*
	FOLLOW HANDLERS
*/

func (f *federatingDB) activityFollow(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	follow, ok := asType.(vocab.ActivityStreamsFollow)
	if !ok {
		return errors.New("activityFollow: could not convert type to follow")
	}

	followRequest, err := f.converter.ASFollowToFollowRequest(ctx, follow)
	if err != nil {
		return fmt.Errorf("activityFollow: could not convert Follow to follow request: %s", err)
	}

	followRequest.ID = id.NewULID()

	if err := f.state.DB.PutFollowRequest(ctx, followRequest); err != nil {
		return fmt.Errorf("activityFollow: database error inserting follow request: %s", err)
	}

	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ActivityFollow,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         followRequest,
		ReceivingAccount: receivingAccount,
	})

	return nil
}

/*
	LIKE HANDLERS
*/

func (f *federatingDB) activityLike(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	like, ok := asType.(vocab.ActivityStreamsLike)
	if !ok {
		return errors.New("activityLike: could not convert type to like")
	}

	fave, err := f.converter.ASLikeToFave(ctx, like)
	if err != nil {
		return fmt.Errorf("activityLike: could not convert Like to fave: %w", err)
	}

	fave.ID = id.NewULID()

	if err := f.state.DB.PutStatusFave(ctx, fave); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// The Like already exists in the database, which
			// means we've already handled side effects. We can
			// just return nil here and be done with it.
			return nil
		}
		return fmt.Errorf("activityLike: database error inserting fave: %w", err)
	}

	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ActivityLike,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         fave,
		ReceivingAccount: receivingAccount,
	})

	return nil
}

/*
	FLAG HANDLERS
*/

func (f *federatingDB) activityFlag(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	flag, ok := asType.(vocab.ActivityStreamsFlag)
	if !ok {
		return errors.New("activityFlag: could not convert type to flag")
	}

	report, err := f.converter.ASFlagToReport(ctx, flag)
	if err != nil {
		return fmt.Errorf("activityFlag: could not convert Flag to report: %w", err)
	}

	report.ID = id.NewULID()

	if err := f.state.DB.PutReport(ctx, report); err != nil {
		return fmt.Errorf("activityFlag: database error inserting report: %w", err)
	}

	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ActivityFlag,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         report,
		ReceivingAccount: receivingAccount,
	})

	return nil
}
