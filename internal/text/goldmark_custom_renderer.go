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
	"errors"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	mdutil "github.com/yuin/goldmark/util"
)

// customRenderer fulfils the following goldmark interfaces:
//
//   - renderer.NodeRenderer
//   - goldmark.Extender.
//
// It is used as a goldmark extension by FromMarkdown and
// (variants of) FromPlain.
//
// The custom renderer extracts and re-renders mentions, hashtags,
// and emojis that are encountered during parsing, writing out valid
// HTML representations of these elements.
//
// The customRenderer has the following side effects:
//
//   - May use its db connection to retrieve existing and/or
//     store new mentions, hashtags, and emojis.
//   - May update its *FormatResult to append discovered
//     mentions, hashtags, and emojis to it.
type customRenderer struct {
	ctx          context.Context
	db           db.DB
	parseMention gtsmodel.ParseMentionFunc
	accountID    string
	statusID     string
	emojiOnly    bool
	result       *FormatResult
}

func (cr *customRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMention, cr.renderMention)
	reg.Register(kindHashtag, cr.renderHashtag)
	reg.Register(kindEmoji, cr.renderEmoji)
}

func (cr *customRenderer) Extend(markdown goldmark.Markdown) {
	// 1000 is set as the lowest
	// priority, but it's arbitrary.
	const prio = 1000

	if cr.emojiOnly {
		// Parse + render only emojis.
		markdown.Parser().AddOptions(
			parser.WithInlineParsers(
				mdutil.Prioritized(new(emojiParser), prio),
			),
		)
	} else {
		// Parse + render emojis, mentions, hashtags.
		markdown.Parser().AddOptions(parser.WithInlineParsers(
			mdutil.Prioritized(new(emojiParser), prio),
			mdutil.Prioritized(new(mentionParser), prio),
			mdutil.Prioritized(new(hashtagParser), prio),
		))
	}

	// Add this custom renderer.
	markdown.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			mdutil.Prioritized(cr, prio),
		),
	)
}

/*
	MENTION RENDERING STUFF
*/

// renderMention takes a mention
// ast.Node and renders it as HTML.
func (cr *customRenderer) renderMention(
	w mdutil.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	// This function is registered
	// only for kindMention, and
	// should not be called for
	// any other node type.
	n, ok := node.(*mention)
	if !ok {
		log.Panic(cr.ctx, "type assertion failed")
	}

	// Get raw mention string eg., '@someone@domain.org'.
	text := string(n.Segment.Value(source))

	// Handle mention and get text to render.
	text = cr.handleMention(text)

	// Write returned text into HTML.
	if _, err := w.WriteString(text); err != nil {
		// We don't have much recourse if this fails.
		log.Errorf(cr.ctx, "error writing HTML: %s", err)
	}

	return ast.WalkSkipChildren, nil
}

// handleMention takes a string in the form '@username@domain.com'
// or '@localusername', and does the following:
//
//   - Parse the mention string into a *gtsmodel.Mention.
//   - Insert mention into database if necessary.
//   - Add mention to cr.results.Mentions slice.
//   - Return mention rendered as nice HTML.
//
// If the mention is invalid or cannot be created,
// the unaltered input text will be returned instead.
func (cr *customRenderer) handleMention(text string) string {
	mention, err := cr.parseMention(cr.ctx, text, cr.accountID, cr.statusID)
	if err != nil {
		log.Errorf(cr.ctx, "error parsing mention %s from status: %s", text, err)
		return text
	}

	// Store mention if it's from a
	// status and wasn't stored before.
	if cr.statusID != "" && mention.IsNew {
		if err := cr.db.PutMention(cr.ctx, mention); err != nil {
			log.Errorf(cr.ctx, "error putting mention in db: %s", err)
			return text
		}
	}

	// Append mention to result if not done already.
	//
	// This prevents multiple occurences of mention
	// in the same status generating multiple
	// entries for the same mention in result.
	func() {
		for _, m := range cr.result.Mentions {
			if mention.TargetAccountID == m.TargetAccountID {
				// Already appended.
				return
			}
		}

		// Not appended yet.
		cr.result.Mentions = append(cr.result.Mentions, mention)
	}()

	if mention.TargetAccount == nil {
		// Fetch mention target account if not yet populated.
		mention.TargetAccount, err = cr.db.GetAccountByID(
			gtscontext.SetBarebones(cr.ctx),
			mention.TargetAccountID,
		)
		if err != nil {
			log.Errorf(cr.ctx, "error populating mention target account: %v", err)
			return text
		}
	}

	// Replace the mention with the formatted mention content,
	// eg. `@someone@domain.org` becomes:
	// `<span class="h-card"><a href="https://domain.org/@someone" class="u-url mention">@<span>someone</span></a></span>`
	var b strings.Builder
	b.WriteString(`<span class="h-card"><a href="`)
	b.WriteString(mention.TargetAccount.URL)
	b.WriteString(`" class="u-url mention">@<span>`)
	b.WriteString(mention.TargetAccount.Username)
	b.WriteString(`</span></a></span>`)
	return b.String()
}

/*
	HASHTAG RENDERING STUFF
*/

