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

package typeutils_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

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

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), testPerson, "", false)
	suite.NoError(err)

	suite.Equal("https://unknown-instance.com/users/brand_new_person", acct.URI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/following", acct.FollowingURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/followers", acct.FollowersURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/inbox", acct.InboxURI)
	suite.Nil(acct.SharedInboxURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/outbox", acct.OutboxURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/collections/featured", acct.FeaturedCollectionURI)
	suite.Equal("brand_new_person", acct.Username)
	suite.Equal("Geoff Brando New Personson", acct.DisplayName)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", acct.Note)
	suite.Equal("https://unknown-instance.com/@brand_new_person", acct.URL)
	suite.True(*acct.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person#main-key", acct.PublicKeyURI)
	suite.False(*acct.Locked)
}

func (suite *ASToInternalTestSuite) TestParsePersonWithSharedInbox() {
	testPerson := suite.testPeople["https://turnip.farm/users/turniplover6969"]

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), testPerson, "", false)
	suite.NoError(err)

	suite.Equal("https://turnip.farm/users/turniplover6969", acct.URI)
	suite.Equal("https://turnip.farm/users/turniplover6969/following", acct.FollowingURI)
	suite.Equal("https://turnip.farm/users/turniplover6969/followers", acct.FollowersURI)
	suite.Equal("https://turnip.farm/users/turniplover6969/inbox", acct.InboxURI)
	suite.Equal("https://turnip.farm/sharedInbox", *acct.SharedInboxURI)
	suite.Equal("https://turnip.farm/users/turniplover6969/outbox", acct.OutboxURI)
	suite.Equal("https://turnip.farm/users/turniplover6969/collections/featured", acct.FeaturedCollectionURI)
	suite.Equal("turniplover6969", acct.Username)
	suite.Equal("Turnip Lover 6969", acct.DisplayName)
	suite.Equal("I just think they're neat", acct.Note)
	suite.Equal("https://turnip.farm/@turniplover6969", acct.URL)
	suite.True(*acct.Discoverable)
	suite.Equal("https://turnip.farm/users/turniplover6969#main-key", acct.PublicKeyURI)
	suite.False(*acct.Locked)
}

func (suite *ASToInternalTestSuite) TestParsePublicStatus() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(publicStatusActivityJson), &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	rep, ok := t.(ap.Statusable)
	suite.True(ok)

	status, err := suite.typeconverter.ASStatusToStatus(context.Background(), rep)
	suite.NoError(err)

	suite.Equal("reading: Punishment and Reward in the Corporate University", status.ContentWarning)
	suite.Equal(`<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>`, status.Content)
}

func (suite *ASToInternalTestSuite) TestParsePublicStatusNoURL() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(publicStatusActivityJsonNoURL), &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	rep, ok := t.(ap.Statusable)
	suite.True(ok)

	status, err := suite.typeconverter.ASStatusToStatus(context.Background(), rep)
	suite.NoError(err)

	suite.Equal("reading: Punishment and Reward in the Corporate University", status.ContentWarning)
	suite.Equal(`<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>`, status.Content)

	// on statuses with no URL in them (like ones we get from pleroma sometimes) we should use the AP URI of the status as URL
	suite.Equal("http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167", status.URL)
}

func (suite *ASToInternalTestSuite) TestParseGargron() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(gargronAsActivityJson), &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	rep, ok := t.(ap.Accountable)
	suite.True(ok)

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), rep, "", false)
	suite.NoError(err)

	suite.Equal("https://mastodon.social/inbox", *acct.SharedInboxURI)

	fmt.Printf("%+v", acct)
	// TODO: write assertions here, rn we're just eyeballing the output
}

func (suite *ASToInternalTestSuite) TestParseReplyWithMention() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(statusWithMentionsActivityJson), &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

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
	suite.True(*status.Federated)
	suite.True(*status.Boostable)
	suite.True(*status.Replyable)
	suite.True(*status.Likeable)
	suite.Equal(`<p><span class="h-card"><a href="http://localhost:8080/@the_mighty_zork" class="u-url mention">@<span>the_mighty_zork</span></a></span> nice there it is:</p><p><a href="http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity" rel="nofollow noopener noreferrer" target="_blank"><span class="invisible">https://</span><span class="ellipsis">social.pixie.town/users/f0x/st</span><span class="invisible">atuses/106221628567855262/activity</span></a></p>`, status.Content)
	suite.Len(status.Mentions, 1)
	m1 := status.Mentions[0]
	suite.Equal(inReplyToAccount.URI, m1.TargetAccountURI)
	suite.Equal("@the_mighty_zork@localhost:8080", m1.NameString)
	suite.Equal(gtsmodel.VisibilityUnlocked, status.Visibility)
}

func (suite *ASToInternalTestSuite) TestParseOwncastService() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(owncastService), &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	rep, ok := t.(ap.Accountable)
	suite.True(ok)

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), rep, "", false)
	suite.NoError(err)

	suite.Equal("rgh", acct.Username)
	suite.Equal("owncast.example.org", acct.Domain)
	suite.Equal("https://owncast.example.org/logo/external", acct.AvatarRemoteURL)
	suite.Equal("https://owncast.example.org/logo/external", acct.HeaderRemoteURL)
	suite.Equal("Rob's Owncast Server", acct.DisplayName)
	suite.Equal("linux audio stuff ", acct.Note)
	suite.True(*acct.Bot)
	suite.False(*acct.Locked)
	suite.True(*acct.Discoverable)
	suite.Equal("https://owncast.example.org/federation/user/rgh", acct.URI)
	suite.Equal("https://owncast.example.org/federation/user/rgh", acct.URL)
	suite.Equal("https://owncast.example.org/federation/user/rgh/inbox", acct.InboxURI)
	suite.Equal("https://owncast.example.org/federation/user/rgh/outbox", acct.OutboxURI)
	suite.Equal("https://owncast.example.org/federation/user/rgh/followers", acct.FollowersURI)
	suite.Equal("Service", acct.ActorType)
	suite.Equal("https://owncast.example.org/federation/user/rgh#main-key", acct.PublicKeyURI)

	acct.ID = "01G42D57DTCJQE8XT9KD4K88RK"

	apiAcct, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), acct)
	suite.NoError(err)
	suite.NotNil(apiAcct)

	b, err := json.Marshal(apiAcct)
	suite.NoError(err)

	fmt.Printf("\n\n\n%s\n\n\n", string(b))
}

func TestASToInternalTestSuite(t *testing.T) {
	suite.Run(t, new(ASToInternalTestSuite))
}
