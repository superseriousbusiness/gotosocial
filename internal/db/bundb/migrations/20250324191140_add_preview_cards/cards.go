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

// Package gtsmodel contains types used *internally* by GoToSocial and added/removed/selected from the database.
// These types should never be serialized and/or sent out via public APIs, as they contain sensitive information.
// The annotation used on these structs is for handling them via the bun-db ORM.
// See here for more info on bun model annotations: https://bun.uptrace.dev/guide/models.html
package gtsmodel

type Card struct {
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // Unique identity string.
	// Location of linked resource.
	// example: https://buzzfeed.com/some/fuckin/buzzfeed/article
	URL string `bun:",nullzero"`
	// Title of linked resource.
	// example: Buzzfeed - Is Water Wet?
	Title string `bun:",nullzero"`
	// Description of preview.
	// example: Is water wet? We're not sure. In this article, we ask an expert...
	Description string `bun:",nullzero"`
	// The type of the preview card.
	// enum:
	// - link
	// - photo
	// - video
	// - rich
	// example: link
	Type string `bun:",nullzero"`
	// The author of the original resource.
	// example: weewee@buzzfeed.com
	AuthorName string `bun:"author_name,nullzero"`
	// A link to the author of the original resource.
	// example: https://buzzfeed.com/authors/weewee
	AuthorURL string `bun:"author_url,nullzero"`
	// The provider of the original resource.
	// example: Buzzfeed
	ProviderName string `bun:"provider_name,nullzero"`
	// A link to the provider of the original resource.
	// example: https://buzzfeed.com
	ProviderURL string `bun:"provider_url,nullzero"`
	// HTML to be used for generating the preview card.
	HTML string `bun:",nullzero"`
	// Width of preview, in pixels.
	Width int `bun:",nullzero"`
	// Height of preview, in pixels.
	Height int `bun:",nullzero"`
	// Preview thumbnail.
	// example: https://example.org/fileserver/preview/thumb.jpg
	Image string `bun:",nullzero"`
	// Used for photo embeds, instead of custom html.
	EmbedURL string `bun:",nullzero"`
	// A hash computed by the BlurHash algorithm, for generating colorful preview thumbnails when media has not been downloaded yet.
	Blurhash string `bun:",nullzero"`
}
