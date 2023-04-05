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
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (f *federatingDB) Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
	l := log.WithContext(ctx)

	if log.Level() >= level.DEBUG {
		i, err := marshalItem(undo)
		if err != nil {
			return err
		}
		l = l.WithField("undo", i)
		l.Debug("entering Undo")
	}

	receivingAccount, _ := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	undoObject := undo.GetActivityStreamsObject()
	if undoObject == nil {
		return errors.New("UNDO: no object set on vocab.ActivityStreamsUndo")
	}

	for iter := undoObject.Begin(); iter != undoObject.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}

		switch t.GetTypeName() {
		case ap.ActivityFollow:
			if err := f.undoFollow(ctx, receivingAccount, undo, t); err != nil {
				return err
			}
		case ap.ActivityLike:
			if err := f.undoLike(ctx, receivingAccount, undo, t); err != nil {
				return err
			}
		case ap.ActivityAnnounce:
			// todo: UNDO BOOST/REBLOG/ANNOUNCE
		case ap.ActivityBlock:
			if err := f.undoBlock(ctx, receivingAccount, undo, t); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *federatingDB) undoFollow(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	Follow, ok := t.(vocab.ActivityStreamsFollow)
	if !ok {
		return errors.New("undoFollow: couldn't parse vocab.Type into vocab.ActivityStreamsFollow")
	}

	// Make sure the undo actor owns the target.
	if !sameActor(undo.GetActivityStreamsActor(), Follow.GetActivityStreamsActor()) {
		// Ignore this Activity.
		return nil
	}

	follow, err := f.typeConverter.ASFollowToFollow(ctx, Follow)
	if err != nil {
		return fmt.Errorf("undoFollow: error converting ActivityStreams Follow to follow: %w", err)
	}

	// Ensure addressee is follow target.
	if follow.TargetAccountID != receivingAccount.ID {
		// Ignore this Activity.
		return nil
	}

	// Delete any existing follow.
	if err := f.state.DB.DeleteFollowByURI(ctx, follow.URI); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("undoFollow: db error removing follow: %w", err)
	}

	// Delete any existing follow request.
	if err := f.state.DB.DeleteFollowRequestByURI(ctx, follow.URI); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("undoFollow: db error removing follow request: %w", err)
	}

	log.Debug(ctx, "follow undone")
	return nil
}

func (f *federatingDB) undoLike(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	Like, ok := t.(vocab.ActivityStreamsLike)
	if !ok {
		return errors.New("undoLike: couldn't parse vocab.Type into vocab.ActivityStreamsLike")
	}

	// Make sure the undo actor owns the target.
	if !sameActor(undo.GetActivityStreamsActor(), Like.GetActivityStreamsActor()) {
		// Ignore this Activity.
		return nil
	}

	fave, err := f.typeConverter.ASLikeToFave(ctx, Like)
	if err != nil {
		return fmt.Errorf("undoLike: error converting ActivityStreams Like to fave: %w", err)
	}

	// Ensure addressee is follow target.
	if fave.TargetAccountID != receivingAccount.ID {
		// Ignore this Activity.
		return nil
	}

	// Ignore URI on likes, since we often
	// get multiple likes with the same target
	// with different URIs. Instead, we'll
	// select using account and target status.
	fave, err = f.state.DB.GetStatusFave(ctx, fave.AccountID, fave.StatusID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// We didn't have a like/fave
			// for this combo anyway, ignore.
			return nil
		}
		aaaaaaaaaa
	}

	// Delete any existing FOLLOW.
	if err := f.state.DB.DeleteFollowByURI(ctx, fave.URI); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("undoFollow: db error removing follow: %w", err)
	}

	// Delete any existing FOLLOW REQUEST.
	if err := f.state.DB.DeleteFollowRequestByURI(ctx, fave.URI); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("undoFollow: db error removing follow request: %w", err)
	}

	log.Debug(ctx, "like undone")
	return nil
}

func (f *federatingDB) undoBlock(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	Block, ok := t.(vocab.ActivityStreamsBlock)
	if !ok {
		return errors.New("undoBlock: couldn't parse vocab.Type into vocab.ActivityStreamsBlock")
	}

	// Make sure the undo actor owns the target.
	if !sameActor(undo.GetActivityStreamsActor(), Block.GetActivityStreamsActor()) {
		// Ignore this Activity.
		return nil
	}

	block, err := f.typeConverter.ASBlockToBlock(ctx, Block)
	if err != nil {
		return fmt.Errorf("undoBlock: error converting ActivityStreams Block to block: %w", err)
	}

	// Ensure addressee is block target.
	if block.TargetAccountID != receivingAccount.ID {
		// Ignore this Activity.
		return nil
	}

	// Delete any existing BLOCK
	if err := f.state.DB.DeleteBlockByURI(ctx, block.URI); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("undoBlock: db error removing block: %w", err)
	}

	log.Debug(ctx, "block undone")
	return nil
}
