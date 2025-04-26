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

package text_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/text"
	"github.com/stretchr/testify/suite"
)

const (
	sanitizeHTML      = `here's some naughty html: <script>alert(ahhhh)</script> !!!`
	sanitizedHTML     = `here&#39;s some naughty html:  !!!`
	sanitizeOutgoing  = `<p>gotta test some fucking &#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39; marks</p>`
	sanitizedOutgoing = `<p>gotta test some fucking &#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39; marks</p>`
)

type SanitizeTestSuite struct {
	suite.Suite
}

func (suite *SanitizeTestSuite) TestSanitizeOutgoing() {
	s := text.SanitizeHTML(sanitizeOutgoing)
	suite.Equal(sanitizedOutgoing, s)
}

func (suite *SanitizeTestSuite) TestSanitizeHTML() {
	s := text.SanitizeHTML(sanitizeHTML)
	suite.Equal(sanitizedHTML, s)
}

func (suite *SanitizeTestSuite) TestSanitizeInlineImg() {
	withInlineImg := "<p>Here's an inline image: <img class=\"fixed-size-img svelte-uci8eb\" aria-hidden=\"false\" alt=\"A black-and-white photo of an Oblique Strategy card. The card reads: 'Define an area as 'safe' and use it as an anchor'.\" title=\"A black-and-white photo of an Oblique Strategy card. The card reads: 'Define an area as 'safe' and use it as an anchor'.\" width=\"0\" height=\"0\" src=\"https://example.org/fileserver/01H7J83147QMCE17C0RS9P10Y9/attachment/small/01H7J8365XXRTCP6CAMGEM49ZE.jpg\" style=\"object-position: 50% 50%;\"></p>"
	sanitized := text.SanitizeHTML(withInlineImg)
	suite.Equal(`<p>Here&#39;s an inline image: </p>`, sanitized)
}

func TestSanitizeTestSuite(t *testing.T) {
	suite.Run(t, new(SanitizeTestSuite))
}
