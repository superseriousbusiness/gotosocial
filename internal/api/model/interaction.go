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

// InteractionRequest represents a pending/requested interaction of type favourite, reply, or reblog, awaiting approval by the user being interacted with.
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
	InteractingAccount *Account `json:"interacting_account"`
	// Status targeted by the requested interaction.
	InteractedStatus *Status `json:"interacted_status"`
	// If type=reply, this field will be set to the reply that is awaiting approval. If type=favourite, or type=reblog, the field will be omitted.
	Reply *Status `json:"reply,omitempty"`
}

// InteractionApproval represents an interaction of type favourite, reply, or reblog which has been approved by the user being interacted with.
//
// swagger:model interactionApproval
type InteractionApproval struct {
	// The id of the interaction approval in the database.
	ID string `json:"id"`
	// The type of interaction that this interaction approval pertains to.
	//
	//	`favourite` - Someone favourited a status.
	//	`reply` - Someone replied to a status.
	//	`reblog` - Someone reblogged / boosted a status.
	Type string `json:"type"`
	// The timestamp of the interaction approval (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The account that performed the interaction.
	InteractingAccount *Account `json:"interacting_account"`
	// Status targeted by the approved interaction.
	InteractedStatus *Status `json:"interacted_status"`
	// If type=reply, this field will be set to the approved reply. If type=favourite, or type=reblog, the field will be omitted.
	Reply *Status `json:"reply,omitempty"`
}

// InteractionRejection represents an interaction of type favourite, reply, or reblog which has been rejected by the user being interacted with.
//
// swagger:model interactionRejection
type InteractionRejection struct {
	// The id of the interaction rejection in the database.
	ID string `json:"id"`
	// The type of interaction that this interaction rejection pertains to.
	//
	//	`favourite` - Someone favourited a status.
	//	`reply` - Someone replied to a status.
	//	`reblog` - Someone reblogged / boosted a status.
	Type string `json:"type"`
	// The timestamp of the interaction rejection (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The account that performed the interaction.
	InteractingAccount *Account `json:"interacting_account"`
	// Status targeted by the rejected interaction.
	InteractedStatus *Status `json:"interacted_status"`
}
