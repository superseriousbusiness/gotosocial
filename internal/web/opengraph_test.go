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

package web

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OpenGraphTestSuite struct {
	suite.Suite
}

func (suite *OpenGraphTestSuite) TestParseDescription() {
	tests := []struct {
		name, in, exp string
	}{
		{name: "shellcmd", in: `echo '\e]8;;http://example.com\e\This is a link\e]8;;\e'`, exp: `echo &#39;&bsol;e]8;;http://example.com&bsol;e&bsol;This is a link&bsol;e]8;;&bsol;e&#39;`},
		{name: "newlines", in: "test\n\ntest\ntest", exp: "test test test"},
	}

	for _, tt := range tests {
		tt := tt
		suite.Run(tt.name, func() {
			suite.Equal(fmt.Sprintf("content=\"%s\"", tt.exp), parseDescription(tt.in))
		})
	}
}

func TestOpenGraphTestSuite(t *testing.T) {
	suite.Run(t, &OpenGraphTestSuite{})
}
