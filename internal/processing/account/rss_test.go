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
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gorilla/feeds"
	"github.com/stretchr/testify/suite"
)

type GetRSSTestSuite struct {
	AccountStandardTestSuite
}

func (suite *GetRSSTestSuite) TestGetAccountRSSAdmin() {
	suite.testGetFeedSerializedAs("admin", &paging.Page{Limit: 20}, (*feeds.Feed).ToRss, 1634726497,
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>Posts from @admin@localhost:8080</title>
    <link>http://localhost:8080/@admin</link>
    <description>Posts from @admin@localhost:8080</description>
    <pubDate>Wed, 20 Oct 2021 10:41:37 +0000</pubDate>
    <lastBuildDate>Wed, 20 Oct 2021 10:41:37 +0000</lastBuildDate>
    <item>
      <title>open to see some &lt;strong&gt;puppies&lt;/strong&gt;</title>
      <link>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</link>
      <description>@admin@localhost:8080 made a new post: &#34;🐕🐕🐕🐕🐕&#34;</description>
      <content:encoded><![CDATA[<p>🐕🐕🐕🐕🐕</p>]]></content:encoded>
      <author>@admin@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</guid>
      <pubDate>Wed, 20 Oct 2021 12:36:45 +0000</pubDate>
      <source>http://localhost:8080/@admin/feed.rss</source>
    </item>
    <item>
      <title>hello world! #welcome ! first post on the instance :rainbow: !</title>
      <link>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</link>
      <description>@admin@localhost:8080 posted 1 attachment: &#34;hello world! #welcome ! first post on the instance :rainbow: !&#34;</description>
      <content:encoded><![CDATA[<p>hello world! <a href="http://localhost:8080/tags/welcome" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>welcome</span></a> ! first post on the instance <img src="http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png" title=":rainbow:" alt=":rainbow:" width="25" height="25" /> !</p>]]></content:encoded>
      <author>@admin@localhost:8080</author>
      <enclosure url="http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg" length="62529" type="image/jpeg"></enclosure>
      <guid isPermaLink="true">http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</guid>
      <pubDate>Wed, 20 Oct 2021 11:36:45 +0000</pubDate>
      <source>http://localhost:8080/@admin/feed.rss</source>
    </item>
  </channel>
</rss>`)
}

func (suite *GetRSSTestSuite) TestGetAccountAtomAdmin() {
	suite.testGetFeedSerializedAs("admin", &paging.Page{Limit: 20}, (*feeds.Feed).ToAtom, 1634726497,
		`<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom">
  <title>Posts from @admin@localhost:8080</title>
  <id>http://localhost:8080/@admin</id>
  <updated>2021-10-20T10:41:37Z</updated>
  <subtitle>Posts from @admin@localhost:8080</subtitle>
  <link href="http://localhost:8080/@admin"></link>
  <entry>
    <title>open to see some &lt;strong&gt;puppies&lt;/strong&gt;</title>
    <updated>2021-10-20T12:36:45Z</updated>
    <id>http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37</id>
    <content type="html">&lt;p&gt;🐕🐕🐕🐕🐕&lt;/p&gt;</content>
    <link href="http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37" rel="alternate"></link>
    <link href="" rel="enclosure"></link>
    <summary type="html">@admin@localhost:8080 made a new post: &#34;🐕🐕🐕🐕🐕&#34;</summary>
    <author>
      <name>@admin@localhost:8080</name>
    </author>
  </entry>
  <entry>
    <title>hello world! #welcome ! first post on the instance :rainbow: !</title>
    <updated>2021-10-20T11:36:45Z</updated>
    <id>http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R</id>
    <content type="html">&lt;p&gt;hello world! &lt;a href=&#34;http://localhost:8080/tags/welcome&#34; class=&#34;mention hashtag&#34; rel=&#34;tag nofollow noreferrer noopener&#34; target=&#34;_blank&#34;&gt;#&lt;span&gt;welcome&lt;/span&gt;&lt;/a&gt; ! first post on the instance &lt;img src=&#34;http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png&#34; title=&#34;:rainbow:&#34; alt=&#34;:rainbow:&#34; width=&#34;25&#34; height=&#34;25&#34; /&gt; !&lt;/p&gt;</content>
    <link href="http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R" rel="alternate"></link>
    <link href="http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg" rel="enclosure" type="image/jpeg" length="62529"></link>
    <summary type="html">@admin@localhost:8080 posted 1 attachment: &#34;hello world! #welcome ! first post on the instance :rainbow: !&#34;</summary>
    <author>
      <name>@admin@localhost:8080</name>
    </author>
  </entry>
</feed>`)
}

func (suite *GetRSSTestSuite) TestGetAccountJSONAdmin() {
	suite.testGetFeedSerializedAs("admin", &paging.Page{Limit: 20}, (*feeds.Feed).ToJSON, 1634726497,
		`{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "Posts from @admin@localhost:8080",
  "home_page_url": "http://localhost:8080/@admin",
  "description": "Posts from @admin@localhost:8080",
  "items": [
    {
      "id": "http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
      "url": "http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
      "external_url": "http://localhost:8080/@admin/feed.rss",
      "title": "open to see some \u003cstrong\u003epuppies\u003c/strong\u003e",
      "content_html": "\u003cp\u003e🐕🐕🐕🐕🐕\u003c/p\u003e",
      "summary": "@admin@localhost:8080 made a new post: \"🐕🐕🐕🐕🐕\"",
      "date_published": "2021-10-20T12:36:45Z",
      "author": {
        "name": "@admin@localhost:8080"
      },
      "authors": [
        {
          "name": "@admin@localhost:8080"
        }
      ]
    },
    {
      "id": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
      "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
      "external_url": "http://localhost:8080/@admin/feed.rss",
      "title": "hello world! #welcome ! first post on the instance :rainbow: !",
      "content_html": "\u003cp\u003ehello world! \u003ca href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\"\u003e#\u003cspan\u003ewelcome\u003c/span\u003e\u003c/a\u003e ! first post on the instance \u003cimg src=\"http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png\" title=\":rainbow:\" alt=\":rainbow:\" width=\"25\" height=\"25\" /\u003e !\u003c/p\u003e",
      "summary": "@admin@localhost:8080 posted 1 attachment: \"hello world! #welcome ! first post on the instance :rainbow: !\"",
      "image": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "date_published": "2021-10-20T11:36:45Z",
      "author": {
        "name": "@admin@localhost:8080"
      },
      "authors": [
        {
          "name": "@admin@localhost:8080"
        }
      ]
    }
  ]
}`)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZork() {
	suite.testGetFeedSerializedAs("the_mighty_zork", &paging.Page{Limit: 20}, (*feeds.Feed).ToRss, 1730451600,
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>Posts from @the_mighty_zork@localhost:8080</title>
    <link>http://localhost:8080/@the_mighty_zork</link>
    <description>Posts from @the_mighty_zork@localhost:8080</description>
    <pubDate>Fri, 01 Nov 2024 09:00:00 +0000</pubDate>
    <lastBuildDate>Fri, 01 Nov 2024 09:00:00 +0000</lastBuildDate>
    <image>
      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp</url>
      <title>Avatar for @the_mighty_zork@localhost:8080</title>
      <link>http://localhost:8080/@the_mighty_zork</link>
    </image>
    <item>
      <title>edited status</title>
      <link>http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR</link>
      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;this is the latest revision of the status, with a content-warning&#34;</description>
      <content:encoded><![CDATA[<p>this is the latest revision of the status, with a content-warning</p>]]></content:encoded>
      <author>@the_mighty_zork@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR</guid>
      <pubDate>Fri, 01 Nov 2024 09:00:00 +0000</pubDate>
      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>
    </item>
    <item>
      <title>HTML in post</title>
      <link>http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40</link>
      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;Here&#39;s a bunch of HTML, read it and weep, weep then!&#xA;&#xA;`+"```"+`html&#xA;&lt;section class=&#34;about-user&#34;&gt;&#xA; &lt;div class=&#34;col-header&#34;&gt;&#xA; &lt;h2&gt;About&lt;/h2&gt;&#xA; &lt;/div&gt; &#xA; &lt;div class=&#34;fields&#34;&gt;&#xA; &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;&#xA; &lt;dl&gt;&#xA;...</description>
      <content:encoded><![CDATA[<p>Here's a bunch of HTML, read it and weep, weep then!</p><pre><code class="language-html">&lt;section class=&#34;about-user&#34;&gt;
    &lt;div class=&#34;col-header&#34;&gt;
        &lt;h2&gt;About&lt;/h2&gt;
    &lt;/div&gt;            
    &lt;div class=&#34;fields&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;
        &lt;dl&gt;
            &lt;div class=&#34;field&#34;&gt;
                &lt;dt&gt;should you follow me?&lt;/dt&gt;
                &lt;dd&gt;maybe!&lt;/dd&gt;
            &lt;/div&gt;
            &lt;div class=&#34;field&#34;&gt;
                &lt;dt&gt;age&lt;/dt&gt;
                &lt;dd&gt;120&lt;/dd&gt;
            &lt;/div&gt;
        &lt;/dl&gt;
    &lt;/div&gt;
    &lt;div class=&#34;bio&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;
        &lt;p&gt;i post about things that concern me&lt;/p&gt;
    &lt;/div&gt;
    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;
        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;
        &lt;span&gt;8 posts.&lt;/span&gt;
        &lt;span&gt;Followed by 1.&lt;/span&gt;
        &lt;span&gt;Following 1.&lt;/span&gt;
    &lt;/div&gt;
    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;
        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;
        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;
        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
    &lt;/div&gt;
&lt;/section&gt;
</code></pre><p>There, hope you liked that!</p>]]></content:encoded>
      <author>@the_mighty_zork@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40</guid>
      <pubDate>Sun, 10 Dec 2023 09:24:00 +0000</pubDate>
      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>
    </item>
    <item>
      <title>introduction post</title>
      <link>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</link>
      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;hello everyone!&#34;</description>
      <content:encoded><![CDATA[<p>hello everyone!</p>]]></content:encoded>
      <author>@the_mighty_zork@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</guid>
      <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>
      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>
    </item>
  </channel>
</rss>`)
}

