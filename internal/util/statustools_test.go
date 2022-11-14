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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type StatusTestSuite struct {
	suite.Suite
}

func (suite *StatusTestSuite) TestLinkNoMention() {
	statusText := `here's a link to a post by zork:

https://localhost:8080/@the_mighty_zork/statuses/01FGVP55XMF2K6316MQRX6PFG1

that link shouldn't come out formatted as a mention!`

	menchies := util.DeriveMentionNamesFromText(statusText)
	suite.Empty(menchies)
}

func (suite *StatusTestSuite) TestDeriveMentionsOK() {
	statusText := `@dumpsterqueer@example.org testing testing

	is this thing on?

	@someone_else@testing.best-horse.com can you confirm? @hello@test.lgbt

	@thisisalocaluser!

	here is a duplicate mention: @hello@test.lgbt @hello@test.lgbt

	@account1@whatever.com @account2@whatever.com

	`

	menchies := util.DeriveMentionNamesFromText(statusText)
	assert.Len(suite.T(), menchies, 6)
	assert.Equal(suite.T(), "@dumpsterqueer@example.org", menchies[0])
	assert.Equal(suite.T(), "@someone_else@testing.best-horse.com", menchies[1])
	assert.Equal(suite.T(), "@hello@test.lgbt", menchies[2])
	assert.Equal(suite.T(), "@thisisalocaluser", menchies[3])
	assert.Equal(suite.T(), "@account1@whatever.com", menchies[4])
	assert.Equal(suite.T(), "@account2@whatever.com", menchies[5])
}

func (suite *StatusTestSuite) TestDeriveMentionsEmpty() {
	statusText := ``
	menchies := util.DeriveMentionNamesFromText(statusText)
	assert.Len(suite.T(), menchies, 0)
}

func (suite *StatusTestSuite) TestDeriveHashtagsOK() {
	statusText := `weeeeeeee #testing123 #also testing

# testing this one shouldn't work

			#thisshouldwork #dupe #dupe!! #dupe

	here's a link with a fragment: https://example.org/whatever#ahhh
	here's another link with a fragment: https://example.org/whatever/#ahhh

(#ThisShouldAlsoWork) #not_this_though

#111111 thisalsoshouldn'twork#### ##

#alimentación, #saúde, #lävistää, #ö, #네
#ThisOneIsThirtyOneCharactersLon...  ...ng
#ThisOneIsThirteyCharactersLong
`

	tags := util.DeriveHashtagsFromText(statusText)
	assert.Len(suite.T(), tags, 12)
	assert.Equal(suite.T(), "testing123", tags[0])
	assert.Equal(suite.T(), "also", tags[1])
	assert.Equal(suite.T(), "thisshouldwork", tags[2])
	assert.Equal(suite.T(), "dupe", tags[3])
	assert.Equal(suite.T(), "ThisShouldAlsoWork", tags[4])
	assert.Equal(suite.T(), "111111", tags[5])
	assert.Equal(suite.T(), "alimentación", tags[6])
	assert.Equal(suite.T(), "saúde", tags[7])
	assert.Equal(suite.T(), "lävistää", tags[8])
	assert.Equal(suite.T(), "ö", tags[9])
	assert.Equal(suite.T(), "네", tags[10])
	assert.Equal(suite.T(), "ThisOneIsThirteyCharactersLong", tags[11])

	statusText = `#올빼미 hej`
	tags = util.DeriveHashtagsFromText(statusText)
	assert.Equal(suite.T(), "올빼미", tags[0])
}

func (suite *StatusTestSuite) TestHashtagSpansOK() {
	statusText := `#0 #3   #8aa`

	spans := util.FindHashtagSpansInText(statusText)
	assert.Equal(suite.T(), 0, spans[0].First)
	assert.Equal(suite.T(), 2, spans[0].Second)
	assert.Equal(suite.T(), 3, spans[1].First)
	assert.Equal(suite.T(), 5, spans[1].Second)
	assert.Equal(suite.T(), 8, spans[2].First)
	assert.Equal(suite.T(), 12, spans[2].Second)
}

func (suite *StatusTestSuite) TestDeriveEmojiOK() {
	statusText := `:test: :another:

Here's some normal text with an :emoji: at the end

:spaces shouldnt work:

:emoji1::emoji2:

:anotheremoji:emoji2:
:anotheremoji::anotheremoji::anotheremoji::anotheremoji:
:underscores_ok_too:
`

	tags := util.DeriveEmojisFromText(statusText)
	assert.Len(suite.T(), tags, 7)
	assert.Equal(suite.T(), "test", tags[0])
	assert.Equal(suite.T(), "another", tags[1])
	assert.Equal(suite.T(), "emoji", tags[2])
	assert.Equal(suite.T(), "emoji1", tags[3])
	assert.Equal(suite.T(), "emoji2", tags[4])
	assert.Equal(suite.T(), "anotheremoji", tags[5])
	assert.Equal(suite.T(), "underscores_ok_too", tags[6])
}

func (suite *StatusTestSuite) TestDeriveMultiple() {
	statusText := `Another test @foss_satan@fossbros-anonymous.io

	#HashTag

	Text`

	ms := util.DeriveMentionNamesFromText(statusText)
	hs := util.DeriveHashtagsFromText(statusText)
	es := util.DeriveEmojisFromText(statusText)

	assert.Len(suite.T(), ms, 1)
	assert.Equal(suite.T(), "@foss_satan@fossbros-anonymous.io", ms[0])

	assert.Len(suite.T(), hs, 1)
	assert.Contains(suite.T(), hs, "HashTag")

	assert.Len(suite.T(), es, 0)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
