/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
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

type TextTestSuite struct {
	suite.Suite
	log *logrus.Logger
}

func (suite *TextTestSuite) SetupSuite() {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log
}

func (suite *TextTestSuite) TearDownSuite() {

}

func (suite *TextTestSuite) SetupTest() {

}

func (suite *TextTestSuite) TearDownTest() {

}

func (suite *TextTestSuite) TestParseURLsFromText1() {
	urls, err := text.FindLinks(text1)

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "https://example.org/link/to/something#fragment", urls[0].String())
	assert.Equal(suite.T(), "http://test.example.org?q=bahhhhhhhhhhhh", urls[1].String())
	assert.Equal(suite.T(), "https://another.link.example.org/with/a/pretty/long/path/at/the/end/of/it", urls[2].String())
	assert.Equal(suite.T(), "https://example.orghttps://google.com", urls[3].String())
}

func (suite *TextTestSuite) TestParseURLsFromText2() {
	urls, err := text.FindLinks(text2)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), urls, 1)
}

func (suite *TextTestSuite) TestParseURLsFromText3() {
	urls, err := text.FindLinks(text3)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), urls, 1)
	assert.Equal(suite.T(), "mailto:whatever@test.org", urls[0].String())
}

func (suite *TextTestSuite) TestReplaceLinksFromText1() {
	replaced := text.ReplaceLinks(text1)
	fmt.Println(replaced)
}

func (suite *TextTestSuite) TestReplaceLinksFromText2() {
	replaced := text.ReplaceLinks(text2)
	fmt.Println(replaced)
}

func (suite *TextTestSuite) TestReplaceLinksFromText4() {
	replaced := text.ReplaceLinks(text3)
	fmt.Println(replaced)
}

func (suite *TextTestSuite) TestReplaceLinksFromText5() {
	replaced := text.ReplaceLinks(text5)
	fmt.Println(replaced)
}

func TestTextTestSuite(t *testing.T) {
	suite.Run(t, new(TextTestSuite))
}