func (suite *GetRSSTestSuite) TestGetAccountAtomZork() {
	suite.testGetFeedSerializedAs("the_mighty_zork", &paging.Page{Limit: 20}, (*feeds.Feed).ToRss, 1730451600,
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>Posts from @the_mighty_zork@localhost:8080</title>
    <link>http://localhost:8080/@the_mighty_zork</link>
    <description>Posts from @the_mighty_zork@localhost:8080</description>
    <pubDate>Fri, 01 Nov 2024 09:00:00 +0000</pubDate>
    <lastBuildDate>Fri, 01 Nov 2024 09:00:00 +0000</lastBuildDate>
    <image>
      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp</url>
      <title>Avatar for @the_mighty_zork@localhost:8080</title>
      <link>http://localhost:8080/@the_mighty_zork</link>
    </image>
    <item>
      <title>edited status</title>
      <link>http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR</link>
      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;this is the latest revision of the status, with a content-warning&#34;</description>
      <content:encoded><![CDATA[<p>this is the latest revision of the status, with a content-warning</p>]]></content:encoded>
      <author>@the_mighty_zork@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR</guid>
      <pubDate>Fri, 01 Nov 2024 09:00:00 +0000</pubDate>
      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>
    </item>
    <item>
      <title>HTML in post</title>
      <link>http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40</link>
      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;Here&#39;s a bunch of HTML, read it and weep, weep then!&#xA;&#xA;`+"```"+`html&#xA;&lt;section class=&#34;about-user&#34;&gt;&#xA; &lt;div class=&#34;col-header&#34;&gt;&#xA; &lt;h2&gt;About&lt;/h2&gt;&#xA; &lt;/div&gt; &#xA; &lt;div class=&#34;fields&#34;&gt;&#xA; &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;&#xA; &lt;dl&gt;&#xA;...</description>
      <content:encoded><![CDATA[<p>Here's a bunch of HTML, read it and weep, weep then!</p><pre><code class="language-html">&lt;section class=&#34;about-user&#34;&gt;
    &lt;div class=&#34;col-header&#34;&gt;
        &lt;h2&gt;About&lt;/h2&gt;
    &lt;/div&gt;            
    &lt;div class=&#34;fields&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;
        &lt;dl&gt;
            &lt;div class=&#34;field&#34;&gt;
                &lt;dt&gt;should you follow me?&lt;/dt&gt;
                &lt;dd&gt;maybe!&lt;/dd&gt;
            &lt;/div&gt;
            &lt;div class=&#34;field&#34;&gt;
                &lt;dt&gt;age&lt;/dt&gt;
                &lt;dd&gt;120&lt;/dd&gt;
            &lt;/div&gt;
        &lt;/dl&gt;
    &lt;/div&gt;
    &lt;div class=&#34;bio&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;
        &lt;p&gt;i post about things that concern me&lt;/p&gt;
    &lt;/div&gt;
    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;
        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;
        &lt;span&gt;8 posts.&lt;/span&gt;
        &lt;span&gt;Followed by 1.&lt;/span&gt;
        &lt;span&gt;Following 1.&lt;/span&gt;
    &lt;/div&gt;
    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;
        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;
        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;
        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
    &lt;/div&gt;
&lt;/section&gt;
</code></pre><p>There, hope you liked that!</p>]]></content:encoded>
      <author>@the_mighty_zork@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40</guid>
      <pubDate>Sun, 10 Dec 2023 09:24:00 +0000</pubDate>
      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>
    </item>
    <item>
      <title>introduction post</title>
      <link>http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</link>
      <description>@the_mighty_zork@localhost:8080 made a new post: &#34;hello everyone!&#34;</description>
      <content:encoded><![CDATA[<p>hello everyone!</p>]]></content:encoded>
      <author>@the_mighty_zork@localhost:8080</author>
      <guid isPermaLink="true">http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY</guid>
      <pubDate>Wed, 20 Oct 2021 10:40:37 +0000</pubDate>
      <source>http://localhost:8080/@the_mighty_zork/feed.rss</source>
    </item>
  </channel>
</rss>`)
}

func (suite *GetRSSTestSuite) TestGetAccountJSONZork() {
	suite.testGetFeedSerializedAs("the_mighty_zork", &paging.Page{Limit: 20}, (*feeds.Feed).ToJSON, 1730451600,
		`{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "Posts from @the_mighty_zork@localhost:8080",
  "home_page_url": "http://localhost:8080/@the_mighty_zork",
  "description": "Posts from @the_mighty_zork@localhost:8080",
  "items": [
    {
      "id": "http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR",
      "url": "http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR",
      "external_url": "http://localhost:8080/@the_mighty_zork/feed.rss",
      "title": "edited status",
      "content_html": "\u003cp\u003ethis is the latest revision of the status, with a content-warning\u003c/p\u003e",
      "summary": "@the_mighty_zork@localhost:8080 made a new post: \"this is the latest revision of the status, with a content-warning\"",
      "date_published": "2024-11-01T09:00:00Z",
      "date_modified": "2024-11-01T09:02:00Z",
      "author": {
        "name": "@the_mighty_zork@localhost:8080"
      },
      "authors": [
        {
          "name": "@the_mighty_zork@localhost:8080"
        }
      ]
    },
    {
      "id": "http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40",
      "url": "http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40",
      "external_url": "http://localhost:8080/@the_mighty_zork/feed.rss",
      "title": "HTML in post",
      "content_html": "\u003cp\u003eHere's a bunch of HTML, read it and weep, weep then!\u003c/p\u003e\u003cpre\u003e\u003ccode class=\"language-html\"\u003e\u0026lt;section class=\u0026#34;about-user\u0026#34;\u0026gt;\n    \u0026lt;div class=\u0026#34;col-header\u0026#34;\u0026gt;\n        \u0026lt;h2\u0026gt;About\u0026lt;/h2\u0026gt;\n    \u0026lt;/div\u0026gt;            \n    \u0026lt;div class=\u0026#34;fields\u0026#34;\u0026gt;\n        \u0026lt;h3 class=\u0026#34;sr-only\u0026#34;\u0026gt;Fields\u0026lt;/h3\u0026gt;\n        \u0026lt;dl\u0026gt;\n            \u0026lt;div class=\u0026#34;field\u0026#34;\u0026gt;\n                \u0026lt;dt\u0026gt;should you follow me?\u0026lt;/dt\u0026gt;\n                \u0026lt;dd\u0026gt;maybe!\u0026lt;/dd\u0026gt;\n            \u0026lt;/div\u0026gt;\n            \u0026lt;div class=\u0026#34;field\u0026#34;\u0026gt;\n                \u0026lt;dt\u0026gt;age\u0026lt;/dt\u0026gt;\n                \u0026lt;dd\u0026gt;120\u0026lt;/dd\u0026gt;\n            \u0026lt;/div\u0026gt;\n        \u0026lt;/dl\u0026gt;\n    \u0026lt;/div\u0026gt;\n    \u0026lt;div class=\u0026#34;bio\u0026#34;\u0026gt;\n        \u0026lt;h3 class=\u0026#34;sr-only\u0026#34;\u0026gt;Bio\u0026lt;/h3\u0026gt;\n        \u0026lt;p\u0026gt;i post about things that concern me\u0026lt;/p\u0026gt;\n    \u0026lt;/div\u0026gt;\n    \u0026lt;div class=\u0026#34;sr-only\u0026#34; role=\u0026#34;group\u0026#34;\u0026gt;\n        \u0026lt;h3 class=\u0026#34;sr-only\u0026#34;\u0026gt;Stats\u0026lt;/h3\u0026gt;\n        \u0026lt;span\u0026gt;Joined in Jun, 2022.\u0026lt;/span\u0026gt;\n        \u0026lt;span\u0026gt;8 posts.\u0026lt;/span\u0026gt;\n        \u0026lt;span\u0026gt;Followed by 1.\u0026lt;/span\u0026gt;\n        \u0026lt;span\u0026gt;Following 1.\u0026lt;/span\u0026gt;\n    \u0026lt;/div\u0026gt;\n    \u0026lt;div class=\u0026#34;accountstats\u0026#34; aria-hidden=\u0026#34;true\u0026#34;\u0026gt;\n        \u0026lt;b\u0026gt;Joined\u0026lt;/b\u0026gt;\u0026lt;time datetime=\u0026#34;2022-06-04T13:12:00.000Z\u0026#34;\u0026gt;Jun, 2022\u0026lt;/time\u0026gt;\n        \u0026lt;b\u0026gt;Posts\u0026lt;/b\u0026gt;\u0026lt;span\u0026gt;8\u0026lt;/span\u0026gt;\n        \u0026lt;b\u0026gt;Followed by\u0026lt;/b\u0026gt;\u0026lt;span\u0026gt;1\u0026lt;/span\u0026gt;\n        \u0026lt;b\u0026gt;Following\u0026lt;/b\u0026gt;\u0026lt;span\u0026gt;1\u0026lt;/span\u0026gt;\n    \u0026lt;/div\u0026gt;\n\u0026lt;/section\u0026gt;\n\u003c/code\u003e\u003c/pre\u003e\u003cp\u003eThere, hope you liked that!\u003c/p\u003e",
      "summary": "@the_mighty_zork@localhost:8080 made a new post: \"Here's a bunch of HTML, read it and weep, weep then!\n\n`+"```"+`html\n\u003csection class=\"about-user\"\u003e\n \u003cdiv class=\"col-header\"\u003e\n \u003ch2\u003eAbout\u003c/h2\u003e\n \u003c/div\u003e \n \u003cdiv class=\"fields\"\u003e\n \u003ch3 class=\"sr-only\"\u003eFields\u003c/h3\u003e\n \u003cdl\u003e\n...",
      "date_published": "2023-12-10T09:24:00Z",
      "author": {
        "name": "@the_mighty_zork@localhost:8080"
      },
      "authors": [
        {
          "name": "@the_mighty_zork@localhost:8080"
        }
      ]
    },
    {
      "id": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
      "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
      "external_url": "http://localhost:8080/@the_mighty_zork/feed.rss",
      "title": "introduction post",
      "content_html": "\u003cp\u003ehello everyone!\u003c/p\u003e",
      "summary": "@the_mighty_zork@localhost:8080 made a new post: \"hello everyone!\"",
      "date_published": "2021-10-20T10:40:37Z",
      "author": {
        "name": "@the_mighty_zork@localhost:8080"
      },
      "authors": [
        {
          "name": "@the_mighty_zork@localhost:8080"
        }
      ]
    }
  ]
}`)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZorkNoPosts() {
	ctx := suite.T().Context()

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

	var zeroTime time.Time

	suite.testGetFeedSerializedAs("the_mighty_zork", &paging.Page{Limit: 20}, (*feeds.Feed).ToRss, zeroTime.Unix(),
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>Posts from @the_mighty_zork@localhost:8080</title>
    <link>http://localhost:8080/@the_mighty_zork</link>
    <description>Posts from @the_mighty_zork@localhost:8080</description>
    <pubDate>Fri, 20 May 2022 11:09:18 +0000</pubDate>
    <lastBuildDate>Fri, 20 May 2022 11:09:18 +0000</lastBuildDate>
    <image>
      <url>http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp</url>
      <title>Avatar for @the_mighty_zork@localhost:8080</title>
      <link>http://localhost:8080/@the_mighty_zork</link>
    </image>
  </channel>
</rss>`)
}

// func (suite *GetRSSTestSuite) testGetAccountRSSPaging(username string, page *paging.Page, expectIDs []string) {
// 	ctx := suite.T().Context()

// 	getFeed, _, errWithCode := suite.accountProcessor.GetRSSFeedForUsername(ctx, username, page)
// 	suite.NoError(errWithCode)

// 	feed, errWithCode := getFeed()
// 	suite.NoError(errWithCode)

// }

func (suite *GetRSSTestSuite) testGetFeedSerializedAs(username string, page *paging.Page, serialize func(*feeds.Feed) (string, error), expectLastMod int64, expectSerialized string) {
	ctx := suite.T().Context()

	getFeed, lastMod, errWithCode := suite.accountProcessor.GetRSSFeedForUsername(ctx, username, page)
	suite.NoError(errWithCode)
	suite.Equal(expectLastMod, lastMod.Unix())

	feed, errWithCode := getFeed()
	suite.NoError(errWithCode)

	feedStr, err := serialize(feed)
	suite.NoError(err)
	suite.Equal(expectSerialized, feedStr)
}

func TestGetRSSTestSuite(t *testing.T) {
	suite.Run(t, new(GetRSSTestSuite))
}
