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

package emoji

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	// EmojiIDKey is for emoji IDs
	EmojiIDKey = "id"
	// EmojiBasePath is the base path for serving information about Emojis eg https://example.org/emoji
	EmojiWithIDPath = "/" + uris.EmojiPath + "/:" + EmojiIDKey
)

type Module struct {
	processor processing.Processor
}

// New returns a emoji module
func New(processor processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, EmojiWithIDPath, m.EmojiGetHandler)
	return nil
}
