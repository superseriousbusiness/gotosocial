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
	"fmt"
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

	fmt.Println(feed)

	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>Posts from @admin@localhost:8080</title>\n    <link>http://localhost:8080/@admin</link>\n    <description>Posts from @admin@localhost:8080</description>\n    <pubDate>Wed, 20 Oct 2021 12:36:45 +0000</pubDate>\n    <lastBuildDate>Wed, 20 Oct 2021 12:36:45 +0000</lastBuildDate>\n    <item>\n      <title>open to see some puppies</title>\n      <link>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</link>\n      <description>@admin@localhost:8080 made a new post: &#34;🐕🐕🐕🐕🐕&#34;</description>\n      <content:encoded><![CDATA[🐕🐕🐕🐕🐕]]></content:encoded>\n      <author>@admin@localhost:8080</author>\n      <guid>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</guid>\n      <pubDate>Wed, 20 Oct 2021 12:36:45 +0000</pubDate>\n      <source>http://localhost:8080/@admin/feed.rss</source>\n    </item>\n    <item>\n      <title>hello world! #welcome ! first post on the instance :rainbow: !</title>\n      <link>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</link>\n      <description>@admin@localhost:8080 posted 1 attachment: &#34;hello world! #welcome ! first post on the instance :rainbow: !&#34;</description>\n      <content:encoded><![CDATA[hello world! #welcome ! first post on the instance <img src=\"http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png\" title=\":rainbow:\" alt=\":rainbow:\" class=\"emoji\"/> !]]></content:encoded>\n      <author>@admin@localhost:8080</author>\n      <enclosure url=\"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg\" length=\"62529\" type=\"image/jpeg\"></enclosure>\n      <guid>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</guid>\n      <pubDate>Wed, 20 Oct 2021 11:36:45 +0000</pubDate>\n      <source>http://localhost:8080/@admin/feed.rss</source>\n    </item>\n  </channel>\n</rss>", feed)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZork() {
	getFeed, lastModified, err := suite.accountProcessor.GetRSSFeedForUsername(context.Background(), "the_mighty_zork")
	suite.NoError(err)
	suite.EqualValues(1634726437, lastModified.Unix())

	feed, err := getFeed()
	suite.NoError(err)

	fmt.Println(feed)

	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>Posts from @the_mighty_zork@localhost:8080</title>\n    <link>http://localhost:8080/@the_mighty_zork</link>\n    <description>Posts from @the_mighty_zork@localhost:8080</description>\n    <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>\n    <lastBuildDate>Wed, 20 Oct 2021 10:40:37 +0000</lastBuildDate>\n    <image>\n      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg</url>\n      <title>Avatar for @the_mighty_zork@localhost:8080</title>\n      <link>http://localhost:8080/@the_mighty_zork</link>\n    </image>\n    <item>\n      <title>introduction post</title>\n      <link>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</link>\n      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;hello everyone!&#34;</description>\n      <content:encoded><![CDATA[hello everyone!]]></content:encoded>\n      <author>@the_mighty_zork@localhost:8080</author>\n      <guid>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</guid>\n      <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>\n      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>\n    </item>\n  </channel>\n</rss>", feed)
}

func TestGetRSSTestSuite(t *testing.T) {
	suite.Run(t, new(GetRSSTestSuite))
}
