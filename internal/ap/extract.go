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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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

// ExtractName returns the first string representation of an interface's name property,
// or an empty string if this is not found.
func ExtractName(i WithName) string {
	nameProp := i.GetActivityStreamsName()
	if nameProp == nil {
		return ""
	}

	// Take the first useful value for the name string we can find.
	for iter := nameProp.Begin(); iter != nameProp.End(); iter = iter.Next() {
		switch {
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString()
		case iter.IsIRI():
			return iter.GetIRI().String()
		}
	}

	return ""
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

// ExtractAttributedTo returns the URL of the actor
// that the withAttributedTo is attributed to.
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
// Will return an empty string if no summary was present.
func ExtractSummary(i WithSummary) string {
	summaryProp := i.GetActivityStreamsSummary()
	if summaryProp == nil || summaryProp.Len() == 0 {
		// no summary to speak of
		return ""
	}

	for iter := summaryProp.Begin(); iter != summaryProp.End(); iter = iter.Next() {
		switch {
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString()
		case iter.IsIRI():
			return iter.GetIRI().String()
		}
	}

	return ""
}

func ExtractFields(i WithAttachment) []*gtsmodel.Field {
	attachmentProp := i.GetActivityStreamsAttachment()
	if attachmentProp == nil {
		// Nothing to do.
		return nil
	}

	l := attachmentProp.Len()
	if l == 0 {
		// Nothing to do.
		return nil
	}

	fields := make([]*gtsmodel.Field, 0, l)
	for iter := attachmentProp.Begin(); iter != attachmentProp.End(); iter = iter.Next() {
		if !iter.IsSchemaPropertyValue() {
			continue
		}

		propertyValue := iter.GetSchemaPropertyValue()
		if propertyValue == nil {
			continue
		}

		nameProp := propertyValue.GetActivityStreamsName()
		if nameProp == nil || nameProp.Len() != 1 {
			continue
		}

		name := nameProp.At(0).GetXMLSchemaString()
		if name == "" {
			continue
		}

		valueProp := propertyValue.GetSchemaValue()
		if valueProp == nil || !valueProp.IsXMLSchemaString() {
			continue
		}

		value := valueProp.Get()
		if value == "" {
			continue
		}

		fields = append(fields, &gtsmodel.Field{
			Name:  name,
			Value: value,
		})
	}

	return fields
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

// ExtractPublicKey extracts the public key, public key ID, and public
// key owner ID from an interface, or an error if something goes wrong.
func ExtractPublicKey(i WithPublicKey) (
	*rsa.PublicKey, // pubkey
	*url.URL, // pubkey ID
	*url.URL, // pubkey owner
	error,
) {
	pubKeyProp := i.GetW3IDSecurityV1PublicKey()
	if pubKeyProp == nil {
		return nil, nil, nil, gtserror.New("public key property was nil")
	}

	for iter := pubKeyProp.Begin(); iter != pubKeyProp.End(); iter = iter.Next() {
		if !iter.IsW3IDSecurityV1PublicKey() {
			continue
		}

		pkey := iter.Get()
		if pkey == nil {
			continue
		}

		pubKeyID, err := pub.GetId(pkey)
		if err != nil {
			continue
		}

		pubKeyOwnerProp := pkey.GetW3IDSecurityV1Owner()
		if pubKeyOwnerProp == nil {
			continue
		}

		pubKeyOwner := pubKeyOwnerProp.GetIRI()
		if pubKeyOwner == nil {
			continue
		}

		pubKeyPemProp := pkey.GetW3IDSecurityV1PublicKeyPem()
		if pubKeyPemProp == nil {
			continue
		}

		pkeyPem := pubKeyPemProp.Get()
		if pkeyPem == "" {
			continue
		}

		block, _ := pem.Decode([]byte(pkeyPem))
		if block == nil {
			continue
		}

		var p crypto.PublicKey
		switch block.Type {
		case "PUBLIC KEY":
			p, err = x509.ParsePKIXPublicKey(block.Bytes)
		case "RSA PUBLIC KEY":
			p, err = x509.ParsePKCS1PublicKey(block.Bytes)
		default:
			err = fmt.Errorf("unknown block type: %q", block.Type)
		}
		if err != nil {
			err = gtserror.Newf("could not parse public key from block bytes: %w", err)
			return nil, nil, nil, err
		}

		if p == nil {
			return nil, nil, nil, gtserror.New("returned public key was empty")
		}

		pubKey, ok := p.(*rsa.PublicKey)
		if !ok {
			continue
		}

		return pubKey, pubKeyID, pubKeyOwner, nil
	}

	return nil, nil, nil, gtserror.New("couldn't find public key")
}

// ExtractContent returns a string representation of the interface's Content property,
// or an empty string if no Content is found.
func ExtractContent(i WithContent) string {
	contentProperty := i.GetActivityStreamsContent()
	if contentProperty == nil {
		return ""
	}

	for iter := contentProperty.Begin(); iter != contentProperty.End(); iter = iter.Next() {
		switch {
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString()
		case iter.IsIRI():
			return iter.GetIRI().String()
		}
	}

	return ""
}

// ExtractAttachment returns a gts model of an attachment from an attachmentable interface.
func ExtractAttachment(i Attachmentable) (*gtsmodel.MediaAttachment, error) {
	attachment := &gtsmodel.MediaAttachment{}

	attachmentURL, err := ExtractURL(i)
	if err != nil {
		return nil, err
	}
	attachment.RemoteURL = attachmentURL.String()

	mediaType := i.GetActivityStreamsMediaType()
	if mediaType != nil {
		attachment.File.ContentType = mediaType.Get()
	}
	attachment.Type = gtsmodel.FileTypeImage

	attachment.Description = ExtractName(i)
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

// ExtractHashtag returns a gtsmodel.Tag from a hashtaggable.
func ExtractHashtag(i Hashtaggable) (*gtsmodel.Tag, error) {
	tag := &gtsmodel.Tag{}

	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, errors.New("no href prop")
	}
	tag.URL = hrefProp.GetIRI().String()

	name := ExtractName(i)
	if name == "" {
		return nil, errors.New("name prop empty")
	}

	tag.Name = strings.TrimPrefix(name, "#")

	return tag, nil
}

// ExtractEmojis extracts a slice of gtsmodel.Emojis
// from a WithTag. If an entry in the WithTag is not an emoji,
// it will be quietly ignored.
func ExtractEmojis(i WithTag) ([]*gtsmodel.Emoji, error) {
	tagsProp := i.GetActivityStreamsTag()
	if tagsProp == nil {
		return nil, nil
	}

	var (
		l      = tagsProp.Len()
		emojis = make([]*gtsmodel.Emoji, 0, l)
		keys   = make(map[string]any, l) // Use map to dedupe items.
	)

	for iter := tagsProp.Begin(); iter != tagsProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil || t.GetTypeName() != ObjectEmoji {
			// Not an emoji we can work with.
			continue
		}

		emojiable, ok := t.(Emojiable)
		if !ok {
			continue
		}

		emoji, err := ExtractEmoji(emojiable)
		if err != nil {
			return nil, err
		}

		// Only append this emoji if we haven't
		// seen it already, to avoid duplicates
		// in the slice.
		if _, set := keys[emoji.URI]; !set {
			keys[emoji.URI] = nil // Value doesn't matter.
			emojis = append(emojis, emoji)
		}
	}

	return emojis, nil
}

// ExtractEmoji extracts a minimal gtsmodel.Emoji from an Emojiable.
func ExtractEmoji(i Emojiable) (*gtsmodel.Emoji, error) {
	// Use AP ID as emoji URI.
	idProp := i.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id for emoji")
	}
	uri := idProp.GetIRI()

	// Extract emoji last updated time (optional).
	var updatedAt time.Time
	updatedProp := i.GetActivityStreamsUpdated()
	if updatedProp != nil && updatedProp.IsXMLSchemaDateTime() {
		updatedAt = updatedProp.Get()
	}

	// Extract emoji name aka shortcode.
	name := ExtractName(i)
	if name == "" {
		return nil, errors.New("name prop empty")
	}
	shortcode := strings.Trim(name, ":")

	// Extract emoji image URL from Icon property.
	imageRemoteURL, err := ExtractIconURL(i)
	if err != nil {
		return nil, errors.New("no url for emoji image")
	}
	imageRemoteURLStr := imageRemoteURL.String()

	return &gtsmodel.Emoji{
		UpdatedAt:       updatedAt,
		Shortcode:       shortcode,
		Domain:          uri.Host,
		ImageRemoteURL:  imageRemoteURLStr,
		URI:             uri.String(),
		Disabled:        new(bool), // Assume false by default.
		VisibleInPicker: new(bool), // Assume false by default.
	}, nil
}

