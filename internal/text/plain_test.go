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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	simple              = "this is a plain and simple status"
	simpleExpected      = "<p>this is a plain and simple status</p>"
	withTag             = "here's a simple status that uses hashtag #welcome!"
	withTagExpected     = "<p>here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!</p>"
	withHTML            = "<div>blah this should just be html escaped blah</div>"
	withHTMLExpected    = "<p>&lt;div>blah this should just be html escaped blah&lt;/div></p>"
	moreComplex         = "Another test @foss_satan@fossbros-anonymous.io\n\n#Hashtag\n\nText\n\n:rainbow:"
	moreComplexExpected = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br>Text<br><br>:rainbow:</p>"
)

type PlainTestSuite struct {
	TextStandardTestSuite
}

func (suite *PlainTestSuite) TestParseSimple() {
	formatted := suite.FromPlain(simple)
	suite.Equal(simpleExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithTag() {
	formatted := suite.FromPlain(withTag)
	suite.Equal(withTagExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithHTML() {
	formatted := suite.FromPlain(withHTML)
	suite.Equal(withHTMLExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseMoreComplex() {
	formatted := suite.FromPlain(moreComplex)
	suite.Equal(moreComplexExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestLinkNoMention() {
	statusText := `here's a link to a post by zork

https://example.com/@the_mighty_zork/statuses/01FGVP55XMF2K6316MQRX6PFG1

that link shouldn't come out formatted as a mention!`

	menchies := suite.FromPlain(statusText).Mentions
	suite.Empty(menchies)
}

func (suite *PlainTestSuite) TestDeriveMentionsEmpty() {
	statusText := ``
	menchies := suite.FromPlain(statusText).Mentions
	assert.Len(suite.T(), menchies, 0)
}

func (suite *PlainTestSuite) TestDeriveHashtagsOK() {
	statusText := `weeeeeeee #testing123 #also testing

# testing this one shouldn't work

			#thisshouldwork #dupe #dupe!! #dupe

	here's a link with a fragment: https://example.org/whatever#ahhh
	here's another link with a fragment: https://example.org/whatever/#ahhh

(#ThisShouldAlsoWork) #this_should_be_split

#111111 thisalsoshouldn'twork#### ##

#alimentación, #saúde, #lävistää, #ö, #네
#ThisOneIsThirtyOneCharactersLon...  ...ng
#ThisOneIsThirteyCharactersLong
`

	tags := suite.FromPlain(statusText).Tags
	assert.Len(suite.T(), tags, 13)
	assert.Equal(suite.T(), "testing123", tags[0].Name)
	assert.Equal(suite.T(), "also", tags[1].Name)
	assert.Equal(suite.T(), "thisshouldwork", tags[2].Name)
	assert.Equal(suite.T(), "dupe", tags[3].Name)
	assert.Equal(suite.T(), "ThisShouldAlsoWork", tags[4].Name)
	assert.Equal(suite.T(), "this", tags[5].Name)
	assert.Equal(suite.T(), "111111", tags[6].Name)
	assert.Equal(suite.T(), "alimentación", tags[7].Name)
	assert.Equal(suite.T(), "saúde", tags[8].Name)
	assert.Equal(suite.T(), "lävistää", tags[9].Name)
	assert.Equal(suite.T(), "ö", tags[10].Name)
	assert.Equal(suite.T(), "네", tags[11].Name)
	assert.Equal(suite.T(), "ThisOneIsThirteyCharactersLong", tags[12].Name)

	statusText = `#올빼미 hej`
	tags = suite.FromPlain(statusText).Tags
	assert.Equal(suite.T(), "올빼미", tags[0].Name)
}

func (suite *PlainTestSuite) TestDeriveMultiple() {
	statusText := `Another test @foss_satan@fossbros-anonymous.io

	#Hashtag

	Text`

	f := suite.FromPlain(statusText)

	assert.Len(suite.T(), f.Mentions, 1)
	assert.Equal(suite.T(), "@foss_satan@fossbros-anonymous.io", f.Mentions[0].NameString)

	assert.Len(suite.T(), f.Tags, 1)
	assert.Equal(suite.T(), "Hashtag", f.Tags[0].Name)

	assert.Len(suite.T(), f.Emojis, 0)
}

func (suite *PlainTestSuite) TestZalgoHashtag() {
	statusText := `yo who else loves #praying to #z̸͉̅a̸͚͋l̵͈̊g̸̫͌ỏ̷̪?`
	f := suite.FromPlain(statusText)
	assert.Len(suite.T(), f.Tags, 1)
	assert.Equal(suite.T(), "praying", f.Tags[0].Name)
}

func TestPlainTestSuite(t *testing.T) {
	suite.Run(t, new(PlainTestSuite))
}
