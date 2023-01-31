/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package text

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Formatter wraps some logic and functions for parsing statuses and other text input into nice html.
// Each of the member functions returns a struct containing the formatted HTML and any tags, mentions, and
// emoji that were found in the text.
type Formatter interface {
	// FromPlain parses an HTML text from a plaintext.
	FromPlain(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, plain string) *FormatResult
	// FromMarkdown parses an HTML text from a markdown-formatted text.
	FromMarkdown(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, md string) *FormatResult
	// FromPlainEmojiOnly parses an HTML text from a plaintext, only parsing emojis and not mentions etc.
	FromPlainEmojiOnly(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, plain string) *FormatResult
}

type FormatFunc func(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, text string) *FormatResult

type formatter struct {
	db db.DB
}

// NewFormatter returns a new Formatter interface for parsing statuses and other text input into nice html.
func NewFormatter(db db.DB) Formatter {
	return &formatter{
		db: db,
	}
}

type FormatResult struct {
	HTML     string
	Mentions []*gtsmodel.Mention
	Tags     []*gtsmodel.Tag
	Emojis   []*gtsmodel.Emoji
}
