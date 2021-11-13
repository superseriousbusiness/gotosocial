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

package typeutils_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type ASToInternalTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *ASToInternalTestSuite) TestParsePerson() {
	testPerson := suite.testPeople["https://unknown-instance.com/users/brand_new_person"]

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), testPerson, false)
	assert.NoError(suite.T(), err)

	suite.Equal("https://unknown-instance.com/users/brand_new_person", acct.URI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/following", acct.FollowingURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/followers", acct.FollowersURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/inbox", acct.InboxURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/outbox", acct.OutboxURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/collections/featured", acct.FeaturedCollectionURI)
	suite.Equal("brand_new_person", acct.Username)
	suite.Equal("Geoff Brando New Personson", acct.DisplayName)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", acct.Note)
	suite.Equal("https://unknown-instance.com/@brand_new_person", acct.URL)
	suite.True(acct.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person#main-key", acct.PublicKeyURI)
	suite.False(acct.Locked)
}

func (suite *ASToInternalTestSuite) TestParseGargron() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(gargronAsActivityJson), &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	rep, ok := t.(ap.Accountable)
	assert.True(suite.T(), ok)

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), rep, false)
	assert.NoError(suite.T(), err)

	fmt.Printf("%+v", acct)
	// TODO: write assertions here, rn we're just eyeballing the output
}

func (suite *ASToInternalTestSuite) TestParseReplyWithMention() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(statusWithMentionsActivityJson), &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	create, ok := t.(vocab.ActivityStreamsCreate)
	suite.True(ok)

	object := create.GetActivityStreamsObject()
	var status *gtsmodel.Status
	for i := object.Begin(); i != nil; i = i.Next() {
		statusable := i.GetActivityStreamsNote()
		s, err := suite.typeconverter.ASStatusToStatus(context.Background(), statusable)
		suite.NoError(err)
		status = s
		break
	}
	suite.NotNil(status)

	postingAccount := suite.testAccounts["remote_account_1"]
	inReplyToAccount := suite.testAccounts["local_account_1"]
	inReplyToStatus := suite.testStatuses["local_account_1_status_1"]

	suite.Equal("http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552", status.URI)
	suite.Equal(postingAccount.ID, status.AccountID)
	suite.Equal(postingAccount.URI, status.AccountURI)
	suite.Equal(inReplyToAccount.ID, status.InReplyToAccountID)
	suite.Equal(inReplyToStatus.ID, status.InReplyToID)
	suite.Equal(inReplyToStatus.URI, status.InReplyToURI)
	suite.True(status.Federated)
	suite.True(status.Boostable)
	suite.True(status.Replyable)
	suite.True(status.Likeable)
	suite.Equal(`<p><span class="h-card"><a href="http://localhost:8080/@the_mighty_zork" class="u-url mention">@<span>the_mighty_zork</span></a></span> nice there it is:</p><p><a href="http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity" rel="nofollow noopener noreferrer" target="_blank"><span class="invisible">https://</span><span class="ellipsis">social.pixie.town/users/f0x/st</span><span class="invisible">atuses/106221628567855262/activity</span></a></p>`, status.Content)
	suite.Len(status.Mentions, 1)
	m1 := status.Mentions[0]
	suite.Equal(inReplyToAccount.URI, m1.TargetAccountURI)
	suite.Equal("@the_mighty_zork@localhost:8080", m1.NameString)
	suite.Equal(gtsmodel.VisibilityUnlocked, status.Visibility)
}

func TestASToInternalTestSuite(t *testing.T) {
	suite.Run(t, new(ASToInternalTestSuite))
}
