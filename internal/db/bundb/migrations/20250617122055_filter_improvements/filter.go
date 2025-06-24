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

import (
	"regexp"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// smallint is the largest size supported
// by a PostgreSQL SMALLINT, since an SQLite
// SMALLINT is actually variable in size.
type smallint int16

// enumType is the type we (at least, should) use
// for database enum types, as smallest int size.
type enumType smallint

// bitFieldType is the type we use
// for database int bit fields, at
// least where the smallest int size
// will suffice for number of fields.
type bitFieldType smallint

// FilterContext represents the
// context in which a Filter applies.
//
// These are used as bit-field masks to determine
// which are enabled in a FilterContexts bit field,
// as well as to signify internally any particular
// context in which a status should be filtered in.
type FilterContext bitFieldType

const (
	// FilterContextNone means no filters should
	// be applied, this is for internal use only.
	FilterContextNone FilterContext = 0

	// FilterContextHome means this status is being
	// filtered as part of a home or list timeline.
	FilterContextHome FilterContext = 1 << 1

	// FilterContextNotifications means this status is
	// being filtered as part of the notifications timeline.
	FilterContextNotifications FilterContext = 1 << 2

	// FilterContextPublic means this status is
	// being filtered as part of a public or tag timeline.
	FilterContextPublic FilterContext = 1 << 3

	// FilterContextThread means this status is
	// being filtered as part of a thread's context.
	FilterContextThread FilterContext = 1 << 4

	// FilterContextAccount means this status is
	// being filtered as part of an account's statuses.
	FilterContextAccount FilterContext = 1 << 5
)

// FilterContexts stores multiple contexts
// in which a Filter applies as bits in an int.
type FilterContexts bitFieldType

// Applies returns whether receiving FilterContexts applies in FilterContexts.
func (ctxs FilterContexts) Applies(ctx FilterContext) bool {
	switch ctx {
	case FilterContextHome:
		return ctxs.Home()
	case FilterContextNotifications:
		return ctxs.Notifications()
	case FilterContextPublic:
		return ctxs.Public()
	case FilterContextThread:
		return ctxs.Thread()
	case FilterContextAccount:
		return ctxs.Account()
	default:
		return false
	}
}

// Home returns whether FilterContextHome is set.
func (ctxs FilterContexts) Home() bool {
	return ctxs&FilterContexts(FilterContextHome) != 0
}

// SetHome will set the FilterContextHome bit.
func (ctxs *FilterContexts) SetHome() {
	*ctxs |= FilterContexts(FilterContextHome)
}

// UnsetHome will unset the FilterContextHome bit.
func (ctxs *FilterContexts) UnsetHome() {
	*ctxs &= ^FilterContexts(FilterContextHome)
}

// Notifications returns whether FilterContextNotifications is set.
func (ctxs FilterContexts) Notifications() bool {
	return ctxs&FilterContexts(FilterContextNotifications) != 0
}

// SetNotifications will set the FilterContextNotifications bit.
func (ctxs *FilterContexts) SetNotifications() {
	*ctxs |= FilterContexts(FilterContextNotifications)
}

// UnsetNotifications will unset the FilterContextNotifications bit.
func (ctxs *FilterContexts) UnsetNotifications() {
	*ctxs &= ^FilterContexts(FilterContextNotifications)
}

// Public returns whether FilterContextPublic is set.
func (ctxs FilterContexts) Public() bool {
	return ctxs&FilterContexts(FilterContextPublic) != 0
}

// SetPublic will set the FilterContextPublic bit.
func (ctxs *FilterContexts) SetPublic() {
	*ctxs |= FilterContexts(FilterContextPublic)
}

// UnsetPublic will unset the FilterContextPublic bit.
func (ctxs *FilterContexts) UnsetPublic() {
	*ctxs &= ^FilterContexts(FilterContextPublic)
}

// Thread returns whether FilterContextThread is set.
func (ctxs FilterContexts) Thread() bool {
	return ctxs&FilterContexts(FilterContextThread) != 0
}

// SetThread will set the FilterContextThread bit.
func (ctxs *FilterContexts) SetThread() {
	*ctxs |= FilterContexts(FilterContextThread)
}

// UnsetThread will unset the FilterContextThread bit.
func (ctxs *FilterContexts) UnsetThread() {
	*ctxs &= ^FilterContexts(FilterContextThread)
}

// Account returns whether FilterContextAccount is set.
func (ctxs FilterContexts) Account() bool {
	return ctxs&FilterContexts(FilterContextAccount) != 0
}

// SetAccount will set / unset the FilterContextAccount bit.
func (ctxs *FilterContexts) SetAccount() {
	*ctxs |= FilterContexts(FilterContextAccount)
}

// UnsetAccount will unset the FilterContextAccount bit.
func (ctxs *FilterContexts) UnsetAccount() {
	*ctxs &= ^FilterContexts(FilterContextAccount)
}

// FilterAction represents the action
// to take on a filtered status.
type FilterAction enumType

const (
	// FilterActionNone filters should not exist, except
	// internally, for partially constructed or invalid filters.
	FilterActionNone FilterAction = 0

	// FilterActionWarn means that the
	// status should be shown behind a warning.
	FilterActionWarn FilterAction = 1

	// FilterActionHide means that the status should
	// be removed from timeline results entirely.
	FilterActionHide FilterAction = 2
)

// Filter stores a filter created by a local account.
type Filter struct {
	ID         string           `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                            // id of this item in the database
	ExpiresAt  time.Time        `bun:"type:timestamptz,nullzero"`                                           // Time filter should expire. If null, should not expire.
	AccountID  string           `bun:"type:CHAR(26),notnull,nullzero,unique:filters_account_id_title_uniq"` // ID of the local account that created the filter.
	Title      string           `bun:",nullzero,notnull,unique:filters_account_id_title_uniq"`              // The name of the filter.
	Action     FilterAction     `bun:",nullzero,notnull,default:0"`                                         // The action to take.
	Keywords   []*FilterKeyword `bun:"-"`                                                                   // Keywords for this filter.
	KeywordIDs []string         `bun:"keywords,array"`                                                      //
	Statuses   []*FilterStatus  `bun:"-"`                                                                   // Statuses for this filter.
	StatusIDs  []string         `bun:"statuses,array"`                                                      //
	Contexts   FilterContexts   `bun:",nullzero,notnull,default:0"`                                         // Which contexts does this filter apply in?
}

// FilterKeyword stores a single keyword to filter statuses against.
type FilterKeyword struct {
	ID        string         `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                     // id of this item in the database
	FilterID  string         `bun:"type:CHAR(26),notnull,nullzero,unique:filter_keywords_filter_id_keyword_uniq"` // ID of the filter that this keyword belongs to.
	Keyword   string         `bun:",nullzero,notnull,unique:filter_keywords_filter_id_keyword_uniq"`              // The keyword or phrase to filter against.
	WholeWord *bool          `bun:",nullzero,notnull,default:false"`                                              // Should the filter consider word boundaries?
	Regexp    *regexp.Regexp `bun:"-"`                                                                            // pre-prepared regular expression
}

// Compile will compile this FilterKeyword as a prepared regular expression.
func (k *FilterKeyword) Compile() (err error) {
	var (
		wordBreakStart string
		wordBreakEnd   string
	)

	if util.PtrOrZero(k.WholeWord) {
		// Either word boundary or
		// whitespace or start of line.
		wordBreakStart = `(?:\b|\s|^)`

		// Either word boundary or
		// whitespace or end of line.
		wordBreakEnd = `(?:\b|\s|$)`
	}

	// Compile keyword filter regexp.
	quoted := regexp.QuoteMeta(k.Keyword)
	k.Regexp, err = regexp.Compile(`(?i)` + wordBreakStart + quoted + wordBreakEnd)
	return // caller is expected to wrap this error
}

// FilterStatus stores a single status to filter.
type FilterStatus struct {
	ID       string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                       // id of this item in the database
	FilterID string `bun:"type:CHAR(26),notnull,nullzero,unique:filter_statuses_filter_id_status_id_uniq"` // ID of the filter that this keyword belongs to.
	StatusID string `bun:"type:CHAR(26),notnull,nullzero,unique:filter_statuses_filter_id_status_id_uniq"` // ID of the status to filter.
}
