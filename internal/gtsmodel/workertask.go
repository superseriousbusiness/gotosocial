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

package gtsmodel

import "time"

type WorkerType uint8

const (
	DeliveryWorker  WorkerType = 1
	FederatorWorker WorkerType = 2
	ClientWorker    WorkerType = 3
)

// WorkerTask represents a queued worker task
// that was persisted to the database on shutdown.
// This is only ever used on startup to pickup
// where we left off, and on shutdown to prevent
// queued tasks from being lost. It is simply a
// means to store a blob of serialized task data.
type WorkerTask struct {
	ID         uint       `bun:",pk,autoincrement"`
	WorkerType WorkerType `bun:",notnull"`
	TaskData   []byte     `bun:",nullzero,notnull"`
	CreatedAt  time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
}
