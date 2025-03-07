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

package status

import (
	"context"
	"errors"
	"fmt"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// validateStatusContent will validate the common
// content fields across status write endpoints against
// current server configuration (e.g. max char counts).
func validateStatusContent(
	status string,
	spoiler string,
	mediaIDs []string,
	poll *apimodel.PollRequest,
) gtserror.WithCode {
	totalChars := len([]rune(status)) +
		len([]rune(spoiler))

	if totalChars == 0 && len(mediaIDs) == 0 && poll == nil {
		const text = "status contains no text, media or poll"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if max := config.GetStatusesMaxChars(); totalChars > max {
		text := fmt.Sprintf("text with spoiler exceed max chars (%d)", max)
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if max := config.GetStatusesMediaMaxFiles(); len(mediaIDs) > max {
		text := fmt.Sprintf("media files exceed max count (%d)", max)
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if poll != nil {
		switch max := config.GetStatusesPollMaxOptions(); {
		case len(poll.Options) == 0:
			const text = "poll cannot have no options"
			return gtserror.NewErrorBadRequest(errors.New(text), text)

		case len(poll.Options) > max:
			text := fmt.Sprintf("poll options exceed max count (%d)", max)
			return gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		max := config.GetStatusesPollOptionMaxChars()
		for i, option := range poll.Options {
			switch l := len([]rune(option)); {
			case l == 0:
				const text = "poll option cannot be empty"
				return gtserror.NewErrorBadRequest(errors.New(text), text)

			case l > max:
				text := fmt.Sprintf("poll option %d exceed max chars (%d)", i, max)
				return gtserror.NewErrorBadRequest(errors.New(text), text)
			}
		}
	}

	return nil
}

// statusContent encompasses the set of common processed
// status content fields from status write operations for
// an easily returnable type, without needing to allocate
// an entire gtsmodel.Status{} model.
type statusContent struct {
	Content        string
	ContentWarning string
	PollOptions    []string
	Language       string
	MentionIDs     []string
	Mentions       []*gtsmodel.Mention
	EmojiIDs       []string
	Emojis         []*gtsmodel.Emoji
	TagIDs         []string
	Tags           []*gtsmodel.Tag
}

// Returns the final content type to use when creating or editing a status.
func processContentType(
	requestContentType apimodel.StatusContentType,
	existingStatus *gtsmodel.Status,
	accountDefaultContentType string,
) gtsmodel.StatusContentType {
	switch {
	// Content type set in the request, return the new value.
	case requestContentType != "":
		return typeutils.APIContentTypeToContentType(requestContentType)

	// No content type in the request, return the existing
	// status's current content type if we know of one.
	case existingStatus != nil && existingStatus.ContentType != 0:
		return existingStatus.ContentType

	// We aren't editing an existing status, or if we are
	// it's an old one that doesn't have a saved content
	// type. Use the user's default content type setting.
	case accountDefaultContentType != "":
		return typeutils.APIContentTypeToContentType(apimodel.StatusContentType(accountDefaultContentType))

	// uhh.. Fall back to global default.
	default:
		return gtsmodel.StatusContentTypeDefault
	}
}

func (p *Processor) processContent(
	ctx context.Context,
	author *gtsmodel.Account,
	statusID string,
	contentType gtsmodel.StatusContentType,
	content string,
	contentWarning string,
	language string,
	poll *apimodel.PollRequest,
) (
	*statusContent,
	gtserror.WithCode,
) {
	if language == "" {
		// Ensure we have a status language.
		language = author.Settings.Language
		if language == "" {
			const text = "account default language unset"
			return nil, gtserror.NewErrorInternalError(
				errors.New(text),
			)
		}
	}

	var err error

	// Validate + normalize determined language.
	language, err = validate.Language(language)
	if err != nil {
		text := fmt.Sprintf("invalid language tag: %v", err)
		return nil, gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
	}

	var (
		// format is the currently set text formatting
		// function, according to the provided content-type.
		format text.FormatFunc
		// formatCW is like format, but for content warning.
		formatCW text.FormatFunc
	)

	switch contentType {

	// Format status according to text/plain.
	case gtsmodel.StatusContentTypePlain:
		format = p.formatter.FromPlain
		formatCW = p.formatter.FromPlainBasic

	// Format status according to text/markdown.
	case gtsmodel.StatusContentTypeMarkdown:
		format = p.formatter.FromMarkdown
		formatCW = p.formatter.FromMarkdownBasic

	// Unknown.
	default:
		const text = "invalid status format"
		return nil, gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
	}

	// Allocate a structure to hold the
	// majority of formatted content without
	// needing to alloc a whole gtsmodel.Status{}.
	var status statusContent
	status.Language = language

	// formatInput is a shorthand function to format the given input string with the
	// currently set 'formatFunc', passing in all required args and returning result.
	formatInput := func(formatFunc text.FormatFunc, input string) *text.FormatResult {
		return formatFunc(ctx, p.parseMention, author.ID, statusID, input)
	}

	// Sanitize input status text and format.
	contentRes := formatInput(format, content)

	// Gather results of formatted.
	status.Content = contentRes.HTML
	status.Mentions = contentRes.Mentions
	status.Emojis = contentRes.Emojis
	status.Tags = contentRes.Tags

	// Sanitize content warning and format.
	cwRes := formatInput(formatCW, contentWarning)

	// Gather results of the formatted.
	status.ContentWarning = cwRes.HTML
	status.Emojis = append(status.Emojis, cwRes.Emojis...)

	if poll != nil {
		// Pre-allocate slice of poll options of expected length.
		status.PollOptions = make([]string, len(poll.Options))
		for i, option := range poll.Options {

			// Strip each poll option and format.
			//
			// For polls just use basic formatting.
			option = text.StripHTMLFromText(option)
			optionRes := formatInput(p.formatter.FromPlainBasic, option)

			// Gather results of the formatted.
			status.PollOptions[i] = optionRes.HTML
			status.Emojis = append(status.Emojis, optionRes.Emojis...)
		}

		// Also update options on the form.
		poll.Options = status.PollOptions
	}

	// We may have received multiple copies of the same emoji, deduplicate these first.
	status.Emojis = xslices.DeduplicateFunc(status.Emojis, func(e *gtsmodel.Emoji) string {
		return e.ID
	})

	// Gather up the IDs of mentions from parsed content.
	status.MentionIDs = xslices.Gather(nil, status.Mentions,
		func(m *gtsmodel.Mention) string {
			return m.ID
		},
	)

	// Gather up the IDs of tags from parsed content.
	status.TagIDs = xslices.Gather(nil, status.Tags,
		func(t *gtsmodel.Tag) string {
			return t.ID
		},
	)

	// Gather up the IDs of emojis in updated content.
	status.EmojiIDs = xslices.Gather(nil, status.Emojis,
		func(e *gtsmodel.Emoji) string {
			return e.ID
		},
	)

	return &status, nil
}

func (p *Processor) processMedia(
	ctx context.Context,
	authorID string,
	statusID string,
	mediaIDs []string,
) (
	[]*gtsmodel.MediaAttachment,
	gtserror.WithCode,
) {
	// No media provided!
	if len(mediaIDs) == 0 {
		return nil, nil
	}

	// Get configured min/max supported descr chars.
	minChars := config.GetMediaDescriptionMinChars()
	maxChars := config.GetMediaDescriptionMaxChars()

	// Pre-allocate slice of media attachments of expected length.
	attachments := make([]*gtsmodel.MediaAttachment, len(mediaIDs))
	for i, id := range mediaIDs {

		// Look for media attachment by ID in database.
		media, err := p.state.DB.GetAttachmentByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting media from db: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Check media exists and is owned by author
		// (this masks finding out media ownership info).
		if media == nil || media.AccountID != authorID {
			text := fmt.Sprintf("media not found: %s", id)
			return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// Check media isn't already attached to another status.
		if (media.StatusID != "" && media.StatusID != statusID) ||
			(media.ScheduledStatusID != "" && media.ScheduledStatusID != statusID) {
			text := fmt.Sprintf("media already attached to status: %s", id)
			return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// Check media description chars within range,
		// this needs to be done here as lots of clients
		// only update media description on status post.
		switch chars := len([]rune(media.Description)); {
		case chars < minChars:
			text := fmt.Sprintf("media description less than min chars (%d)", minChars)
			return nil, gtserror.NewErrorBadRequest(errors.New(text), text)

		case chars > maxChars:
			text := fmt.Sprintf("media description exceeds max chars (%d)", maxChars)
			return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// Set media at index.
		attachments[i] = media
	}

	return attachments, nil
}

func (p *Processor) processPoll(
	ctx context.Context,
	statusID string,
	form *apimodel.PollRequest,
	now time.Time, // used for expiry time
) (
	*gtsmodel.Poll,
	gtserror.WithCode,
) {
	var expiresAt time.Time

	// Set an expiry time if one given.
	if in := form.ExpiresIn; in > 0 {
		expiresIn := time.Duration(in)
		expiresAt = now.Add(expiresIn * time.Second)
	}

	// Create new poll model.
	poll := &gtsmodel.Poll{
		ID:         id.NewULIDFromTime(now),
		Multiple:   &form.Multiple,
		HideCounts: &form.HideTotals,
		Options:    form.Options,
		StatusID:   statusID,
		ExpiresAt:  expiresAt,
	}

	// Insert the newly created poll model in the database.
	if err := p.state.DB.PutPoll(ctx, poll); err != nil {
		err := gtserror.Newf("error inserting poll in db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return poll, nil
}
