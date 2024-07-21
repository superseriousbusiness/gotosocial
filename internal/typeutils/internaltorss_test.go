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

package typeutils_test

import (
	"context"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type InternalToRSSTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToRSSTestSuite) TestStatusToRSSItem1() {
	s := suite.testStatuses["local_account_1_status_1"]
	item, err := suite.typeconverter.StatusToRSSItem(context.Background(), s)
	suite.NoError(err)

	suite.Equal("introduction post", item.Title)
	suite.Equal("http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY", item.Link.Href)
	suite.Equal("", item.Link.Length)
	suite.Equal("", item.Link.Rel)
	suite.Equal("", item.Link.Type)
	suite.Equal("http://localhost:8080/@the_mighty_zork/feed.rss", item.Source.Href)
	suite.Equal("", item.Source.Length)
	suite.Equal("", item.Source.Rel)
	suite.Equal("", item.Source.Type)
	suite.Equal("", item.Author.Email)
	suite.Equal("@the_mighty_zork@localhost:8080", item.Author.Name)
	suite.Equal("@the_mighty_zork@localhost:8080 made a new post: \"hello everyone!\"", item.Description)
	suite.Equal("http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY", item.Id)
	suite.EqualValues(1634726437, item.Updated.Unix())
	suite.EqualValues(1634726437, item.Created.Unix())
	suite.Equal("", item.Enclosure.Length)
	suite.Equal("", item.Enclosure.Type)
	suite.Equal("", item.Enclosure.Url)
	suite.Equal("hello everyone!", item.Content)
}

func (suite *InternalToRSSTestSuite) TestStatusToRSSItem2() {
	s := suite.testStatuses["admin_account_status_1"]
	item, err := suite.typeconverter.StatusToRSSItem(context.Background(), s)
	suite.NoError(err)

	suite.Equal("hello world! #welcome ! first post on the instance :rainbow: !", item.Title)
	suite.Equal("http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R", item.Link.Href)
	suite.Equal("", item.Link.Length)
	suite.Equal("", item.Link.Rel)
	suite.Equal("", item.Link.Type)
	suite.Equal("http://localhost:8080/@admin/feed.rss", item.Source.Href)
	suite.Equal("", item.Source.Length)
	suite.Equal("", item.Source.Rel)
	suite.Equal("", item.Source.Type)
	suite.Equal("", item.Author.Email)
	suite.Equal("@admin@localhost:8080", item.Author.Name)
	suite.Equal("@admin@localhost:8080 posted 1 attachment: \"hello world! #welcome ! first post on the instance :rainbow: !\"", item.Description)
	suite.Equal("http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R", item.Id)
	suite.EqualValues(1634729805, item.Updated.Unix())
	suite.EqualValues(1634729805, item.Created.Unix())
	suite.Equal("62529", item.Enclosure.Length)
	suite.Equal("image/jpeg", item.Enclosure.Type)
	suite.Equal("http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg", item.Enclosure.Url)
	suite.Equal("hello world! #welcome ! first post on the instance <img src=\"http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png\" title=\":rainbow:\" alt=\":rainbow:\" width=\"25\" height=\"25\" /> !", item.Content)
}

