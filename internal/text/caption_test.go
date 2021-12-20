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

type CaptionTestSuite struct {
	suite.Suite
}

func (suite *CaptionTestSuite) TestSanitizeCaption1() {
	dodgyCaption := "<script>console.log('haha!')</script>this is just a normal caption ;)"
	sanitized := text.SanitizeCaption(dodgyCaption)
	suite.Equal("this is just a normal caption ;)", sanitized)
}

func (suite *CaptionTestSuite) TestSanitizeCaption2() {
	dodgyCaption := "<em>here's a LOUD caption</em>"
	sanitized := text.SanitizeCaption(dodgyCaption)
	suite.Equal("here's a LOUD caption", sanitized)
}

func (suite *CaptionTestSuite) TestSanitizeCaption3() {
	dodgyCaption := ""
	sanitized := text.SanitizeCaption(dodgyCaption)
	suite.Equal("", sanitized)
}

func (suite *CaptionTestSuite) TestSanitizeCaption4() {
	dodgyCaption := `


here is
a multi line
caption
with some newlines



`
	sanitized := text.SanitizeCaption(dodgyCaption)
	suite.Equal("here is\na multi line\ncaption\nwith some newlines", sanitized)
}

func (suite *CaptionTestSuite) TestSanitizeCaption5() {
	// html-escaped: "<script>console.log('aha!')</script> hello world"
	dodgyCaption := `&lt;script&gt;console.log(&apos;aha!&apos;)&lt;/script&gt; hello world`
	sanitized := text.SanitizeCaption(dodgyCaption)
	suite.Equal("hello world", sanitized)
}

func (suite *CaptionTestSuite) TestSanitizeCaption6() {
	// html-encoded: "<script>console.log('aha!')</script> hello world"
	dodgyCaption := `&lt;&#115;&#99;&#114;&#105;&#112;&#116;&gt;&#99;&#111;&#110;&#115;&#111;&#108;&#101;&period;&#108;&#111;&#103;&lpar;&apos;&#97;&#104;&#97;&excl;&apos;&rpar;&lt;&sol;&#115;&#99;&#114;&#105;&#112;&#116;&gt;&#32;&#104;&#101;&#108;&#108;&#111;&#32;&#119;&#111;&#114;&#108;&#100;`
	sanitized := text.SanitizeCaption(dodgyCaption)
	suite.Equal("hello world", sanitized)
}

func TestCaptionTestSuite(t *testing.T) {
	suite.Run(t, new(CaptionTestSuite))
}
