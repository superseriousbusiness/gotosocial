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
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MediaUtilTestSuite struct {
	suite.Suite
	log *logrus.Logger
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *MediaUtilTestSuite) SetupSuite() {
	// some of our subsequent entities need a log so create this here
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log
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
	assert.Nil(suite.T(), err)
	ct, err := parseContentType(f)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "image/jpeg", ct)
}

func (suite *MediaUtilTestSuite) TestParseContentTypeNotOK() {
	f, err := ioutil.ReadFile("./test/test-corrupted.jpg")
	assert.Nil(suite.T(), err)
	ct, err := parseContentType(f)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "", ct)
	assert.Equal(suite.T(), "filetype unknown", err.Error())
}

func (suite *MediaUtilTestSuite) TestRemoveEXIF() {
	// load and validate image
	b, err := ioutil.ReadFile("./test/test-with-exif.jpg")
	assert.Nil(suite.T(), err)

	// clean it up and validate the clean version
	clean, err := purgeExif(b)
	assert.Nil(suite.T(), err)

	// compare it to our stored sample
	sampleBytes, err := ioutil.ReadFile("./test/test-without-exif.jpg")
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), sampleBytes, clean)
}

func (suite *MediaUtilTestSuite) TestDeriveImageFromJPEG() {
	// load image
	b, err := ioutil.ReadFile("./test/test-jpeg.jpg")
	assert.Nil(suite.T(), err)

	// clean it up and validate the clean version
	imageAndMeta, err := deriveImage(b, "image/jpeg")
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), 1920, imageAndMeta.width)
	assert.Equal(suite.T(), 1080, imageAndMeta.height)
	assert.Equal(suite.T(), 1.7777777777777777, imageAndMeta.aspect)
	assert.Equal(suite.T(), 2073600, imageAndMeta.size)
	assert.Equal(suite.T(), "LjCZnlvyRkRn_NvzRjWF?urqV@f9", imageAndMeta.blurhash)

	// assert that the final image is what we would expect
	sampleBytes, err := ioutil.ReadFile("./test/test-jpeg-processed.jpg")
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), sampleBytes, imageAndMeta.image)
}

func (suite *MediaUtilTestSuite) TestDeriveThumbnailFromJPEG() {
	// load image
	b, err := ioutil.ReadFile("./test/test-jpeg.jpg")
	assert.Nil(suite.T(), err)

	// clean it up and validate the clean version
	imageAndMeta, err := deriveThumbnail(b, "image/jpeg")
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), 256, imageAndMeta.width)
	assert.Equal(suite.T(), 144, imageAndMeta.height)
	assert.Equal(suite.T(), 1.7777777777777777, imageAndMeta.aspect)
	assert.Equal(suite.T(), 36864, imageAndMeta.size)

	sampleBytes, err := ioutil.ReadFile("./test/test-jpeg-thumbnail.jpg")
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), sampleBytes, imageAndMeta.image)
}

func (suite *MediaUtilTestSuite) TestSupportedImageTypes() {
	ok := supportedImageType("image/jpeg")
	assert.True(suite.T(), ok)

	ok = supportedImageType("image/bmp")
	assert.False(suite.T(), ok)
}

func TestMediaUtilTestSuite(t *testing.T) {
	suite.Run(t, new(MediaUtilTestSuite))
}
