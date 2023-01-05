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

// Package ap contains models and utilities for working with activitypub/activitystreams representations.
//
// It is built on top of go-fed/activity.
package ap

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ExtractPreferredUsername returns a string representation of an interface's preferredUsername property.
func ExtractPreferredUsername(i WithPreferredUsername) (string, error) {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return "", errors.New("preferredUsername was not a string")
	}
	if u.GetXMLSchemaString() == "" {
		return "", errors.New("preferredUsername was empty")
	}
	return u.GetXMLSchemaString(), nil
}

// ExtractName returns a string representation of an interface's name property.
func ExtractName(i WithName) (string, error) {
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

// ExtractInReplyToURI extracts the inReplyToURI property (if present) from an interface.
func ExtractInReplyToURI(i WithInReplyTo) *url.URL {
	inReplyToProp := i.GetActivityStreamsInReplyTo()
	if inReplyToProp == nil {
		// the property just wasn't set
		return nil
	}
	for iter := inReplyToProp.Begin(); iter != inReplyToProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				return iter.GetIRI()
			}
		}
	}
	// couldn't find a URI
	return nil
}

// ExtractURLItems extracts a slice of URLs from a property that has withItems.
func ExtractURLItems(i WithItems) []*url.URL {
	urls := []*url.URL{}
	items := i.GetActivityStreamsItems()
	if items == nil || items.Len() == 0 {
		return urls
	}

	for iter := items.Begin(); iter != items.End(); iter = iter.Next() {
		if iter.IsIRI() {
			urls = append(urls, iter.GetIRI())
		}
	}
	return urls
}

// ExtractTos returns a list of URIs that the activity addresses as To.
func ExtractTos(i WithTo) ([]*url.URL, error) {
	to := []*url.URL{}
	toProp := i.GetActivityStreamsTo()
	if toProp == nil {
		return nil, errors.New("toProp was nil")
	}
	for iter := toProp.Begin(); iter != toProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				to = append(to, iter.GetIRI())
			}
		}
	}
	return to, nil
}

// ExtractCCs returns a list of URIs that the activity addresses as CC.
func ExtractCCs(i WithCC) ([]*url.URL, error) {
	cc := []*url.URL{}
	ccProp := i.GetActivityStreamsCc()
	if ccProp == nil {
		return cc, nil
	}
	for iter := ccProp.Begin(); iter != ccProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				cc = append(cc, iter.GetIRI())
			}
		}
	}
	return cc, nil
}

// ExtractAttributedTo returns the URL of the actor that the withAttributedTo is attributed to.
func ExtractAttributedTo(i WithAttributedTo) (*url.URL, error) {
	attributedToProp := i.GetActivityStreamsAttributedTo()
	if attributedToProp == nil {
		return nil, errors.New("attributedToProp was nil")
	}
	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			if iter.GetIRI() != nil {
				return iter.GetIRI(), nil
			}
		}
	}
	return nil, errors.New("couldn't find iri for attributed to")
}

// ExtractPublished extracts the publication time of an activity.
func ExtractPublished(i WithPublished) (time.Time, error) {
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

// ExtractIconURL extracts a URL to a supported image file from something like:
//
//	"icon": {
//	  "mediaType": "image/jpeg",
//	  "type": "Image",
//	  "url": "http://example.org/path/to/some/file.jpeg"
//	},
func ExtractIconURL(i WithIcon) (*url.URL, error) {
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
		url, err := ExtractURL(imageValue)
		if err == nil && url != nil {
			return url, nil
		}
	}
	// if we get to this point we didn't find an icon meeting our criteria :'(
	return nil, errors.New("could not extract valid image from icon")
}

// ExtractImageURL extracts a URL to a supported image file from something like:
//
//	"image": {
//	  "mediaType": "image/jpeg",
//	  "type": "Image",
//	  "url": "http://example.org/path/to/some/file.jpeg"
//	},
func ExtractImageURL(i WithImage) (*url.URL, error) {
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
		url, err := ExtractURL(imageValue)
		if err == nil && url != nil {
			return url, nil
		}
	}
	// if we get to this point we didn't find an image meeting our criteria :'(
	return nil, errors.New("could not extract valid image from image property")
}

