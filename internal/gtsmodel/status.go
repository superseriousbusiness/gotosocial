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

package gtsmodel

import (
	"slices"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
)

// Status represents a user-created 'post' or 'status' in the database, either remote or local
type Status struct {
	ID                       string             `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt                time.Time          `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	EditedAt                 time.Time          `bun:"type:timestamptz,nullzero"`                                   // when this status was last edited (if set)
	FetchedAt                time.Time          `bun:"type:timestamptz,nullzero"`                                   // when was item (remote) last fetched.
	PinnedAt                 time.Time          `bun:"type:timestamptz,nullzero"`                                   // Status was pinned by owning account at this time.
	URI                      string             `bun:",unique,nullzero,notnull"`                                    // activitypub URI of this status
	URL                      string             `bun:",nullzero"`                                                   // web url for viewing this status
	Content                  string             `bun:""`                                                            // Content HTML for this status.
	AttachmentIDs            []string           `bun:"attachments,array"`                                           // Database IDs of any media attachments associated with this status
	Attachments              []*MediaAttachment `bun:"attached_media,rel:has-many"`                                 // Attachments corresponding to attachmentIDs
	TagIDs                   []string           `bun:"tags,array"`                                                  // Database IDs of any tags used in this status
	Tags                     []*Tag             `bun:"attached_tags,m2m:status_to_tags"`                            // Tags corresponding to tagIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	MentionIDs               []string           `bun:"mentions,array"`                                              // Database IDs of any mentions in this status
	Mentions                 []*Mention         `bun:"attached_mentions,rel:has-many"`                              // Mentions corresponding to mentionIDs
	EmojiIDs                 []string           `bun:"emojis,array"`                                                // Database IDs of any emojis used in this status
	Emojis                   []*Emoji           `bun:"attached_emojis,m2m:status_to_emojis"`                        // Emojis corresponding to emojiIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	Local                    *bool              `bun:",nullzero,notnull,default:false"`                             // is this status from a local account?
	AccountID                string             `bun:"type:CHAR(26),nullzero,notnull"`                              // which account posted this status?
	Account                  *Account           `bun:"rel:belongs-to"`                                              // account corresponding to accountID
	AccountURI               string             `bun:",nullzero,notnull"`                                           // activitypub uri of the owner of this status
	InReplyToID              string             `bun:"type:CHAR(26),nullzero"`                                      // id of the status this status replies to
	InReplyToURI             string             `bun:",nullzero"`                                                   // activitypub uri of the status this status is a reply to
	InReplyToAccountID       string             `bun:"type:CHAR(26),nullzero"`                                      // id of the account that this status replies to
	InReplyTo                *Status            `bun:"-"`                                                           // status corresponding to inReplyToID
	InReplyToAccount         *Account           `bun:"rel:belongs-to"`                                              // account corresponding to inReplyToAccountID
	BoostOfID                string             `bun:"type:CHAR(26),nullzero"`                                      // id of the status this status is a boost of
	BoostOfURI               string             `bun:"-"`                                                           // URI of the status this status is a boost of; field not inserted in the db, just for dereferencing purposes.
	BoostOfAccountID         string             `bun:"type:CHAR(26),nullzero"`                                      // id of the account that owns the boosted status
	BoostOf                  *Status            `bun:"-"`                                                           // status that corresponds to boostOfID
	BoostOfAccount           *Account           `bun:"rel:belongs-to"`                                              // account that corresponds to boostOfAccountID
	ThreadID                 string             `bun:"type:CHAR(26),nullzero"`                                      // id of the thread to which this status belongs; only set for remote statuses if a local account is involved at some point in the thread, otherwise null
	EditIDs                  []string           `bun:"edits,array"`                                                 //
	Edits                    []*StatusEdit      `bun:"-"`                                                           //
	PollID                   string             `bun:"type:CHAR(26),nullzero"`                                      //
	Poll                     *Poll              `bun:"-"`                                                           //
	ContentWarning           string             `bun:",nullzero"`                                                   // Content warning HTML for this status.
	ContentWarningText       string             `bun:""`                                                            // Original text of the content warning without formatting
	Visibility               Visibility         `bun:",nullzero,notnull"`                                           // visibility entry for this status
	Sensitive                *bool              `bun:",nullzero,notnull,default:false"`                             // mark the status as sensitive?
	Language                 string             `bun:",nullzero"`                                                   // what language is this status written in?
	CreatedWithApplicationID string             `bun:"type:CHAR(26),nullzero"`                                      // Which application was used to create this status?
	CreatedWithApplication   *Application       `bun:"rel:belongs-to"`                                              // application corresponding to createdWithApplicationID
	ActivityStreamsType      string             `bun:",nullzero,notnull"`                                           // What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types. Will probably almost always be Note but who knows!.
	Text                     string             `bun:""`                                                            // Original text of the status without formatting
	ContentType              StatusContentType  `bun:",nullzero"`                                                   // Content type used to process the original text of the status
	Federated                *bool              `bun:",notnull"`                                                    // This status will be federated beyond the local timeline(s)
	InteractionPolicy        *InteractionPolicy `bun:""`                                                            // InteractionPolicy for this status. If null then the default InteractionPolicy should be assumed for this status's Visibility. Always null for boost wrappers.
	PendingApproval          *bool              `bun:",nullzero,notnull,default:false"`                             // If true then status is a reply or boost wrapper that must be Approved by the reply-ee or boost-ee before being fully distributed.
	PreApproved              bool               `bun:"-"`                                                           // If true, then status is a reply to or boost wrapper of a status on our instance, has permission to do the interaction, and an Accept should be sent out for it immediately. Field not stored in the DB.
	ApprovedByURI            string             `bun:",nullzero"`                                                   // URI of an Accept Activity that approves the Announce or Create Activity that this status was/will be attached to.
}

