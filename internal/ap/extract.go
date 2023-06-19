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

package ap

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ExtractPreferredUsername returns a string representation of
// an interface's preferredUsername property. Will return an
// error if preferredUsername is nil, not a string, or empty.
func ExtractPreferredUsername(i WithPreferredUsername) (string, error) {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return "", gtserror.New("preferredUsername nil or not a string")
	}

	if u.GetXMLSchemaString() == "" {
		return "", gtserror.New("preferredUsername was empty")
	}

	return u.GetXMLSchemaString(), nil
}

// ExtractName returns the first string representation it
// can find of an interface's name property, or an empty
// string if this is not found.
func ExtractName(i WithName) string {
	nameProp := i.GetActivityStreamsName()
	if nameProp == nil {
		return ""
	}

	for iter := nameProp.Begin(); iter != nameProp.End(); iter = iter.Next() {
		// Name may be parsed as IRI, depending on
		// how it's formatted, so account for this.
		switch {
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString()
		case iter.IsIRI():
			return iter.GetIRI().String()
		}
	}

	return ""
}

// ExtractInReplyToURI extracts the first inReplyTo URI
// property it can find from an interface. Will return
// nil if no valid URI can be found.
func ExtractInReplyToURI(i WithInReplyTo) *url.URL {
	inReplyToProp := i.GetActivityStreamsInReplyTo()
	if inReplyToProp == nil {
		return nil
	}

	for iter := inReplyToProp.Begin(); iter != inReplyToProp.End(); iter = iter.Next() {
		iri, err := pub.ToId(iter)
		if err == nil && iri != nil {
			// Found one we can use.
			return iri
		}
	}

	return nil
}

// ExtractItemsURIs extracts each URI it can
// find for an item from the provided WithItems.
func ExtractItemsURIs(i WithItems) []*url.URL {
	itemsProp := i.GetActivityStreamsItems()
	if itemsProp == nil {
		return nil
	}

	uris := make([]*url.URL, 0, itemsProp.Len())
	for iter := itemsProp.Begin(); iter != itemsProp.End(); iter = iter.Next() {
		uri, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			uris = append(uris, uri)
		}
	}

	return uris
}

// ExtractToURIs returns a slice of URIs
// that the given WithTo addresses as To.
func ExtractToURIs(i WithTo) []*url.URL {
	toProp := i.GetActivityStreamsTo()
	if toProp == nil {
		return nil
	}

	uris := make([]*url.URL, 0, toProp.Len())
	for iter := toProp.Begin(); iter != toProp.End(); iter = iter.Next() {
		uri, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			uris = append(uris, uri)
		}
	}

	return uris
}

// ExtractCcURIs returns a slice of URIs
// that the given WithCC addresses as Cc.
func ExtractCcURIs(i WithCC) []*url.URL {
	ccProp := i.GetActivityStreamsCc()
	if ccProp == nil {
		return nil
	}

	urls := make([]*url.URL, 0, ccProp.Len())
	for iter := ccProp.Begin(); iter != ccProp.End(); iter = iter.Next() {
		uri, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			urls = append(urls, uri)
		}
	}

	return urls
}

// ExtractAttributedToURI returns the first URI it can find in the
// given WithAttributedTo, or an error if no URI can be found.
func ExtractAttributedToURI(i WithAttributedTo) (*url.URL, error) {
	attributedToProp := i.GetActivityStreamsAttributedTo()
	if attributedToProp == nil {
		return nil, gtserror.New("attributedToProp was nil")
	}

	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			return id, nil
		}
	}

	return nil, gtserror.New("couldn't find iri for attributed to")
}

// ExtractPublished extracts the published time from the given
// WithPublished. Will return an error if the published property
// is not set, is not a time.Time, or is zero.
func ExtractPublished(i WithPublished) (time.Time, error) {
	t := time.Time{}

	publishedProp := i.GetActivityStreamsPublished()
	if publishedProp == nil {
		return t, gtserror.New("published prop was nil")
	}

	if !publishedProp.IsXMLSchemaDateTime() {
		return t, gtserror.New("published prop was not date time")
	}

	t = publishedProp.Get()
	if t.IsZero() {
		return t, gtserror.New("published time was zero")
	}

	return t, nil
}

// ExtractIconURI extracts the first URI it can find from
// the given WithIcon which links to a supported image file.
// Input will look something like this:
//
//	"icon": {
//	  "mediaType": "image/jpeg",
//	  "type": "Image",
//	  "url": "http://example.org/path/to/some/file.jpeg"
//	},
//
// If no valid URI can be found, this will return an error.
func ExtractIconURI(i WithIcon) (*url.URL, error) {
	iconProp := i.GetActivityStreamsIcon()
	if iconProp == nil {
		return nil, gtserror.New("icon property was nil")
	}

	// Icon can potentially contain multiple entries,
	// so we iterate through all of them here in order
	// to find the first one that meets these criteria:
	//
	//   1. Is an image.
	//   2. Has a URL that we can use to derefereince it.
	for iter := iconProp.Begin(); iter != iconProp.End(); iter = iter.Next() {
		if !iter.IsActivityStreamsImage() {
			continue
		}

		image := iter.GetActivityStreamsImage()
		if image == nil {
			continue
		}

		imageURL, err := ExtractURL(image)
		if err == nil && imageURL != nil {
			return imageURL, nil
		}
	}

	return nil, gtserror.New("could not extract valid image URI from icon")
}

