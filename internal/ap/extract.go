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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"codeberg.org/superseriousbusiness/activity/pub"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ExtractObjects will extract object vocab.Types from given implementing interface.
func ExtractObjects(with WithObject) []TypeOrIRI {
	// Extract the attached object (if any).
	objProp := with.GetActivityStreamsObject()
	if objProp == nil {
		return nil
	}

	// Check for zero len.
	if objProp.Len() == 0 {
		return nil
	}

	// Accumulate all of the objects into a slice.
	objs := make([]TypeOrIRI, objProp.Len())
	for i := 0; i < objProp.Len(); i++ {
		objs[i] = objProp.At(i)
	}

	return objs
}

// ExtractActivityData will extract the usable data type (e.g. Note, Question, etc) and corresponding JSON, from activity.
func ExtractActivityData(activity pub.Activity, rawJSON map[string]any) ([]TypeOrIRI, []any, bool) {
	switch typeName := activity.GetTypeName(); {
	// Activity (has "object").
	case isActivity(typeName):
		objTypes := ExtractObjects(activity)
		if len(objTypes) == 0 {
			return nil, nil, false
		}

		var objJSON []any
		switch json := rawJSON["object"].(type) {
		case nil:
			// do nothing
		case map[string]any:
			// Wrap map in slice.
			objJSON = []any{json}
		case []any:
			// Use existing slice.
			objJSON = json
		}

		return objTypes, objJSON, true

	// IntransitiveAcitivity (no "object").
	case isIntransitiveActivity(typeName):
		asTypeOrIRI := _TypeOrIRI{activity} // wrap activity.
		return []TypeOrIRI{&asTypeOrIRI}, []any{rawJSON}, true

	// Unknown.
	default:
		return nil, nil, false
	}
}

// ExtractAccountables extracts Accountable objects from a slice TypeOrIRI, returning extracted and remaining TypeOrIRIs.
func ExtractAccountables(arr []TypeOrIRI) ([]Accountable, []TypeOrIRI) {
	var accounts []Accountable

	for i := 0; i < len(arr); i++ {
		elem := arr[i]

		if elem.IsIRI() {
			// skip IRIs
			continue
		}

		// Extract AS vocab type
		// associated with elem.
		t := elem.GetType()

		// Try cast AS type as Accountable.
		account, ok := ToAccountable(t)
		if !ok {
			continue
		}

		// Add casted accountable type.
		accounts = append(accounts, account)

		// Drop elem from slice.
		copy(arr[:i], arr[i+1:])
		arr = arr[:len(arr)-1]
	}

	return accounts, arr
}

// ExtractStatusables extracts Statusable objects from a slice TypeOrIRI, returning extracted and remaining TypeOrIRIs.
func ExtractStatusables(arr []TypeOrIRI) ([]Statusable, []TypeOrIRI) {
	var statuses []Statusable

	for i := 0; i < len(arr); i++ {
		elem := arr[i]

		if elem.IsIRI() {
			// skip IRIs
			continue
		}

		// Extract AS vocab type
		// associated with elem.
		t := elem.GetType()

		// Try cast AS type as Statusable.
		status, ok := ToStatusable(t)
		if !ok {
			continue
		}

		// Add casted Statusable type.
		statuses = append(statuses, status)

		// Drop elem from slice.
		copy(arr[:i], arr[i+1:])
		arr = arr[:len(arr)-1]
	}

	return statuses, arr
}

// ExtractPollOptionables extracts PollOptionable objects from a slice TypeOrIRI, returning extracted and remaining TypeOrIRIs.
func ExtractPollOptionables(arr []TypeOrIRI) ([]PollOptionable, []TypeOrIRI) {
	var options []PollOptionable

	for i := 0; i < len(arr); i++ {
		elem := arr[i]

		if elem.IsIRI() {
			// skip IRIs
			continue
		}

		// Extract AS vocab type
		// associated with elem.
		t := elem.GetType()

		// Try cast as PollOptionable.
		option, ok := ToPollOptionable(t)
		if !ok {
			continue
		}

		// Add casted PollOptionable type.
		options = append(options, option)

		// Drop elem from slice.
		copy(arr[:i], arr[i+1:])
		arr = arr[:len(arr)-1]
	}

	return options, arr
}

