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

package typeutils

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

const defaultHeaderPath = "/assets/default_header.webp"

// populateDefaultAvatars returns a slice of standard avatars found
// in the path [web-assets-base-dir]/default_avatars. The slice
// entries correspond to the relative url via which they can be
// retrieved from the server.
//
// So for example, an avatar called default.jpeg would be returned
// in the slice as "/assets/default_avatars/default.jpeg".
func populateDefaultAvatars() (defaultAvatars []string) {
	webAssetsAbsFilePath, err := filepath.Abs(config.GetWebAssetBaseDir())
	if err != nil {
		log.Panicf(nil, "error getting abs path for web assets: %s", err)
	}

	defaultAvatarsAbsFilePath := filepath.Join(webAssetsAbsFilePath, "default_avatars")
	defaultAvatarFiles, err := os.ReadDir(defaultAvatarsAbsFilePath)
	if err != nil {
		log.Warnf(nil, "error reading default avatars at %s: %s", defaultAvatarsAbsFilePath, err)
		return
	}

	for _, f := range defaultAvatarFiles {
		// ignore directories
		if f.IsDir() {
			continue
		}

		// ignore files bigger than 50kb
		if i, err := f.Info(); err != nil || i.Size() > 50000 {
			continue
		}

		// get the name of the file, eg avatar.jpeg
		fileName := f.Name()

		// get just the .jpeg, for example, from avatar.jpeg
		extensionWithDot := filepath.Ext(fileName)

		// remove the leading . to just get, eg, jpeg
		extension := strings.TrimPrefix(extensionWithDot, ".")

		// take only files with simple extensions
		// that we know will work OK as avatars
		switch strings.ToLower(extension) {
		case "jpeg", "jpg", "gif", "png", "webp":
			avatarURL := config.GetProtocol() + "://" + config.GetHost() + "/assets/default_avatars/" + fileName
			defaultAvatars = append(defaultAvatars, avatarURL)
		default:
			continue
		}
	}

	return
}

// ensureAvatar ensures that the given account has a value set
// for the avatar URL.
//
// If no value is set, an avatar will be selected at random from
// the available default avatars. This selection is 'sticky', so
// the same account will get the same result on subsequent calls.
//
// If a value for the avatar URL is already set, this function is
// a no-op.
//
// If there are no default avatars available, this function is a
// no-op.
func (c *Converter) ensureAvatar(account *apimodel.Account) {
	if (account.Avatar != "" && account.AvatarStatic != "") || len(c.defaultAvatars) == 0 {
		return
	}

	var avatar string
	if avatarI, ok := c.randAvatars.Load(account.ID); ok {
		// we already have a default avatar stored for this account
		avatar, ok = avatarI.(string)
		if !ok {
			panic("avatarI was not a string")
		}
	} else {
		// select + store a default avatar for this account at random
		randomIndex := rand.Intn(len(c.defaultAvatars)) //nolint:gosec
		avatar = c.defaultAvatars[randomIndex]
		c.randAvatars.Store(account.ID, avatar)
	}

	account.Avatar = avatar
	account.AvatarStatic = avatar

	const defaultAviDesc = "Grayed-out line drawing of a cute sloth (default avatar)."
	account.AvatarDescription = defaultAviDesc
}

// ensureHeader ensures that the given account has a value set
// for the header URL.
//
// If no value is set, the default header will be set.
//
// If a value for the header URL is already set, this function is
// a no-op.
func (c *Converter) ensureHeader(account *apimodel.Account) {
	if account.Header != "" && account.HeaderStatic != "" {
		return
	}

	h := config.GetProtocol() + "://" + config.GetHost() + defaultHeaderPath
	account.Header = h
	account.HeaderStatic = h

	const defaultHeaderDesc = "Flat gray background (default header)."
	account.HeaderDescription = defaultHeaderDesc
}
