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
	"testing"

	"github.com/stretchr/testify/suite"
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
	suite.Equal("hello world! #welcome ! first post on the instance <img src=\"http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png\" title=\":rainbow:\" alt=\":rainbow:\" class=\"emoji\"/> !", item.Content)
}

func TestInternalToRSSTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToRSSTestSuite))
}