// ExtractSummary extracts the summary/content warning of an interface.
func ExtractSummary(i WithSummary) (string, error) {
	summaryProp := i.GetActivityStreamsSummary()
	if summaryProp == nil || summaryProp.Len() == 0 {
		// no summary to speak of
		return "", nil
	}

	for iter := summaryProp.Begin(); iter != summaryProp.End(); iter = iter.Next() {
		switch {
		case iter.IsIRI():
			return iter.GetIRI().String(), nil
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString(), nil
		}
	}

	return "", nil
}

// ExtractDiscoverable extracts the Discoverable boolean of an interface.
func ExtractDiscoverable(i WithDiscoverable) (bool, error) {
	if i.GetTootDiscoverable() == nil {
		return false, errors.New("discoverable was nil")
	}
	return i.GetTootDiscoverable().Get(), nil
}

// ExtractURL extracts the URL property of an interface.
func ExtractURL(i WithURL) (*url.URL, error) {
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

// ExtractPublicKeyForOwner extracts the public key from an interface, as long as it belongs to the specified owner.
// It will return the public key itself, the id/URL of the public key, or an error if something goes wrong.
func ExtractPublicKeyForOwner(i WithPublicKey, forOwner *url.URL) (*rsa.PublicKey, *url.URL, error) {
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
		if block == nil {
			return nil, nil, errors.New("could not decode publicKeyPem: no PEM data")
		}
		var p crypto.PublicKey
		switch block.Type {
		case "PUBLIC KEY":
			p, err = x509.ParsePKIXPublicKey(block.Bytes)
		case "RSA PUBLIC KEY":
			p, err = x509.ParsePKCS1PublicKey(block.Bytes)
		default:
			return nil, nil, fmt.Errorf("could not parse public key: unknown block type: %q", block.Type)
		}
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

// ExtractContent returns a string representation of the interface's Content property,
// or an empty string if no Content is found.
func ExtractContent(i WithContent) string {
	contentProperty := i.GetActivityStreamsContent()
	if contentProperty == nil {
		return ""
	}

	for iter := contentProperty.Begin(); iter != contentProperty.End(); iter = iter.Next() {
		if iter.IsXMLSchemaString() {
			return iter.GetXMLSchemaString()
		}
		if iter.IsIRI() && iter.GetIRI() != nil {
			return iter.GetIRI().String()
		}
	}

	return ""
}

// ExtractAttachments returns a slice of attachments on the interface.
func ExtractAttachments(i WithAttachment) ([]*gtsmodel.MediaAttachment, error) {
	attachments := []*gtsmodel.MediaAttachment{}
	attachmentProp := i.GetActivityStreamsAttachment()
	if attachmentProp == nil {
		return attachments, nil
	}
	for iter := attachmentProp.Begin(); iter != attachmentProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}
		attachmentable, ok := t.(Attachmentable)
		if !ok {
			continue
		}
		attachment, err := ExtractAttachment(attachmentable)
		if err != nil {
			continue
		}
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}

// ExtractAttachment returns a gts model of an attachment from an attachmentable interface.
func ExtractAttachment(i Attachmentable) (*gtsmodel.MediaAttachment, error) {
	attachment := &gtsmodel.MediaAttachment{
		File: gtsmodel.File{},
	}

	attachmentURL, err := ExtractURL(i)
	if err != nil {
		return nil, err
	}
	attachment.RemoteURL = attachmentURL.String()

	mediaType := i.GetActivityStreamsMediaType()
	if mediaType == nil || mediaType.Get() == "" {
		return nil, errors.New("no media type")
	}
	attachment.File.ContentType = mediaType.Get()
	attachment.Type = gtsmodel.FileTypeImage

	name, err := ExtractName(i)
	if err == nil {
		attachment.Description = name
	}

	attachment.Blurhash = ExtractBlurhash(i)

	attachment.Processing = gtsmodel.ProcessingStatusReceived

	return attachment, nil
}

// ExtractBlurhash extracts the blurhash value (if present) from a WithBlurhash interface.
func ExtractBlurhash(i WithBlurhash) string {
	if i.GetTootBlurhash() == nil {
		return ""
	}
	return i.GetTootBlurhash().Get()
}

// ExtractHashtags returns a slice of tags on the interface.
func ExtractHashtags(i WithTag) ([]*gtsmodel.Tag, error) {
	tags := []*gtsmodel.Tag{}
	tagsProp := i.GetActivityStreamsTag()
	if tagsProp == nil {
		return tags, nil
	}
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

		tag, err := ExtractHashtag(hashtaggable)
		if err != nil {
			continue
		}

		tags = append(tags, tag)
	}
	return tags, nil
}

