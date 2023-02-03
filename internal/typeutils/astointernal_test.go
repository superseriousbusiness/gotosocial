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

func (suite *ASToInternalTestSuite) jsonToType(in string) vocab.Type {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(in), &m); err != nil {
		suite.FailNow(err.Error())
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return t
}

func (suite *ASToInternalTestSuite) TestParsePerson() {
	testPerson := suite.testPeople["https://unknown-instance.com/users/brand_new_person"]

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), testPerson, "")
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

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), testPerson, "")
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
	t := suite.jsonToType(publicStatusActivityJson)
	rep, ok := t.(ap.Statusable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	status, err := suite.typeconverter.ASStatusToStatus(context.Background(), rep)
	suite.NoError(err)

	suite.Equal("reading: Punishment and Reward in the Corporate University", status.ContentWarning)
	suite.Equal(`<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>`, status.Content)
}

func (suite *ASToInternalTestSuite) TestParsePublicStatusNoURL() {
	t := suite.jsonToType(publicStatusActivityJsonNoURL)
	rep, ok := t.(ap.Statusable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	status, err := suite.typeconverter.ASStatusToStatus(context.Background(), rep)
	suite.NoError(err)

	suite.Equal("reading: Punishment and Reward in the Corporate University", status.ContentWarning)
	suite.Equal(`<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>`, status.Content)

	// on statuses with no URL in them (like ones we get from pleroma sometimes) we should use the AP URI of the status as URL
	suite.Equal("http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167", status.URL)
}

func (suite *ASToInternalTestSuite) TestParseGargron() {
	t := suite.jsonToType(gargronAsActivityJson)
	rep, ok := t.(ap.Accountable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), rep, "")
	suite.NoError(err)
	suite.Equal("https://mastodon.social/inbox", *acct.SharedInboxURI)
}

func (suite *ASToInternalTestSuite) TestParseReplyWithMention() {
	t := suite.jsonToType(statusWithMentionsActivityJson)
	create, ok := t.(vocab.ActivityStreamsCreate)
	if !ok {
		suite.FailNow("type not coercible")
	}

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
	t := suite.jsonToType(owncastService)
	rep, ok := t.(ap.Accountable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), rep, "")
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