// GetID implements timeline.Timelineable{}.
func (s *Status) GetID() string {
	return s.ID
}

// GetAccountID implements timeline.Timelineable{}.
func (s *Status) GetAccountID() string {
	return s.AccountID
}

// GetAccount returns the account that owns
// this status. May be nil if status not populated.
// Fulfils Interaction interface.
func (s *Status) GetAccount() *Account {
	return s.Account
}

// GetBoostOfID implements timeline.Timelineable{}.
func (s *Status) GetBoostOfID() string {
	return s.BoostOfID
}

// GetBoostOfAccountID implements timeline.Timelineable{}.
func (s *Status) GetBoostOfAccountID() string {
	return s.BoostOfAccountID
}

// AttachmentsPopulated returns whether media attachments
// are populated according to current AttachmentIDs.
func (s *Status) AttachmentsPopulated() bool {
	if len(s.AttachmentIDs) != len(s.Attachments) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range s.AttachmentIDs {
		if s.Attachments[i].ID != id {
			return false
		}
	}
	return true
}

// TagsPopulated returns whether tags are
// populated according to current TagIDs.
func (s *Status) TagsPopulated() bool {
	if len(s.TagIDs) != len(s.Tags) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range s.TagIDs {
		if s.Tags[i].ID != id {
			return false
		}
	}
	return true
}

// MentionsPopulated returns whether mentions are
// populated according to current MentionIDs.
func (s *Status) MentionsPopulated() bool {
	if len(s.MentionIDs) != len(s.Mentions) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range s.MentionIDs {
		if s.Mentions[i].ID != id {
			return false
		}
	}
	return true
}

// EmojisPopulated returns whether emojis are
// populated according to current EmojiIDs.
func (s *Status) EmojisPopulated() bool {
	if len(s.EmojiIDs) != len(s.Emojis) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range s.EmojiIDs {
		if s.Emojis[i].ID != id {
			return false
		}
	}
	return true
}

