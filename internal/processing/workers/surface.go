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

package workers

import (
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/filter/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/processing/conversations"
	"code.superseriousbusiness.org/gotosocial/internal/processing/stream"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/webpush"
)

// Surface wraps functions for 'surfacing' the result
// of processing a message, eg:
//   - timelining a status
//   - removing a status from timelines
//   - sending a notification to a user
//   - sending an email
type Surface struct {
	State         *state.State
	Converter     *typeutils.Converter
	Stream        *stream.Processor
	VisFilter     *visibility.Filter
	MuteFilter    *mutes.Filter
	StatusFilter  *status.Filter
	EmailSender   email.Sender
	WebPushSender webpush.Sender
	Conversations *conversations.Processor
}
