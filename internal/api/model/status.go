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

package model

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
	MediaAttachments []Attachment `json:"media_attachments"`
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
// swagger:model statusCreateRequest
type StatusCreateRequest struct {
	// Text content of the status.
	// If media_ids is provided, this becomes optional.
	// Attaching a poll is optional while status is provided.
	// in: formData
	Status string `form:"status" json:"status" xml:"status"`
	// Array of Attachment ids to be attached as media.
	// If provided, status becomes optional, and poll cannot be used.
	//
	// If the status is being submitted as a form, the key is 'media_ids[]',
	// but if it's json or xml, the key is 'media_ids'.
	//
	// in: formData
	MediaIDs []string `form:"media_ids[]" json:"media_ids" xml:"media_ids"`
	// Poll to include with this status.
	// swagger:ignore
	Poll *PollRequest `form:"poll" json:"poll" xml:"poll"`
	// ID of the status being replied to, if status is a reply.
	// in: formData
	InReplyToID string `form:"in_reply_to_id" json:"in_reply_to_id" xml:"in_reply_to_id"`
	// Status and attached media should be marked as sensitive.
	// in: formData
	Sensitive bool `form:"sensitive" json:"sensitive" xml:"sensitive"`
	// Text to be shown as a warning or subject before the actual content.
	// Statuses are generally collapsed behind this field.
	// in: formData
	SpoilerText string `form:"spoiler_text" json:"spoiler_text" xml:"spoiler_text"`
	// Visibility of the posted status.
	// in: formData
	Visibility Visibility `form:"visibility" json:"visibility" xml:"visibility"`
	// ISO 8601 Datetime at which to schedule a status.
	// Providing this parameter will cause ScheduledStatus to be returned instead of Status.
	// Must be at least 5 minutes in the future.
	// in: formData
	ScheduledAt string `form:"scheduled_at" json:"scheduled_at" xml:"scheduled_at"`
	// ISO 639 language code for this status.
	// in: formData
	Language string `form:"language" json:"language" xml:"language"`
	// Format to use when parsing this status.
	// in: formData
	Format StatusFormat `form:"format" json:"format" xml:"format"`
}

// Visibility models the visibility of a status.
//
// swagger:enum statusVisibility
// swagger:type string
type Visibility string

const (
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

// AdvancedStatusCreateForm wraps the mastodon-compatible status create form along with the GTS advanced
// visibility settings.
//
// swagger:parameters statusCreate
type AdvancedStatusCreateForm struct {
	StatusCreateRequest
	AdvancedVisibilityFlagsForm
}

// AdvancedVisibilityFlagsForm allows a few more advanced flags to be set on new statuses, in addition
// to the standard mastodon-compatible ones.
//
// swagger:model advancedVisibilityFlagsForm
type AdvancedVisibilityFlagsForm struct {
	// This status will be federated beyond the local timeline(s).
	Federated *bool `form:"federated" json:"federated" xml:"federated"`
	// This status can be boosted/reblogged.
	Boostable *bool `form:"boostable" json:"boostable" xml:"boostable"`
	// This status can be replied to.
	Replyable *bool `form:"replyable" json:"replyable" xml:"replyable"`
	// This status can be liked/faved.
	Likeable *bool `form:"likeable" json:"likeable" xml:"likeable"`
}

// StatusFormat is the format in which to parse the submitted status.
// Can be either plain or markdown. Empty will default to plain.
//
// swagger:enum statusFormat
// swagger:type string
type StatusFormat string

// Format to use when parsing submitted status into an html-formatted status
const (
	StatusFormatPlain    StatusFormat = "plain"
	StatusFormatMarkdown StatusFormat = "markdown"
	StatusFormatDefault  StatusFormat = StatusFormatPlain
)