// ExtractPreferredUsername returns a string representation of
// an interface's preferredUsername property. Will return an
// error if preferredUsername is nil, not a string, or empty.
func ExtractPreferredUsername(i WithPreferredUsername) string {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return ""
	}
	return u.GetXMLSchemaString()
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
func ExtractCcURIs(i WithCc) []*url.URL {
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

// ExtractPubKeyFromActor extracts the public key, public key ID, and public
// key owner ID from an interface, or an error if something goes wrong.
func ExtractPubKeyFromActor(i WithPublicKey) (
	*rsa.PublicKey, // pubkey
	*url.URL, // pubkey ID
	*url.URL, // pubkey owner
	error,
) {
	pubKeyProp := i.GetW3IDSecurityV1PublicKey()
	if pubKeyProp == nil {
		return nil, nil, nil, gtserror.New("public key property was nil")
	}

	// Take the first public key we can find.
	for iter := pubKeyProp.Begin(); iter != pubKeyProp.End(); iter = iter.Next() {
		if !iter.IsW3IDSecurityV1PublicKey() {
			continue
		}

		pkey := iter.Get()
		if pkey == nil {
			continue
		}

		return ExtractPubKeyFromKey(pkey)
	}

	return nil, nil, nil, gtserror.New("couldn't find valid public key")
}

// ExtractPubKeyFromActor extracts the public key, public key ID, and public
// key owner ID from an interface, or an error if something goes wrong.
func ExtractPubKeyFromKey(pkey vocab.W3IDSecurityV1PublicKey) (
	*rsa.PublicKey, // pubkey
	*url.URL, // pubkey ID
	*url.URL, // pubkey owner
	error,
) {
	pubKeyID, err := pub.GetId(pkey)
	if err != nil {
		return nil, nil, nil, errors.New("no id set on public key")
	}

	pubKeyOwnerProp := pkey.GetW3IDSecurityV1Owner()
	if pubKeyOwnerProp == nil {
		return nil, nil, nil, errors.New("nil pubKeyOwnerProp")
	}

	pubKeyOwner := pubKeyOwnerProp.GetIRI()
	if pubKeyOwner == nil {
		return nil, nil, nil, errors.New("nil iri on pubKeyOwnerProp")
	}

	pubKeyPemProp := pkey.GetW3IDSecurityV1PublicKeyPem()
	if pubKeyPemProp == nil {
		return nil, nil, nil, errors.New("nil pubKeyPemProp")
	}

	pkeyPem := pubKeyPemProp.Get()
	if pkeyPem == "" {
		return nil, nil, nil, errors.New("empty pubKeyPemProp")
	}

	block, _ := pem.Decode([]byte(pkeyPem))
	if block == nil {
		return nil, nil, nil, errors.New("nil pubKeyPem")
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
		err = fmt.Errorf("could not parse public key from block bytes: %w", err)
		return nil, nil, nil, err
	}

	if p == nil {
		return nil, nil, nil, fmt.Errorf("returned public key was empty")
	}

	pubKey, ok := p.(*rsa.PublicKey)
	if !ok {
		return nil, nil, nil, fmt.Errorf("could not type pubKey to *rsa.PublicKey")
	}

	return pubKey, pubKeyID, pubKeyOwner, nil
}

// ExtractContent returns an intermediary representation of
// the given interface's Content and/or ContentMap property.
func ExtractContent(i WithContent) gtsmodel.Content {
	content := gtsmodel.Content{}

	contentProp := i.GetActivityStreamsContent()
	if contentProp == nil {
		// No content at all.
		return content
	}

	for iter := contentProp.Begin(); iter != contentProp.End(); iter = iter.Next() {
		switch {
		case iter.IsRDFLangString() &&
			len(content.ContentMap) == 0:
			content.ContentMap = iter.GetRDFLangString()

		case iter.IsXMLSchemaString() &&
			content.Content == "":
			content.Content = iter.GetXMLSchemaString()

		case iter.IsIRI() &&
			content.Content == "":
			content.Content = iter.GetIRI().String()
		}
	}

	return content
}

// ExtractAttachments attempts to extract barebones MediaAttachment objects from given AS interface type.
func ExtractAttachments(i WithAttachment) ([]*gtsmodel.MediaAttachment, error) {
	attachmentProp := i.GetActivityStreamsAttachment()
	if attachmentProp == nil {
		return nil, nil
	}

	var errs gtserror.MultiError

	attachments := make([]*gtsmodel.MediaAttachment, 0, attachmentProp.Len())
	for iter := attachmentProp.Begin(); iter != attachmentProp.End(); iter = iter.Next() {
		t := iter.GetType()
		if t == nil {
			errs.Appendf("nil attachment type")
			continue
		}
		attachmentable, ok := t.(Attachmentable)
		if !ok {
			errs.Appendf("incorrect attachment type: %T", t)
			continue
		}
		attachment, err := ExtractAttachment(attachmentable)
		if err != nil {
			errs.Appendf("error extracting attachment: %w", err)
			continue
		}
		attachments = append(attachments, attachment)
	}

	return attachments, errs.Combine()
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
		Description: ExtractDescription(i),
		Blurhash:    ExtractBlurhash(i),
		Processing:  gtsmodel.ProcessingStatusReceived,
	}, nil
}

