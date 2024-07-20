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

package model

// ThreadContext models the tree or
// "thread" around a given status.
//
// swagger:model threadContext
type ThreadContext struct {
	// Parents in the thread.
	Ancestors []Status `json:"ancestors"`
	// Children in the thread.
	Descendants []Status `json:"descendants"`
}

type WebThreadContext struct {
	// Status around which this
	// thread ctx was constructed.
	Status *WebStatus

	// Ordered slice of statuses
	// for rendering in template.
	//
	// Includes ancestors, target
	// status, and descendants.
	Statuses []*WebStatus

	// Total length of
	// the main thread.
	ThreadLength int

	// Number of entries in
	// the main thread shown.
	ThreadShown int

	// Number of statuses hidden
	// from the main thread (not
	// visible to requester etc).
	ThreadHidden int

	// Total number of replies
	// in the replies section.
	ThreadReplies int

	// Number of replies shown.
	ThreadRepliesShown int

	// Number of replies hidden.
	ThreadRepliesHidden int
}
