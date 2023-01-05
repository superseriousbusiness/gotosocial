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
	"unicode"

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

var kindMention = ast.NewNodeKind("Mention")
var kindHashtag = ast.NewNodeKind("Hashtag")

func (n *mention) Kind() ast.NodeKind {
	return kindMention
}

func (n *hashtag) Kind() ast.NodeKind {
	return kindHashtag
}

// Dump is used by goldmark for debugging. It is implemented only minimally because
// it is not used in our code.
func (n *mention) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func (n *hashtag) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
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

// mentionParser and hashtagParser fulfil the goldmark parser.InlineParser interface.
type mentionParser struct {
}

type hashtagParser struct {
}

func (p *mentionParser) Trigger() []byte {
	return []byte{'@'}
}

func (p *hashtagParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()

	if !unicode.IsSpace(before) {
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

	if !util.IsHashtagBoundary(before) {
		return nil
	}

	for i, r := range s {
		switch {
		case r == '#' && i == 0:
			continue
		case !util.IsPermittedInHashtag(r) && !util.IsHashtagBoundary(r):
			// Fake hashtag, don't trust it
			return nil
		case util.IsHashtagBoundary(r):
			// End of hashtag
			block.Advance(i)
			return newHashtag(segment.WithStop(segment.Start + i))
		}
	}
	// If we don't find invalid characters before the end of the line then it's good
	block.Advance(len(s))
	return newHashtag(segment)
}

// customRenderer fulfils both the renderer.NodeRenderer and goldmark.Extender interfaces.
// It is created in FromMarkdown to be used a goldmark extension, and the fields are used
// when rendering mentions and tags.
type customRenderer struct {
	f        *formatter
	ctx      context.Context
	mentions []*gtsmodel.Mention
	tags     []*gtsmodel.Tag
}

func (r *customRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMention, r.renderMention)
	reg.Register(kindHashtag, r.renderHashtag)
}

func (r *customRenderer) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		// 500 is pretty arbitrary here, it was copied from example goldmark extension code.
		// https://github.com/yuin/goldmark/blob/75d8cce5b78c7e1d5d9c4ca32c1164f0a1e57b53/extension/strikethrough.go#L111
		mdutil.Prioritized(&mentionParser{}, 500),
		mdutil.Prioritized(&hashtagParser{}, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		mdutil.Prioritized(r, 500),
	))
}

// renderMention and renderHashtag take a mention or a hashtag ast.Node and render it as HTML.
func (r *customRenderer) renderMention(w mdutil.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n, ok := node.(*mention) // this function is only registered for kindMention
	if !ok {
		log.Errorf("type assertion failed")
	}
	text := string(n.Segment.Value(source))

	html := r.f.ReplaceMentions(r.ctx, text, r.mentions)

	// we don't have much recourse if this fails
	if _, err := w.WriteString(html); err != nil {
		log.Errorf("error outputting markdown text: %s", err)
	}
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderHashtag(w mdutil.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n, ok := node.(*hashtag) // this function is only registered for kindHashtag
	if !ok {
		log.Errorf("type assertion failed")
	}
	text := string(n.Segment.Value(source))

	html := r.f.ReplaceTags(r.ctx, text, r.tags)

	// we don't have much recourse if this fails
	if _, err := w.WriteString(html); err != nil {
		log.Errorf("error outputting markdown text: %s", err)
	}
	return ast.WalkContinue, nil
}
