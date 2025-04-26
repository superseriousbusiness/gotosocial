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

package model

import (
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/language"
)

// Status models a status or post.
//
// swagger:model status
type Status struct {
	// ID of the status.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	ID string `json:"id"`
	// The date when this status was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`
	// Timestamp of when the status was last edited (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	// nullable: true
	EditedAt *string `json:"edited_at"`
	// ID of the status being replied to.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	// nullable: true
	InReplyToID *string `json:"in_reply_to_id"`
	// ID of the account being replied to.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	// nullable: true
	InReplyToAccountID *string `json:"in_reply_to_account_id"`
	// Status contains sensitive content.
	// example: false
	Sensitive bool `json:"sensitive"`
	// Subject, summary, or content warning for the status.
	// example: warning nsfw
	SpoilerText string `json:"spoiler_text"`
	// Visibility of this status.
	// example: unlisted
	Visibility Visibility `json:"visibility"`
	// Set to "true" if status is not federated, ie., a "local only" status; omitted from response otherwise.
	LocalOnly bool `json:"local_only,omitempty"`
	// Primary language of this status (ISO 639 Part 1 two-letter language code).
	// Will be null if language is not known.
	// example: en
	Language *string `json:"language"`
	// ActivityPub URI of the status. Equivalent to the status's activitypub ID.
	// example: https://example.org/users/some_user/statuses/01FBVD42CQ3ZEEVMW180SBX03B
	URI string `json:"uri"`
	// The status's publicly available web URL. This link will only work if the visibility of the status is 'public'.
	// example: https://example.org/@some_user/statuses/01FBVD42CQ3ZEEVMW180SBX03B
	URL string `json:"url"`
	// Number of replies to this status, according to our instance.
	RepliesCount int `json:"replies_count"`
	// Number of times this status has been boosted/reblogged, according to our instance.
	ReblogsCount int `json:"reblogs_count"`
	// Number of favourites/likes this status has received, according to our instance.
	FavouritesCount int `json:"favourites_count"`
	// This status has been favourited by the account viewing it.
	Favourited bool `json:"favourited"`
	// This status has been boosted/reblogged by the account viewing it.
	Reblogged bool `json:"reblogged"`
	// Replies to this status have been muted by the account viewing it.
	Muted bool `json:"muted"`
	// This status has been bookmarked by the account viewing it.
	Bookmarked bool `json:"bookmarked"`
	// This status has been pinned by the account viewing it (only relevant for your own statuses).
	Pinned bool `json:"pinned"`
	// The content of this status. Should be HTML, but might also be plaintext in some cases.
	// example: <p>Hey this is a status!</p>
	Content string `json:"content"`
	// The status that this status reblogs/boosts.
	// nullable: true
	Reblog *StatusReblogged `json:"reblog"`
	// The application used to post this status, if visible.
	Application *Application `json:"application,omitempty"`
	// The account that authored this status.
	Account *Account `json:"account"`
	// Media that is attached to this status.
	MediaAttachments []*Attachment `json:"media_attachments"`
	// Mentions of users within the status content.
	Mentions []Mention `json:"mentions"`
	// Hashtags used within the status content.
	Tags []Tag `json:"tags"`
	// Custom emoji to be used when rendering status content.
	Emojis []Emoji `json:"emojis"`
	// Preview card for links included within status content.
	// nullable: true
	Card *Card `json:"card"`
	// The poll attached to the status.
	// nullable: true
	Poll *Poll `json:"poll"`
	// Plain-text source of a status. Returned instead of content when status is deleted,
	// so the user may redraft from the source text without the client having to reverse-engineer
	// the original text from the HTML content.
	Text string `json:"text,omitempty"`
	// Content type that was used to parse the status's text. Returned when
	// status is deleted, so if the user is redrafting the message the client
	// can default to the same content type.
	ContentType StatusContentType `json:"content_type,omitempty"`
	// A list of filters that matched this status and why they matched, if there are any such filters.
	Filtered []FilterResult `json:"filtered,omitempty"`
	// The interaction policy for this status, as set by the status author.
	InteractionPolicy InteractionPolicy `json:"interaction_policy"`
}

