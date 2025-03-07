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
	"math"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/language"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

// toAPISize converts a set of media dimensions
// to mastodon API compatible size string.
func toAPISize(width, height int) string {
	return strconv.Itoa(width) +
		"x" +
		strconv.Itoa(height)
}

// toAPIFrameRate converts a media framerate ptr
// to mastodon API compatible framerate string.
func toAPIFrameRate(framerate *float32) string {
	if framerate == nil {
		return ""
	}
	// The masto api expects this as a string in
	// the format `integer/1`, so 30fps is `30/1`.
	round := math.Round(float64(*framerate))
	return strconv.Itoa(int(round)) + "/1"
}

type statusInteractions struct {
	Favourited bool
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
		si.Favourited = faved

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

		bookmarked, err := c.state.DB.IsStatusBookmarkedBy(ctx, requestingAccount.ID, s.ID)
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

// placeholderAttachments separates any attachments with missing local URL
// out of the given slice, and returns a piece of text containing links to
// those attachments, as well as the slice of remaining "known" attachments.
// If there are no unknown-type attachments in the provided slice, an empty
// string and the original slice will be returned.
//
// Returned text will be run through the sanitizer before being returned, to
// ensure that malicious links don't cause issues.
//
// Example:
//
//	<div class="gts-system-message gts-placeholder-attachments">
//		<hr>
//		<p><i lang="en">ℹ️ Note from your.instance.com: 2 attachment(s) in this status were not downloaded. Treat the following external link(s) with care:</i></p>
//		<ul>
//			<li><a href="http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7ZGJYTSYMXF927GF9353KR.svg" rel="nofollow noreferrer noopener" target="_blank">01HE7ZGJYTSYMXF927GF9353KR.svg</a> [SVG line art of a sloth, public domain]</li>
//			<li><a href="http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE892Y8ZS68TQCNPX7J888P3.mp3" rel="nofollow noreferrer noopener" target="_blank">01HE892Y8ZS68TQCNPX7J888P3.mp3</a> [Jolly salsa song, public domain.]</li>
//		</ul>
//	</div>
func placeholderAttachments(arr []*apimodel.Attachment) (string, []*apimodel.Attachment) {

	// Extract non-locally stored attachments into a
	// separate slice, deleting them from input slice.
	var nonLocal []*apimodel.Attachment
	arr = slices.DeleteFunc(arr, func(elem *apimodel.Attachment) bool {
		if elem.URL == nil {
			nonLocal = append(nonLocal, elem)
			return true
		}
		return false
	})

	if len(nonLocal) == 0 {
		// No non-locally
		// stored media.
		return "", arr
	}

	var note strings.Builder
	note.WriteString(`<hr>`)
	note.WriteString(`<p><i lang="en">ℹ️ Note from `)
	note.WriteString(config.GetHost())
	note.WriteString(`: `)
	note.WriteString(strconv.Itoa(len(nonLocal)))

	if len(nonLocal) > 1 {
		// Use plural word form.
		note.WriteString(` attachments in this status were not downloaded. ` +
			`Treat the following external links with care:`)
	} else {
		// Use singular word form.
		note.WriteString(` attachment in this status was not downloaded. ` +
			`Treat the following external link with care:`)
	}

	note.WriteString(`</i></p><ul>`)
	for _, a := range nonLocal {
		note.WriteString(`<li>`)
		note.WriteString(`<a href="`)
		note.WriteString(*a.RemoteURL)
		note.WriteString(`">`)
		note.WriteString(path.Base(*a.RemoteURL))
		note.WriteString(`</a>`)
		if d := a.Description; d != nil && *d != "" {
			note.WriteString(` [`)
			note.WriteString(*d)
			note.WriteString(`]`)
		}
		note.WriteString(`</li>`)
	}
	note.WriteString(`</ul>`)

	return systemMessage("gts-placeholder-attachments", note.String()), arr
}

func (c *Converter) pendingReplyNote(
	ctx context.Context,
	s *gtsmodel.Status,
) (string, error) {
	intReq, err := c.state.DB.GetInteractionRequestByInteractionURI(ctx, s.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Something's gone wrong.
		err := gtserror.Newf("db error getting interaction request for %s: %w", s.URI, err)
		return "", err
	}

	// No interaction request present
	// for this status. Race condition?
	if intReq == nil {
		return "", nil
	}

	var (
		proto = config.GetProtocol()
		host  = config.GetHost()

		// Build the settings panel URL at which the user
		// can view + approve/reject the interaction request.
		//
		// Eg., https://example.org/settings/user/interaction_requests/01J5QVXCCEATJYSXM9H6MZT4JR
		settingsURL = proto + "://" + host + "/settings/user/interaction_requests/" + intReq.ID
	)

	var note strings.Builder
	note.WriteString(`<hr>`)
	note.WriteString(`<p><i lang="en">ℹ️ Note from ` + host + `: `)
	note.WriteString(`This reply is pending your approval. You can quickly accept it by liking, boosting or replying to it. You can also accept or reject it at the following link: `)
	note.WriteString(`<a href="` + settingsURL + `" `)
	note.WriteString(`rel="noreferrer noopener" target="_blank">`)
	note.WriteString(settingsURL)
	note.WriteString(`</a>.`)
	note.WriteString(`</i></p>`)

	return systemMessage("gts-pending-reply", note.String()), nil
}

// systemMessage wraps a note with a div with semantic classes that aren't allowed through the sanitizer,
// but may be emitted to the client as an addition to the status's actual content.
// Clients may want to display these specially or suppress them in favor of their own UI.
//
// messageClass must be valid inside an HTML attribute and should be one or more classes starting with `gts-`.
func systemMessage(
	messageClass string,
	unsanitizedNoteHTML string,
) string {
	var wrappedNote strings.Builder

	wrappedNote.WriteString(`<div class="gts-system-message `)
	wrappedNote.WriteString(messageClass)
	wrappedNote.WriteString(`">`)
	wrappedNote.WriteString(text.SanitizeHTML(unsanitizedNoteHTML))
	wrappedNote.WriteString(`</div>`)

	return wrappedNote.String()
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

	switch contentMap := content.ContentMap; {
	// Simplest case: no `contentMap`.
	// Return `content`, even if empty.
	case contentMap == nil:
		return content.Content, ""

	// `content` and `contentMap` set.
	// Try to infer "primary" language.
	case content.Content != "":
		// Assume `content` is intended
		// primary content, and look for
		// corresponding language tag.
		contentStr = content.Content

		for t, c := range contentMap {
			if contentStr == c {
				langTagStr = t
				break
			}
		}

	// `content` not set; `contentMap`
	// is set with only one value.
	// This must be the "primary" lang.
	case len(contentMap) == 1:
		// Use an empty loop to
		// get the values we want.
		// nolint:revive
		for langTagStr, contentStr = range contentMap {
		}

	// Only `contentMap` is set, with more
	// than one value. Map order is not
	// guaranteed so we can't know the
	// "primary" language.
	//
	// Try to select content using our
	// instance's configured languages.
	//
	// In case of no hits, just take the
	// first tag and content in the map.
	default:
		instanceLangs := config.GetInstanceLanguages()
		for _, langTagStr = range instanceLangs.TagStrs() {
			if contentStr = contentMap[langTagStr]; contentStr != "" {
				// Hit!
				break
			}
		}

		// If nothing found, just take
		// the first entry we can get by
		// breaking after the first iter.
		if contentStr == "" {
			for langTagStr, contentStr = range contentMap {
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

// filterableFields returns text fields from
// a status that we might want to filter on:
//
//   - content warning
//   - content (converted to plaintext from HTML)
//   - media descriptions
//   - poll options
//
// Each field should be filtered separately.
// This avoids scenarios where false-positive
// multiple-word matches can be made by matching
// the last word of one field + the first word
// of the next field together.
func filterableFields(s *gtsmodel.Status) []string {
	// Estimate length of fields.
	fieldCount := 2 + len(s.Attachments)
	if s.Poll != nil {
		fieldCount += len(s.Poll.Options)
	}
	fields := make([]string, 0, fieldCount)

	// Content warning / title.
	if s.ContentWarning != "" {
		fields = append(fields, s.ContentWarning)
	}

	// Status content. Though we have raw text
	// available for statuses created on our
	// instance, use the plaintext version to
	// remove markdown-formatting characters
	// and ensure more consistent filtering.
	if s.Content != "" {
		text := text.ParseHTMLToPlain(s.Content)
		if text != "" {
			fields = append(fields, text)
		}
	}

	// Media descriptions.
	for _, attachment := range s.Attachments {
		if attachment.Description != "" {
			fields = append(fields, attachment.Description)
		}
	}

	// Poll options.
	if s.Poll != nil {
		for _, opt := range s.Poll.Options {
			if opt != "" {
				fields = append(fields, opt)
			}
		}
	}

	return fields
}
