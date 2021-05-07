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

package typeutils

import (
	"errors"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

type usernameable interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
}

type iconable interface {
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
}

type displaynameable interface {
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
}

type imageable interface {
	GetActivityStreamsImage() vocab.ActivityStreamsImageProperty
}

func extractPreferredUsername(i usernameable) (string, error) {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return "", errors.New("preferredUsername was not a string")
	}
	if u.GetXMLSchemaString() == "" {
		return "", errors.New("preferredUsername was empty")
	}
	return u.GetXMLSchemaString(), nil
}

func extractName(i displaynameable) (string, error) {
	nameProp := i.GetActivityStreamsName()
	if nameProp == nil {
		return "", errors.New("activityStreamsName not found")
	}

	// take the first name string we can find
	for nameIter := nameProp.Begin(); nameIter != nameProp.End(); nameIter = nameIter.Next() {
		if nameIter.IsXMLSchemaString() && nameIter.GetXMLSchemaString() != "" {
			return nameIter.GetXMLSchemaString(), nil
		}
	}

	return "", errors.New("activityStreamsName not found")
}

// extractIconURL extracts a URL to a supported image file from something like:
//   "icon": {
//     "mediaType": "image/jpeg",
//     "type": "Image",
//     "url": "http://example.org/path/to/some/file.jpeg"
//   },
func extractIconURL(i iconable) (*url.URL, error) {
	iconProp := i.GetActivityStreamsIcon()
	if iconProp == nil {
		return nil, errors.New("icon property was nil")
	}

	// icon can potentially contain multiple entries, so we iterate through all of them
	// here in order to find the first one that meets these criteria:
	// 1. is an image
	// 2. is a supported type
	// 3. has a URL so we can grab it
	for iconIter := iconProp.Begin(); iconIter != iconProp.End(); iconIter = iconIter.Next() {
		// 1. is an image
		if !iconIter.IsActivityStreamsImage() {
			continue
		}
		imageValue := iconIter.GetActivityStreamsImage()
		if imageValue == nil {
			continue
		}

		// 2. is a supported type
		imageType := imageValue.GetActivityStreamsMediaType()
		if imageType == nil || !media.SupportedImageType(imageType.Get()) {
			continue
		}

		// 3. has a URL so we can grab it
		imageURLProp := imageValue.GetActivityStreamsUrl()
		if imageURLProp == nil {
			continue
		}

		// URL is also an iterable!
		// so let's take the first valid one we can find
		for urlIter := imageURLProp.Begin(); urlIter != imageURLProp.End(); urlIter = urlIter.Next() {
			if !urlIter.IsIRI() {
				continue
			}
			if urlIter.GetIRI() == nil {
				continue
			}
			// found it!!!
			return urlIter.GetIRI(), nil
		}
	}
	// if we get to this point we didn't find an icon meeting our criteria :'(
	return nil, errors.New("could not extract valid image from icon")
}


// extractImageURL extracts a URL to a supported image file from something like:
//   "image": {
//     "mediaType": "image/jpeg",
//     "type": "Image",
//     "url": "http://example.org/path/to/some/file.jpeg"
//   },
func extractImageURL(i imageable) (*url.URL, error) {
	imageProp := i.GetActivityStreamsImage()
	if imageProp == nil {
		return nil, errors.New("icon property was nil")
	}

	// icon can potentially contain multiple entries, so we iterate through all of them
	// here in order to find the first one that meets these criteria:
	// 1. is an image
	// 2. is a supported type
	// 3. has a URL so we can grab it
	for imageIter := imageProp.Begin(); imageIter != imageProp.End(); imageIter = imageIter.Next() {
		// 1. is an image
		if !imageIter.IsActivityStreamsImage() {
			continue
		}
		imageValue := imageIter.GetActivityStreamsImage()
		if imageValue == nil {
			continue
		}

		// 2. is a supported type
		imageType := imageValue.GetActivityStreamsMediaType()
		if imageType == nil || !media.SupportedImageType(imageType.Get()) {
			continue
		}

		// 3. has a URL so we can grab it
		imageURLProp := imageValue.GetActivityStreamsUrl()
		if imageURLProp == nil {
			continue
		}

		// URL is also an iterable!
		// so let's take the first valid one we can find
		for urlIter := imageURLProp.Begin(); urlIter != imageURLProp.End(); urlIter = urlIter.Next() {
			if !urlIter.IsIRI() {
				continue
			}
			if urlIter.GetIRI() == nil {
				continue
			}
			// found it!!!
			return urlIter.GetIRI(), nil
		}
	}
	// if we get to this point we didn't find an image meeting our criteria :'(
	return nil, errors.New("could not extract valid image from image property")
}
