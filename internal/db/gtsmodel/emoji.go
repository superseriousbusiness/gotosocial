/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package gtsmodel

import "time"

type Emoji struct {
	// database ID of this emoji
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	// String shortcode for this emoji -- the part that's between colons. This should be lowercase a-z_
	// eg., 'blob_hug' 'purple_heart' Must be unique with domain.
	Shortcode string `pg:"notnull,unique:shortcodedomain"`
	// Origin domain of this emoji, eg 'example.org', 'queer.party'. Null for local emojis.
	Domain string `pg:",unique:shortcodedomain"`
	// When was this emoji created. Must be unique with shortcode.
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this emoji updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Where can this emoji be retrieved remotely? Null for local emojis.
	// For remote emojis, it'll be something like:
	// https://hackers.town/system/custom_emojis/images/000/049/842/original/1b74481204feabfd.png
	ImageRemoteURL string
	// Where can a static / non-animated version of this emoji be retrieved remotely? Null for local emojis.
	// For remote emojis, it'll be something like:
	// https://hackers.town/system/custom_emojis/images/000/049/842/static/1b74481204feabfd.png
	ImageStaticRemoteURL string
	// Where can this emoji be retrieved from the local server? Null for remote emojis.
	// Assuming our server is hosted at 'example.org', this will be something like:
	// 'https://example.org/fileserver/6339820e-ef65-4166-a262-5a9f46adb1a7/emoji/original/bfa6c9c5-6c25-4ea4-98b4-d78b8126fb52.png'
	ImageURL string
	// Where can a static version of this emoji be retrieved from the local server? Null for remote emojis.
	// Assuming our server is hosted at 'example.org', this will be something like:
	// 'https://example.org/fileserver/6339820e-ef65-4166-a262-5a9f46adb1a7/emoji/small/bfa6c9c5-6c25-4ea4-98b4-d78b8126fb52.png'
	ImageStaticURL string
	// Path of the emoji image in the server storage system. Will be something like:
	// '/gotosocial/storage/6339820e-ef65-4166-a262-5a9f46adb1a7/emoji/original/bfa6c9c5-6c25-4ea4-98b4-d78b8126fb52.png'
	ImagePath string `pg:",notnull"`
	// Path of a static version of the emoji image in the server storage system. Will be something like:
	// '/gotosocial/storage/6339820e-ef65-4166-a262-5a9f46adb1a7/emoji/small/bfa6c9c5-6c25-4ea4-98b4-d78b8126fb52.png'
	ImageStaticPath string `pg:",notnull"`
	// MIME content type of the emoji image
	// Probably "image/png"
	ImageContentType string `pg:",notnull"`
	// Size of the emoji image file in bytes, for serving purposes.
	ImageFileSize int `pg:",notnull"`
	// Size of the static version of the emoji image file in bytes, for serving purposes.
	ImageStaticFileSize int `pg:",notnull"`
	// When was the emoji image last updated?
	ImageUpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Has a moderation action disabled this emoji from being shown?
	Disabled bool `pg:",notnull,default:false"`
	// ActivityStreams uri of this emoji. Something like 'https://example.org/emojis/1234'
	URI string `pg:",notnull,unique"`
	// Is this emoji visible in the admin emoji picker?
	VisibleInPicker bool `pg:",notnull,default:true"`
	// In which emoji category is this emoji visible?
	CategoryID string
}
