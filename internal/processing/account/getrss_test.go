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

package account_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GetRSSTestSuite struct {
	AccountStandardTestSuite
}

func (suite *GetRSSTestSuite) TestGetAccountRSSAdmin() {
	getFeed, lastModified, err := suite.accountProcessor.GetRSSFeedForUsername(context.Background(), "admin")
	suite.NoError(err)
	suite.EqualValues(1634733405, lastModified.Unix())

	feed, err := getFeed()
	suite.NoError(err)
	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>admin</title>\n    <link>http://localhost:8080/@admin</link>\n    <description>Public posts from @admin@localhost:8080</description>\n    <pubDate>Wed, 20 Oct 2021 12:36:45 +0000</pubDate>\n    <lastBuildDate>Wed, 20 Oct 2021 12:36:45 +0000</lastBuildDate>\n    <item>\n      <title>open to see some puppies</title>\n      <link>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</link>\n      <description>🐕🐕🐕🐕🐕</description>\n      <content:encoded><![CDATA[🐕🐕🐕🐕🐕]]></content:encoded>\n      <guid>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</guid>\n      <pubDate>Wed, 20 Oct 2021 12:36:45 +0000</pubDate>\n    </item>\n    <item>\n      <title>Posted 1 attachment; hello world! #welcome ! first post on the i</title>\n      <link>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</link>\n      <description>Posted 1 attachment; hello world! #welcome ! first post on the instance :rainbow: !</description>\n      <content:encoded><![CDATA[hello world! #welcome ! first post on the instance :rainbow: !]]></content:encoded>\n      <guid>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</guid>\n      <pubDate>Wed, 20 Oct 2021 11:36:45 +0000</pubDate>\n    </item>\n  </channel>\n</rss>", feed)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZork() {
	getFeed, lastModified, err := suite.accountProcessor.GetRSSFeedForUsername(context.Background(), "the_mighty_zork")
	suite.NoError(err)
	suite.EqualValues(1634726437, lastModified.Unix())

	feed, err := getFeed()
	suite.NoError(err)
	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>original zork (he/they)</title>\n    <link>http://localhost:8080/@the_mighty_zork</link>\n    <description>Public posts from @the_mighty_zork@localhost:8080</description>\n    <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>\n    <lastBuildDate>Wed, 20 Oct 2021 10:40:37 +0000</lastBuildDate>\n    <image>\n      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg</url>\n      <title>Avatar for @the_mighty_zork@localhost:8080</title>\n      <link>http://localhost:8080/@the_mighty_zork</link>\n    </image>\n    <item>\n      <title>introduction post</title>\n      <link>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</link>\n      <description>hello everyone!</description>\n      <content:encoded><![CDATA[hello everyone!]]></content:encoded>\n      <guid>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</guid>\n      <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>\n    </item>\n  </channel>\n</rss>", feed)
}

func TestGetRSSTestSuite(t *testing.T) {
	suite.Run(t, new(GetRSSTestSuite))
}