// EditsPopulated returns whether edits are
// populated according to current EditIDs.
func (s *Status) EditsPopulated() bool {
	if len(s.EditIDs) != len(s.Edits) {
		// this is quickest indicator.
		return false
	}
	for i, id := range s.EditIDs {
		if s.Edits[i].ID != id {
			return false
		}
	}
	return true
}

// EmojisUpToDate returns whether status emoji attachments of receiving status are up-to-date
// according to emoji attachments of the passed status, by comparing their emoji URIs. We don't
// use IDs as this is used to determine whether there are new emojis to fetch.
func (s *Status) EmojisUpToDate(other *Status) bool {
	if len(s.Emojis) != len(other.Emojis) {
		// this is the quickest indicator.
		return false
	}
	for i := range s.Emojis {
		if s.Emojis[i].URI != other.Emojis[i].URI {
			return false
		}
	}
	return true
}

// GetAttachmentByRemoteURL searches status for MediaAttachment{} with remote URL.
func (s *Status) GetAttachmentByRemoteURL(url string) (*MediaAttachment, bool) {
	for _, media := range s.Attachments {
		if media.RemoteURL == url {
			return media, true
		}
	}
	return nil, false
}

// GetMentionByTargetURI searches status for Mention{} with target URI.
func (s *Status) GetMentionByTargetURI(uri string) (*Mention, bool) {
	for _, mention := range s.Mentions {
		if mention.TargetAccountURI == uri {
			return mention, true
		}
	}
	return nil, false
}

// GetMentionByTargetID searches status for Mention{} with target ID.
func (s *Status) GetMentionByTargetID(id string) (*Mention, bool) {
	for _, mention := range s.Mentions {
		if mention.TargetAccountID == id {
			return mention, true
		}
	}
	return nil, false
}

// GetMentionByUsernameDomain fetches the Mention associated with given
// username and domains, typically extracted from a mention Namestring.
func (s *Status) GetMentionByUsernameDomain(username, domain string) (*Mention, bool) {
	for _, mention := range s.Mentions {

		// We can only check if target
		// account is set on the mention.
		account := mention.TargetAccount
		if account == nil {
			continue
		}

		// Usernames must always match.
		if account.Username != username {
			continue
		}

		// Finally, either domains must
		// match or an empty domain may
		// be permitted if account local.
		if account.Domain == domain ||
			(domain == "" && account.IsLocal()) {
			return mention, true
		}
	}

	return nil, false
}

// GetTagByName searches status for Tag{} with name.
func (s *Status) GetTagByName(name string) (*Tag, bool) {
	for _, tag := range s.Tags {
		if tag.Name == name {
			return tag, true
		}
	}
	return nil, false
}

// MentionsAccount returns whether status mentions the given account ID.
func (s *Status) MentionsAccount(accountID string) bool {
	return slices.ContainsFunc(s.Mentions, func(m *Mention) bool {
		return m.TargetAccountID == accountID
	})
}

// BelongsToAccount returns whether status belongs to the given account ID.
func (s *Status) BelongsToAccount(accountID string) bool {
	return s.AccountID == accountID
}

// IsLocal returns true if this is a local
// status (ie., originating from this instance).
func (s *Status) IsLocal() bool {
	return s.Local != nil && *s.Local
}

// IsLocalOnly returns true if this status
// is "local-only" ie., unfederated.
func (s *Status) IsLocalOnly() bool {
	return s.Federated == nil || !*s.Federated
}

