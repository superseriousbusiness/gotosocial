package typeutils

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (c *converter) ASPersonToAccount(person vocab.ActivityStreamsPerson) (*gtsmodel.Account, error) {
	// first check if we actually already know this person
	uriProp := person.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found on person, or id was not an iri")
	}
	uri := uriProp.GetIRI()

	acct := &gtsmodel.Account{}
	if err := c.db.GetWhere("uri", uri.String(), acct); err == nil {
		// we already know this account so we can skip generating it
		return acct, nil
	} else {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// we don't know the account and there's been a real error
			return nil, fmt.Errorf("error getting account with uri %s from the database: %s", uri.String(), err)
		}
	}

	// we don't know the account so we need to generate it from the person -- at least we already have the URI!
	acct = &gtsmodel.Account{}
	acct.URI = uri.String()

	// Username
	// We need this one so bail if it's not set.
	username, err := extractUsername(person)
	if err != nil {
		return nil, fmt.Errorf("couldn't extract username: %s", err)
	}
	acct.Username = username

	// Domain
	// We need this one as well
	acct.Domain = uri.Host

	// avatar aka icon
	// if this one isn't extractable in a format we recognise we'll just skip it
	if avatarURL, err := extractIconURL(person); err == nil {
		acct.AvatarRemoteURL = avatarURL.String()
	}

	// header aka image
	// if this one isn't extractable in a format we recognise we'll just skip it
	if headerURL, err := extractImageURL(person); err == nil {
		acct.HeaderRemoteURL = headerURL.String()
	}

	return acct, nil
}

type usernameable interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
}

func extractUsername(i usernameable) (string, error) {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return "", errors.New("preferredUsername was not a string")
	}
	if u.GetXMLSchemaString() == "" {
		return "", errors.New("preferredUsername was empty")
	}
	return u.GetXMLSchemaString(), nil
}

type iconable interface {
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
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

type imageable interface {
	GetActivityStreamsImage() vocab.ActivityStreamsImageProperty
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
