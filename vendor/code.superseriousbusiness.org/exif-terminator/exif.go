/*
   exif-terminator
   Copyright (C) 2022 SuperSeriousBusiness admin@gotosocial.org

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

package terminator

import (
	"strings"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type withEXIF interface {
	Exif() (rootIfd *exif.Ifd, data []byte, err error)
	SetExif(ib *exif.IfdBuilder) (err error)
}

func terminateEXIF(data withEXIF) error {
	ifd, _, err := data.Exif()
	if err != nil {
		if strings.Contains(err.Error(), "no exif data") {
			err = nil
		}
		return err
	}

	ifdb := exif.NewIfdBuilderFromExistingChain(ifd)
	orientation, _ := ifdb.FindTagWithName("Orientation")

	im, ti := exifcommon.NewIfdMapping(), exif.NewTagIndex()
	ifdb = exif.NewIfdBuilder(im, ti, ifd.IfdIdentity(), ifd.ByteOrder())

	if orientation != nil {
		ifdb.Add(orientation)
	}

	return data.SetExif(ifdb)
}
