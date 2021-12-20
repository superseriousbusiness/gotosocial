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

package text_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

const (
	removeHTML  = `<p>Another test <span class="h-card"><a href="http://fossbros-anonymous.io/@foss_satan" class="u-url mention" rel="nofollow noreferrer noopener" target="_blank">@<span>foss_satan</span></a></span><br/><br/><a href="http://localhost:8080/tags/Hashtag" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>Hashtag</span></a><br/><br/>Text</p>`
	removedHTML = `Another test @foss_satan#HashtagText`

	sanitizeHTML  = `here's some naughty html: <script>alert(ahhhh)</script> !!!`
	sanitizedHTML = `here&#39;s some naughty html:  !!!`

	withEscapedLiteral         = `it\u0026amp;#39;s its it is`
	withEscapedLiteralExpected = `it\u0026amp;#39;s its it is`
	withEscaped                = "it\u0026amp;#39;s its it is"
	withEscapedExpected        = "it&amp;#39;s its it is"

	sanitizeOutgoing  = `<p>gotta test some fucking &#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39; marks</p>`
	sanitizedOutgoing = `<p>gotta test some fucking &#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39;&#39; marks</p>`
)

type SanitizeTestSuite struct {
	suite.Suite
}

func (suite *SanitizeTestSuite) TestRemoveHTML() {
	s := text.RemoveHTML(removeHTML)
	suite.Equal(removedHTML, s)
}

func (suite *SanitizeTestSuite) TestSanitizeOutgoing() {
	s := text.SanitizeHTML(sanitizeOutgoing)
	suite.Equal(sanitizedOutgoing, s)
}

func (suite *SanitizeTestSuite) TestSanitizeHTML() {
	s := text.SanitizeHTML(sanitizeHTML)
	suite.Equal(sanitizedHTML, s)
}

func (suite *SanitizeTestSuite) TestSanitizeWithEscapedLiteral() {
	s := text.RemoveHTML(withEscapedLiteral)
	suite.Equal(withEscapedLiteralExpected, s)
}

func (suite *SanitizeTestSuite) TestSanitizeWithEscaped() {
	s := text.RemoveHTML(withEscaped)
	suite.Equal(withEscapedExpected, s)
}

func TestSanitizeTestSuite(t *testing.T) {
	suite.Run(t, new(SanitizeTestSuite))
}
