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

package advancedmigrations

import (
	"context"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/processing/conversations"
)

// Processor holds references to any other processor that has migrations to run.
type Processor struct {
	conversations *conversations.Processor
}

func New(
	conversations *conversations.Processor,
) Processor {
	return Processor{
		conversations: conversations,
	}
}

// Migrate runs all advanced migrations.
// Errors should be in the same format thrown by other server or testrig startup failures.
func (p *Processor) Migrate(ctx context.Context) error {
	if err := p.conversations.MigrateDMsToConversations(ctx); err != nil {
		return fmt.Errorf("error running conversations advanced migration: %w", err)
	}

	return nil
}
