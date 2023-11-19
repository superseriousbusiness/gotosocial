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

package typeutils

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/language"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

type statusInteractions struct {
	Faved      bool
	Muted      bool
	Bookmarked bool
	Reblogged  bool
	Pinned     bool
}

func (c *Converter) interactionsWithStatusForAccount(ctx context.Context, s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*statusInteractions, error) {
	si := &statusInteractions{}

	if requestingAccount != nil {
		faved, err := c.state.DB.IsStatusFavedBy(ctx, s.ID, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has faved status: %s", err)
		}
		si.Faved = faved

		reblogged, err := c.state.DB.IsStatusBoostedBy(ctx, s.ID, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has reblogged status: %s", err)
		}
		si.Reblogged = reblogged

		muted, err := c.state.DB.IsThreadMutedByAccount(ctx, s.ThreadID, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has muted status: %s", err)
		}
		si.Muted = muted

		bookmarked, err := c.state.DB.IsStatusBookmarkedBy(ctx, s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has bookmarked status: %s", err)
		}
		si.Bookmarked = bookmarked

		// The only time 'pinned' should be true is if the
		// requesting account is looking at its OWN status.
		if s.AccountID == requestingAccount.ID {
			si.Pinned = !s.PinnedAt.IsZero()
		}
	}
	return si, nil
}

func misskeyReportInlineURLs(content string) []*url.URL {
	m := regexes.MisskeyReportNotes.FindAllStringSubmatch(content, -1)
	urls := make([]*url.URL, 0, len(m))
	for _, sm := range m {
		url, err := url.Parse(sm[1])
		if err == nil && url != nil {
			urls = append(urls, url)
		}
	}
	return urls
}

// getURI is a shortcut/util function for extracting
// the JSONLDId URI of an Activity or Object.
func getURI(withID ap.WithJSONLDId) (*url.URL, string, error) {
	idProp := withID.GetJSONLDId()
	if idProp == nil {
		return nil, "", errors.New("id prop was nil")
	}

	if !idProp.IsIRI() {
		return nil, "", errors.New("id prop was not an IRI")
	}

	id := idProp.Get()
	return id, id.String(), nil
}

// placeholdUnknownAttachments separates any attachments with type `unknown`
// out of the given slice, and returns an `<aside>` tag containing links to
// those attachments, as well as the slice of remaining "known" attachments.
// If there are no unknown-type attachments in the provided slice, an empty
// string and the original slice will be returned.
//
// If an aside is created, it will be run through the sanitizer before being
// returned, to ensure that malicious links don't cause issues.
//
// Example:
//
//	<aside>
//	   <p>Note from your.instance.com: 2 attachments in this status could not be downloaded. Treat the following external links with care:
//	      <ul>
//	         <li><a href="http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7ZGJYTSYMXF927GF9353KR.svg" rel="nofollow noreferrer noopener" target="_blank">01HE7ZGJYTSYMXF927GF9353KR.svg</a> [SVG line art of a sloth, public domain]</li>
//	         <li><a href="http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE892Y8ZS68TQCNPX7J888P3.mp3" rel="nofollow noreferrer noopener" target="_blank">01HE892Y8ZS68TQCNPX7J888P3.mp3</a> [Jolly salsa song, public domain.]</li>
//	      </ul>
//	   </p>
//	</aside>
func placeholdUnknownAttachments(arr []apimodel.Attachment) (string, []apimodel.Attachment) {
	// Extract unknown-type attachments into a separate
	// slice, deleting them from arr in the process.
	var unknowns []apimodel.Attachment
	arr = slices.DeleteFunc(arr, func(elem apimodel.Attachment) bool {
		unknown := elem.Type == "unknown"
		if unknown {
			// Set aside unknown-type attachment.
			unknowns = append(unknowns, elem)
		}

		return unknown
	})

	unknownsLen := len(unknowns)
	if unknownsLen == 0 {
		// No unknown attachments,
		// nothing to do.
		return "", arr
	}

	// Plural / singular.
	var (
		attachments string
		links       string
	)

	if unknownsLen == 1 {
		attachments = "1 attachment"
		links = "link"
	} else {
		attachments = strconv.Itoa(unknownsLen) + " attachments"
		links = "links"
	}

	var aside strings.Builder
	aside.WriteString(`<aside>`)
	aside.WriteString(`<p>`)
	aside.WriteString(`Note from ` + config.GetHost() + `: ` + attachments + ` in this status could not be downloaded. Treat the following external ` + links + ` with care:`)
	aside.WriteString(`<ul>`)
	for _, a := range unknowns {
		var (
			remoteURL = *a.RemoteURL
			base      = path.Base(remoteURL)
			entry     = fmt.Sprintf(`<a href="%s">%s</a>`, remoteURL, base)
		)
		if d := a.Description; d != nil && *d != "" {
			entry += ` [` + *d + `]`
		}
		aside.WriteString(`<li>` + entry + `</li>`)
	}
	aside.WriteString(`</ul>`)
	aside.WriteString(`</p>`)
	aside.WriteString(`</aside>`)

	return text.SanitizeToHTML(aside.String()), arr
}

// ContentToContentLanguage tries to
// extract a content string and language
// tag string from the given intermediary
// content.
//
// Either/both of the returned strings may
// be empty, depending on how things go.
func ContentToContentLanguage(
	ctx context.Context,
	content gtsmodel.Content,
) (
	string, // content
	string, // language
) {
	var (
		contentStr string
		langTagStr string
	)

	switch cMap := content.ContentMap; {
	// Simplest case: no `contentMap`.
	// Return `content`, even if empty.
	case cMap == nil:
		return content.Content, ""

	// `content` and `contentMap` set.
	// Try to infer "primary" language.
	case content.Content != "":
		// Assume `content` is intended
		// primary content, and look for
		// corresponding language tag.
		contentStr = content.Content

		for t, c := range cMap {
			if contentStr == c {
				langTagStr = t
			}
		}

	// `content` not set; `contentMap`
	// is set with only one value.
	// This must be the "primary" lang.
	case len(cMap) == 1:
		// Use an empty loop to
		// get the values we want.
		// nolint:revive
		for langTagStr, contentStr = range cMap {
		}

	// Only `contentMap` is set, with
	// more than one value. Map order
	// is not guaranteed so we can't
	// know the "primary" language.
	//
	// Just stop at the first non-empty
	// langTag and langCnt we can find.
	default:
		for langTagStr, contentStr = range cMap {
			if langTagStr != "" &&
				contentStr != "" {
				break
			}
		}
	}

	if langTagStr != "" {
		// Found a lang tag for this content,
		// make sure it's valid / parseable.
		lang, err := language.Parse(langTagStr)
		if err != nil {
			log.Warnf(
				ctx,
				"could not parse %s as BCP47 language tag in status contentMap: %v",
				langTagStr, err,
			)
		} else {
			// Inferred the language!
			// Use normalized version.
			langTagStr = lang.TagStr
		}
	}

	return contentStr, langTagStr
}
