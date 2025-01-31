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
)

const (
	simple                     = "this is a plain and simple status"
	simpleExpected             = "<p>this is a plain and simple status</p>"
	simpleExpectedNoParagraph  = "this is a plain and simple status"
	withTag                    = "here's a simple status that uses hashtag #welcome!"
	withTagExpected            = "<p>here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!</p>"
	withTagExpectedNoParagraph = "here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!"
	withHTML                   = "<div>blah this should just be html escaped blah</div>"
	withHTMLExpected           = "<p>&lt;div>blah this should just be html escaped blah&lt;/div></p>"
	moreComplex                = "Another test @foss_satan@fossbros-anonymous.io\n\n#Hashtag\n\nText\n\n:rainbow:"
	moreComplexExpected        = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br>Text<br><br>:rainbow:</p>"
	withUTF8Link               = "here's a link with utf-8 characters in it: https://example.org/söme_url"
	withUTF8LinkExpected       = "<p>here's a link with utf-8 characters in it: <a href=\"https://example.org/s%C3%B6me_url\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://example.org/söme_url</a></p>"
	withFunkyTags              = "#hashtag1 pee #hashtag2\u200Bpee #hashtag3|poo #hashtag4\uFEFFpoo"
	withFunkyTagsExpected      = "<p><a href=\"http://localhost:8080/tags/hashtag1\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag1</span></a> pee <a href=\"http://localhost:8080/tags/hashtag2\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag2</span></a>\u200bpee <a href=\"http://localhost:8080/tags/hashtag3\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag3</span></a>|poo <a href=\"http://localhost:8080/tags/hashtag4\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag4</span></a>\ufeffpoo</p>"
)

type PlainTestSuite struct {
	TextStandardTestSuite
}

func (suite *PlainTestSuite) TestParseSimple() {
	formatted := suite.FromPlain(simple)
	suite.Equal(simpleExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseSimpleNoParagraph() {
	formatted := suite.FromPlainNoParagraph(simple)
	suite.Equal(simpleExpectedNoParagraph, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithTag() {
	formatted := suite.FromPlain(withTag)
	suite.Equal(withTagExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithTagNoParagraph() {
	formatted := suite.FromPlainNoParagraph(withTag)
	suite.Equal(withTagExpectedNoParagraph, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithHTML() {
	formatted := suite.FromPlain(withHTML)
	suite.Equal(withHTMLExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseMoreComplex() {
	formatted := suite.FromPlain(moreComplex)
	suite.Equal(moreComplexExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestWithUTF8Link() {
	formatted := suite.FromPlain(withUTF8Link)
	suite.Equal(withUTF8LinkExpected, formatted.HTML)
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
	suite.Len(menchies, 0)
}

func (suite *PlainTestSuite) TestDeriveHashtagsOK() {
	statusText := `weeeeeeee #testing123 #also testing

# testing this one shouldn't work

			#thisshouldwork #dupe #dupe!! #dupe

	here's a link with a fragment: https://example.org/whatever#ahhh
	here's another link with a fragment: https://example.org/whatever/#ahhh

(#ThisShouldAlsoWork) #this_should_not_be_split

#__ <- just underscores, shouldn't work

#111111 thisalsoshouldn'twork#### ##

#alimentación, #saúde, #lävistää, #ö, #네
#ThisOneIsOneHundredAndOneCharactersLongWhichIsReallyJustWayWayTooLongDefinitelyLongerThanYouWouldNeed...
#ThisOneIsThirteyCharactersLong
`

	tags := suite.FromPlain(statusText).Tags
	if suite.Len(tags, 12) {
		suite.Equal("testing123", tags[0].Name)
		suite.Equal("also", tags[1].Name)
		suite.Equal("thisshouldwork", tags[2].Name)
		suite.Equal("dupe", tags[3].Name)
		suite.Equal("ThisShouldAlsoWork", tags[4].Name)
		suite.Equal("this_should_not_be_split", tags[5].Name)
		suite.Equal("alimentación", tags[6].Name)
		suite.Equal("saúde", tags[7].Name)
		suite.Equal("lävistää", tags[8].Name)
		suite.Equal("ö", tags[9].Name)
		suite.Equal("네", tags[10].Name)
		suite.Equal("ThisOneIsThirteyCharactersLong", tags[11].Name)
	}

	statusText = `#올빼미 hej`
	tags = suite.FromPlain(statusText).Tags
	suite.Equal("올빼미", tags[0].Name)
}

func (suite *PlainTestSuite) TestFunkyTags() {
	formatted := suite.FromPlain(withFunkyTags)
	suite.Equal(withFunkyTagsExpected, formatted.HTML)

	tags := formatted.Tags
	suite.Equal("hashtag1", tags[0].Name)
	suite.Equal("hashtag2", tags[1].Name)
	suite.Equal("hashtag3", tags[2].Name)
	suite.Equal("hashtag4", tags[3].Name)
}

func (suite *PlainTestSuite) TestDeriveMultiple() {
	statusText := `Another test @foss_satan@fossbros-anonymous.io

	#Hashtag

	Text`

	f := suite.FromPlain(statusText)

	suite.Len(f.Mentions, 1)
	suite.Equal("@foss_satan@fossbros-anonymous.io", f.Mentions[0].NameString)

	suite.Len(f.Tags, 1)
	suite.Equal("hashtag", f.Tags[0].Name)

	suite.Len(f.Emojis, 0)
}

func (suite *PlainTestSuite) TestZalgoHashtag() {
	statusText := `yo who else loves #praying to #z̸͉̅a̸͚͋l̵͈̊g̸̫͌ỏ̷̪?`
	f := suite.FromPlain(statusText)
	if suite.Len(f.Tags, 2) {
		suite.Equal("praying", f.Tags[0].Name)
		// NFC doesn't do much for Zalgo text, but it's difficult to strip marks without affecting non-Latin text.
		suite.Equal("z̸͉̅a̸͚͋l̵͈̊g̸̫͌ỏ̷̪", f.Tags[1].Name)
	}
}

func (suite *PlainTestSuite) TestNumbersAreNotHashtags() {
	statusText := `yo who else thinks #19_98 is #1?`
	f := suite.FromPlain(statusText)
	suite.Len(f.Tags, 0)
}

func TestPlainTestSuite(t *testing.T) {
	suite.Run(t, new(PlainTestSuite))
}
