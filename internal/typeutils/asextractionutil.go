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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
)

func extractPreferredUsername(i withPreferredUsername) (string, error) {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return "", errors.New("preferredUsername was not a string")
	}
	if u.GetXMLSchemaString() == "" {
		return "", errors.New("preferredUsername was empty")
	}
	return u.GetXMLSchemaString(), nil
}

func extractName(i withDisplayName) (string, error) {
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

func extractInReplyToURI(i withInReplyTo) (*url.URL, error) {
	inReplyToProp := i.GetActivityStreamsInReplyTo()
	for i := inReplyToProp.Begin(); i != inReplyToProp.End(); i = i.Next() {
		if i.IsIRI() {
			if i.GetIRI() != nil {
				return i.GetIRI(), nil
			}
		}
	}
	return nil, errors.New("couldn't find iri for in reply to")
}

func extractTos(i withTo) ([]*url.URL, error) {
	to := []*url.URL{}
	toProp := i.GetActivityStreamsTo()
	for i := toProp.Begin(); i != toProp.End(); i = i.Next() {
		if i.IsIRI() {
			if i.GetIRI() != nil {
				to = append(to, i.GetIRI())
			}
		}
	}
	if len(to) == 0 {
		return nil, errors.New("found no to entries")
	}
	return to, nil
}

func extractCCs(i withCC) ([]*url.URL, error) {
	cc := []*url.URL{}
	ccProp := i.GetActivityStreamsCc()
	for i := ccProp.Begin(); i != ccProp.End(); i = i.Next() {
		if i.IsIRI() {
			if i.GetIRI() != nil {
				cc = append(cc, i.GetIRI())
			}
		}
	}
	if len(cc) == 0 {
		return nil, errors.New("found no cc entries")
	}
	return cc, nil
}

func extractAttributedTo(i withAttributedTo) (*url.URL, error) {
	attributedToProp := i.GetActivityStreamsAttributedTo()
	for aIter := attributedToProp.Begin(); aIter != attributedToProp.End(); aIter = aIter.Next() {
		if aIter.IsIRI() {
			if aIter.GetIRI() != nil {
				return aIter.GetIRI(), nil
			}
		}
	}
	return nil, errors.New("couldn't find iri for attributed to")
}

func extractPublished(i withPublished) (time.Time, error) {
	publishedProp := i.GetActivityStreamsPublished()
	if publishedProp == nil {
		return time.Time{}, errors.New("published prop was nil")
	}

	if !publishedProp.IsXMLSchemaDateTime() {
		return time.Time{}, errors.New("published prop was not date time")
	}

	t := publishedProp.Get()
	if t.IsZero() {
		return time.Time{}, errors.New("published time was zero")
	}
	return t, nil
}

// extractIconURL extracts a URL to a supported image file from something like:
//   "icon": {
//     "mediaType": "image/jpeg",
//     "type": "Image",
//     "url": "http://example.org/path/to/some/file.jpeg"
//   },
func extractIconURL(i withIcon) (*url.URL, error) {
	iconProp := i.GetActivityStreamsIcon()
	if iconProp == nil {
		return nil, errors.New("icon property was nil")
	}

	// icon can potentially contain multiple entries, so we iterate through all of them
	// here in order to find the first one that meets these criteria:
	// 1. is an image
	// 2. has a URL so we can grab it
	for iconIter := iconProp.Begin(); iconIter != iconProp.End(); iconIter = iconIter.Next() {
		// 1. is an image
		if !iconIter.IsActivityStreamsImage() {
			continue
		}
		imageValue := iconIter.GetActivityStreamsImage()
		if imageValue == nil {
			continue
		}

		// 2. has a URL so we can grab it
		url, err := extractURL(imageValue)
		if err == nil && url != nil {
			return url, nil
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
func extractImageURL(i withImage) (*url.URL, error) {
	imageProp := i.GetActivityStreamsImage()
	if imageProp == nil {
		return nil, errors.New("icon property was nil")
	}

	// icon can potentially contain multiple entries, so we iterate through all of them
	// here in order to find the first one that meets these criteria:
	// 1. is an image
	// 2. has a URL so we can grab it
	for imageIter := imageProp.Begin(); imageIter != imageProp.End(); imageIter = imageIter.Next() {
		// 1. is an image
		if !imageIter.IsActivityStreamsImage() {
			continue
		}
		imageValue := imageIter.GetActivityStreamsImage()
		if imageValue == nil {
			continue
		}

		// 2. has a URL so we can grab it
		url, err := extractURL(imageValue)
		if err == nil && url != nil {
			return url, nil
		}
	}
	// if we get to this point we didn't find an image meeting our criteria :'(
	return nil, errors.New("could not extract valid image from image property")
}

func extractSummary(i withSummary) (string, error) {
	summaryProp := i.GetActivityStreamsSummary()
	if summaryProp == nil {
		return "", errors.New("summary property was nil")
	}

	for summaryIter := summaryProp.Begin(); summaryIter != summaryProp.End(); summaryIter = summaryIter.Next() {
		if summaryIter.IsXMLSchemaString() && summaryIter.GetXMLSchemaString() != "" {
			return summaryIter.GetXMLSchemaString(), nil
		}
	}

	return "", errors.New("could not extract summary")
}

func extractDiscoverable(i withDiscoverable) (bool, error) {
	if i.GetTootDiscoverable() == nil {
		return false, errors.New("discoverable was nil")
	}
	return i.GetTootDiscoverable().Get(), nil
}

func extractURL(i withURL) (*url.URL, error) {
	urlProp := i.GetActivityStreamsUrl()
	if urlProp == nil {
		return nil, errors.New("url property was nil")
	}

	for urlIter := urlProp.Begin(); urlIter != urlProp.End(); urlIter = urlIter.Next() {
		if urlIter.IsIRI() && urlIter.GetIRI() != nil {
			return urlIter.GetIRI(), nil
		}
	}

	return nil, errors.New("could not extract url")
}

func extractPublicKeyForOwner(i withPublicKey, forOwner *url.URL) (*rsa.PublicKey, *url.URL, error) {
	publicKeyProp := i.GetW3IDSecurityV1PublicKey()
	if publicKeyProp == nil {
		return nil, nil, errors.New("public key property was nil")
	}

	for publicKeyIter := publicKeyProp.Begin(); publicKeyIter != publicKeyProp.End(); publicKeyIter = publicKeyIter.Next() {
		pkey := publicKeyIter.Get()
		if pkey == nil {
			continue
		}

		pkeyID, err := pub.GetId(pkey)
		if err != nil || pkeyID == nil {
			continue
		}

		if pkey.GetW3IDSecurityV1Owner() == nil || pkey.GetW3IDSecurityV1Owner().Get() == nil || pkey.GetW3IDSecurityV1Owner().Get().String() != forOwner.String() {
			continue
		}

		if pkey.GetW3IDSecurityV1PublicKeyPem() == nil {
			continue
		}

		pkeyPem := pkey.GetW3IDSecurityV1PublicKeyPem().Get()
		if pkeyPem == "" {
			continue
		}

		block, _ := pem.Decode([]byte(pkeyPem))
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, nil, errors.New("could not decode publicKeyPem to PUBLIC KEY pem block type")
		}

		p, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse public key from block bytes: %s", err)
		}
		if p == nil {
			return nil, nil, errors.New("returned public key was empty")
		}

		if publicKey, ok := p.(*rsa.PublicKey); ok {
			return publicKey, pkeyID, nil
		}
	}
	return nil, nil, errors.New("couldn't find public key")
}
