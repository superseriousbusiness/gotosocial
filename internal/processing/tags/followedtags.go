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

package tags

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	state     *state.State
	converter *typeutils.Converter
}

func New(state *state.State, converter *typeutils.Converter) Processor {
	return Processor{
		state:     state,
		converter: converter,
	}
}

// apiTag is a shortcut to return the API version of the given tag,
// or return an appropriate error if conversion fails.
func (p *Processor) apiTag(ctx context.Context, tag *gtsmodel.Tag, following bool) (*apimodel.Tag, gtserror.WithCode) {
	apiTag, err := p.converter.TagToAPITag(ctx, tag, true, &following)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("error converting tag %s to API representation: %w", tag.Name, err),
		)
	}

	return &apiTag, nil
}