// WebStatus is like *model.Status, but contains
// additional fields used only for HTML templating.
//
// swagger:ignore
type WebStatus struct {
	*Status

	// HTML version of spoiler content
	// (ie., not converted to plaintext).
	SpoilerContent string `json:"-"`

	// Override API account with web account.
	Account *WebAccount `json:"account"`

	// Web version of media
	// attached to this status.
	MediaAttachments []*WebAttachment `json:"media_attachments"`

	// Template-ready language tag and
	// string, based on *status.Language.
	LanguageTag *language.Language

	// Template-ready poll options with vote shares
	// calculated as a percentage of total votes.
	PollOptions []WebPollOption

	// Status is from a local account.
	Local bool

	// Level of indentation at which to
	// display this status in the web view.
	Indent int

	// This status is the last visible status
	// in the main thread, so everything below
	// can be considered "replies".
	ThreadLastMain bool

	// This status is the one around which
	// the thread context was constructed.
	ThreadContextStatus bool

	// This status is the first visibile status
	// after the "main" thread, so it and everything
	// below it can be considered "replies".
	ThreadFirstReply bool

	// Sorted slice of StatusEdit times for
	// this status, from latest to oldest.
	// Only set if status has been edited.
	// Last entry is always creation time.
	EditTimeline []string `json:"-"`
}

/*
** The below functions are added onto the API model status so that it satisfies
** the Preparable interface in internal/timeline.
 */

func (s *Status) GetID() string {
	return s.ID
}

func (s *Status) GetAccountID() string {
	if s.Account != nil {
		return s.Account.ID
	}
	return ""
}

func (s *Status) GetBoostOfID() string {
	if s.Reblog != nil {
		return s.Reblog.ID
	}
	return ""
}

func (s *Status) GetBoostOfAccountID() string {
	if s.Reblog != nil && s.Reblog.Account != nil {
		return s.Reblog.Account.ID
	}
	return ""
}

// StatusReblogged represents a reblogged status.
//
// swagger:model statusReblogged
type StatusReblogged struct {
	*Status
}

// StatusCreateRequest models status creation parameters.
//
// swagger:ignore
type StatusCreateRequest struct {

	// Text content of the status.
	// If media_ids is provided, this becomes optional.
	// Attaching a poll is optional while status is provided.
	Status string `form:"status" json:"status"`

	// Array of Attachment ids to be attached as media.
	// If provided, status becomes optional, and poll cannot be used.
	MediaIDs []string `form:"media_ids[]" json:"media_ids"`

	// Poll to include with this status.
	Poll *PollRequest `form:"poll" json:"poll"`

	// ID of the status being replied to, if status is a reply.
	InReplyToID string `form:"in_reply_to_id" json:"in_reply_to_id"`

	// Status and attached media should be marked as sensitive.
	Sensitive bool `form:"sensitive" json:"sensitive"`

	// Text to be shown as a warning or subject before the actual content.
	// Statuses are generally collapsed behind this field.
	SpoilerText string `form:"spoiler_text" json:"spoiler_text"`

	// Visibility of the posted status.
	Visibility Visibility `form:"visibility" json:"visibility"`

	// Set to "true" if this status should not be
	// federated,ie. it should be a "local only" status.
	LocalOnly *bool `form:"local_only" json:"local_only"`

	// Deprecated: Only used if LocalOnly is not set.
	Federated *bool `form:"federated" json:"federated"`

	// ISO 8601 Datetime at which to schedule a status.
	//
	// Providing this parameter with a *future* time will cause ScheduledStatus to be returned instead of Status.
	// Must be at least 5 minutes in the future.
	// This feature isn't implemented yet.
	//
	// Providing this parameter with a *past* time will cause the status to be backdated,
	// and will not push it to the user's followers. This is intended for importing old statuses.
	ScheduledAt *time.Time `form:"scheduled_at" json:"scheduled_at"`

	// ISO 639 language code for this status.
	Language string `form:"language" json:"language"`

	// Content type to use when parsing this status.
	ContentType StatusContentType `form:"content_type" json:"content_type"`

	// Interaction policy to use for this status.
	InteractionPolicy *InteractionPolicy `form:"-" json:"interaction_policy"`
}

// Separate form for parsing interaction
// policy on status create requests.
//
// swagger:ignore
type StatusInteractionPolicyForm struct {

	// Interaction policy to use for this status.
	InteractionPolicy *InteractionPolicy `form:"interaction_policy" json:"-"`
}

