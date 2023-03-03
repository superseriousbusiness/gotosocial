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

package status

import (
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

type Processor struct {
	state        *state.State
	tc           typeutils.TypeConverter
	filter       visibility.Filter
	formatter    text.Formatter
	parseMention gtsmodel.ParseMentionFunc
}

// New returns a new status processor.
func New(state *state.State, tc typeutils.TypeConverter, parseMention gtsmodel.ParseMentionFunc) Processor {
	return Processor{
		state:        state,
		tc:           tc,
		filter:       visibility.NewFilter(state.DB),
		formatter:    text.NewFormatter(state.DB),
		parseMention: parseMention,
	}
}
