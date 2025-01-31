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
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

/*
	MENTION PARSER STUFF
*/

// mention fulfils the goldmark
// ast.Node interface.
type mention struct {
	ast.BaseInline
	Segment text.Segment
}

var kindMention = ast.NewNodeKind("Mention")

func (n *mention) Kind() ast.NodeKind {
	return kindMention
}

func (n *mention) Dump(source []byte, level int) {
	fmt.Printf("%sMention: %s\n", strings.Repeat("    ", level), string(n.Segment.Value(source)))
}

// newMention creates a goldmark ast.Node
// from a text.Segment. The contained segment
// is used in rendering.
func newMention(s text.Segment) *mention {
	return &mention{
		BaseInline: ast.BaseInline{},
		Segment:    s,
	}
}

// mentionParser fulfils the goldmark
// parser.InlineParser interface.
type mentionParser struct{}

// Mention parsing is triggered by the `@` symbol
// which appears at the beginning of a mention.
func (p *mentionParser) Trigger() []byte {
	return []byte{'@'}
}

func (p *mentionParser) Parse(
	_ ast.Node,
	block text.Reader,
	_ parser.Context,
) ast.Node {
	// If preceding character is not a valid boundary
	// character, then this cannot be a valid mention.
	if !isMentionBoundary(block.PrecendingCharacter()) {
		return nil
	}

	line, segment := block.PeekLine()

	// Ascertain location of mention in the line
	// that starts with the trigger character.
	loc := regexes.MentionFinder.FindIndex(line)
	if loc == nil || loc[0] != 0 {
		// Noop if not found or
		// not found at start.
		return nil
	}

	// Advance the block to
	// the end of the mention.
	block.Advance(loc[1])

	// mention ast.Node spans from the
	// beginning of this segment up to
	// the last character of the mention.
	return newMention(
		segment.WithStop(
			segment.Start + loc[1],
		),
	)
}

/*
	HASHTAG PARSER STUFF
*/

// hashtag fulfils the goldmark
// ast.Node interface.
type hashtag struct {
	ast.BaseInline
	Segment text.Segment
}

var kindHashtag = ast.NewNodeKind("Hashtag")

func (n *hashtag) Kind() ast.NodeKind {
	return kindHashtag
}

func (n *hashtag) Dump(source []byte, level int) {
	fmt.Printf("%sHashtag: %s\n", strings.Repeat("    ", level), string(n.Segment.Value(source)))
}

// newHashtag creates a goldmark ast.Node
// from a text.Segment. The contained segment
// is used in rendering.
func newHashtag(s text.Segment) *hashtag {
	return &hashtag{
		BaseInline: ast.BaseInline{},
		Segment:    s,
	}
}

type hashtagParser struct{}

// Hashtag parsing is triggered by a '#' symbol
// which appears at the beginning of a hashtag.
func (p *hashtagParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *hashtagParser) Parse(
	_ ast.Node,
	block text.Reader,
	_ parser.Context,
) ast.Node {
	// If preceding character is not a valid boundary
	// character, then this cannot be a valid hashtag.
	if !isHashtagBoundary(block.PrecendingCharacter()) {
		return nil
	}

	var (
		line, segment = block.PeekLine()
		lineStr       = string(line)
		lineStrLen    = len(lineStr)
	)

	if lineStrLen <= 1 {
		// This is probably just
		// a lonely '#' char.
		return nil
	}

	// Iterate through the runes in the detected
	// hashtag string until we reach either:
	//   - A weird character (bad).
	//   - The end of the hashtag (ok).
	//   - The end of the string (also ok).
	for i, r := range lineStr {
		switch {
		case r == '#' && i == 0:
			// Ignore initial '#'.
			continue

		case !isPermittedInHashtag(r) &&
			!isHashtagBoundary(r):
			// Weird non-boundary character
			// in the hashtag. Don't trust it.
			return nil

		case isHashtagBoundary(r):
			// Reached closing hashtag
			// boundary. Advance block
			// to the end of the hashtag.
			block.Advance(i)

			// hashtag ast.Node spans from
			// the beginning of this segment
			// up to the boundary character.
			return newHashtag(
				segment.WithStop(
					segment.Start + i,
				),
			)
		}
	}

	// No invalid or boundary characters before the
	// end of the line: it's all hashtag, baby 😎
	//
	// Advance block to the end of the segment.
	block.Advance(segment.Len())

	// hashtag ast.Node spans
	// the entire segment.
	return newHashtag(segment)
}

/*
	EMOJI PARSER STUFF
*/

// emoji fulfils the goldmark
// ast.Node interface.
type emoji struct {
	ast.BaseInline
	Segment text.Segment
}

var kindEmoji = ast.NewNodeKind("Emoji")

func (n *emoji) Kind() ast.NodeKind {
	return kindEmoji
}

func (n *emoji) Dump(source []byte, level int) {
	fmt.Printf("%sEmoji: %s\n", strings.Repeat("    ", level), string(n.Segment.Value(source)))
}

// newEmoji creates a goldmark ast.Node
// from a text.Segment. The contained
// segment is used in rendering.
func newEmoji(s text.Segment) *emoji {
	return &emoji{
		BaseInline: ast.BaseInline{},
		Segment:    s,
	}
}

type emojiParser struct{}

// Emoji parsing is triggered by a ':' char
// which appears at the start of the emoji.
func (p *emojiParser) Trigger() []byte {
	return []byte{':'}
}

func (p *emojiParser) Parse(
	_ ast.Node,
	block text.Reader,
	_ parser.Context,
) ast.Node {
	line, segment := block.PeekLine()

	// Ascertain location of emoji in the line
	// that starts with the trigger character.
	loc := regexes.EmojiFinder.FindIndex(line)
	if loc == nil || loc[0] != 0 {
		// Noop if not found or
		// not found at start.
		return nil
	}

	// Advance the block to
	// the end of the emoji.
	block.Advance(loc[1])

	// emoji ast.Node spans from the
	// beginning of this segment up to
	// the last character of the emoji.
	return newEmoji(
		segment.WithStop(
			segment.Start + loc[1],
		),
	)
}
