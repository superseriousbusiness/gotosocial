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
	"fmt"
	"regexp"
	"strconv"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-byteutil"
)

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

// String returns human-readable form of FilterContext.
func (ctx FilterContext) String() string {
	switch ctx {
	case FilterContextNone:
		return ""
	case FilterContextHome:
		return "home"
	case FilterContextNotifications:
		return "notifications"
	case FilterContextPublic:
		return "public"
	case FilterContextThread:
		return "thread"
	case FilterContextAccount:
		return "account"
	default:
		panic(fmt.Sprintf("invalid filter context: %d", ctx))
	}
}

// FilterContexts stores multiple contexts
// in which a Filter applies as bits in an int.
type FilterContexts bitFieldType

// Applies returns whether receiving FilterContexts applies in FilterContexts.
func (ctxs FilterContexts) Applies(ctx FilterContext) bool {
	return ctxs&FilterContexts(ctx) != 0
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

// String returns a single human-readable form of FilterContexts.
func (ctxs FilterContexts) String() string {
	var buf byteutil.Buffer
	buf.Guarantee(72) // worst-case estimate
	buf.B = append(buf.B, '{')
	buf.B = append(buf.B, "home="...)
	buf.B = strconv.AppendBool(buf.B, ctxs.Home())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "notifications="...)
	buf.B = strconv.AppendBool(buf.B, ctxs.Notifications())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "public="...)
	buf.B = strconv.AppendBool(buf.B, ctxs.Public())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "thread="...)
	buf.B = strconv.AppendBool(buf.B, ctxs.Thread())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "account="...)
	buf.B = strconv.AppendBool(buf.B, ctxs.Account())
	buf.B = append(buf.B, '}')
	return buf.String()
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

	// FilterActionWarn means that the status should
	// be shown with its media attachments hidden/blurred.
	FilterActionBlur FilterAction = 3
)

// String returns human-readable form of FilterAction.
func (act FilterAction) String() string {
	switch act {
	case FilterActionNone:
		return ""
	case FilterActionWarn:
		return "warn"
	case FilterActionHide:
		return "hide"
	case FilterActionBlur:
		return "blur"
	default:
		panic(fmt.Sprintf("invalid filter action: %d", act))
	}
}

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

// KeywordsPopulated returns whether keywords
// are populated according to current KeywordIDs.
func (f *Filter) KeywordsPopulated() bool {
	if len(f.KeywordIDs) != len(f.Keywords) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range f.KeywordIDs {
		if f.Keywords[i].ID != id {
			return false
		}
	}
	return true
}

// StatusesPopulated returns whether statuses
// are populated according to current StatusIDs.
func (f *Filter) StatusesPopulated() bool {
	if len(f.StatusIDs) != len(f.Statuses) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range f.StatusIDs {
		if f.Statuses[i].ID != id {
			return false
		}
	}
	return true
}

// Expired returns whether the filter has expired at a given time.
// Filters without an expiration timestamp never expire.
func (f *Filter) Expired(now time.Time) bool {
	return !f.ExpiresAt.IsZero() && !f.ExpiresAt.After(now)
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
