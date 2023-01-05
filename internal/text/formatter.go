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
type Formatter interface {
	// FromPlain parses an HTML text from a plaintext.
	FromPlain(ctx context.Context, plain string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string
	// FromMarkdown parses an HTML text from a markdown-formatted text.
	FromMarkdown(ctx context.Context, md string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag, emojis []*gtsmodel.Emoji) string

	// ReplaceTags takes a piece of text and a slice of tags, and returns the same text with the tags nicely formatted as hrefs.
	ReplaceTags(ctx context.Context, in string, tags []*gtsmodel.Tag) string
	// ReplaceMentions takes a piece of text and a slice of mentions, and returns the same text with the mentions nicely formatted as hrefs.
	ReplaceMentions(ctx context.Context, in string, mentions []*gtsmodel.Mention) string
	// ReplaceLinks takes a piece of text, finds all recognizable links in that text, and replaces them with hrefs.
	ReplaceLinks(ctx context.Context, in string) string
}

type formatter struct {
	db db.DB
}

// NewFormatter returns a new Formatter interface for parsing statuses and other text input into nice html.
func NewFormatter(db db.DB) Formatter {
	return &formatter{
		db: db,
	}
}
