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
	suite.EqualValues(1634726497, lastModified.Unix())

	feed, err := getFeed()
	suite.NoError(err)
	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>Posts from @admin@localhost:8080</title>\n    <link>http://localhost:8080/@admin</link>\n    <description>Posts from @admin@localhost:8080</description>\n    <pubDate>Wed, 20 Oct 2021 10:41:37 +0000</pubDate>\n    <lastBuildDate>Wed, 20 Oct 2021 10:41:37 +0000</lastBuildDate>\n    <item>\n      <title>open to see some puppies</title>\n      <link>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</link>\n      <description>@admin@localhost:8080 made a new post: &#34;🐕🐕🐕🐕🐕&#34;</description>\n      <content:encoded><![CDATA[🐕🐕🐕🐕🐕]]></content:encoded>\n      <author>@admin@localhost:8080</author>\n      <guid>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</guid>\n      <pubDate>Wed, 20 Oct 2021 12:36:45 +0000</pubDate>\n      <source>http://localhost:8080/@admin/feed.rss</source>\n    </item>\n    <item>\n      <title>hello world! #welcome ! first post on the instance :rainbow: !</title>\n      <link>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</link>\n      <description>@admin@localhost:8080 posted 1 attachment: &#34;hello world! #welcome ! first post on the instance :rainbow: !&#34;</description>\n      <content:encoded><![CDATA[hello world! #welcome ! first post on the instance <img src=\"http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png\" title=\":rainbow:\" alt=\":rainbow:\" width=\"25\" height=\"25\"/> !]]></content:encoded>\n      <author>@admin@localhost:8080</author>\n      <enclosure url=\"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg\" length=\"62529\" type=\"image/jpeg\"></enclosure>\n      <guid>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</guid>\n      <pubDate>Wed, 20 Oct 2021 11:36:45 +0000</pubDate>\n      <source>http://localhost:8080/@admin/feed.rss</source>\n    </item>\n  </channel>\n</rss>", feed)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZork() {
	getFeed, lastModified, err := suite.accountProcessor.GetRSSFeedForUsername(context.Background(), "the_mighty_zork")
	suite.NoError(err)
	suite.EqualValues(1702200240, lastModified.Unix())

	feed, err := getFeed()
	suite.NoError(err)
	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>Posts from @the_mighty_zork@localhost:8080</title>\n    <link>http://localhost:8080/@the_mighty_zork</link>\n    <description>Posts from @the_mighty_zork@localhost:8080</description>\n    <pubDate>Sun, 10 Dec 2023 09:24:00 +0000</pubDate>\n    <lastBuildDate>Sun, 10 Dec 2023 09:24:00 +0000</lastBuildDate>\n    <image>\n      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpg</url>\n      <title>Avatar for @the_mighty_zork@localhost:8080</title>\n      <link>http://localhost:8080/@the_mighty_zork</link>\n    </image>\n    <item>\n      <title>HTML in post</title>\n      <link>http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40</link>\n      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;Here&#39;s a bunch of HTML, read it and weep, weep then!&#xA;&#xA;```html&#xA;&lt;section class=&#34;about-user&#34;&gt;&#xA; &lt;div class=&#34;col-header&#34;&gt;&#xA; &lt;h2&gt;About&lt;/h2&gt;&#xA; &lt;/div&gt; &#xA; &lt;div class=&#34;fields&#34;&gt;&#xA; &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;&#xA; &lt;dl&gt;&#xA;...</description>\n      <content:encoded><![CDATA[<p>Here's a bunch of HTML, read it and weep, weep then!</p><pre><code class=\"language-html\">&lt;section class=&#34;about-user&#34;&gt;\n    &lt;div class=&#34;col-header&#34;&gt;\n        &lt;h2&gt;About&lt;/h2&gt;\n    &lt;/div&gt;            \n    &lt;div class=&#34;fields&#34;&gt;\n        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;\n        &lt;dl&gt;\n            &lt;div class=&#34;field&#34;&gt;\n                &lt;dt&gt;should you follow me?&lt;/dt&gt;\n                &lt;dd&gt;maybe!&lt;/dd&gt;\n            &lt;/div&gt;\n            &lt;div class=&#34;field&#34;&gt;\n                &lt;dt&gt;age&lt;/dt&gt;\n                &lt;dd&gt;120&lt;/dd&gt;\n            &lt;/div&gt;\n        &lt;/dl&gt;\n    &lt;/div&gt;\n    &lt;div class=&#34;bio&#34;&gt;\n        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;\n        &lt;p&gt;i post about things that concern me&lt;/p&gt;\n    &lt;/div&gt;\n    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;\n        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;\n        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;\n        &lt;span&gt;8 posts.&lt;/span&gt;\n        &lt;span&gt;Followed by 1.&lt;/span&gt;\n        &lt;span&gt;Following 1.&lt;/span&gt;\n    &lt;/div&gt;\n    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;\n        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;\n        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;\n        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;\n        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;\n    &lt;/div&gt;\n&lt;/section&gt;\n</code></pre><p>There, hope you liked that!</p>]]></content:encoded>\n      <author>@the_mighty_zork@localhost:8080</author>\n      <guid>http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40</guid>\n      <pubDate>Sun, 10 Dec 2023 09:24:00 +0000</pubDate>\n      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>\n    </item>\n    <item>\n      <title>introduction post</title>\n      <link>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</link>\n      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;hello everyone!&#34;</description>\n      <content:encoded><![CDATA[hello everyone!]]></content:encoded>\n      <author>@the_mighty_zork@localhost:8080</author>\n      <guid>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</guid>\n      <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>\n      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>\n    </item>\n  </channel>\n</rss>", feed)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZorkNoPosts() {
	ctx := context.Background()

	// Get all of zork's posts.
	statuses, err := suite.db.GetAccountStatuses(ctx, suite.testAccounts["local_account_1"].ID, 0, false, false, "", "", false, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Now delete them! Hahaha!
	for _, status := range statuses {
		if err := suite.db.DeleteStatusByID(ctx, status.ID); err != nil {
			suite.FailNow(err.Error())
		}
	}

	getFeed, lastModified, err := suite.accountProcessor.GetRSSFeedForUsername(ctx, "the_mighty_zork")
	suite.NoError(err)
	suite.Empty(lastModified)

	feed, err := getFeed()
	suite.NoError(err)
	suite.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?><rss version=\"2.0\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n  <channel>\n    <title>Posts from @the_mighty_zork@localhost:8080</title>\n    <link>http://localhost:8080/@the_mighty_zork</link>\n    <description>Posts from @the_mighty_zork@localhost:8080</description>\n    <pubDate>Fri, 20 May 2022 11:09:18 +0000</pubDate>\n    <lastBuildDate>Fri, 20 May 2022 11:09:18 +0000</lastBuildDate>\n    <image>\n      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpg</url>\n      <title>Avatar for @the_mighty_zork@localhost:8080</title>\n      <link>http://localhost:8080/@the_mighty_zork</link>\n    </image>\n  </channel>\n</rss>", feed)
}

func TestGetRSSTestSuite(t *testing.T) {
	suite.Run(t, new(GetRSSTestSuite))
}
