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
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	mdutil "github.com/yuin/goldmark/util"
)

// A goldmark extension that parses potential mentions and hashtags separately from regular
// text, so that they stay as one contiguous text fragment in the AST, and then renders
// them separately too, to avoid scanning normal text for mentions and tags.

// mention and hashtag fulfil the goldmark ast.Node interface.
type mention struct {
	ast.BaseInline
	Segment text.Segment
}

type hashtag struct {
	ast.BaseInline
	Segment text.Segment
}

type emoji struct {
	ast.BaseInline
	Segment text.Segment
}

var (
	kindMention = ast.NewNodeKind("Mention")
	kindHashtag = ast.NewNodeKind("Hashtag")
	kindEmoji   = ast.NewNodeKind("Emoji")
)

func (n *mention) Kind() ast.NodeKind {
	return kindMention
}

func (n *hashtag) Kind() ast.NodeKind {
	return kindHashtag
}

func (n *emoji) Kind() ast.NodeKind {
	return kindEmoji
}

// Dump can be used for debugging.
func (n *mention) Dump(source []byte, level int) {
	fmt.Printf("%sMention: %s\n", strings.Repeat("    ", level), string(n.Segment.Value(source)))
}

func (n *hashtag) Dump(source []byte, level int) {
	fmt.Printf("%sHashtag: %s\n", strings.Repeat("    ", level), string(n.Segment.Value(source)))
}

func (n *emoji) Dump(source []byte, level int) {
	fmt.Printf("%sEmoji: %s\n", strings.Repeat("    ", level), string(n.Segment.Value(source)))
}

// newMention and newHashtag create a goldmark ast.Node from a goldmark text.Segment.
// The contained segment is used in rendering.
func newMention(s text.Segment) *mention {
	return &mention{
		BaseInline: ast.BaseInline{},
		Segment:    s,
	}
}

func newHashtag(s text.Segment) *hashtag {
	return &hashtag{
		BaseInline: ast.BaseInline{},
		Segment:    s,
	}
}

func newEmoji(s text.Segment) *emoji {
	return &emoji{
		BaseInline: ast.BaseInline{},
		Segment:    s,
	}
}

// mentionParser and hashtagParser fulfil the goldmark parser.InlineParser interface.
type mentionParser struct{}

type hashtagParser struct{}

type emojiParser struct{}

func (p *mentionParser) Trigger() []byte {
	return []byte{'@'}
}

func (p *hashtagParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *emojiParser) Trigger() []byte {
	return []byte{':'}
}

func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()

	if !util.IsMentionOrHashtagBoundary(before) {
		return nil
	}

	// unideal for performance but makes use of existing regex
	loc := regexes.MentionFinder.FindIndex(line)
	switch {
	case loc == nil:
		fallthrough
	case loc[0] != 0: // fail if not found at start
		return nil
	default:
		block.Advance(loc[1])
		return newMention(segment.WithStop(segment.Start + loc[1]))
	}
}

func (p *hashtagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	s := string(line)

	if !util.IsMentionOrHashtagBoundary(before) || len(s) == 1 {
		return nil
	}

	for i, r := range s {
		switch {
		case r == '#' && i == 0:
			// ignore initial #
			continue
		case !util.IsPlausiblyInHashtag(r) && !util.IsMentionOrHashtagBoundary(r):
			// Fake hashtag, don't trust it
			return nil
		case util.IsMentionOrHashtagBoundary(r):
			if i <= 1 {
				// empty
				return nil
			}
			// End of hashtag
			block.Advance(i)
			return newHashtag(segment.WithStop(segment.Start + i))
		}
	}
	// If we don't find invalid characters before the end of the line then it's all hashtag, babey
	block.Advance(segment.Len())
	return newHashtag(segment)
}

func (p *emojiParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// unideal for performance but makes use of existing regex
	loc := regexes.EmojiFinder.FindIndex(line)
	switch {
	case loc == nil:
		fallthrough
	case loc[0] != 0: // fail if not found at start
		return nil
	default:
		block.Advance(loc[1])
		return newEmoji(segment.WithStop(segment.Start + loc[1]))
	}
}

// customRenderer fulfils both the renderer.NodeRenderer and goldmark.Extender interfaces.
// It is created in FromMarkdown and FromPlain to be used as a goldmark extension, and the
// fields are used to report tags and mentions to the caller for use as metadata.
type customRenderer struct {
	f            *formatter
	ctx          context.Context
	parseMention gtsmodel.ParseMentionFunc
	accountID    string
	statusID     string
	emojiOnly    bool
	result       *FormatResult
}

func (r *customRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMention, r.renderMention)
	reg.Register(kindHashtag, r.renderHashtag)
	reg.Register(kindEmoji, r.renderEmoji)
}

func (r *customRenderer) Extend(m goldmark.Markdown) {
	// 1000 is set as the lowest priority, but it's arbitrary
	m.Parser().AddOptions(parser.WithInlineParsers(
		mdutil.Prioritized(&emojiParser{}, 1000),
	))
	if !r.emojiOnly {
		m.Parser().AddOptions(parser.WithInlineParsers(
			mdutil.Prioritized(&mentionParser{}, 1000),
			mdutil.Prioritized(&hashtagParser{}, 1000),
		))
	}
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		mdutil.Prioritized(r, 1000),
	))
}

// renderMention and renderHashtag take a mention or a hashtag ast.Node and render it as HTML.
func (r *customRenderer) renderMention(w mdutil.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	n, ok := node.(*mention) // this function is only registered for kindMention
	if !ok {
		log.Panic(nil, "type assertion failed")
	}
	text := string(n.Segment.Value(source))

	html := r.replaceMention(text)

	// we don't have much recourse if this fails
	if _, err := w.WriteString(html); err != nil {
		log.Errorf(nil, "error writing HTML: %s", err)
	}
	return ast.WalkSkipChildren, nil
}

func (r *customRenderer) renderHashtag(w mdutil.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	n, ok := node.(*hashtag) // this function is only registered for kindHashtag
	if !ok {
		log.Panic(nil, "type assertion failed")
	}
	text := string(n.Segment.Value(source))

	html := r.replaceHashtag(text)

	_, err := w.WriteString(html)
	// we don't have much recourse if this fails
	if err != nil {
		log.Errorf(nil, "error writing HTML: %s", err)
	}
	return ast.WalkSkipChildren, nil
}

// renderEmoji doesn't turn an emoji into HTML, but adds it to the metadata.
func (r *customRenderer) renderEmoji(w mdutil.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	n, ok := node.(*emoji) // this function is only registered for kindEmoji
	if !ok {
		log.Panic(nil, "type assertion failed")
	}
	text := string(n.Segment.Value(source))
	shortcode := text[1 : len(text)-1]

	emoji, err := r.f.db.GetEmojiByShortcodeDomain(r.ctx, shortcode, "")
	if err != nil {
		if err != db.ErrNoEntries {
			log.Errorf(nil, "error getting local emoji with shortcode %s: %s", shortcode, err)
		}
	} else if *emoji.VisibleInPicker && !*emoji.Disabled {
		listed := false
		for _, e := range r.result.Emojis {
			if e.Shortcode == emoji.Shortcode {
				listed = true
				break
			}
		}
		if !listed {
			r.result.Emojis = append(r.result.Emojis, emoji)
		}
	}

	// we don't have much recourse if this fails
	if _, err := w.WriteString(text); err != nil {
		log.Errorf(nil, "error writing HTML: %s", err)
	}
	return ast.WalkSkipChildren, nil
}
