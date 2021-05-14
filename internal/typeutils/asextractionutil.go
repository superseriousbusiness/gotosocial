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
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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

func extractName(i withName) (string, error) {
	nameProp := i.GetActivityStreamsName()
	if nameProp == nil {
		return "", errors.New("activityStreamsName not found")
	}

	// take the first name string we can find
	for iter := nameProp.Begin(); iter != nameProp.End(); iter = iter.Next() {
		if iter.IsXMLSchemaString() && iter.GetXMLSchemaString() != "" {
			return iter.GetXMLSchemaString(), nil
		}
	}

	return "", errors.New("activityStreamsName not found")
}

func extractInReplyToURI(i withInReplyTo) (*url.URL, error) {
	inReplyToProp := i.GetActivityStreamsInReplyTo()
	for iter := inReplyToProp.Begin(); iter != inReplyToProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				return iter.GetIRI(), nil
			}
		}
	}
	return nil, errors.New("couldn't find iri for in reply to")
}

func extractTos(i withTo) ([]*url.URL, error) {
	to := []*url.URL{}
	toProp := i.GetActivityStreamsTo()
	for iter := toProp.Begin(); iter != toProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				to = append(to, iter.GetIRI())
			}
		}
	}
	return to, nil
}

func extractCCs(i withCC) ([]*url.URL, error) {
	cc := []*url.URL{}
	ccProp := i.GetActivityStreamsCc()
	for iter := ccProp.Begin(); iter != ccProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				cc = append(cc, iter.GetIRI())
			}
		}
	}
	return cc, nil
}

