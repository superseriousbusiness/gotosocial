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

import "time"

// Emoji represents a custom emoji that's been uploaded
// through the admin UI or downloaded from a remote instance.
type Emoji struct {
	ID                     string         `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt              time.Time      `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time      `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Shortcode              string         `bun:",nullzero,notnull,unique:domainshortcode"`                    // String shortcode for this emoji -- the part that's between colons. This should be a-zA-Z_  eg., 'blob_hug' 'purple_heart' 'Gay_Otter' Must be unique with domain.
	Domain                 string         `bun:",nullzero,unique:domainshortcode"`                            // Origin domain of this emoji, eg 'example.org', 'queer.party'. empty string for local emojis.
	ImageRemoteURL         string         `bun:",nullzero"`                                                   // Where can this emoji be retrieved remotely? Null for local emojis.
	ImageStaticRemoteURL   string         `bun:",nullzero"`                                                   // Where can a static / non-animated version of this emoji be retrieved remotely? Null for local emojis.
	ImageURL               string         `bun:",nullzero"`                                                   // Where can this emoji be retrieved from the local server? Null for remote emojis.
	ImageStaticURL         string         `bun:",nullzero"`                                                   // Where can a static version of this emoji be retrieved from the local server? Null for remote emojis.
	ImagePath              string         `bun:",notnull"`                                                    // Path of the emoji image in the server storage system.
	ImageStaticPath        string         `bun:",notnull"`                                                    // Path of a static version of the emoji image in the server storage system
	ImageContentType       string         `bun:",notnull"`                                                    // MIME content type of the emoji image
	ImageStaticContentType string         `bun:",notnull"`                                                    // MIME content type of the static version of the emoji image.
	ImageFileSize          int            `bun:",notnull"`                                                    // Size of the emoji image file in bytes, for serving purposes.
	ImageStaticFileSize    int            `bun:",notnull"`                                                    // Size of the static version of the emoji image file in bytes, for serving purposes.
	Disabled               *bool          `bun:",nullzero,notnull,default:false"`                             // Has a moderation action disabled this emoji from being shown?
	URI                    string         `bun:",nullzero,notnull,unique"`                                    // ActivityPub uri of this emoji. Something like 'https://example.org/emojis/1234'
	VisibleInPicker        *bool          `bun:",nullzero,notnull,default:true"`                              // Is this emoji visible in the admin emoji picker?
	Category               *EmojiCategory `bun:"rel:belongs-to"`                                              // In which emoji category is this emoji visible?
	CategoryID             string         `bun:"type:CHAR(26),nullzero"`                                      // ID of the category this emoji belongs to.
	Cached                 *bool          `bun:",nullzero,notnull,default:false"`                             // whether emoji is cached in locally in gotosocial storage.
}

// IsLocal returns true if the emoji is
// local to this instance., ie., it did
// not originate from a remote instance.
func (e *Emoji) IsLocal() bool {
	return e.Domain == ""
}

// ShortcodeDomain returns the [shortcode]@[domain] for the given emoji.
func (e *Emoji) ShortcodeDomain() string {
	return e.Shortcode + "@" + e.Domain
}
