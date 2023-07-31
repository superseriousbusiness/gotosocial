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
	"errors"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// replaceMention takes a string in the form @username@domain.com or @localusername
func (r *customRenderer) replaceMention(text string) string {
	mention, err := r.parseMention(r.ctx, text, r.accountID, r.statusID)
	if err != nil {
		log.Errorf(r.ctx, "error parsing mention %s from status: %s", text, err)
		return text
	}

	if r.statusID != "" {
		if err := r.f.db.PutMention(r.ctx, mention); err != nil {
			log.Errorf(r.ctx, "error putting mention in db: %s", err)
			return text
		}
	}

	// only append if it's not been listed yet
	listed := false
	for _, m := range r.result.Mentions {
		if mention.ID == m.ID {
			listed = true
			break
		}
	}
	if !listed {
		r.result.Mentions = append(r.result.Mentions, mention)
	}

	if mention.TargetAccount == nil {
		// Fetch mention target account if not yet populated.
		mention.TargetAccount, err = r.f.db.GetAccountByID(
			gtscontext.SetBarebones(r.ctx),
			mention.TargetAccountID,
		)
		if err != nil {
			log.Errorf(r.ctx, "error populating mention target account: %v", err)
			return text
		}
	}

	// The mention's target is our target
	targetAccount := mention.TargetAccount

	var b strings.Builder

	// replace the mention with the formatted mention content
	// <span class="h-card"><a href="targetAccount.URL" class="u-url mention">@<span>targetAccount.Username</span></a></span>
	b.WriteString(`<span class="h-card"><a href="`)
	b.WriteString(targetAccount.URL)
	b.WriteString(`" class="u-url mention">@<span>`)
	b.WriteString(targetAccount.Username)
	b.WriteString(`</span></a></span>`)
	return b.String()
}

// replaceHashtag takes a string in the form #SomeHashtag, and will normalize
// it before adding it to the db (or just getting it from the db if it already
// exists) and turning it into HTML.
func (r *customRenderer) replaceHashtag(text string) string {
	normalized, ok := NormalizeHashtag(text)
	if !ok {
		// Not a valid hashtag.
		return text
	}

	tag, err := r.getOrCreateHashtag(normalized)
	if err != nil {
		log.Errorf(r.ctx, "error generating hashtags from status: %s", err)
		return text
	}

	// Append tag to result if not done already.
	//
	// This prevents multiple uses of a tag in
	// the same status generating multiple
	// entries for the same tag in result.
	func() {
		for _, t := range r.result.Tags {
			if tag.ID == t.ID {
				// Already appended.
				return
			}
		}

		// Not appended yet.
		r.result.Tags = append(r.result.Tags, tag)
	}()

	// Replace tag with the formatted tag content, eg. `#SomeHashtag` becomes:
	// `<a href="https://example.org/tags/somehashtag" class="mention hashtag" rel="tag">#<span>SomeHashtag</span></a>`
	var b strings.Builder
	b.WriteString(`<a href="`)
	b.WriteString(uris.GenerateURIForTag(normalized))
	b.WriteString(`" class="mention hashtag" rel="tag">#<span>`)
	b.WriteString(normalized)
	b.WriteString(`</span></a>`)

	return b.String()
}

func (r *customRenderer) getOrCreateHashtag(name string) (*gtsmodel.Tag, error) {
	var (
		tag *gtsmodel.Tag
		err error
	)

	// Check if we have a tag with this name already.
	tag, err = r.f.db.GetTagByName(r.ctx, name)
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

	if err = r.f.db.PutTag(r.ctx, tag); err != nil {
		return nil, gtserror.Newf("db error putting new tag %s: %w", name, err)
	}

	return tag, nil
}