// ExtractHashtag returns a gtsmodel tag from a hashtaggable.
func ExtractHashtag(i Hashtaggable) (*gtsmodel.Tag, error) {
	tag := &gtsmodel.Tag{}

	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, errors.New("no href prop")
	}
	tag.URL = hrefProp.GetIRI().String()

	name, err := ExtractName(i)
	if err != nil {
		return nil, err
	}
	tag.Name = strings.TrimPrefix(name, "#")

	return tag, nil
}

// ExtractEmojis returns a slice of emojis on the interface.
func ExtractEmojis(i WithTag) ([]*gtsmodel.Emoji, error) {
	emojis := []*gtsmodel.Emoji{}
	tagsProp := i.GetActivityStreamsTag()
	if tagsProp == nil {
		return emojis, nil
	}
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

		emoji, err := ExtractEmoji(emojiable)
		if err != nil {
			continue
		}

		emojis = append(emojis, emoji)
	}
	return emojis, nil
}

// ExtractEmoji ...
func ExtractEmoji(i Emojiable) (*gtsmodel.Emoji, error) {
	emoji := &gtsmodel.Emoji{}

	idProp := i.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id for emoji")
	}
	uri := idProp.GetIRI()
	emoji.URI = uri.String()
	emoji.Domain = uri.Host

	name, err := ExtractName(i)
	if err != nil {
		return nil, err
	}
	emoji.Shortcode = strings.Trim(name, ":")

	if i.GetActivityStreamsIcon() == nil {
		return nil, errors.New("no icon for emoji")
	}
	imageURL, err := ExtractIconURL(i)
	if err != nil {
		return nil, errors.New("no url for emoji image")
	}
	emoji.ImageRemoteURL = imageURL.String()

	// assume false for both to begin
	emoji.Disabled = new(bool)
	emoji.VisibleInPicker = new(bool)

	updatedProp := i.GetActivityStreamsUpdated()
	if updatedProp != nil && updatedProp.IsXMLSchemaDateTime() {
		emoji.UpdatedAt = updatedProp.Get()
	}

	return emoji, nil
}

// ExtractMentions extracts a slice of gtsmodel Mentions from a WithTag interface.
func ExtractMentions(i WithTag) ([]*gtsmodel.Mention, error) {
	mentions := []*gtsmodel.Mention{}
	tagsProp := i.GetActivityStreamsTag()
	if tagsProp == nil {
		return mentions, nil
	}
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
			return nil, errors.New("mention was not convertable to ap.Mentionable")
		}

		mention, err := ExtractMention(mentionable)
		if err != nil {
			return nil, err
		}

		mentions = append(mentions, mention)
	}
	return mentions, nil
}

// ExtractMention extracts a gts model mention from a Mentionable.
func ExtractMention(i Mentionable) (*gtsmodel.Mention, error) {
	mention := &gtsmodel.Mention{}

	mentionString, err := ExtractName(i)
	if err != nil {
		return nil, err
	}

	// just make sure the mention string is valid so we can handle it properly later on...
	_, _, err = util.ExtractNamestringParts(mentionString)
	if err != nil {
		return nil, err
	}
	mention.NameString = mentionString

	// the href prop should be the AP URI of a user we know, eg https://example.org/users/whatever_user
	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, errors.New("no href prop")
	}
	mention.TargetAccountURI = hrefProp.GetIRI().String()
	return mention, nil
}

// ExtractActor extracts the actor ID/IRI from an interface WithActor.
func ExtractActor(i WithActor) (*url.URL, error) {
	actorProp := i.GetActivityStreamsActor()
	if actorProp == nil {
		return nil, errors.New("actor property was nil")
	}
	for iter := actorProp.Begin(); iter != actorProp.End(); iter = iter.Next() {
		if iter.IsIRI() && iter.GetIRI() != nil {
			return iter.GetIRI(), nil
		}
	}
	return nil, errors.New("no iri found for actor prop")
}