func extractAttributedTo(i withAttributedTo) (*url.URL, error) {
	attributedToProp := i.GetActivityStreamsAttributedTo()
	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				return iter.GetIRI(), nil
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
	for iter := iconProp.Begin(); iter != iconProp.End(); iter = iter.Next() {
		// 1. is an image
		if !iter.IsActivityStreamsImage() {
			continue
		}
		imageValue := iter.GetActivityStreamsImage()
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
	for iter := imageProp.Begin(); iter != imageProp.End(); iter = iter.Next() {
		// 1. is an image
		if !iter.IsActivityStreamsImage() {
			continue
		}
		imageValue := iter.GetActivityStreamsImage()
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

	for iter := summaryProp.Begin(); iter != summaryProp.End(); iter = iter.Next() {
		if iter.IsXMLSchemaString() && iter.GetXMLSchemaString() != "" {
			return iter.GetXMLSchemaString(), nil
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

	for iter := urlProp.Begin(); iter != urlProp.End(); iter = iter.Next() {
		if iter.IsIRI() && iter.GetIRI() != nil {
			return iter.GetIRI(), nil
		}
	}

	return nil, errors.New("could not extract url")
}

func extractPublicKeyForOwner(i withPublicKey, forOwner *url.URL) (*rsa.PublicKey, *url.URL, error) {
	publicKeyProp := i.GetW3IDSecurityV1PublicKey()
	if publicKeyProp == nil {
		return nil, nil, errors.New("public key property was nil")
	}

	for iter := publicKeyProp.Begin(); iter != publicKeyProp.End(); iter = iter.Next() {
		pkey := iter.Get()
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

func extractContent(i withContent) (string, error) {
	contentProperty := i.GetActivityStreamsContent()
	if contentProperty == nil {
		return "", errors.New("content property was nil")
	}
	for iter := contentProperty.Begin(); iter != contentProperty.End(); iter = iter.Next() {
		if iter.IsXMLSchemaString() && iter.GetXMLSchemaString() != "" {
			return iter.GetXMLSchemaString(), nil
		}
	}
	return "", errors.New("no content found")
}

func extractAttachments(i withAttachment) ([]*gtsmodel.MediaAttachment, error) {
	attachments := []*gtsmodel.MediaAttachment{}

	attachmentProp := i.GetActivityStreamsAttachment()
	for iter := attachmentProp.Begin(); iter != attachmentProp.End(); iter = iter.Next() {
		attachmentable, ok := iter.(Attachmentable)
		if !ok {
			continue
		}
		attachment, err := extractAttachment(attachmentable)
		if err != nil {
			continue
		}
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}

func extractAttachment(i Attachmentable) (*gtsmodel.MediaAttachment, error) {
	attachment := &gtsmodel.MediaAttachment{
		File: gtsmodel.File{},
	}

	attachmentURL, err := extractURL(i)
	if err != nil {
		return nil, err
	}
	attachment.RemoteURL = attachmentURL.String()

	mediaType := i.GetActivityStreamsMediaType()
	if mediaType == nil {
		return nil, errors.New("no media type")
	}
	if mediaType.Get() == "" {
		return nil, errors.New("no media type")
	}
	attachment.File.ContentType = mediaType.Get()
	attachment.Type = gtsmodel.FileTypeImage

	name, err := extractName(i)
	if err == nil {
		attachment.Description = name
	}

	blurhash, err := extractBlurhash(i)
	if err == nil {
		attachment.Blurhash = blurhash
	}

	return attachment, nil
}

func extractBlurhash(i withBlurhash) (string, error) {
	if i.GetTootBlurhashProperty() == nil {
		return "", errors.New("blurhash property was nil")
	}
	if i.GetTootBlurhashProperty().Get() == "" {
		return "", errors.New("empty blurhash string")
	}
	return i.GetTootBlurhashProperty().Get(), nil
}

func extractHashtags(i withTag) ([]*gtsmodel.Tag, error) {
	tags := []*gtsmodel.Tag{}

	tagsProp := i.GetActivityStreamsTag()
	for iter := tagsProp.Begin(); iter != tagsProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}

		if t.GetTypeName() != "Hashtag" {
			continue
		}

		hashtaggable, ok := t.(Hashtaggable)
		if !ok {
			continue
		}

		tag, err := extractHashtag(hashtaggable)
		if err != nil {
			continue
		}

		tags = append(tags, tag)
	}
	return tags, nil
}

func extractHashtag(i Hashtaggable) (*gtsmodel.Tag, error) {
	tag := &gtsmodel.Tag{}

	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, errors.New("no href prop")
	}
	tag.URL = hrefProp.GetIRI().String()

	name, err := extractName(i)
	if err != nil {
		return nil, err
	}
	tag.Name = strings.TrimPrefix(name, "#")

	return tag, nil
}

func extractEmojis(i withTag) ([]*gtsmodel.Emoji, error) {
	emojis := []*gtsmodel.Emoji{}
	tagsProp := i.GetActivityStreamsTag()
	for iter := tagsProp.Begin(); iter != tagsProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}

		if t.GetTypeName() != "Emoji" {
			continue
		}

		emojiable, ok := t.(Emojiable)
		if !ok {
			continue
		}

		emoji, err := extractEmoji(emojiable)
		if err != nil {
			continue
		}

		emojis = append(emojis, emoji)
	}
	return emojis, nil
}

func extractEmoji(i Emojiable) (*gtsmodel.Emoji, error) {
	emoji := &gtsmodel.Emoji{}

	idProp := i.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id for emoji")
	}
	uri := idProp.GetIRI()
	emoji.URI = uri.String()
	emoji.Domain = uri.Host

	name, err := extractName(i)
	if err != nil {
		return nil, err
	}
	emoji.Shortcode = strings.Trim(name, ":")

	if i.GetActivityStreamsIcon() == nil {
		return nil, errors.New("no icon for emoji")
	}
	imageURL, err := extractIconURL(i)
	if err != nil {
		return nil, errors.New("no url for emoji image")
	}
	emoji.ImageRemoteURL = imageURL.String()

	return emoji, nil
}

func extractMentions(i withTag) ([]*gtsmodel.Mention, error) {
	mentions := []*gtsmodel.Mention{}
	tagsProp := i.GetActivityStreamsTag()
	for iter := tagsProp.Begin(); iter != tagsProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}

		if t.GetTypeName() != "Mention" {
			continue
		}

		mentionable, ok := t.(Mentionable)
		if !ok {
			continue
		}

		mention, err := extractMention(mentionable)
		if err != nil {
			continue
		}

		mentions = append(mentions, mention)
	}
	return mentions, nil
}

func extractMention(i Mentionable) (*gtsmodel.Mention, error) {
	mention := &gtsmodel.Mention{}

	mentionString, err := extractName(i)
	if err != nil {
		return nil, err
	}

	// just make sure the mention string is valid so we can handle it properly later on...
	username, domain, err := util.ExtractMentionParts(mentionString)
	if err != nil {
		return nil, err
	}
	if username == "" || domain == "" {
		return nil, errors.New("username or domain was empty")
	}
	mention.NameString = mentionString

	// the href prop should be the AP URI of a user we know, eg https://example.org/users/whatever_user
	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, errors.New("no href prop")
	}
	mention.MentionedAccountURI = hrefProp.GetIRI().String()
	return mention, nil
}