// Visibility models the visibility of a status.
//
// swagger:enum statusVisibility
// swagger:type string
type Visibility string

const (
	// VisibilityNone is visible to nobody. This is only used for the visibility of web statuses.
	VisibilityNone Visibility = "none"
	// VisibilityPublic is visible to everyone, and will be available via the web even for nonauthenticated users.

	VisibilityPublic Visibility = "public"

	// VisibilityUnlisted is visible to everyone, but only on home timelines, lists, etc.
	VisibilityUnlisted Visibility = "unlisted"

	// VisibilityPrivate is visible only to followers of the account that posted the status.
	VisibilityPrivate Visibility = "private"

	// VisibilityMutualsOnly is visible only to mutual followers of the account that posted the status.
	VisibilityMutualsOnly Visibility = "mutuals_only"

	// VisibilityDirect is visible only to accounts tagged in the status. It is equivalent to a direct message.
	VisibilityDirect Visibility = "direct"
)

// StatusContentType is the content type with which to parse the submitted status.
// Can be either text/plain or text/markdown. Empty will default to text/plain.
//
// swagger:enum statusContentType
// swagger:type string
type StatusContentType string

// Content type to use when parsing submitted
// status into an html-formatted status.
const (
	StatusContentTypePlain    StatusContentType = "text/plain"
	StatusContentTypeMarkdown StatusContentType = "text/markdown"
	StatusContentTypeDefault                    = StatusContentTypePlain
)

// StatusSource represents the source text of a
// status as submitted to the API when it was created.
//
// swagger:model statusSource
type StatusSource struct {

	// ID of the status.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	ID string `json:"id"`

	// Plain-text source of a status.
	Text string `json:"text"`

	// Plain-text version of spoiler text.
	SpoilerText string `json:"spoiler_text"`

	// Content type that was used to parse the text.
	ContentType StatusContentType `json:"content_type,omitempty"`
}

// StatusEdit represents one historical revision of a status, containing
// partial information about the state of the status at that revision.
//
// swagger:model statusEdit
type StatusEdit struct {

	// The content of this status at this revision.
	// Should be HTML, but might also be plaintext in some cases.
	// example: <p>Hey this is a status!</p>
	Content string `json:"content"`

	// Subject, summary, or content warning for the status at this revision.
	// example: warning nsfw
	SpoilerText string `json:"spoiler_text"`

	// Status marked sensitive at this revision.
	// example: false
	Sensitive bool `json:"sensitive"`

	// The date when this revision was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`

	// The account that authored this status.
	Account *Account `json:"account"`

	// The poll attached to the status at this revision.
	// Note that edits changing the poll options will be collapsed together into one edit, since this action resets the poll.
	// nullable: true
	Poll *Poll `json:"poll"`

	// Media that is attached to this status.
	MediaAttachments []*Attachment `json:"media_attachments"`

	// Custom emoji to be used when rendering status content.
	Emojis []Emoji `json:"emojis"`
}

// StatusEditRequest models status edit parameters.
//
// swagger:ignore
type StatusEditRequest struct {

	// Text content of the status.
	// If media_ids is provided, this becomes optional.
	// Attaching a poll is optional while status is provided.
	Status string `form:"status" json:"status"`

	// Text to be shown as a warning or subject before the actual content.
	// Statuses are generally collapsed behind this field.
	SpoilerText string `form:"spoiler_text" json:"spoiler_text"`

	// Content type to use when parsing this status.
	ContentType StatusContentType `form:"content_type" json:"content_type"`

	// Status and attached media should be marked as sensitive.
	Sensitive bool `form:"sensitive" json:"sensitive"`

	// ISO 639 language code for this status.
	Language string `form:"language" json:"language"`

	// Array of Attachment ids to be attached as media.
	// If provided, status becomes optional, and poll cannot be used.
	MediaIDs []string `form:"media_ids[]" json:"media_ids"`

	// Array of Attachment attributes to be updated in attached media.
	MediaAttributes []AttachmentAttributesRequest `form:"media_attributes[]" json:"media_attributes"`

	// Poll to include with this status.
	Poll *PollRequest `form:"poll" json:"poll"`
}
