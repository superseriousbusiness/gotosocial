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

package media

import (
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MediaUtilTestSuite struct {
	suite.Suite
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *MediaUtilTestSuite) SetupSuite() {
	// doesn't use testrig.InitTestLog() helper to prevent import cycle
	err := log.Initialize(logrus.TraceLevel.String())
	if err != nil {
		panic(err)
	}

}

func (suite *MediaUtilTestSuite) TearDownSuite() {

}

// SetupTest creates a db connection and creates necessary tables before each test
func (suite *MediaUtilTestSuite) SetupTest() {

}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *MediaUtilTestSuite) TearDownTest() {

}

/*
	ACTUAL TESTS
*/

func (suite *MediaUtilTestSuite) TestParseContentTypeOK() {
	f, err := ioutil.ReadFile("./test/test-jpeg.jpg")
	suite.NoError(err)
	ct, err := parseContentType(f)
	suite.NoError(err)
	suite.Equal("image/jpeg", ct)
}

func (suite *MediaUtilTestSuite) TestParseContentTypeNotOK() {
	f, err := ioutil.ReadFile("./test/test-corrupted.jpg")
	suite.NoError(err)
	ct, err := parseContentType(f)
	suite.NotNil(err)
	suite.Equal("", ct)
	suite.Equal("filetype unknown", err.Error())
}

func (suite *MediaUtilTestSuite) TestRemoveEXIF() {
	// load and validate image
	b, err := ioutil.ReadFile("./test/test-with-exif.jpg")
	suite.NoError(err)

	// clean it up and validate the clean version
	clean, err := purgeExif(b)
	suite.NoError(err)

	// compare it to our stored sample
	sampleBytes, err := ioutil.ReadFile("./test/test-without-exif.jpg")
	suite.NoError(err)
	suite.EqualValues(sampleBytes, clean)
}

func (suite *MediaUtilTestSuite) TestDeriveImageFromJPEG() {
	// load image
	b, err := ioutil.ReadFile("./test/test-jpeg.jpg")
	suite.NoError(err)

	// clean it up and validate the clean version
	imageAndMeta, err := deriveImage(b, "image/jpeg")
	suite.NoError(err)

	suite.Equal(1920, imageAndMeta.width)
	suite.Equal(1080, imageAndMeta.height)
	suite.Equal(1.7777777777777777, imageAndMeta.aspect)
	suite.Equal(2073600, imageAndMeta.size)

	// assert that the final image is what we would expect
	sampleBytes, err := ioutil.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.EqualValues(sampleBytes, imageAndMeta.image)
}

func (suite *MediaUtilTestSuite) TestDeriveThumbnailFromJPEG() {
	// load image
	b, err := ioutil.ReadFile("./test/test-jpeg.jpg")
	suite.NoError(err)

	// clean it up and validate the clean version
	imageAndMeta, err := deriveThumbnail(b, "image/jpeg", 512, 512)
	suite.NoError(err)

	suite.Equal(512, imageAndMeta.width)
	suite.Equal(288, imageAndMeta.height)
	suite.Equal(1.7777777777777777, imageAndMeta.aspect)
	suite.Equal(147456, imageAndMeta.size)
	suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", imageAndMeta.blurhash)

	sampleBytes, err := ioutil.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.EqualValues(sampleBytes, imageAndMeta.image)
}

func (suite *MediaUtilTestSuite) TestSupportedImageTypes() {
	ok := SupportedImageType("image/jpeg")
	suite.True(ok)

	ok = SupportedImageType("image/bmp")
	suite.False(ok)
}

func TestMediaUtilTestSuite(t *testing.T) {
	suite.Run(t, new(MediaUtilTestSuite))
}
