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

package text_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

const text1 = `
This is a text with some links in it. Here's link number one: https://example.org/link/to/something#fragment

Here's link number two: http://test.example.org?q=bahhhhhhhhhhhh

https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it

really.cool.website <-- this one shouldn't be parsed as a link because it doesn't contain the scheme

https://example.orghttps://google.com <-- this shouldn't work either, but it does?! OK
`

const text2 = `
this is one link: https://example.org

this is the same link again: https://example.org

these should be deduplicated
`

const text3 = `
here's a mailto link: mailto:whatever@test.org
`

const text4 = `
two similar links:

https://example.org

https://example.org/test
`

const text5 = `
what happens when we already have a link within an href?

<a href="https://example.org">https://example.org</a>
`

type LinkTestSuite struct {
	TextStandardTestSuite
}

func (suite *LinkTestSuite) TestParseSimple() {
	f := suite.formatter.FromPlain(context.Background(), simple, nil, nil)
	suite.Equal(simpleExpected, f)
}

func (suite *LinkTestSuite) TestParseURLsFromText1() {
	urls := text.FindLinks(text1)

	suite.Equal("https://example.org/link/to/something#fragment", urls[0].String())
	suite.Equal("http://test.example.org?q=bahhhhhhhhhhhh", urls[1].String())
	suite.Equal("https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it", urls[2].String())
	suite.Equal("https://example.orghttps://google.com", urls[3].String())
}

func (suite *LinkTestSuite) TestParseURLsFromText2() {
	urls := text.FindLinks(text2)

	// assert length 1 because the found links will be deduplicated
	assert.Len(suite.T(), urls, 1)
}

func (suite *LinkTestSuite) TestParseURLsFromText3() {
	urls := text.FindLinks(text3)

	// assert length 0 because `mailto:` isn't accepted
	assert.Len(suite.T(), urls, 0)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText1() {
	replaced := suite.formatter.ReplaceLinks(context.Background(), text1)
	suite.Equal(`
This is a text with some links in it. Here's link number one: <a href="https://example.org/link/to/something#fragment" rel="noopener">example.org/link/to/something#fragment</a>

Here's link number two: <a href="http://test.example.org?q=bahhhhhhhhhhhh" rel="noopener">test.example.org?q=bahhhhhhhhhhhh</a>

<a href="https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it" rel="noopener">another.link.example.org/with/a/pretty/long/path/at/the/end/of/it</a>

really.cool.website <-- this one shouldn't be parsed as a link because it doesn't contain the scheme

<a href="https://example.orghttps://google.com" rel="noopener">example.orghttps://google.com</a> <-- this shouldn't work either, but it does?! OK
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText2() {
	replaced := suite.formatter.ReplaceLinks(context.Background(), text2)
	suite.Equal(`
this is one link: <a href="https://example.org" rel="noopener">example.org</a>

this is the same link again: <a href="https://example.org" rel="noopener">example.org</a>

these should be deduplicated
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText3() {
	// we know mailto links won't be replaced with hrefs -- we only accept https and http
	replaced := suite.formatter.ReplaceLinks(context.Background(), text3)
	suite.Equal(`
here's a mailto link: mailto:whatever@test.org
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText4() {
	replaced := suite.formatter.ReplaceLinks(context.Background(), text4)
	suite.Equal(`
two similar links:

<a href="https://example.org" rel="noopener">example.org</a>

<a href="https://example.org/test" rel="noopener">example.org/test</a>
`, replaced)
}

func (suite *LinkTestSuite) TestReplaceLinksFromText5() {
	// we know this one doesn't work properly, which is why html should always be sanitized before being passed into the ReplaceLinks function
	replaced := suite.formatter.ReplaceLinks(context.Background(), text5)
	suite.Equal(`
what happens when we already have a link within an href?

<a href="<a href="https://example.org" rel="noopener">example.org</a>"><a href="https://example.org" rel="noopener">example.org</a></a>
`, replaced)
}

func TestLinkTestSuite(t *testing.T) {
	suite.Run(t, new(LinkTestSuite))
}