// ExtractImageURI extracts the first URI it can find from
// the given WithImage which links to a supported image file.
// Input will look something like this:
//
//	"image": {
//	  "mediaType": "image/jpeg",
//	  "type": "Image",
//	  "url": "http://example.org/path/to/some/file.jpeg"
//	},
//
// If no valid URI can be found, this will return an error.
func ExtractImageURI(i WithImage) (*url.URL, error) {
	imageProp := i.GetActivityStreamsImage()
	if imageProp == nil {
		return nil, gtserror.New("image property was nil")
	}

	// Image can potentially contain multiple entries,
	// so we iterate through all of them here in order
	// to find the first one that meets these criteria:
	//
	//   1. Is an image.
	//   2. Has a URL that we can use to derefereince it.
	for iter := imageProp.Begin(); iter != imageProp.End(); iter = iter.Next() {
		if !iter.IsActivityStreamsImage() {
			continue
		}

		image := iter.GetActivityStreamsImage()
		if image == nil {
			continue
		}

		imageURL, err := ExtractURL(image)
		if err == nil && imageURL != nil {
			return imageURL, nil
		}
	}

	return nil, gtserror.New("could not extract valid image URI from image")
}

// ExtractSummary extracts the summary/content warning of
// the given WithSummary interface. Will return an empty
// string if no summary/content warning was present.
func ExtractSummary(i WithSummary) string {
	summaryProp := i.GetActivityStreamsSummary()
	if summaryProp == nil {
		return ""
	}

	for iter := summaryProp.Begin(); iter != summaryProp.End(); iter = iter.Next() {
		// Summary may be parsed as IRI, depending on
		// how it's formatted, so account for this.
		switch {
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString()
		case iter.IsIRI():
			return iter.GetIRI().String()
		}
	}

	return ""
}

// ExtractFields extracts property/value fields from the given
// WithAttachment interface. Will return an empty slice if no
// property/value fields can be found. Attachments that are not
// (well-formed) PropertyValues will be ignored.
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

// ExtractDiscoverable extracts the Discoverable boolean
// of the given WithDiscoverable interface. Will return
// an error if Discoverable was nil.
func ExtractDiscoverable(i WithDiscoverable) (bool, error) {
	discoverableProp := i.GetTootDiscoverable()
	if discoverableProp == nil {
		return false, gtserror.New("discoverable was nil")
	}

	return discoverableProp.Get(), nil
}

