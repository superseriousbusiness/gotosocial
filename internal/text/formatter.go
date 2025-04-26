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

package text

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// FormatFunc is fulfilled by FromPlain,
// FromPlainNoParagraph, and FromMarkdown.
type FormatFunc func(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	text string,
) *FormatResult

// Formatter wraps logic and functions for parsing
// statuses and other text input into nice html.
type Formatter struct {
	db db.DB
}

// NewFormatter returns a new Formatter.
func NewFormatter(db db.DB) *Formatter {
	return &Formatter{
		db: db,
	}
}

type FormatResult struct {
	HTML     string
	Mentions []*gtsmodel.Mention
	Tags     []*gtsmodel.Tag
	Emojis   []*gtsmodel.Emoji
}
