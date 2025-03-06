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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/text"
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

func (suite *SanitizeTestSuite) TestSanitizeCaption1() {
	dodgyCaption := "<script>console.log('haha!')</script>this is just a normal caption ;)"
	sanitized := text.RemoveHTML(dodgyCaption)
	suite.Equal("this is just a normal caption ;)", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeCaption2() {
	dodgyCaption := "<em>here's a LOUD caption</em>"
	sanitized := text.RemoveHTML(dodgyCaption)
	suite.Equal("here's a LOUD caption", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeCaption3() {
	dodgyCaption := ""
	sanitized := text.RemoveHTML(dodgyCaption)
	suite.Equal("", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeCaption4() {
	dodgyCaption := `


here is
a multi line
caption
with some newlines



`
	sanitized := text.RemoveHTML(dodgyCaption)
	suite.Equal("here is\na multi line\ncaption\nwith some newlines", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeCaption5() {
	// html-escaped: "<script>console.log('aha!')</script> hello world"
	dodgyCaption := `&lt;script&gt;console.log(&apos;aha!&apos;)&lt;/script&gt; hello world`
	sanitized := text.RemoveHTML(dodgyCaption)
	suite.Equal("hello world", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeCaption6() {
	// html-encoded: "<script>console.log('aha!')</script> hello world"
	dodgyCaption := `&lt;&#115;&#99;&#114;&#105;&#112;&#116;&gt;&#99;&#111;&#110;&#115;&#111;&#108;&#101;&period;&#108;&#111;&#103;&lpar;&apos;&#97;&#104;&#97;&excl;&apos;&rpar;&lt;&sol;&#115;&#99;&#114;&#105;&#112;&#116;&gt;&#32;&#104;&#101;&#108;&#108;&#111;&#32;&#119;&#111;&#114;&#108;&#100;`
	sanitized := text.RemoveHTML(dodgyCaption)
	suite.Equal("hello world", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeCustomCSS() {
	customCSS := `.toot .username {
	color: var(--link_fg);
	line-height: 2rem;
	margin-top: -0.5rem;
	align-self: start;
	
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}`
	sanitized := text.RemoveHTML(customCSS)
	suite.Equal(customCSS, sanitized) // should be the same as it was before
}

func (suite *SanitizeTestSuite) TestSanitizeNaughtyCustomCSS1() {
	// try to break out of <style> into <head> and change the document title
	customCSS := "</style><title>pee pee poo poo</title><style>"
	sanitized := text.RemoveHTML(customCSS)
	suite.Empty(sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeNaughtyCustomCSS2() {
	// try to break out of <style> into <head> and change the document title
	customCSS := "pee pee poo poo</style><title></title><style>"
	sanitized := text.RemoveHTML(customCSS)
	suite.Equal("pee pee poo poo", sanitized)
}

func (suite *SanitizeTestSuite) TestSanitizeInlineImg() {
	withInlineImg := "<p>Here's an inline image: <img class=\"fixed-size-img svelte-uci8eb\" aria-hidden=\"false\" alt=\"A black-and-white photo of an Oblique Strategy card. The card reads: 'Define an area as 'safe' and use it as an anchor'.\" title=\"A black-and-white photo of an Oblique Strategy card. The card reads: 'Define an area as 'safe' and use it as an anchor'.\" width=\"0\" height=\"0\" src=\"https://example.org/fileserver/01H7J83147QMCE17C0RS9P10Y9/attachment/small/01H7J8365XXRTCP6CAMGEM49ZE.jpg\" style=\"object-position: 50% 50%;\"></p>"
	sanitized := text.SanitizeHTML(withInlineImg)
	suite.Equal(`<p>Here&#39;s an inline image: </p>`, sanitized)
}

func TestSanitizeTestSuite(t *testing.T) {
	suite.Run(t, new(SanitizeTestSuite))
}