// renderHashtag takes a hashtag
// ast.Node and renders it as HTML.
func (cr *customRenderer) renderHashtag(
	w mdutil.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	// This function is registered
	// only for kindHashtag, and
	// should not be called for
	// any other node type.
	n, ok := node.(*hashtag)
	if !ok {
		log.Panic(cr.ctx, "type assertion failed")
	}

	// Get raw hashtag string eg., '#SomeHashtag'.
	text := string(n.Segment.Value(source))

	// Handle hashtag and get text to render.
	text = cr.handleHashtag(text)

	// Write returned text into HTML.
	if _, err := w.WriteString(text); err != nil {
		// We don't have much recourse if this fails.
		log.Errorf(cr.ctx, "error writing HTML: %s", err)
	}

	return ast.WalkSkipChildren, nil
}

// handleHashtag takes a string in the form '#SomeHashtag',
// and does the following:
//
//   - Normalize + validate the hashtag.
//   - Get or create hashtag in the db.
//   - Add hashtag to cr.results.Tags slice.
//   - Return hashtag rendered as nice HTML.
//
// If the hashtag is invalid or cannot be retrieved,
// the unaltered input text will be returned instead.
func (cr *customRenderer) handleHashtag(text string) string {
	normalized, ok := NormalizeHashtag(text)
	if !ok {
		// Not a valid hashtag.
		return text
	}

	getOrCreateHashtag := func(name string) (*gtsmodel.Tag, error) {
		var (
			tag *gtsmodel.Tag
			err error
		)

		// Check if we have a tag with this name already.
		tag, err = cr.db.GetTagByName(cr.ctx, name)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf("db error getting tag %s: %w", name, err)
		}

		if tag != nil {
			// We had it!
			return tag, nil
		}

		// We didn't have a tag with
		// this name, create one.
		tag = &gtsmodel.Tag{
			ID:   id.NewULID(),
			Name: name,
		}

		if err = cr.db.PutTag(cr.ctx, tag); err != nil {
			return nil, gtserror.Newf("db error putting new tag %s: %w", name, err)
		}

		return tag, nil
	}

	tag, err := getOrCreateHashtag(normalized)
	if err != nil {
		log.Errorf(cr.ctx, "error generating hashtags from status: %s", err)
		return text
	}

	// Append tag to result if not done already.
	//
	// This prevents multiple uses of a tag in
	// the same status generating multiple
	// entries for the same tag in result.
	func() {
		for _, t := range cr.result.Tags {
			if tag.ID == t.ID {
				// Already appended.
				return
			}
		}

		// Not appended yet.
		cr.result.Tags = append(cr.result.Tags, tag)
	}()

	// Replace tag with the formatted tag content, eg. `#SomeHashtag` becomes:
	// `<a href="https://example.org/tags/somehashtag" class="mention hashtag" rel="tag">#<span>SomeHashtag</span></a>`
	var b strings.Builder
	b.WriteString(`<a href="`)
	b.WriteString(uris.URIForTag(normalized))
	b.WriteString(`" class="mention hashtag" rel="tag">#<span>`)
	b.WriteString(normalized)
	b.WriteString(`</span></a>`)

	return b.String()
}

/*
	EMOJI RENDERING STUFF
*/

// renderEmoji doesn't actually turn an emoji
// ast.Node into HTML, but instead only adds it to
// the custom renderer results for later processing.
func (cr *customRenderer) renderEmoji(
	w mdutil.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}

	// This function is registered
	// only for kindEmoji, and
	// should not be called for
	// any other node type.
	n, ok := node.(*emoji)
	if !ok {
		log.Panic(cr.ctx, "type assertion failed")
	}

	// Get raw emoji string eg., ':boobs:'.
	text := string(n.Segment.Value(source))

	// Handle emoji and get text to render.
	text = cr.handleEmoji(text)

	// Write returned text into HTML.
	if _, err := w.WriteString(text); err != nil {
		// We don't have much recourse if this fails.
		log.Errorf(cr.ctx, "error writing HTML: %s", err)
	}

	return ast.WalkSkipChildren, nil
}

// handleEmoji takes a string in the form ':some_emoji:',
// and does the following:
//
//   - Try to get emoji from the db.
//   - Add emoji to cr.results.Emojis slice if found and useable.
//
// This function will always return the unaltered input
// text, since emojification is handled elsewhere.
func (cr *customRenderer) handleEmoji(text string) string {
	// Check if text points to a valid
	// local emoji by using its shortcode.
	//
	// The shortcode is the text
	// between enclosing ':' chars.
	shortcode := strings.Trim(text, ":")

	// Try to fetch emoji as a locally stored emoji.
	emoji, err := cr.db.GetEmojiByShortcodeDomain(cr.ctx, shortcode, "")
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf(nil, "db error getting local emoji with shortcode %s: %s", shortcode, err)
	}

	if emoji == nil {
		// No emoji found for this
		// shortcode, oh well!
		return text
	}

	if *emoji.Disabled || !*emoji.VisibleInPicker {
		// Emoji was found but not useable.
		return text
	}

	// Emoji was found and useable.
	// Append to result if not done already.
	//
	// This prevents multiple uses of an emoji
	// in the same status generating multiple
	// entries for the same emoji in result.
	func() {
		for _, e := range cr.result.Emojis {
			if emoji.Shortcode == e.Shortcode {
				// Already appended.
				return
			}
		}

		// Not appended yet.
		cr.result.Emojis = append(cr.result.Emojis, emoji)
	}()

	return text
}
