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

// FilterV2 represents a user-defined filter for determining which statuses should not be shown to the user.
// v2 filters have names and can include multiple phrases and status IDs to filter.
//
// swagger:model filterV2
//
// ---
// tags:
// - filters
type FilterV2 struct {
	// The ID of the filter in the database.
	ID string `json:"id"`
	// The name of the filter.
	//
	// Example: Linux Words
	Title string `json:"title"`
	// The contexts in which the filter should be applied.
	//
	// Minimum items: 1
	// Unique: true
	// Enum:
	//	- home
	//	- notifications
	//	- public
	//	- thread
	//	- account
	// Example: ["home", "public"]
	Context []FilterContext `json:"context"`
	// When the filter should no longer be applied. Null if the filter does not expire.
	//
	// Example: 2024-02-01T02:57:49Z
	ExpiresAt *string `json:"expires_at"`
	// The action to be taken when a status matches this filter.
	// Enum:
	//	- warn
	//	- hide
	FilterAction FilterAction `json:"filter_action"`
	// The keywords grouped under this filter.
	Keywords []FilterKeyword `json:"keywords"`
	// The statuses grouped under this filter.
	Statuses []FilterStatus `json:"statuses"`
}

// FilterAction is the action to apply to statuses matching a filter.
type FilterAction string

const (
	// FilterActionNone filters should not exist, except internally, for partially constructed or invalid filters.
	FilterActionNone FilterAction = ""
	// FilterActionWarn filters will include this status in API results with a warning.
	FilterActionWarn FilterAction = "warn"
	// FilterActionHide filters will remove this status from API results.
	FilterActionHide FilterAction = "hide"
)

// FilterKeyword represents text to filter within a v2 filter.
//
// swagger:model filterKeyword
//
// ---
// tags:
// - filters
type FilterKeyword struct {
	// The ID of the filter keyword entry in the database.
	ID string `json:"id"`
	// The text to be filtered.
	//
	// Example: fnord
	Keyword string `json:"keyword"`
	// Should the filter keyword consider word boundaries?
	//
	// Example: true
	WholeWord bool `json:"whole_word"`
}

// FilterStatus represents a single status to filter within a v2 filter.
//
// swagger:model filterStatus
//
// ---
// tags:
// - filters
type FilterStatus struct {
	// The ID of the filter status entry in the database.
	ID string `json:"id"`
	// The status ID to be filtered.
	StatusID string `json:"phrase"`
}

// FilterCreateRequestV2 captures params for creating a v2 filter.
//
// swagger:ignore
type FilterCreateRequestV2 struct {
	// The name of the filter.
	//
	// Required: true
	// Example: fnord
	Title string `form:"title" json:"title" xml:"title"`
	// The contexts in which the filter should be applied.
	//
	// Required: true
	// Minimum length: 1
	// Unique: true
	// Enum: home,notifications,public,thread,account
	// Example: ["home", "public"]
	Context []FilterContext `form:"context[]" json:"context" xml:"context"`
	// The action to be taken when a status matches this filter. If omitted, defaults to warn.
	// Enum:
	//	- warn
	//	- hide
	// Example: warn
	FilterAction *FilterAction `form:"filter_action" json:"filter_action" xml:"filter_action"`

	// Number of seconds from now that the filter should expire. If omitted, filter never expires.
	ExpiresIn *int `json:"-" form:"expires_in" xml:"expires_in"`
	// Number of seconds from now that the filter should expire. If omitted, filter never expires.
	//
	// Example: 86400
	ExpiresInI Nullable[any] `json:"expires_in"`

	// Keywords to be added to the newly created filter.
	Keywords []FilterKeywordCreateUpdateRequest `form:"-" json:"keywords_attributes" xml:"keywords_attributes"`
	// Form data version of Keywords[].Keyword.
	KeywordsAttributesKeyword []string `form:"keywords_attributes[][keyword]" json:"-" xml:"-"`
	// Form data version of Keywords[].WholeWord.
	KeywordsAttributesWholeWord []bool `form:"keywords_attributes[][whole_word]" json:"-" xml:"-"`

	// Statuses to be added to the newly created filter.
	Statuses []FilterStatusCreateRequest `form:"-" json:"statuses_attributes" xml:"statuses_attributes"`
	// Form data version of Statuses[].StatusID.
	StatusesAttributesStatusID []string `form:"statuses_attributes[][status_id]" json:"-" xml:"-"`
}

