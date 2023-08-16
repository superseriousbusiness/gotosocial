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

// InstanceRule represents a single instance rule.
//
// swagger:model instanceRule
type InstanceRule struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// InstanceRuleCreateRequest represents a request to create a new instance rule, made through the admin API.
//
// swagger:model instanceRuleCreateRequest
type InstanceRuleCreateRequest struct {
	Text string `form:"text" validation:"required"`
}

// InstanceRuleUpdateRequest represents a request to update the text of an instance rule, made through the admin API.
//
// swagger:model instanceRuleUpdateRequest
type InstanceRuleUpdateRequest struct {
	ID   string `form:"id"`
	Text string `form:"text"`
}
