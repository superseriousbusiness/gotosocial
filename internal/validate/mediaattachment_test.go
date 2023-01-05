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

func happyMediaAttachment() *gtsmodel.MediaAttachment {
	// the file validator actually runs os.Stat on given paths, so we need to just create small
	// temp files for both the main attachment file and the thumbnail

	mainFile, err := os.CreateTemp("", "gts_test_mainfile")
	if err != nil {
		panic(err)
	}
	if _, err := mainFile.WriteString("main"); err != nil {
		panic(err)
	}
	mainPath := mainFile.Name()
	if err := mainFile.Close(); err != nil {
		panic(err)
	}

	thumbnailFile, err := os.CreateTemp("", "gts_test_thumbnail")
	if err != nil {
		panic(err)
	}
	if _, err := thumbnailFile.WriteString("thumbnail"); err != nil {
		panic(err)
	}
	thumbnailPath := thumbnailFile.Name()
	if err := thumbnailFile.Close(); err != nil {
		panic(err)
	}

	return &gtsmodel.MediaAttachment{
		ID:        "01F8MH6NEM8D7527KZAECTCR76",
		CreatedAt: time.Now().Add(-71 * time.Hour),
		UpdatedAt: time.Now().Add(-71 * time.Hour),
		StatusID:  "01F8MH75CBF9JFX4ZAD54N0W0R",
		URL:       "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg",
		RemoteURL: "",
		Type:      gtsmodel.FileTypeImage,
		FileMeta: gtsmodel.FileMeta{
			Original: gtsmodel.Original{
				Width:  1200,
				Height: 630,
				Size:   756000,
				Aspect: 1.9047619047619047,
			},
			Small: gtsmodel.Small{
				Width:  256,
				Height: 134,
				Size:   34304,
				Aspect: 1.9104477611940298,
			},
		},
		AccountID:         "01F8MH17FWEB39HZJ76B6VXSKF",
		Description:       "Black and white image of some 50's style text saying: Welcome On Board",
		ScheduledStatusID: "",
		Blurhash:          "LNJRdVM{00Rj%Mayt7j[4nWBofRj",
		Processing:        2,
		File: gtsmodel.File{
			Path:        mainPath,
			ContentType: "image/jpeg",
			FileSize:    62529,
			UpdatedAt:   time.Now().Add(-71 * time.Hour),
		},
		Thumbnail: gtsmodel.Thumbnail{
			Path:        thumbnailPath,
			ContentType: "image/jpeg",
			FileSize:    6872,
			UpdatedAt:   time.Now().Add(-71 * time.Hour),
			URL:         "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpeg",
			RemoteURL:   "",
		},
		Avatar: testrig.FalseBool(),
		Header: testrig.FalseBool(),
	}
}

type MediaAttachmentValidateTestSuite struct {
	suite.Suite
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentHappyPath() {
	// no problem here
	m := happyMediaAttachment()
	err := validate.Struct(m)
	suite.NoError(err)
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentBadFilePaths() {
	m := happyMediaAttachment()

	m.File.Path = "/tmp/nonexistent/file/for/gotosocial/test"
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.File.Path' Error:Field validation for 'Path' failed on the 'file' tag")

	m.File.Path = ""
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.File.Path' Error:Field validation for 'Path' failed on the 'required' tag")

	m.File.Path = "???????????thisnot a valid path####"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.File.Path' Error:Field validation for 'Path' failed on the 'file' tag")

	m.Thumbnail.Path = "/tmp/nonexistent/file/for/gotosocial/test"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.File.Path' Error:Field validation for 'Path' failed on the 'file' tag\nKey: 'MediaAttachment.Thumbnail.Path' Error:Field validation for 'Path' failed on the 'file' tag")

	m.Thumbnail.Path = ""
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.File.Path' Error:Field validation for 'Path' failed on the 'file' tag\nKey: 'MediaAttachment.Thumbnail.Path' Error:Field validation for 'Path' failed on the 'required' tag")

	m.Thumbnail.Path = "???????????thisnot a valid path####"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.File.Path' Error:Field validation for 'Path' failed on the 'file' tag\nKey: 'MediaAttachment.Thumbnail.Path' Error:Field validation for 'Path' failed on the 'file' tag")
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentBadType() {
	m := happyMediaAttachment()

	m.Type = ""
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.Type' Error:Field validation for 'Type' failed on the 'oneof' tag")

	m.Type = "Not Supported"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.Type' Error:Field validation for 'Type' failed on the 'oneof' tag")
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentBadFileMeta() {
	m := happyMediaAttachment()

	m.FileMeta.Original.Aspect = 0
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.FileMeta.Original.Aspect' Error:Field validation for 'Aspect' failed on the 'required_with' tag")

	m.FileMeta.Original.Height = 0
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.FileMeta.Original.Height' Error:Field validation for 'Height' failed on the 'required_with' tag\nKey: 'MediaAttachment.FileMeta.Original.Aspect' Error:Field validation for 'Aspect' failed on the 'required_with' tag")

	m.FileMeta.Original = gtsmodel.Original{}
	err = validate.Struct(m)
	suite.NoError(err)

	m.FileMeta.Focus.X = 3.6
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.FileMeta.Focus.X' Error:Field validation for 'X' failed on the 'max' tag")

	m.FileMeta.Focus.Y = -50
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.FileMeta.Focus.X' Error:Field validation for 'X' failed on the 'max' tag\nKey: 'MediaAttachment.FileMeta.Focus.Y' Error:Field validation for 'Y' failed on the 'min' tag")
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentBadURLCombos() {
	m := happyMediaAttachment()

	m.URL = "aaaaaaaaaa"
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.URL' Error:Field validation for 'URL' failed on the 'url' tag")

	m.URL = ""
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.URL' Error:Field validation for 'URL' failed on the 'required_without' tag\nKey: 'MediaAttachment.RemoteURL' Error:Field validation for 'RemoteURL' failed on the 'required_without' tag")

	m.RemoteURL = "oooooooooo"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.RemoteURL' Error:Field validation for 'RemoteURL' failed on the 'url' tag")

	m.RemoteURL = "https://a-valid-url.gay"
	err = validate.Struct(m)
	suite.NoError(err)
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentBlurhash() {
	m := happyMediaAttachment()

	m.Blurhash = ""
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.Blurhash' Error:Field validation for 'Blurhash' failed on the 'required_if' tag")

	m.Type = gtsmodel.FileTypeAudio
	err = validate.Struct(m)
	suite.NoError(err)

	m.Blurhash = "some_blurhash"
	err = validate.Struct(m)
	suite.NoError(err)
}

func (suite *MediaAttachmentValidateTestSuite) TestValidateMediaAttachmentProcessing() {
	m := happyMediaAttachment()

	m.Processing = 420
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.Processing' Error:Field validation for 'Processing' failed on the 'oneof' tag")

	m.Processing = -5
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'MediaAttachment.Processing' Error:Field validation for 'Processing' failed on the 'oneof' tag")
}

func TestMediaAttachmentValidateTestSuite(t *testing.T) {
	suite.Run(t, new(MediaAttachmentValidateTestSuite))
}