func (suite *ASToInternalTestSuite) TestParseBookwyrmStatus() {
	authorAccount := suite.testAccounts["remote_account_1"]

	raw := `{
  "id": "` + authorAccount.URI + `/review/445260",
  "type": "Article",
  "published": "2022-11-09T16:34:28.488375+00:00",
  "attributedTo": "` + authorAccount.URI + `",
  "content": "<p>The original novel is a great read. Not just for the way it codified modern vampire lore. But for the way it's built entirely out of diary entries, letters, news fragments, telegrams and so on. For the way it shows modern science coming to grips with ancient superstition and figuring out how to deal with it. For showing an early example of a woman participating in her own rescue. And for some of the parts that <em>didn't</em> make it into general pop culture. (Count Dracula spends an awful lot of time in a shipping box.)</p>\n<p>In some senses it's the written-word equivalent of the \"found footage\" horror genre. Except the \"sources\" are wildly varying. John and Mina write their journals and letters to each other in shorthand. Business letters are of course written formally. Dr. Seward keeps an audio diary on a phonograph. Van Helsing's speech is rendered with every quirk of his Dutch accent and speech patterns. And then halfway through the book, when all the major characters finally come together...they collate all the documents and Mina transcribes them on a typewriter, and they pass around the first half of the book so they can all read up on what the rest of them have been doing! (Literally getting them all on the same page.)</p>\n<p>That's not to say it's flawless. It's unclear why some victims rise again as vampires while others don't. While the science/superstition contrast works well for the most part, eastern Europeans don't exactly come off very well. Especially when they'd talk about the \"gypsies\" carrying Dracula around Transylvania. I mean, it could have been a lot worse, but it's still jarring.</p>\n<p>Overall, though, it's an engaging read, whether approached as a book or, as Dracula Daily did, one day's letters at a time from May 3 through November 7. </p>\n<p>Dracula Daily: <a href=\"https://draculadaily.example.org/archive\">draculadaily.example.org/archive</a></p>\n<p>This review on my website: <a href=\"https://example.org/reviews/books/dracula/\">example.org/reviews/books/dracula/</a></p>",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "` + authorAccount.FollowersURI + `"
  ],
  "replies": {
    "id": "` + authorAccount.URI + `/review/445260/replies",
    "type": "OrderedCollection",
    "totalItems": 0,
    "first": "` + authorAccount.URI + `/review/445260/replies?page=1",
    "last": "` + authorAccount.URI + `/review/445260/replies?page=1",
    "@context": "https://www.w3.org/ns/activitystreams"
  },
  "tag": [],
  "attachment": [
    {
  	"type": "Document",
  	"url": "` + authorAccount.URI + `/review/445260/images/previews/covers/451118-5f7bd96e-ca03-4865-ab14-baa7addaca8e.jpg",
  	"name": "Dracula (Paperback, 1992, Signet)",
  	"@context": "https://www.w3.org/ns/activitystreams"
    }
  ],
  "sensitive": false,
  "inReplyToBook": "https://bookwyrm.social/book/451118",
  "name": "Review of \"Dracula\" (5 stars): A great read, not just for codifying vampire lore, but the way it's built from letters and diaries.",
  "rating": 5,
  "@context": "https://www.w3.org/ns/activitystreams"
}`

	t := suite.jsonToType(raw)
	asArticle, ok := t.(ap.Statusable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	status, err := suite.typeconverter.ASStatusToStatus(context.Background(), asArticle)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("Review of \"Dracula\" (5 stars): A great read, not just for codifying vampire lore, but the way it's built from letters and diaries.", status.ContentWarning)
	suite.Len(status.Attachments, 1)
}

func (suite *ASToInternalTestSuite) TestParseFlag1() {
	reportedAccount := suite.testAccounts["local_account_1"]
	reportingAccount := suite.testAccounts["remote_account_1"]
	reportedStatus := suite.testStatuses["local_account_1_status_1"]

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "Note: ` + reportedStatus.URL + `\n-----\nban this sick filth ⛔",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": "` + reportedAccount.URI + `",
  "type": "Flag"
}`

	t := suite.jsonToType(raw)
	asFlag, ok := t.(ap.Flaggable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	report, err := suite.typeconverter.ASFlagToReport(context.Background(), asFlag)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(report.AccountID, reportingAccount.ID)
	suite.Equal(report.TargetAccountID, reportedAccount.ID)
	suite.Len(report.StatusIDs, 1)
	suite.Len(report.Statuses, 1)
	suite.Equal(report.Statuses[0].ID, reportedStatus.ID)
	suite.Equal(report.Comment, "Note: "+reportedStatus.URL+"\n-----\nban this sick filth ⛔")
}

func (suite *ASToInternalTestSuite) TestParseFlag2() {
	reportedAccount := suite.testAccounts["local_account_1"]
	reportingAccount := suite.testAccounts["remote_account_1"]
	// report a status that doesn't exist
	reportedStatusURL := "http://localhost:8080/@the_mighty_zork/01GQHR6MCQSTCP85ZG4A0VR316"

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "Note: ` + reportedStatusURL + `\n-----\nban this sick filth ⛔",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": "` + reportedAccount.URI + `",
  "type": "Flag"
}`

	t := suite.jsonToType(raw)
	asFlag, ok := t.(ap.Flaggable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	report, err := suite.typeconverter.ASFlagToReport(context.Background(), asFlag)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(report.AccountID, reportingAccount.ID)
	suite.Equal(report.TargetAccountID, reportedAccount.ID)

	// nonexistent status should just be skipped, it'll still be in the content though
	suite.Len(report.StatusIDs, 0)
	suite.Len(report.Statuses, 0)
	suite.Equal(report.Comment, "Note: "+reportedStatusURL+"\n-----\nban this sick filth ⛔")
}

func (suite *ASToInternalTestSuite) TestParseFlag3() {
	// flag an account that doesn't exist
	reportedAccountURI := "http://localhost:8080/users/mr_e_man"
	reportingAccount := suite.testAccounts["remote_account_1"]

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "ban this sick filth ⛔",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": "` + reportedAccountURI + `",
  "type": "Flag"
}`

	t := suite.jsonToType(raw)
	asFlag, ok := t.(ap.Flaggable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	report, err := suite.typeconverter.ASFlagToReport(context.Background(), asFlag)
	suite.Nil(report)
	suite.EqualError(err, "ASFlagToReport: account with uri http://localhost:8080/users/mr_e_man could not be found in the db")
}

func (suite *ASToInternalTestSuite) TestParseFlag4() {
	// flag an account from another instance
	reportingAccount := suite.testAccounts["remote_account_1"]
	reportedAccountURI := suite.testAccounts["remote_account_2"].URI

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "ban this sick filth ⛔",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": "` + reportedAccountURI + `",
  "type": "Flag"
}`

	t := suite.jsonToType(raw)
	asFlag, ok := t.(ap.Flaggable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	report, err := suite.typeconverter.ASFlagToReport(context.Background(), asFlag)
	suite.Nil(report)
	suite.EqualError(err, "ASFlagToReport: flaggable objects contained no recognizable target account uri")
}

func (suite *ASToInternalTestSuite) TestParseFlag5() {
	reportedAccount := suite.testAccounts["local_account_1"]
	reportingAccount := suite.testAccounts["remote_account_1"]
	reportedStatus := suite.testStatuses["local_account_1_status_1"]

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "misinformation",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": [
    "` + reportedAccount.URI + `",
    "` + reportedStatus.URI + `"
  ],
  "type": "Flag"
  }`

	t := suite.jsonToType(raw)
	asFlag, ok := t.(ap.Flaggable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	report, err := suite.typeconverter.ASFlagToReport(context.Background(), asFlag)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(report.AccountID, reportingAccount.ID)
	suite.Equal(report.TargetAccountID, reportedAccount.ID)
	suite.Len(report.StatusIDs, 1)
	suite.Len(report.Statuses, 1)
	suite.Equal(report.Statuses[0].ID, reportedStatus.ID)
	suite.Equal(report.Comment, "misinformation")
}

func (suite *ASToInternalTestSuite) TestParseFlag6() {
	reportedAccount := suite.testAccounts["local_account_1"]
	reportingAccount := suite.testAccounts["remote_account_1"]
	// flag a status that belongs to another account
	reportedStatus := suite.testStatuses["local_account_2_status_1"]

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "misinformation",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": [
    "` + reportedAccount.URI + `",
    "` + reportedStatus.URI + `"
  ],
  "type": "Flag"
  }`

	t := suite.jsonToType(raw)
	asFlag, ok := t.(ap.Flaggable)
	if !ok {
		suite.FailNow("type not coercible")
	}

	report, err := suite.typeconverter.ASFlagToReport(context.Background(), asFlag)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(report.AccountID, reportingAccount.ID)
	suite.Equal(report.TargetAccountID, reportedAccount.ID)
	suite.Len(report.StatusIDs, 0)
	suite.Len(report.Statuses, 0)
	suite.Equal(report.Comment, "misinformation")
}

func TestASToInternalTestSuite(t *testing.T) {
	suite.Run(t, new(ASToInternalTestSuite))
}
