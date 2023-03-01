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

package ap_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type ExtractAttachmentsTestSuite struct {
	ExtractTestSuite
}

func (suite *ExtractAttachmentsTestSuite) TestExtractAttachmentMissingURL() {
	d1 := suite.document1
	d1.SetActivityStreamsUrl(streams.NewActivityStreamsUrlProperty())

	attachment, err := ap.ExtractAttachment(d1)
	suite.EqualError(err, "could not extract url")
	suite.Nil(attachment)
}

func TestExtractAttachmentsTestSuite(t *testing.T) {
	suite.Run(t, &ExtractAttachmentsTestSuite{})
}
