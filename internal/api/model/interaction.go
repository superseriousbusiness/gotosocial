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

// InteractionRequest represents a pending, approved, or rejected interaction of type favourite, reply, or reblog.
//
// swagger:model interactionRequest
type InteractionRequest struct {
	// The id of the interaction request in the database.
	ID string `json:"id"`
	// The type of interaction that this interaction request pertains to.
	//
	//	`favourite` - Someone favourited a status.
	//	`reply` - Someone replied to a status.
	//	`reblog` - Someone reblogged / boosted a status.
	Type string `json:"type"`
	// The timestamp of the interaction request (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The account that performed the interaction.
	Account *Account `json:"account"`
	// Status targeted by the requested interaction.
	Status *Status `json:"status"`
	// If type=reply, this field will be set to the reply that is awaiting approval. If type=favourite, or type=reblog, the field will be omitted.
	Reply *Status `json:"reply,omitempty"`
	// The timestamp that the interaction request was accepted (ISO 8601 Datetime). Field omitted if request not accepted (yet).
	AcceptedAt string `json:"accepted_at,omitempty"`
	// The timestamp that the interaction request was rejected (ISO 8601 Datetime). Field omitted if request not rejected (yet).
	RejectedAt string `json:"rejected_at,omitempty"`
	// URI of the Accept or Reject. Only set if accepted_at or rejected_at is set, else omitted.
	URI string `json:"uri,omitempty"`
}
