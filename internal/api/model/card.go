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

package model

// Card represents a rich preview card that is generated using OpenGraph tags from a URL. See here: https://docs.joinmastodon.org/entities/card/
type Card struct {
	// REQUIRED

	// Location of linked resource.
	URL string `json:"url"`
	// Title of linked resource.
	Title string `json:"title"`
	// Description of preview.
	Description string `json:"description"`
	// The type of the preview card.
	//    String (Enumerable, oneOf)
	//    link = Link OEmbed
	//    photo = Photo OEmbed
	//    video = Video OEmbed
	//    rich = iframe OEmbed. Not currently accepted, so won't show up in practice.
	Type string `json:"type"`

	// OPTIONAL

	// The author of the original resource.
	AuthorName string `json:"author_name"`
	// A link to the author of the original resource.
	AuthorURL string `json:"author_url"`
	// The provider of the original resource.
	ProviderName string `json:"provider_name"`
	// A link to the provider of the original resource.
	ProviderURL string `json:"provider_url"`
	// HTML to be used for generating the preview card.
	HTML string `json:"html"`
	// Width of preview, in pixels.
	Width int `json:"width"`
	// Height of preview, in pixels.
	Height int `json:"height"`
	// Preview thumbnail.
	Image string `json:"image"`
	// Used for photo embeds, instead of custom html.
	EmbedURL string `json:"embed_url"`
	// A hash computed by the BlurHash algorithm, for generating colorful preview thumbnails when media has not been downloaded yet.
	Blurhash string `json:"blurhash"`
}