// ExtractMentions extracts a slice of gtsmodel.Mentions
// from a WithTag. If an entry in the WithTag is not a mention,
// it will be quietly ignored.
func ExtractMentions(i WithTag) ([]*gtsmodel.Mention, error) {
	tagsProp := i.GetActivityStreamsTag()
	if tagsProp == nil {
		return nil, nil
	}

	var (
		l        = tagsProp.Len()
		mentions = make([]*gtsmodel.Mention, 0, l)
		keys     = make(map[string]any, l) // Use map to dedupe items.
	)

	for iter := tagsProp.Begin(); iter != tagsProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil || t.GetTypeName() != ObjectMention {
			// Not a mention we can work with.
			continue
		}

		mentionable, ok := t.(Mentionable)
		if !ok {
			continue
		}

		mention, err := ExtractMention(mentionable)
		if err != nil {
			return nil, err
		}

		// Only append this mention if we haven't
		// seen it already, to avoid duplicates
		// in the slice.
		if _, set := keys[mention.TargetAccountURI]; !set {
			keys[mention.TargetAccountURI] = nil // Value doesn't matter.
			mentions = append(mentions, mention)
		}
	}

	return mentions, nil
}

// ExtractMention extracts a minimal gtsmodel.Mention from a Mentionable.
func ExtractMention(i Mentionable) (*gtsmodel.Mention, error) {
	nameString := ExtractName(i)
	if nameString == "" {
		return nil, errors.New("name prop empty")
	}

	// Ensure namestring is valid so we
	// can handle it properly later on.
	if _, _, err := util.ExtractNamestringParts(nameString); err != nil {
		return nil, err
	}

	// The href prop should be the AP URI
	// of the target account.
	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, errors.New("no href prop")
	}

	return &gtsmodel.Mention{
		NameString:       nameString,
		TargetAccountURI: hrefProp.GetIRI().String(),
	}, nil
}

