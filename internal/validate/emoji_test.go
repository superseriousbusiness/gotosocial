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

package validate_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func happyEmoji() *gtsmodel.Emoji {
	// the file validator actually runs os.Stat on given paths, so we need to just create small
	// temp files for both the main attachment file and the thumbnail

	imageFile, err := os.CreateTemp("", "gts_test_emoji")
	if err != nil {
		panic(err)
	}
	if _, err := imageFile.WriteString("main"); err != nil {
		panic(err)
	}
	imagePath := imageFile.Name()
	if err := imageFile.Close(); err != nil {
		panic(err)
	}

	staticFile, err := os.CreateTemp("", "gts_test_emoji_static")
	if err != nil {
		panic(err)
	}
	if _, err := staticFile.WriteString("thumbnail"); err != nil {
		panic(err)
	}
	imageStaticPath := staticFile.Name()
	if err := staticFile.Close(); err != nil {
		panic(err)
	}

	return &gtsmodel.Emoji{
		ID:                     "01F8MH6NEM8D7527KZAECTCR76",
		CreatedAt:              time.Now().Add(-71 * time.Hour),
		UpdatedAt:              time.Now().Add(-71 * time.Hour),
		Shortcode:              "blob_test",
		Domain:                 "example.org",
		ImageRemoteURL:         "https://example.org/emojis/blob_test.gif",
		ImageStaticRemoteURL:   "https://example.org/emojis/blob_test.png",
		ImageURL:               "",
		ImageStaticURL:         "",
		ImagePath:              imagePath,
		ImageStaticPath:        imageStaticPath,
		ImageContentType:       "image/gif",
		ImageStaticContentType: "image/png",
		ImageFileSize:          1024,
		ImageStaticFileSize:    256,
		ImageUpdatedAt:         time.Now(),
		Disabled:               testrig.FalseBool(),
		URI:                    "https://example.org/emojis/blob_test",
		VisibleInPicker:        testrig.TrueBool(),
		CategoryID:             "01FEE47ZH70PWDSEAVBRFNX325",
	}
}

type EmojiValidateTestSuite struct {
	suite.Suite
}

func (suite *EmojiValidateTestSuite) TestValidateEmojiHappyPath() {
	// no problem here
	m := happyEmoji()
	err := validate.Struct(*m)
	suite.NoError(err)
}

func (suite *EmojiValidateTestSuite) TestValidateEmojiBadFilePaths() {
	e := happyEmoji()

	e.ImagePath = "/tmp/nonexistent/file/for/gotosocial/test"
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImagePath' Error:Field validation for 'ImagePath' failed on the 'file' tag")

	e.ImagePath = ""
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImagePath' Error:Field validation for 'ImagePath' failed on the 'required' tag")

	e.ImagePath = "???????????thisnot a valid path####"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImagePath' Error:Field validation for 'ImagePath' failed on the 'file' tag")

	e.ImageStaticPath = "/tmp/nonexistent/file/for/gotosocial/test"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImagePath' Error:Field validation for 'ImagePath' failed on the 'file' tag\nKey: 'Emoji.ImageStaticPath' Error:Field validation for 'ImageStaticPath' failed on the 'file' tag")

	e.ImageStaticPath = ""
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImagePath' Error:Field validation for 'ImagePath' failed on the 'file' tag\nKey: 'Emoji.ImageStaticPath' Error:Field validation for 'ImageStaticPath' failed on the 'required' tag")

	e.ImageStaticPath = "???????????thisnot a valid path####"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImagePath' Error:Field validation for 'ImagePath' failed on the 'file' tag\nKey: 'Emoji.ImageStaticPath' Error:Field validation for 'ImageStaticPath' failed on the 'file' tag")
}

func (suite *EmojiValidateTestSuite) TestValidateEmojiURI() {
	e := happyEmoji()

	e.URI = "aaaaaaaaaa"
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.URI' Error:Field validation for 'URI' failed on the 'url' tag")

	e.URI = ""
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.URI' Error:Field validation for 'URI' failed on the 'url' tag")
}

func (suite *EmojiValidateTestSuite) TestValidateEmojiURLCombos() {
	e := happyEmoji()

	e.ImageRemoteURL = ""
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageRemoteURL' Error:Field validation for 'ImageRemoteURL' failed on the 'required_without' tag\nKey: 'Emoji.ImageURL' Error:Field validation for 'ImageURL' failed on the 'required_without' tag")

	e.ImageURL = "https://whatever.org"
	err = validate.Struct(e)
	suite.NoError(err)

	e.ImageStaticRemoteURL = ""
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageStaticRemoteURL' Error:Field validation for 'ImageStaticRemoteURL' failed on the 'required_without' tag\nKey: 'Emoji.ImageStaticURL' Error:Field validation for 'ImageStaticURL' failed on the 'required_without' tag")

	e.ImageStaticURL = "https://whatever.org"
	err = validate.Struct(e)
	suite.NoError(err)

	e.ImageURL = ""
	e.ImageStaticURL = ""
	e.ImageRemoteURL = ""
	e.ImageStaticRemoteURL = ""
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageRemoteURL' Error:Field validation for 'ImageRemoteURL' failed on the 'required_without' tag\nKey: 'Emoji.ImageStaticRemoteURL' Error:Field validation for 'ImageStaticRemoteURL' failed on the 'required_without' tag\nKey: 'Emoji.ImageURL' Error:Field validation for 'ImageURL' failed on the 'required_without' tag\nKey: 'Emoji.ImageStaticURL' Error:Field validation for 'ImageStaticURL' failed on the 'required_without' tag")
}

func (suite *EmojiValidateTestSuite) TestValidateFileSize() {
	e := happyEmoji()

	e.ImageFileSize = 0
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageFileSize' Error:Field validation for 'ImageFileSize' failed on the 'required' tag")

	e.ImageStaticFileSize = 0
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageFileSize' Error:Field validation for 'ImageFileSize' failed on the 'required' tag\nKey: 'Emoji.ImageStaticFileSize' Error:Field validation for 'ImageStaticFileSize' failed on the 'required' tag")

	e.ImageFileSize = -1
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageFileSize' Error:Field validation for 'ImageFileSize' failed on the 'min' tag\nKey: 'Emoji.ImageStaticFileSize' Error:Field validation for 'ImageStaticFileSize' failed on the 'required' tag")

	e.ImageStaticFileSize = -1
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageFileSize' Error:Field validation for 'ImageFileSize' failed on the 'min' tag\nKey: 'Emoji.ImageStaticFileSize' Error:Field validation for 'ImageStaticFileSize' failed on the 'min' tag")
}

func (suite *EmojiValidateTestSuite) TestValidateDomain() {
	e := happyEmoji()

	e.Domain = ""
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.ImageURL' Error:Field validation for 'ImageURL' failed on the 'required_without' tag\nKey: 'Emoji.ImageStaticURL' Error:Field validation for 'ImageStaticURL' failed on the 'required_without' tag")

	e.Domain = "aaaaaaaaa"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'Emoji.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")
}

func TestEmojiValidateTestSuite(t *testing.T) {
	suite.Run(t, new(EmojiValidateTestSuite))
}