// FilterKeywordCreateUpdateRequest captures params for creating or updating a filter keyword while creating a v2 filter or as a standalone operation.
//
// swagger:ignore
type FilterKeywordCreateUpdateRequest struct {
	// The text to be filtered.
	//
	// Example: fnord
	// Maximum length: 40
	Keyword string `form:"keyword" json:"keyword" xml:"keyword"`
	// Should the filter keyword consider word boundaries?
	//
	// Example: true
	WholeWord *bool `form:"whole_word" json:"whole_word" xml:"whole_word"`
}

// FilterStatusCreateRequest captures params for a status while creating a v2 filter or filter status.
//
// swagger:ignore
type FilterStatusCreateRequest struct {
	// The status ID to be filtered.
	StatusID string `form:"status_id" json:"status_id" xml:"status_id"`
}

// FilterUpdateRequestV2 captures params for creating a v2 filter.
//
// swagger:ignore
type FilterUpdateRequestV2 struct {
	// The name of the filter.
	//
	// Example: illuminati nonsense
	Title *string `form:"title" json:"title" xml:"title"`
	// The contexts in which the filter should be applied.
	//
	// Minimum length: 1
	// Unique: true
	// Enum: home,notifications,public,thread,account
	// Example: ["home", "public"]
	Context *[]FilterContext `form:"context[]" json:"context" xml:"context"`
	// The action to be taken when a status matches this filter.
	// Enum:
	//	- warn
	//	- hide
	// Example: warn
	FilterAction *FilterAction `form:"filter_action" json:"filter_action" xml:"filter_action"`

	// Number of seconds from now that the filter should expire. If omitted, filter never expires.
	ExpiresIn *int `json:"-" form:"expires_in" xml:"expires_in"`
	// Number of seconds from now that the filter should expire. If omitted, filter never expires.
	//
	// Example: 86400
	ExpiresInI Nullable[any] `json:"expires_in"`

	// Keywords to be added to the filter, modified, or removed.
	Keywords []FilterKeywordCreateUpdateDeleteRequest `form:"-" json:"keywords_attributes" xml:"keywords_attributes"`
	// Form data version of Keywords[].ID.
	KeywordsAttributesID []string `form:"keywords_attributes[][id]" json:"-" xml:"-"`
	// Form data version of Keywords[].Keyword.
	KeywordsAttributesKeyword []string `form:"keywords_attributes[][keyword]" json:"-" xml:"-"`
	// Form data version of Keywords[].WholeWord.
	KeywordsAttributesWholeWord []bool `form:"keywords_attributes[][whole_word]" json:"-" xml:"-"`
	// Form data version of Keywords[].Destroy.
	KeywordsAttributesDestroy []bool `form:"keywords_attributes[][_destroy]" json:"-" xml:"-"`

	// Statuses to be added to the filter, or removed.
	Statuses []FilterStatusCreateDeleteRequest `form:"-" json:"statuses_attributes" xml:"statuses_attributes"`
	// Form data version of Statuses[].ID.
	StatusesAttributesID []string `form:"statuses_attributes[][id]" json:"-" xml:"-"`
	// Form data version of Statuses[].ID.
	StatusesAttributesStatusID []string `form:"statuses_attributes[][status_id]" json:"-" xml:"-"`
	// Form data version of Statuses[].Destroy.
	StatusesAttributesDestroy []bool `form:"statuses_attributes[][_destroy]" json:"-" xml:"-"`
}

// FilterKeywordCreateUpdateDeleteRequest captures params for creating, updating, or deleting a keyword while updating a v2 filter.
//
// swagger:ignore
type FilterKeywordCreateUpdateDeleteRequest struct {
	// The ID of the filter keyword entry in the database.
	// Optional: use to modify or delete an existing keyword instead of adding a new one.
	ID *string `json:"id" xml:"id"`
	// The text to be filtered.
	//
	// Example: fnord
	// Maximum length: 40
	Keyword *string `json:"keyword" xml:"keyword"`
	// Should the filter keyword consider word boundaries?
	//
	// Example: true
	WholeWord *bool `json:"whole_word" xml:"whole_word"`
	// Remove this filter keyword. Requires an ID.
	Destroy *bool `json:"_destroy" xml:"_destroy"`
}

// FilterStatusCreateDeleteRequest captures params for creating or deleting a status while updating a v2 filter.
//
// swagger:ignore
type FilterStatusCreateDeleteRequest struct {
	// The ID of the filter status entry in the database.
	// Optional: use to delete an existing status instead of adding a new one.
	ID *string `json:"id" xml:"id"`
	// The status ID to be filtered.
	StatusID *string `json:"status_id" xml:"status_id"`
	// Remove this filter status. Requires an ID.
	Destroy *bool `json:"_destroy" xml:"_destroy"`
}
