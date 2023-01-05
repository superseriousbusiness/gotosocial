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

package text

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	test_removeHTML                 = `<p>Another test <span class="h-card"><a href="http://fossbros-anonymous.io/@foss_satan" class="u-url mention" rel="nofollow noreferrer noopener" target="_blank">@<span>foss_satan</span></a></span><br/><br/><a href="http://localhost:8080/tags/Hashtag" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>Hashtag</span></a><br/><br/>Text</p>`
	test_removedHTML                = `Another test @foss_satan#HashtagText`
	test_withEscapedLiteral         = `it\u0026amp;#39;s its it is`
	test_withEscapedLiteralExpected = `it\u0026amp;#39;s its it is`
	test_withEscaped                = "it\u0026amp;#39;s its it is"
	test_withEscapedExpected        = "it&amp;#39;s its it is"
)

type RemoveHTMLTestSuite struct {
	suite.Suite
}

func (suite *RemoveHTMLTestSuite) TestSanitizeWithEscapedLiteral() {
	s := removeHTML(test_withEscapedLiteral)
	suite.Equal(test_withEscapedLiteralExpected, s)
}

func (suite *RemoveHTMLTestSuite) TestSanitizeWithEscaped() {
	s := removeHTML(test_withEscaped)
	suite.Equal(test_withEscapedExpected, s)
}

func (suite *RemoveHTMLTestSuite) TestRemoveHTML() {
	s := removeHTML(test_removeHTML)
	suite.Equal(test_removedHTML, s)
}

func TestRemoveHTMLTestSuite(t *testing.T) {
	suite.Run(t, &RemoveHTMLTestSuite{})
}
