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

package db

const (
	// DBTypePostgres represents an underlying POSTGRES database type.
	DBTypePostgres string = "POSTGRES"
)

// DB provides methods for interacting with an underlying database or other storage mechanism.
type DB interface {
	Account
	Admin
	AdvancedMigration
	Application
	Basic
	Conversation
	Domain
	Emoji
	HeaderFilter
	Instance
	Interaction
	Filter
	List
	Marker
	Media
	Mention
	Move
	Notification
	Poll
	Relationship
	Report
	Rule
	Search
	Session
	SinBinStatus
	Status
	StatusBookmark
	StatusEdit
	StatusFave
	Tag
	Thread
	Timeline
	User
	Tombstone
	WebPush
	WorkerTask
}
