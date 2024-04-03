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

package queue

import (
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Queues ...
type Queues struct {
	// APRequests ...
	APRequests StructQueue[*APRequest]
}

// Init will re(initialize) queues. NOTE: the queue
// MUST NOT be in use anywhere, this is not thread-safe.
func (q *Queues) Init() {
	log.Infof(nil, "init: %p", q)

	q.initHTTPRequest()
}

func (q *Queues) initHTTPRequest() {
	q.APRequests.Init(structr.QueueConfig[*APRequest]{
		Indices: []structr.IndexConfig{
			{Fields: "ObjectID", Multiple: true},
			{Fields: "Request.URL.Host", Multiple: true},
		},
	})
}