// AllAttachmentIDs gathers ALL media attachment IDs from both
// the receiving Status{}, and any historical Status{}.Edits.
func (s *Status) AllAttachmentIDs() []string {
	var total int

	// Check if this is being correctly
	// called on fully populated status.
	if !s.EditsPopulated() {
		log.Warnf(nil, "status edits not populated for %s", s.URI)
	}

	// Get count of attachment IDs.
	total += len(s.AttachmentIDs)
	for _, edit := range s.Edits {
		total += len(edit.AttachmentIDs)
	}

	// Start gathering of all IDs with *current* attachment IDs.
	attachmentIDs := make([]string, len(s.AttachmentIDs), total)
	copy(attachmentIDs, s.AttachmentIDs)

	// Append IDs of historical edits.
	for _, edit := range s.Edits {
		attachmentIDs = append(attachmentIDs, edit.AttachmentIDs...)
	}

	// Deduplicate these IDs in case of shared media.
	return xslices.Deduplicate(attachmentIDs)
}

// UpdatedAt returns latest time this status
// was updated, either EditedAt or CreatedAt.
func (s *Status) UpdatedAt() time.Time {
	if s.EditedAt.IsZero() {
		return s.CreatedAt
	}
	return s.EditedAt
}

// StatusToTag is an intermediate struct to facilitate the many2many relationship between a status and one or more tags.
type StatusToTag struct {
	StatusID string  `bun:"type:CHAR(26),unique:statustag,nullzero,notnull"`
	Status   *Status `bun:"rel:belongs-to"`
	TagID    string  `bun:"type:CHAR(26),unique:statustag,nullzero,notnull"`
	Tag      *Tag    `bun:"rel:belongs-to"`
}

// StatusToEmoji is an intermediate struct to facilitate the many2many relationship between a status and one or more emojis.
type StatusToEmoji struct {
	StatusID string  `bun:"type:CHAR(26),unique:statusemoji,nullzero,notnull"`
	Status   *Status `bun:"rel:belongs-to"`
	EmojiID  string  `bun:"type:CHAR(26),unique:statusemoji,nullzero,notnull"`
	Emoji    *Emoji  `bun:"rel:belongs-to"`
}

// Visibility represents the
// visibility granularity of a status.
type Visibility enumType

const (
	// VisibilityNone means nobody can see this.
	// It's only used for web status visibility.
	VisibilityNone Visibility = 1

	// VisibilityPublic means this status will
	// be visible to everyone on all timelines.
	VisibilityPublic Visibility = 2

	// VisibilityUnlocked means this status will be visible to everyone,
	// but will only show on home timeline to followers, and in lists.
	VisibilityUnlocked Visibility = 3

	// VisibilityFollowersOnly means this status is viewable to followers only.
	VisibilityFollowersOnly Visibility = 4

	// VisibilityMutualsOnly means this status
	// is visible to mutual followers only.
	VisibilityMutualsOnly Visibility = 5

	// VisibilityDirect means this status is
	// visible only to mentioned recipients.
	VisibilityDirect Visibility = 6

	// VisibilityDefault is used when no other setting can be found.
	VisibilityDefault Visibility = VisibilityUnlocked
)

// String returns a stringified, frontend API compatible form of Visibility.
func (v Visibility) String() string {
	switch v {
	case VisibilityNone:
		return "none"
	case VisibilityPublic:
		return "public"
	case VisibilityUnlocked:
		return "unlocked"
	case VisibilityFollowersOnly:
		return "followers_only"
	case VisibilityMutualsOnly:
		return "mutuals_only"
	case VisibilityDirect:
		return "direct"
	default:
		panic("invalid visibility")
	}
}

// StatusContentType is the content type with which a status's text is
// parsed. Can be either plain or markdown. Empty will default to plain.
type StatusContentType enumType

const (
	StatusContentTypePlain    StatusContentType = 1
	StatusContentTypeMarkdown StatusContentType = 2
	StatusContentTypeDefault                    = StatusContentTypePlain
)

// Content models the simple string content
// of a status along with its ContentMap,
// which contains content entries keyed by
// BCP47 language tag.
//
// Content and/or ContentMap may be zero/nil.
type Content struct {
	Content    string
	ContentMap map[string]string
}

// BackfillStatus is a wrapper for creating a status without pushing notifications to followers.
type BackfillStatus struct {
	*Status
}