// ExtractDescription extracts the image description
// of an attachmentable, if present. Will try the
// 'summary' prop first, then fall back to 'name'.
func ExtractDescription(i Attachmentable) string {
	if summary := ExtractSummary(i); summary != "" {
		return summary
	}

	return ExtractName(i)
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
// or has a name that cannot be normalized, it will be ignored.
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

		tag, err := extractHashtag(hashtaggable)
		if err != nil {
			continue
		}

		// "Normalize" this tag by combining diacritics +
		// unicode chars. If this returns false, it means
		// we couldn't normalize it well enough to make it
		// valid on our instance, so just ignore it.
		normalized, ok := text.NormalizeHashtag(tag.Name)
		if !ok {
			continue
		}

		// We store tag names lowercased, might
		// as well change case here already.
		tag.Name = strings.ToLower(normalized)

		// Only append this tag if we haven't
		// seen it already, to avoid duplicates
		// in the slice.
		if _, set := keys[tag.Name]; !set {
			keys[tag.Name] = nil // Value doesn't matter.
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// extractHashtag extracts a minimal gtsmodel.Tag from the given
// Hashtaggable, without yet doing any normalization on it.
func extractHashtag(i Hashtaggable) (*gtsmodel.Tag, error) {
	// Extract name for the tag; trim leading hash
	// character, so '#example' becomes 'example'.
	name := ExtractName(i)
	if name == "" {
		return nil, gtserror.New("name prop empty")
	}
	tagName := strings.TrimPrefix(name, "#")

	// Extract href for the tag, if set.
	//
	// Fine if not, it's only used for spam
	// checking anyway so not critical.
	var href string
	hrefProp := i.GetActivityStreamsHref()
	if hrefProp != nil && hrefProp.IsIRI() {
		href = hrefProp.GetIRI().String()
	}

	return &gtsmodel.Tag{
		Name:     tagName,
		Useable:  util.Ptr(true), // Assume true by default.
		Listable: util.Ptr(true), // Assume true by default.
		Href:     href,
	}, nil
}

// ExtractEmojis extracts a slice of minimal gtsmodel.Emojis
// from a WithTag. If an entry in the WithTag is not an emoji,
// it will be quietly ignored.
func ExtractEmojis(i WithTag, host string) ([]*gtsmodel.Emoji, error) {
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

		emoji, err := ExtractEmoji(tootEmoji, host)
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

// ExtractEmoji extracts a minimal gtsmodel.Emoji from
// the given Emojiable. The host (eg., "example.org")
// of the emoji should be passed in as well, so that a
// dummy URI for the emoji can be constructed in case
// there's no id property or id property is null.
//
// https://github.com/superseriousbusiness/gotosocial/issues/3384)
func ExtractEmoji(
	e Emojiable,
	host string,
) (*gtsmodel.Emoji, error) {
	// Extract emoji name,
	// eg., ":some_emoji".
	name := ExtractName(e)
	if name == "" {
		return nil, gtserror.New("name prop empty")
	}
	name = strings.TrimSpace(name)

	// Derive shortcode from
	// name, eg., "some_emoji".
	shortcode := strings.Trim(name, ":")
	shortcode = strings.TrimSpace(shortcode)

	// Extract emoji image
	// URL from Icon property.
	imageRemoteURL, err := ExtractIconURI(e)
	if err != nil {
		return nil, gtserror.New("no url for emoji image")
	}
	imageRemoteURLStr := imageRemoteURL.String()

	// Use AP ID as emoji URI, or fall
	// back to dummy URI if not present.
	uri := GetJSONLDId(e)
	if uri == nil {
		// No ID was set,
		// construct dummy.
		uri, err = url.Parse(
			// eg., https://example.org/dummy_emoji_path?shortcode=some_emoji
			"https://" + host + "/dummy_emoji_path?shortcode=" + url.QueryEscape(shortcode),
		)
		if err != nil {
			return nil, gtserror.Newf("error constructing dummy path: %w", err)
		}
	}

	return &gtsmodel.Emoji{
		UpdatedAt:       GetUpdated(e),
		Shortcode:       shortcode,
		Domain:          host,
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
	// See if a name has been set in the
	// format `@someone@example.org`.
	nameString := ExtractName(i)

	// The href prop should be the AP URI
	// of the target account; it could also
	// be the URL, but we'll check this later.
	var href string
	hrefProp := i.GetActivityStreamsHref()
	if hrefProp != nil && hrefProp.IsIRI() {
		href = hrefProp.GetIRI().String()
	}

	// One of nameString and hrefProp must be set.
	if nameString == "" && href == "" {
		return nil, gtserror.Newf("neither Name nor Href were set")
	}

	return &gtsmodel.Mention{
		NameString:       nameString,
		TargetAccountURI: href,
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
		return 0, gtserror.Newf("message wasn't TO or CC anyone")
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

// ExtractInteractionPolicy extracts a *gtsmodel.InteractionPolicy
// from the given Statusable created by by the given *gtsmodel.Account.
//
// Will be nil (default policy) for Statusables that have no policy
// set on them, or have a null policy. In such a case, the caller
// should assume the default policy for the status's visibility level.
func ExtractInteractionPolicy(
	statusable Statusable,
	owner *gtsmodel.Account,
) *gtsmodel.InteractionPolicy {
	ipa, ok := statusable.(InteractionPolicyAware)
	if !ok {
		// Not a type with interaction
		// policy properties settable.
		return nil
	}

	policyProp := ipa.GetGoToSocialInteractionPolicy()
	if policyProp == nil || policyProp.Len() != 1 {
		return nil
	}

	policyPropIter := policyProp.At(0)
	if !policyPropIter.IsGoToSocialInteractionPolicy() {
		return nil
	}

	policy := policyPropIter.Get()
	if policy == nil {
		return nil
	}

	return &gtsmodel.InteractionPolicy{
		CanLike:     extractCanLike(policy.GetGoToSocialCanLike(), owner),
		CanReply:    extractCanReply(policy.GetGoToSocialCanReply(), owner),
		CanAnnounce: extractCanAnnounce(policy.GetGoToSocialCanAnnounce(), owner),
	}
}

func extractCanLike(
	prop vocab.GoToSocialCanLikeProperty,
	owner *gtsmodel.Account,
) gtsmodel.PolicyRules {
	if prop == nil || prop.Len() != 1 {
		return gtsmodel.PolicyRules{}
	}

	propIter := prop.At(0)
	if !propIter.IsGoToSocialCanLike() {
		return gtsmodel.PolicyRules{}
	}

	withRules := propIter.Get()
	if withRules == nil {
		return gtsmodel.PolicyRules{}
	}

	return gtsmodel.PolicyRules{
		Always:       extractPolicyValues(withRules.GetGoToSocialAlways(), owner),
		WithApproval: extractPolicyValues(withRules.GetGoToSocialApprovalRequired(), owner),
	}
}

func extractCanReply(
	prop vocab.GoToSocialCanReplyProperty,
	owner *gtsmodel.Account,
) gtsmodel.PolicyRules {
	if prop == nil || prop.Len() != 1 {
		return gtsmodel.PolicyRules{}
	}

	propIter := prop.At(0)
	if !propIter.IsGoToSocialCanReply() {
		return gtsmodel.PolicyRules{}
	}

	withRules := propIter.Get()
	if withRules == nil {
		return gtsmodel.PolicyRules{}
	}

	return gtsmodel.PolicyRules{
		Always:       extractPolicyValues(withRules.GetGoToSocialAlways(), owner),
		WithApproval: extractPolicyValues(withRules.GetGoToSocialApprovalRequired(), owner),
	}
}

func extractCanAnnounce(
	prop vocab.GoToSocialCanAnnounceProperty,
	owner *gtsmodel.Account,
) gtsmodel.PolicyRules {
	if prop == nil || prop.Len() != 1 {
		return gtsmodel.PolicyRules{}
	}

	propIter := prop.At(0)
	if !propIter.IsGoToSocialCanAnnounce() {
		return gtsmodel.PolicyRules{}
	}

	withRules := propIter.Get()
	if withRules == nil {
		return gtsmodel.PolicyRules{}
	}

	return gtsmodel.PolicyRules{
		Always:       extractPolicyValues(withRules.GetGoToSocialAlways(), owner),
		WithApproval: extractPolicyValues(withRules.GetGoToSocialApprovalRequired(), owner),
	}
}

func extractPolicyValues[T WithIRI](
	prop Property[T],
	owner *gtsmodel.Account,
) gtsmodel.PolicyValues {
	iris := getIRIs(prop)
	PolicyValues := make(gtsmodel.PolicyValues, 0, len(iris))

	for _, iri := range iris {
		switch iriStr := iri.String(); iriStr {
		case pub.PublicActivityPubIRI:
			PolicyValues = append(PolicyValues, gtsmodel.PolicyValuePublic)
		case owner.FollowersURI:
			PolicyValues = append(PolicyValues, gtsmodel.PolicyValueFollowers)
		case owner.FollowingURI:
			PolicyValues = append(PolicyValues, gtsmodel.PolicyValueFollowing)
		case owner.URI:
			PolicyValues = append(PolicyValues, gtsmodel.PolicyValueAuthor)
		default:
			if iri.Scheme == "http" || iri.Scheme == "https" {
				PolicyValues = append(PolicyValues, gtsmodel.PolicyValue(iriStr))
			}
		}
	}

	return PolicyValues
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

// ExtractPoll extracts a placeholder Poll from Pollable interface, with available options and flags populated.
func ExtractPoll(poll Pollable) (*gtsmodel.Poll, error) {
	var closed time.Time

	// Extract the options (votes if any) and 'multiple choice' flag.
	options, multi, hideCounts, err := extractPollOptions(poll)
	if err != nil {
		return nil, err
	}

	// Extract the poll closed time,
	// it's okay for this to be zero.
	closedSlice := GetClosed(poll)
	if len(closedSlice) == 1 {
		closed = closedSlice[0]
	}

	// Extract the poll end time, again
	// this isn't necessarily set as some
	// servers support "endless" polls.
	endTime := GetEndTime(poll)

	if endTime.IsZero() && !closed.IsZero() {
		// If no endTime is provided, but the
		// poll is marked as closed, infer the
		// endTime from the closed time.
		endTime = closed
	}

	// Extract the number of voters.
	voters := GetVotersCount(poll)

	return &gtsmodel.Poll{
		Options:    optionNames(options),
		Multiple:   &multi,
		HideCounts: &hideCounts,
		Votes:      optionVotes(options),
		Voters:     &voters,
		ExpiresAt:  endTime,
		ClosedAt:   closed,
	}, nil
}

// pollOption is a simple type
// to unify a poll option name
// with the number of votes.
type pollOption struct {
	Name  string
	Votes int
}

// optionNames extracts name strings from a slice of poll options.
func optionNames(in []pollOption) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = in[i].Name
	}
	return out
}

// optionVotes extracts vote counts from a slice of poll options.
func optionVotes(in []pollOption) []int {
	out := make([]int, len(in))
	for i := range in {
		out[i] = in[i].Votes
	}
	return out
}

// extractPollOptions extracts poll option name strings, the 'multiple choice flag', and 'hideCounts' intrinsic flag properties value from Pollable.
func extractPollOptions(poll Pollable) (options []pollOption, multi bool, hide bool, err error) {
	var errs gtserror.MultiError

	// Iterate the oneOf property and gather poll single-choice options.
	IterateOneOf(poll, func(iter vocab.ActivityStreamsOneOfPropertyIterator) {
		name, votes, err := extractPollOption(iter.GetType())
		if err != nil {
			errs.Append(err)
			return
		}
		if votes == nil {
			hide = true
			votes = new(int)
		}
		options = append(options, pollOption{
			Name:  name,
			Votes: *votes,
		})
	})
	if len(options) > 0 || len(errs) > 0 {
		return options, false, hide, errs.Combine()
	}

	// Iterate the anyOf property and gather poll multi-choice options.
	IterateAnyOf(poll, func(iter vocab.ActivityStreamsAnyOfPropertyIterator) {
		name, votes, err := extractPollOption(iter.GetType())
		if err != nil {
			errs.Append(err)
			return
		}
		if votes == nil {
			hide = true
			votes = new(int)
		}
		options = append(options, pollOption{
			Name:  name,
			Votes: *votes,
		})
	})
	if len(options) > 0 || len(errs) > 0 {
		return options, true, hide, errs.Combine()
	}

	return nil, false, false, errors.New("poll without options")
}

// IterateOneOf will attempt to extract oneOf property from given interface, and passes each iterated item to function.
func IterateOneOf(withOneOf WithOneOf, foreach func(vocab.ActivityStreamsOneOfPropertyIterator)) {
	if foreach == nil {
		// nil check outside loop.
		panic("nil function")
	}

	// Extract the one-of property from interface.
	oneOfProp := withOneOf.GetActivityStreamsOneOf()
	if oneOfProp == nil {
		return
	}

	// Get start and end of iter.
	start := oneOfProp.Begin()
	end := oneOfProp.End()

	// Pass iterated oneOf entries to given function.
	for iter := start; iter != end; iter = iter.Next() {
		foreach(iter)
	}
}

// IterateAnyOf will attempt to extract anyOf property from given interface, and passes each iterated item to function.
func IterateAnyOf(withAnyOf WithAnyOf, foreach func(vocab.ActivityStreamsAnyOfPropertyIterator)) {
	if foreach == nil {
		// nil check outside loop.
		panic("nil function")
	}

	// Extract the any-of property from interface.
	anyOfProp := withAnyOf.GetActivityStreamsAnyOf()
	if anyOfProp == nil {
		return
	}

	// Get start and end of iter.
	start := anyOfProp.Begin()
	end := anyOfProp.End()

	// Pass iterated anyOf entries to given function.
	for iter := start; iter != end; iter = iter.Next() {
		foreach(iter)
	}
}

// extractPollOption extracts a usable poll option name from vocab.Type, or error.
func extractPollOption(t vocab.Type) (name string, votes *int, err error) {
	// Check fulfills PollOptionable type
	// (this accounts for nil input type).
	optionable, ok := t.(PollOptionable)
	if !ok {
		return "", nil, fmt.Errorf("incorrect option type: %T", t)
	}

	// Extract PollOption from interface.
	name = ExtractName(optionable)
	if name == "" {
		return "", nil, errors.New("empty option name")
	}

	// Check PollOptionable for attached 'replies' property.
	repliesProp := optionable.GetActivityStreamsReplies()
	if repliesProp != nil {

		// Get repliesProp as the AS collection type it should be.
		collection := repliesProp.GetActivityStreamsCollection()
		if collection != nil {

			// Extract integer value from the collection 'totalItems' property.
			totalItemsProp := collection.GetActivityStreamsTotalItems()
			if totalItemsProp != nil {
				i := totalItemsProp.Get()
				votes = &i
			}
		}
	}

	return name, votes, nil
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