func (suite *InternalToRSSTestSuite) TestStatusToRSSItem3() {
	account := suite.testAccounts["admin_account"]
	s := &gtsmodel.Status{
		ID:                  "01H7G0VW1ACBZTRHN6RSA4JWVH",
		URI:                 "http://localhost:8080/users/admin/statuses/01H7G0VW1ACBZTRHN6RSA4JWVH",
		URL:                 "http://localhost:8080/@admin/statuses/01H7G0VW1ACBZTRHN6RSA4JWVH",
		ContentWarning:      "这是简体中文帖子的一些示例内容。\n\n我希望我能读到这个，因为与无聊的旧 ASCII 相比，这些字符绝对漂亮。 不幸的是，我是一个愚蠢的西方人。\n\n无论如何，无论是谁读到这篇文章，你今天过得怎么样？ 希望你过得愉快！ 如果您有一段时间没有这样做，请从椅子上站起来，喝一杯水，并将您的眼睛集中在远处的物体上，而不是电脑屏幕上！",
		Content:             "这是另一段，只是为了确保这篇文章足够长。 通过前肢上长而弯曲的爪子的数量可以轻松识别不同的树懒类别。 顾名思义，二趾树懒的前肢上有两个爪子，而三趾树懒的四个肢上都有三个爪子。 二趾树懒也比三趾树懒稍大，并且都属于不同的分类科。 美洲共有六种树懒，主要分布在中美洲和南美洲的热带雨林中。\n\n\n\n\t霍夫曼二趾树懒 (Choloepus hoffmanni)\n\n\t林奈二趾树懒 (Choloepus didactylus)\n\n\t侏儒三趾树懒 (Bradypus pygmaeus)\n\n\t鬃三趾树懒 (Bradypus torquatus)\n\n\t棕喉树懒 (Bradypus variegatus)\n\n\t浅喉树懒 (Bradypus tridactylus)\n\n\n\n目前，有 4 种树懒被 IUCN 濒危物种红色名录列为最不受关注的物种。 鬃毛三趾树懒很脆弱，而侏儒三趾树懒则极度濒危，树懒物种面临最大的灭绝风险。",
		Local:               util.Ptr(true),
		AccountURI:          account.URI,
		AccountID:           account.ID,
		Visibility:          gtsmodel.VisibilityDefault,
		ActivityStreamsType: ap.ObjectNote,
		Federated:           util.Ptr(true),
	}
	item, err := suite.typeconverter.StatusToRSSItem(context.Background(), s)
	suite.NoError(err)

	data, err := xml.MarshalIndent(item, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`<Item>
  <Title>这是简体中文帖子的一些示例内容。&#xA;&#xA;我希望我能读到这个，因为与无聊的旧 ASCII 相比，这些字符绝对漂亮。 不幸的是，我是一个愚蠢的西方人。&#xA;&#xA;无论如何，无论是谁读到这篇文章，你今天过得怎么样？ 希望你过得愉快！ 如果您有一段时间没有这样做，请从椅...</Title>
  <Link>
    <Href>http://localhost:8080/@admin/statuses/01H7G0VW1ACBZTRHN6RSA4JWVH</Href>
    <Rel></Rel>
    <Type></Type>
    <Length></Length>
  </Link>
  <Source>
    <Href>http://localhost:8080/@admin/feed.rss</Href>
    <Rel></Rel>
    <Type></Type>
    <Length></Length>
  </Source>
  <Author>
    <Name>@admin@localhost:8080</Name>
    <Email></Email>
  </Author>
  <Description>@admin@localhost:8080 made a new post</Description>
  <Id>http://localhost:8080/@admin/statuses/01H7G0VW1ACBZTRHN6RSA4JWVH</Id>
  <IsPermaLink>true</IsPermaLink>
  <Updated>0001-01-01T00:00:00Z</Updated>
  <Created>0001-01-01T00:00:00Z</Created>
  <Enclosure>
    <Url></Url>
    <Length></Length>
    <Type></Type>
  </Enclosure>
  <Content>这是另一段，只是为了确保这篇文章足够长。 通过前肢上长而弯曲的爪子的数量可以轻松识别不同的树懒类别。 顾名思义，二趾树懒的前肢上有两个爪子，而三趾树懒的四个肢上都有三个爪子。 二趾树懒也比三趾树懒稍大，并且都属于不同的分类科。 美洲共有六种树懒，主要分布在中美洲和南美洲的热带雨林中。&#xA;&#xA;&#xA;&#xA;&#x9;霍夫曼二趾树懒 (Choloepus hoffmanni)&#xA;&#xA;&#x9;林奈二趾树懒 (Choloepus didactylus)&#xA;&#xA;&#x9;侏儒三趾树懒 (Bradypus pygmaeus)&#xA;&#xA;&#x9;鬃三趾树懒 (Bradypus torquatus)&#xA;&#xA;&#x9;棕喉树懒 (Bradypus variegatus)&#xA;&#xA;&#x9;浅喉树懒 (Bradypus tridactylus)&#xA;&#xA;&#xA;&#xA;目前，有 4 种树懒被 IUCN 濒危物种红色名录列为最不受关注的物种。 鬃毛三趾树懒很脆弱，而侏儒三趾树懒则极度濒危，树懒物种面临最大的灭绝风险。</Content>
</Item>`, string(data))
}

func TestInternalToRSSTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToRSSTestSuite))
}