// ExtractObject extracts a URL object from a WithObject interface.
func ExtractObject(i WithObject) (*url.URL, error) {
	objectProp := i.GetActivityStreamsObject()
	if objectProp == nil {
		return nil, errors.New("object property was nil")
	}
	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		if iter.IsIRI() && iter.GetIRI() != nil {
			return iter.GetIRI(), nil
		}
	}
	return nil, errors.New("no iri found for object prop")
}

// ExtractVisibility extracts the gtsmodel.Visibility of a given addressable with a To and CC property.
//
// ActorFollowersURI is needed to check whether the visibility is FollowersOnly or not. The passed-in value
// should just be the string value representation of the followers URI of the actor who created the activity,
// eg https://example.org/users/whoever/followers.
func ExtractVisibility(addressable Addressable, actorFollowersURI string) (gtsmodel.Visibility, error) {
	to, err := ExtractTos(addressable)
	if err != nil {
		return "", fmt.Errorf("deriveVisibility: error extracting TO values: %s", err)
	}

	cc, err := ExtractCCs(addressable)
	if err != nil {
		return "", fmt.Errorf("deriveVisibility: error extracting CC values: %s", err)
	}

	if len(to) == 0 && len(cc) == 0 {
		return "", errors.New("deriveVisibility: message wasn't TO or CC anyone")
	}

	// for visibility derivation, we start by assuming most restrictive, and work our way to least restrictive
	visibility := gtsmodel.VisibilityDirect

	// if it's got followers in TO and it's not also CC'ed to public, it's followers only
	if isFollowers(to, actorFollowersURI) {
		visibility = gtsmodel.VisibilityFollowersOnly
	}

	// if it's CC'ed to public, it's unlocked
	// mentioned SPECIFIC ACCOUNTS also get added to CC'es if it's not a direct message
	if isPublic(cc) {
		visibility = gtsmodel.VisibilityUnlocked
	}

	// if it's To public, it's just straight up public
	if isPublic(to) {
		visibility = gtsmodel.VisibilityPublic
	}

	return visibility, nil
}

// isPublic checks if at least one entry in the given uris slice equals
// the activitystreams public uri.
func isPublic(uris []*url.URL) bool {
	for _, entry := range uris {
		if strings.EqualFold(entry.String(), pub.PublicActivityPubIRI) {
			return true
		}
	}
	return false
}

// isFollowers checks if at least one entry in the given uris slice equals
// the given followersURI.
func isFollowers(uris []*url.URL, followersURI string) bool {
	for _, entry := range uris {
		if strings.EqualFold(entry.String(), followersURI) {
			return true
		}
	}
	return false
}

// ExtractSensitive extracts whether or not an item is 'sensitive'.
// If no sensitive property is set on the item at all, or if this property
// isn't a boolean, then false will be returned by default.
func ExtractSensitive(withSensitive WithSensitive) bool {
	sensitiveProp := withSensitive.GetActivityStreamsSensitive()
	if sensitiveProp == nil {
		return false
	}

	for iter := sensitiveProp.Begin(); iter != sensitiveProp.End(); iter = iter.Next() {
		if iter.IsXMLSchemaBoolean() {
			return iter.Get()
		}
	}

	return false
}

// ExtractSharedInbox extracts the sharedInbox URI properly from an Actor.
// Returns nil if this property is not set.
func ExtractSharedInbox(withEndpoints WithEndpoints) *url.URL {
	endpointsProp := withEndpoints.GetActivityStreamsEndpoints()
	if endpointsProp == nil {
		return nil
	}

	for iter := endpointsProp.Begin(); iter != endpointsProp.End(); iter = iter.Next() {
		if iter.IsActivityStreamsEndpoints() {
			endpoints := iter.Get()
			if endpoints == nil {
				return nil
			}
			sharedInboxProp := endpoints.GetActivityStreamsSharedInbox()
			if sharedInboxProp == nil {
				return nil
			}

			if !sharedInboxProp.IsIRI() {
				return nil
			}

			return sharedInboxProp.GetIRI()
		}
	}

	return nil
}
