/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package users

// SwaggerCollection represents an activitypub collection.
// swagger:model swaggerCollection
type SwaggerCollection struct {
	// ActivityStreams context.
	// example: https://www.w3.org/ns/activitystreams
	Context string `json:"@context"`
	// ActivityStreams ID.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies
	ID string `json:"id"`
	// ActivityStreams type.
	// example: Collection
	Type string `json:"type"`
	// ActivityStreams first property.
	First SwaggerCollectionPage `json:"first"`
	// ActivityStreams last property.
	Last SwaggerCollectionPage `json:"last,omitempty"`
}

// SwaggerCollectionPage represents one page of a collection.
// swagger:model swaggerCollectionPage
type SwaggerCollectionPage struct {
	// ActivityStreams ID.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies?page=true
	ID string `json:"id"`
	// ActivityStreams type.
	// example: CollectionPage
	Type string `json:"type"`
	// Link to the next page.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies?only_other_accounts=true&page=true
	Next string `json:"next"`
	// Collection this page belongs to.
	// example: https://example.org/users/some_user/statuses/106717595988259568/replies
	PartOf string `json:"partOf"`
	// Items on this page.
	// example: ["https://example.org/users/some_other_user/statuses/086417595981111564", "https://another.example.com/users/another_user/statuses/01FCN8XDV3YG7B4R42QA6YQZ9R"]
	Items []string `json:"items"`
}