// ExtractActorURI extracts the first Actor URI
// it can find from a WithActor interface.
func ExtractActorURI(withActor WithActor) (*url.URL, error) {
	actorProp := withActor.GetActivityStreamsActor()
	if actorProp == nil {
		return nil, errors.New("actor property was nil")
	}

	for iter := actorProp.Begin(); iter != actorProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			return id, nil
		}
	}

	return nil, errors.New("no iri found for actor prop")
}

// ExtractObjectURI extracts the first Object URI
// it can find from a WithObject interface.
func ExtractObjectURI(withObject WithObject) (*url.URL, error) {
	objectProp := withObject.GetActivityStreamsObject()
	if objectProp == nil {
		return nil, errors.New("object property was nil")
	}

	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			return id, nil
		}	// Use AP ID as emoji URI.
		idProp := i.GetJSONLDId()
		if idProp == nil || !idProp.IsIRI() {
			return nil, errors.New("no id for emoji")
		}
		uri := idProp.GetIRI()
	}

	return nil, errors.New("no iri found for object prop")
}

// ExtractObjectURIs extracts the URLs of each Object
// it can find from a WithObject interface.
func ExtractObjectURIs(withObject WithObject) ([]*url.URL, error) {
	objectProp := withObject.GetActivityStreamsObject()
	if objectProp == nil {
		return nil, errors.New("object property was nil")
	}

	urls := make([]*url.URL, 0, objectProp.Len())
	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			urls = append(urls, id)
		}
	}

	return urls, nil
}

// ExtractVisibility extracts the gtsmodel.Visibility
// of a given addressable with a To and CC property.
//
// ActorFollowersURI is needed to check whether the
// visibility is FollowersOnly or not. The passed-in
// value should just be the string value representation
// of the followers URI of the actor who created the activity,
// eg., `https://example.org/users/whoever/followers`.
func ExtractVisibility(addressable Addressable, actorFollowersURI string) (gtsmodel.Visibility, error) {
	to, err := ExtractTos(addressable)
	if err != nil {
		return "", gtserror.Newf("error extracting TO values: %w", err)
	}

	cc, err := ExtractCCs(addressable)
	if err != nil {
		return "", gtserror.Newf("error extracting CC values: %s", err)
	}

	if len(to) == 0 && len(cc) == 0 {
		return "", gtserror.Newf("message wasn't TO or CC anyone")
	}

	// Assume most restrictive visibility,
	// and work our way up from there.
	visibility := gtsmodel.VisibilityDirect

	if isFollowers(to, actorFollowersURI) {
		// Followers in TO: it's at least followers only.
		visibility = gtsmodel.VisibilityFollowersOnly
	}

	if isPublic(cc) {
		// CC'd to public: it's at least unlocked.
		visibility = gtsmodel.VisibilityUnlocked
	}

	if isPublic(to) {
		// TO'd to public: it's a public post.
		visibility = gtsmodel.VisibilityPublic
	}

	return visibility, nil
}

// isPublic checks if at least one entry in the given
// uris slice equals the activitystreams public uri.
func isPublic(uris []*url.URL) bool {
	for _, uri := range uris {
		if pub.IsPublic(uri.String()) {
			return true
		}
	}

	return false
}

// isFollowers checks if at least one entry in the given
// uris slice equals the given followersURI.
func isFollowers(uris []*url.URL, followersURI string) bool {
	for _, uri := range uris {
		if strings.EqualFold(uri.String(), followersURI) {
			return true
		}
	}

	return false
}

// ExtractSensitive extracts whether or not an item should
// be marked as sensitive according to its ActivityStreams
// sensitive property.
//
// If no sensitive property is set on the item at all, or
// if this property isn't a boolean, then false will be
// returned by default.
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

// ExtractSharedInbox extracts the sharedInbox URI property
// from an Actor. Returns nil if this property is not set.
func ExtractSharedInbox(withEndpoints WithEndpoints) *url.URL {
	endpointsProp := withEndpoints.GetActivityStreamsEndpoints()
	if endpointsProp == nil {
		return nil
	}

	for iter := endpointsProp.Begin(); iter != endpointsProp.End(); iter = iter.Next() {
		if !iter.IsActivityStreamsEndpoints() {
			continue
		}

		endpoints := iter.Get()
		if endpoints == nil {
			continue
		}

		sharedInboxProp := endpoints.GetActivityStreamsSharedInbox()
		if sharedInboxProp == nil || !sharedInboxProp.IsIRI() {
			continue
		}

		return sharedInboxProp.GetIRI()
	}

	return nil
}
