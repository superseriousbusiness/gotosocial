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

package ap_test

import (
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

var (
	// NOTE: the below aren't actually tests that are run,
	// we just move them into an _test.go file to declutter
	// the main interfaces.go file, which is already long.

	// Compile-time checks for Activityable interface methods.
	_ ap.Activityable = (vocab.ActivityStreamsAccept)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsTentativeAccept)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsAdd)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsCreate)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsDelete)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsFollow)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsIgnore)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsJoin)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsLeave)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsLike)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsOffer)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsInvite)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsReject)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsTentativeReject)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsRemove)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsUndo)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsUpdate)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsView)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsListen)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsRead)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsMove)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsAnnounce)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsBlock)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsFlag)(nil)
	_ ap.Activityable = (vocab.ActivityStreamsDislike)(nil)

	// the below intransitive activities don't fit the interface definition because they're
	// missing an attached object (as the activity itself contains the details), but we don't
	// actually end up using them so it's  simpler to just comment them out and not have to do
	// a WithObject{} interface check on every single incoming activity:
	//
	// _ Activityable = (vocab.ActivityStreamsArrive)(nil)
	// _ Activityable = (vocab.ActivityStreamsTravel)(nil)
	// _ Activityable = (vocab.ActivityStreamsQuestion)(nil)

	// Compile-time checks for Accountable interface methods.
	_ ap.Accountable = (vocab.ActivityStreamsPerson)(nil)
	_ ap.Accountable = (vocab.ActivityStreamsApplication)(nil)
	_ ap.Accountable = (vocab.ActivityStreamsOrganization)(nil)
	_ ap.Accountable = (vocab.ActivityStreamsService)(nil)
	_ ap.Accountable = (vocab.ActivityStreamsGroup)(nil)

	// Compile-time checks for Statusable interface methods.
	_ ap.Statusable = (vocab.ActivityStreamsArticle)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsDocument)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsImage)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsVideo)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsNote)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsPage)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsEvent)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsPlace)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsProfile)(nil)
	_ ap.Statusable = (vocab.ActivityStreamsQuestion)(nil)

	// Compile-time checks for Pollable interface methods.
	_ ap.Pollable = (vocab.ActivityStreamsQuestion)(nil)

	// Compile-time checks for PollOptionable interface methods.
	_ ap.PollOptionable = (vocab.ActivityStreamsNote)(nil)

	// Compile-time checks for Acceptable interface methods.
	_ ap.Acceptable = (vocab.ActivityStreamsAccept)(nil)
)
