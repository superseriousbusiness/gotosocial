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
	"errors"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"golang.org/x/text/unicode/norm"
)

const (
	maximumHashtagLength = 30
)

// given a mention or a hashtag string, the methods in this file will attempt to parse it,
// add it to the database, and render it as HTML. If any of these steps fails, the method
// will just return the original string and log an error.

// replaceMention takes a string in the form @username@domain.com or @localusername
func (r *customRenderer) replaceMention(text string) string {
	menchie, err := r.parseMention(r.ctx, text, r.accountID, r.statusID)
	if err != nil {
		log.Errorf(nil, "error parsing mention %s from status: %s", text, err)
		return text
	}

	if r.statusID != "" {
		if err := r.f.db.Put(r.ctx, menchie); err != nil {
			log.Errorf(nil, "error putting mention in db: %s", err)
			return text
		}
	}

	// only append if it's not been listed yet
	listed := false
	for _, m := range r.result.Mentions {
		if menchie.ID == m.ID {
			listed = true
			break
		}
	}
	if !listed {
		r.result.Mentions = append(r.result.Mentions, menchie)
	}

	// make sure we have an account attached to this mention
	if menchie.TargetAccount == nil {
		a, err := r.f.db.GetAccountByID(r.ctx, menchie.TargetAccountID)
		if err != nil {
			log.Errorf(nil, "error getting account with id %s from the db: %s", menchie.TargetAccountID, err)
			return text
		}
		menchie.TargetAccount = a
	}

	// The mention's target is our target
	targetAccount := menchie.TargetAccount

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

// replaceMention takes a string in the form #HashedTag, and will normalize it before
// adding it to the db and turning it into HTML.
func (r *customRenderer) replaceHashtag(text string) string {
	// this normalization is specifically to avoid cases where visually-identical
	// hashtags are stored with different unicode representations (e.g. with combining
	// diacritics). It allows a tasteful number of combining diacritics to be used,
	// as long as they can be combined with parent characters to form regular letter
	// symbols.
	normalized := norm.NFC.String(text[1:])

	for i, r := range normalized {
		if i >= maximumHashtagLength || !util.IsPermittedInHashtag(r) {
			return text
		}
	}

	tag, err := r.f.db.TagStringToTag(r.ctx, normalized, r.accountID)
	if err != nil {
		log.Errorf(nil, "error generating hashtags from status: %s", err)
		return text
	}

	// only append if it's not been listed yet
	listed := false
	for _, t := range r.result.Tags {
		if tag.ID == t.ID {
			listed = true
			break
		}
	}
	if !listed {
		err = r.f.db.Put(r.ctx, tag)
		if err != nil {
			if !errors.Is(err, db.ErrAlreadyExists) {
				log.Errorf(nil, "error putting tags in db: %s", err)
				return text
			}
		}
		r.result.Tags = append(r.result.Tags, tag)
	}

	var b strings.Builder
	// replace the #tag with the formatted tag content
	// `<a href="tag.URL" class="mention hashtag" rel="tag">#<span>tagAsEntered</span></a>
	b.WriteString(`<a href="`)
	b.WriteString(tag.URL)
	b.WriteString(`" class="mention hashtag" rel="tag">#<span>`)
	b.WriteString(normalized)
	b.WriteString(`</span></a>`)

	return b.String()
}