// ExtractURL extracts the first URI it can find from the
// given WithURL interface, or an error if no URL was set.
// The ID of a type will not work, this function wants a URI
// specifically.
func ExtractURL(i WithURL) (*url.URL, error) {
	urlProp := i.GetActivityStreamsUrl()
	if urlProp == nil {
		return nil, gtserror.New("url property was nil")
	}

	for iter := urlProp.Begin(); iter != urlProp.End(); iter = iter.Next() {
		if !iter.IsIRI() {
			continue
		}

		// Found it.
		return iter.GetIRI(), nil
	}

	return nil, gtserror.New("no valid URL property found")
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

// ExtractContent returns a string representation of the
// given interface's Content property, or an empty string
// if no Content is found.
func ExtractContent(i WithContent) string {
	contentProperty := i.GetActivityStreamsContent()
	if contentProperty == nil {
		return ""
	}

	for iter := contentProperty.Begin(); iter != contentProperty.End(); iter = iter.Next() {
		switch {
		// Content may be parsed as IRI, depending on
		// how it's formatted, so account for this.
		case iter.IsXMLSchemaString():
			return iter.GetXMLSchemaString()
		case iter.IsIRI():
			return iter.GetIRI().String()
		}
	}

	return ""
}

// ExtractAttachment extracts a minimal gtsmodel.Attachment
// (just remote URL, description, and blurhash) from the given
// Attachmentable interface, or an error if no remote URL is set.
func ExtractAttachment(i Attachmentable) (*gtsmodel.MediaAttachment, error) {
	// Get the URL for the attachment file.
	// If no URL is set, we can't do anything.
	remoteURL, err := ExtractURL(i)
	if err != nil {
		return nil, gtserror.Newf("error extracting attachment URL: %w", err)
	}

	return &gtsmodel.MediaAttachment{
		RemoteURL:   remoteURL.String(),
		Description: ExtractName(i),
		Blurhash:    ExtractBlurhash(i),
		Processing:  gtsmodel.ProcessingStatusReceived,
	}, nil
}

// ExtractBlurhash extracts the blurhash string value
// from the given WithBlurhash interface, or returns
// an empty string if nothing is found.
func ExtractBlurhash(i WithBlurhash) string {
	blurhashProp := i.GetTootBlurhash()
	if blurhashProp == nil {
		return ""
	}

	return blurhashProp.Get()
}

// ExtractHashtags extracts a slice of minimal gtsmodel.Tags
// from a WithTag. If an entry in the WithTag is not a hashtag,
// it will be quietly ignored.
//
// TODO: find a better heuristic for determining if something
// is a hashtag or not, since looking for type name "Hashtag"
// is non-normative. Perhaps look for things that are either
// type "Hashtag" or have no type name set at all?
func ExtractHashtags(i WithTag) ([]*gtsmodel.Tag, error) {
	tagsProp := i.GetActivityStreamsTag()
	if tagsProp == nil {
		return nil, nil
	}

	var (
		l    = tagsProp.Len()
		tags = make([]*gtsmodel.Tag, 0, l)
		keys = make(map[string]any, l) // Use map to dedupe items.
	)

	for iter := tagsProp.Begin(); iter != tagsProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			continue
		}

		if t.GetTypeName() != TagHashtag {
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

		// Only append this tag if we haven't
		// seen it already, to avoid duplicates
		// in the slice.
		if _, set := keys[tag.URL]; !set {
			keys[tag.URL] = nil // Value doesn't matter.
			tags = append(tags, tag)
			tags = append(tags, tag)
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// ExtractEmoji extracts a minimal gtsmodel.Tag
// from the given Hashtaggable.
func ExtractHashtag(i Hashtaggable) (*gtsmodel.Tag, error) {
	// Extract href/link for this tag.
	hrefProp := i.GetActivityStreamsHref()
	if hrefProp == nil || !hrefProp.IsIRI() {
		return nil, gtserror.New("no href prop")
	}
	tagURL := hrefProp.GetIRI().String()

	// Extract name for the tag; trim leading hash
	// character, so '#example' becomes 'example'.
	name := ExtractName(i)
	if name == "" {
		return nil, gtserror.New("name prop empty")
	}
	tagName := strings.TrimPrefix(name, "#")

	return &gtsmodel.Tag{
		URL:  tagURL,
		Name: tagName,
	}, nil
}

// ExtractEmojis extracts a slice of minimal gtsmodel.Emojis
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
		if !iter.IsTootEmoji() {
			continue
		}

		tootEmoji := iter.GetTootEmoji()
		if tootEmoji == nil {
			continue
		}

		emoji, err := ExtractEmoji(tootEmoji)
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

// ExtractEmoji extracts a minimal gtsmodel.Emoji
// from the given Emojiable.
func ExtractEmoji(i Emojiable) (*gtsmodel.Emoji, error) {
	// Use AP ID as emoji URI.
	idProp := i.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, gtserror.New("no id for emoji")
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
		return nil, gtserror.New("name prop empty")
	}
	shortcode := strings.Trim(name, ":")

	// Extract emoji image URL from Icon property.
	imageRemoteURL, err := ExtractIconURI(i)
	if err != nil {
		return nil, gtserror.New("no url for emoji image")
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

// ExtractMentions extracts a slice of minimal gtsmodel.Mentions
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
		if !iter.IsActivityStreamsMention() {
			continue
		}

		asMention := iter.GetActivityStreamsMention()
		if asMention == nil {
			continue
		}

		mention, err := ExtractMention(asMention)
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
		return nil, gtserror.New("name prop empty")
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
		return nil, gtserror.New("no href prop")
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
		return nil, gtserror.New("actor property was nil")
	}

	for iter := actorProp.Begin(); iter != actorProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			return id, nil
		}
	}

	return nil, gtserror.New("no iri found for actor prop")
}

// ExtractObjectURI extracts the first Object URI
// it can find from a WithObject interface.
func ExtractObjectURI(withObject WithObject) (*url.URL, error) {
	objectProp := withObject.GetActivityStreamsObject()
	if objectProp == nil {
		return nil, gtserror.New("object property was nil")
	}

	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			return id, nil
		}
	}

	return nil, gtserror.New("no iri found for object prop")
}

// ExtractObjectURIs extracts the URLs of each Object
// it can find from a WithObject interface.
func ExtractObjectURIs(withObject WithObject) ([]*url.URL, error) {
	objectProp := withObject.GetActivityStreamsObject()
	if objectProp == nil {
		return nil, gtserror.New("object property was nil")
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
	var (
		to = ExtractToURIs(addressable)
		cc = ExtractCcURIs(addressable)
	)

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
