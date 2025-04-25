package pub

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
)

var (
	// ErrObjectRequired indicates the activity needs its object property
	// set. Can be returned by DelegateActor's PostInbox or PostOutbox so a
	// Bad Request response is set.
	ErrObjectRequired = errors.New("object property required on the provided activity")
	// ErrTargetRequired indicates the activity needs its target property
	// set. Can be returned by DelegateActor's PostInbox or PostOutbox so a
	// Bad Request response is set.
	ErrTargetRequired = errors.New("target property required on the provided activity")
)

// activityStreamsMediaTypes contains all of the accepted ActivityStreams media
// types. Generated at init time.
var activityStreamsMediaTypes []string

func init() {
	activityStreamsMediaTypes = []string{
		"application/activity+json",
	}
	jsonLdType := "application/ld+json"
	for _, semi := range []string{";", " ;", " ; ", "; "} {
		for _, profile := range []string{
			"profile=https://www.w3.org/ns/activitystreams",
			"profile=\"https://www.w3.org/ns/activitystreams\"",
		} {
			activityStreamsMediaTypes = append(
				activityStreamsMediaTypes,
				fmt.Sprintf("%s%s%s", jsonLdType, semi, profile))
		}
	}
}

// headerIsActivityPubMediaType returns true if the header string contains one
// of the accepted ActivityStreams media types.
//
// Note we don't try to build a comprehensive parser and instead accept a
// tolerable amount of whitespace since the HTTP specification is ambiguous
// about the format and significance of whitespace.
func headerIsActivityPubMediaType(header string) bool {
	for _, mediaType := range activityStreamsMediaTypes {
		if strings.Contains(header, mediaType) {
			return true
		}
	}
	return false
}

const (
	// The Content-Type header.
	contentTypeHeader = "Content-Type"
	// The Accept header.
	acceptHeader = "Accept"
)

// readActivityPubReq reads ActivityPub data from an incoming request, handling body close.
func readActivityPubReq(req *http.Request) (map[string]interface{}, error) {
	// Ensure closed when done.
	defer req.Body.Close()

	var m map[string]interface{}

	// Wrap body in a JSON decoder.
	dec := json.NewDecoder(req.Body)

	// Decode JSON body as "raw" AP data map.
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	// Perform a final second decode to ensure no trailing
	// garbage data or second JSON value (indicates malformed).
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New("trailing data after json")
	}

	return m, nil
}

// readActivityPubResp reads ActivityPub data from a dereference response, handling media type check and body close.
func readActivityPubResp(resp *http.Response) (map[string]interface{}, error) {
	// Ensure closed when done.
	defer resp.Body.Close()

	// Check the incoming response media type is the expected ActivityPub content-type.
	if mediaType := resp.Header.Get("Content-Type"); !headerIsActivityPubMediaType(mediaType) {
		return nil, fmt.Errorf("%s %s resp was not ActivityPub media type: %s", resp.Request.Method, resp.Request.URL, mediaType)
	}

	var m map[string]interface{}

	// Wrap body in a JSON decoder.
	dec := json.NewDecoder(resp.Body)

	// Decode JSON body as "raw" AP data map.
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	// Perform a final second decode to ensure no trailing
	// garbage data or second JSON value (indicates malformed).
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New("trailing data after json")
	}

	return m, nil
}

// isActivityPubPost returns true if the request is a POST request that has the
// ActivityStreams content type header
func isActivityPubPost(r *http.Request) bool {
	return r.Method == "POST" && headerIsActivityPubMediaType(r.Header.Get(contentTypeHeader))
}

// isActivityPubGet returns true if the request is a GET request that has the
// ActivityStreams content type header
func isActivityPubGet(r *http.Request) bool {
	return r.Method == "GET" && headerIsActivityPubMediaType(r.Header.Get(acceptHeader))
}

// dedupeOrderedItems deduplicates the 'orderedItems' within an ordered
// collection type. Deduplication happens by the 'id' property.
func dedupeOrderedItems(oc orderedItemser) error {
	oi := oc.GetActivityStreamsOrderedItems()
	if oi == nil {
		return nil
	}
	seen := make(map[string]bool, oi.Len())
	for i := 0; i < oi.Len(); {
		var id *url.URL

		iter := oi.At(i)
		asType := iter.GetType()
		if asType != nil {
			var err error
			id, err = GetId(asType)
			if err != nil {
				return err
			}
		} else if iter.IsIRI() {
			id = iter.GetIRI()
		} else {
			return fmt.Errorf("element %d in OrderedCollection does not have an ID nor is an IRI", i)
		}
		if seen[id.String()] {
			oi.Remove(i)
		} else {
			seen[id.String()] = true
			i++
		}
	}
	return nil
}

const (
	// The Location header
	locationHeader = "Location"
	// Contains the ActivityStreams Content-Type value.
	contentTypeHeaderValue = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
	// The Date header.
	dateHeader = "Date"
	// The Digest header.
	digestHeader = "Digest"
	// The delimiter used in the Digest header.
	digestDelimiter = "="
	// SHA-256 string for the Digest header.
	sha256Digest = "SHA-256"
)

// addResponseHeaders sets headers needed in the HTTP response, such but not
// limited to the Content-Type, Date, and Digest headers.
func addResponseHeaders(h http.Header, c Clock, responseContent []byte) {
	h.Set(contentTypeHeader, contentTypeHeaderValue)
	// RFC 7231 §7.1.1.2
	h.Set(dateHeader, c.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	// RFC 3230 and RFC 5843
	var b bytes.Buffer
	b.WriteString(sha256Digest)
	b.WriteString(digestDelimiter)
	hashed := sha256.Sum256(responseContent)
	b.WriteString(base64.StdEncoding.EncodeToString(hashed[:]))
	h.Set(digestHeader, b.String())
}

// IdProperty is a property that can readily have its id obtained
type IdProperty interface {
	// GetIRI returns the IRI of this property. When IsIRI returns false,
	// GetIRI will return an arbitrary value.
	GetIRI() *url.URL
	// GetType returns the value in this property as a Type. Returns nil if
	// the value is not an ActivityStreams type, such as an IRI or another
	// value.
	GetType() vocab.Type
	// IsIRI returns true if this property is an IRI.
	IsIRI() bool
}

// ToId returns an IdProperty's id.
func ToId(i IdProperty) (*url.URL, error) {
	if i.GetType() != nil {
		return GetId(i.GetType())
	} else if i.IsIRI() {
		return i.GetIRI(), nil
	}
	return nil, fmt.Errorf("cannot determine id of activitystreams property")
}

// GetId will attempt to find the 'id' property or, if it happens to be a
// Link or derived from Link type, the 'href' property instead.
//
// Returns an error if the id is not set and either the 'href' property is not
// valid on this type, or it is also not set.
func GetId(t vocab.Type) (*url.URL, error) {
	if id := t.GetJSONLDId(); id != nil {
		return id.Get(), nil
	} else if h, ok := t.(hrefer); ok {
		if href := h.GetActivityStreamsHref(); href != nil {
			return href.Get(), nil
		}
	}
	return nil, fmt.Errorf("cannot determine id of activitystreams value")
}

// getInboxForwardingValues obtains the 'inReplyTo', 'object', 'target', and
// 'tag' values on an ActivityStreams value.
func getInboxForwardingValues(o vocab.Type) (t []vocab.Type, iri []*url.URL) {
	// 'inReplyTo'
	if i, ok := o.(inReplyToer); ok {
		if irt := i.GetActivityStreamsInReplyTo(); irt != nil {
			for iter := irt.Begin(); iter != irt.End(); iter = iter.Next() {
				if tv := iter.GetType(); tv != nil {
					t = append(t, tv)
				} else {
					iri = append(iri, iter.GetIRI())
				}
			}
		}
	}
	// 'tag'
	if i, ok := o.(tagger); ok {
		if tag := i.GetActivityStreamsTag(); tag != nil {
			for iter := tag.Begin(); iter != tag.End(); iter = iter.Next() {
				if tv := iter.GetType(); tv != nil {
					t = append(t, tv)
				} else {
					iri = append(iri, iter.GetIRI())
				}
			}
		}
	}
	// 'object'
	if i, ok := o.(objecter); ok {
		if obj := i.GetActivityStreamsObject(); obj != nil {
			for iter := obj.Begin(); iter != obj.End(); iter = iter.Next() {
				if tv := iter.GetType(); tv != nil {
					t = append(t, tv)
				} else {
					iri = append(iri, iter.GetIRI())
				}
			}
		}
	}
	// 'target'
	if i, ok := o.(targeter); ok {
		if tar := i.GetActivityStreamsTarget(); tar != nil {
			for iter := tar.Begin(); iter != tar.End(); iter = iter.Next() {
				if tv := iter.GetType(); tv != nil {
					t = append(t, tv)
				} else {
					iri = append(iri, iter.GetIRI())
				}
			}
		}
	}
	return
}

// wrapInCreate will automatically wrap the provided object in a Create
// activity. This will copy over the 'to', 'bto', 'cc', 'bcc', and 'audience'
// properties. It will also copy over the published time if present.
func wrapInCreate(ctx context.Context, o vocab.Type, actor *url.URL) (c vocab.ActivityStreamsCreate, err error) {
	c = streams.NewActivityStreamsCreate()
	// Object property
	oProp := streams.NewActivityStreamsObjectProperty()
	oProp.AppendType(o)
	c.SetActivityStreamsObject(oProp)
	// Actor Property
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actor)
	c.SetActivityStreamsActor(actorProp)
	// Published Property
	if v, ok := o.(publisheder); ok {
		c.SetActivityStreamsPublished(v.GetActivityStreamsPublished())
	}
	// Copying over properties.
	if v, ok := o.(toer); ok {
		if to := v.GetActivityStreamsTo(); to != nil {
			activityTo := streams.NewActivityStreamsToProperty()
			for iter := to.Begin(); iter != to.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				activityTo.AppendIRI(id)
			}
			c.SetActivityStreamsTo(activityTo)
		}
	}
	if v, ok := o.(btoer); ok {
		if bto := v.GetActivityStreamsBto(); bto != nil {
			activityBto := streams.NewActivityStreamsBtoProperty()
			for iter := bto.Begin(); iter != bto.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				activityBto.AppendIRI(id)
			}
			c.SetActivityStreamsBto(activityBto)
		}
	}
	if v, ok := o.(ccer); ok {
		if cc := v.GetActivityStreamsCc(); cc != nil {
			activityCc := streams.NewActivityStreamsCcProperty()
			for iter := cc.Begin(); iter != cc.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				activityCc.AppendIRI(id)
			}
			c.SetActivityStreamsCc(activityCc)
		}
	}
	if v, ok := o.(bccer); ok {
		if bcc := v.GetActivityStreamsBcc(); bcc != nil {
			activityBcc := streams.NewActivityStreamsBccProperty()
			for iter := bcc.Begin(); iter != bcc.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				activityBcc.AppendIRI(id)
			}
			c.SetActivityStreamsBcc(activityBcc)
		}
	}
	if v, ok := o.(audiencer); ok {
		if aud := v.GetActivityStreamsAudience(); aud != nil {
			activityAudience := streams.NewActivityStreamsAudienceProperty()
			for iter := aud.Begin(); iter != aud.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				activityAudience.AppendIRI(id)
			}
			c.SetActivityStreamsAudience(activityAudience)
		}
	}
	return
}

const (
	// PublicActivityPubIRI is the IRI that indicates an Activity is meant
	// to be visible for general public consumption.
	PublicActivityPubIRI = "https://www.w3.org/ns/activitystreams#Public"
	publicJsonLD         = "Public"
	publicJsonLDAS       = "as:Public"
)

// IsPublic determines if an IRI string is the Public collection as defined in
// the spec, including JSON-LD compliant collections.
func IsPublic(s string) bool {
	return s == PublicActivityPubIRI || s == publicJsonLD || s == publicJsonLDAS
}

// Derives an ID URI from the given IdProperty and, if it's not the
// magic AP Public IRI, appends it to the actorsAndCollections slice.
func appendToActorsAndCollectionsIRIs(
	iter IdProperty,
	actorsAndCollections []*url.URL,
) ([]*url.URL, error) {
	id, err := ToId(iter)
	if err != nil {
		return nil, err
	}

	// Ignore Public IRI as we
	// can't deliver to it directly.
	if !IsPublic(id.String()) {
		actorsAndCollections = append(actorsAndCollections, id)
	}

	return actorsAndCollections, nil
}

// actorsToInboxIRIs extracts the 'inbox' IRIs from actor types.
func actorsToInboxIRIs(t []vocab.Type) (u []*url.URL, err error) {
	for _, elem := range t {
		var iri *url.URL
		iri, err = getInbox(elem)
		if err != nil {
			return
		}
		u = append(u, iri)
	}
	return
}

// getInbox extracts the 'inbox' IRI from an actor type.
func getInbox(t vocab.Type) (u *url.URL, err error) {
	ib, ok := t.(inboxer)
	if !ok {
		err = fmt.Errorf("actor type %T has no inbox", t)
		return
	}
	inbox := ib.GetActivityStreamsInbox()
	return ToId(inbox)
}

// filterInboxIRIs will deduplicate the given inboxes
// slice, while also leaving out any filtered IRIs.
func filterInboxIRIs(
	inboxes []*url.URL,
	filtered ...*url.URL,
) []*url.URL {
	// Prepopulate the ignored map with each filtered IRI.
	ignored := make(map[string]struct{}, len(filtered)+len(inboxes))
	for _, filteredIRI := range filtered {
		ignored[filteredIRI.String()] = struct{}{}
	}

	deduped := make([]*url.URL, 0, len(inboxes))
	for _, inbox := range inboxes {
		inboxStr := inbox.String()
		_, ignore := ignored[inboxStr]
		if ignore {
			// We already included
			// this URI in out, or
			// we should ignore it.
			continue
		}

		// Include this IRI in output, and
		// add entry to the ignored map to
		// ensure we don't include it again.
		deduped = append(deduped, inbox)
		ignored[inboxStr] = struct{}{}
	}

	return deduped
}

// stripHiddenRecipients removes "bto" and "bcc" from the activity.
//
// Note that this requirement of the specification is under "Section 6: Client
// to Server Interactions", the Social API, and not the Federative API.
func stripHiddenRecipients(activity Activity) {
	activity.SetActivityStreamsBto(nil)
	activity.SetActivityStreamsBcc(nil)
	op := activity.GetActivityStreamsObject()
	if op != nil {
		for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
			if v, ok := iter.GetType().(btoer); ok {
				v.SetActivityStreamsBto(nil)
			}
			if v, ok := iter.GetType().(bccer); ok {
				v.SetActivityStreamsBcc(nil)
			}
		}
	}
}

// mustHaveActivityOriginMatchObjects ensures that the Host in the activity id
// IRI matches all of the Hosts in the object id IRIs.
func mustHaveActivityOriginMatchObjects(a Activity) error {
	originIRI, err := GetId(a)
	if err != nil {
		return err
	}
	originHost := originIRI.Host
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return nil
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		iri, err := ToId(iter)
		if err != nil {
			return err
		}
		if originHost != iri.Host {
			return fmt.Errorf("object %q: not in activity origin", iri)
		}
	}
	return nil
}

// normalizeRecipients ensures the activity and object have the same 'to',
// 'bto', 'cc', 'bcc', and 'audience' properties. Copy the Activity's recipients
// to objects, and the objects to the activity, but does NOT copy objects'
// recipients to each other.
func normalizeRecipients(a vocab.ActivityStreamsCreate) error {
	// Phase 0: Acquire all recipients on the activity.
	//
	// Obtain the actorTo map
	actorToMap := make(map[string]*url.URL)
	actorTo := a.GetActivityStreamsTo()
	if actorTo == nil {
		actorTo = streams.NewActivityStreamsToProperty()
		a.SetActivityStreamsTo(actorTo)
	}
	for iter := actorTo.Begin(); iter != actorTo.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		actorToMap[id.String()] = id
	}
	// Obtain the actorBto map
	actorBtoMap := make(map[string]*url.URL)
	actorBto := a.GetActivityStreamsBto()
	if actorBto == nil {
		actorBto = streams.NewActivityStreamsBtoProperty()
		a.SetActivityStreamsBto(actorBto)
	}
	for iter := actorBto.Begin(); iter != actorBto.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		actorBtoMap[id.String()] = id
	}
	// Obtain the actorCc map
	actorCcMap := make(map[string]*url.URL)
	actorCc := a.GetActivityStreamsCc()
	if actorCc == nil {
		actorCc = streams.NewActivityStreamsCcProperty()
		a.SetActivityStreamsCc(actorCc)
	}
	for iter := actorCc.Begin(); iter != actorCc.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		actorCcMap[id.String()] = id
	}
	// Obtain the actorBcc map
	actorBccMap := make(map[string]*url.URL)
	actorBcc := a.GetActivityStreamsBcc()
	if actorBcc == nil {
		actorBcc = streams.NewActivityStreamsBccProperty()
		a.SetActivityStreamsBcc(actorBcc)
	}
	for iter := actorBcc.Begin(); iter != actorBcc.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		actorBccMap[id.String()] = id
	}
	// Obtain the actorAudience map
	actorAudienceMap := make(map[string]*url.URL)
	actorAudience := a.GetActivityStreamsAudience()
	if actorAudience == nil {
		actorAudience = streams.NewActivityStreamsAudienceProperty()
		a.SetActivityStreamsAudience(actorAudience)
	}
	for iter := actorAudience.Begin(); iter != actorAudience.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		actorAudienceMap[id.String()] = id
	}
	// Obtain the objects maps for each recipient type.
	o := a.GetActivityStreamsObject()
	objsTo := make([]map[string]*url.URL, o.Len())
	objsBto := make([]map[string]*url.URL, o.Len())
	objsCc := make([]map[string]*url.URL, o.Len())
	objsBcc := make([]map[string]*url.URL, o.Len())
	objsAudience := make([]map[string]*url.URL, o.Len())
	for i := 0; i < o.Len(); i++ {
		iter := o.At(i)
		// Phase 1: Acquire all existing recipients on the object.
		//
		// Object to
		objsTo[i] = make(map[string]*url.URL)
		var oTo vocab.ActivityStreamsToProperty
		if tr, ok := iter.GetType().(toer); !ok {
			return fmt.Errorf("the Create object at %d has no 'to' property", i)
		} else {
			oTo = tr.GetActivityStreamsTo()
			if oTo == nil {
				oTo = streams.NewActivityStreamsToProperty()
				tr.SetActivityStreamsTo(oTo)
			}
		}
		for iter := oTo.Begin(); iter != oTo.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			objsTo[i][id.String()] = id
		}
		// Object bto
		objsBto[i] = make(map[string]*url.URL)
		var oBto vocab.ActivityStreamsBtoProperty
		if tr, ok := iter.GetType().(btoer); !ok {
			return fmt.Errorf("the Create object at %d has no 'bto' property", i)
		} else {
			oBto = tr.GetActivityStreamsBto()
			if oBto == nil {
				oBto = streams.NewActivityStreamsBtoProperty()
				tr.SetActivityStreamsBto(oBto)
			}
		}
		for iter := oBto.Begin(); iter != oBto.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			objsBto[i][id.String()] = id
		}
		// Object cc
		objsCc[i] = make(map[string]*url.URL)
		var oCc vocab.ActivityStreamsCcProperty
		if tr, ok := iter.GetType().(ccer); !ok {
			return fmt.Errorf("the Create object at %d has no 'cc' property", i)
		} else {
			oCc = tr.GetActivityStreamsCc()
			if oCc == nil {
				oCc = streams.NewActivityStreamsCcProperty()
				tr.SetActivityStreamsCc(oCc)
			}
		}
		for iter := oCc.Begin(); iter != oCc.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			objsCc[i][id.String()] = id
		}
		// Object bcc
		objsBcc[i] = make(map[string]*url.URL)
		var oBcc vocab.ActivityStreamsBccProperty
		if tr, ok := iter.GetType().(bccer); !ok {
			return fmt.Errorf("the Create object at %d has no 'bcc' property", i)
		} else {
			oBcc = tr.GetActivityStreamsBcc()
			if oBcc == nil {
				oBcc = streams.NewActivityStreamsBccProperty()
				tr.SetActivityStreamsBcc(oBcc)
			}
		}
		for iter := oBcc.Begin(); iter != oBcc.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			objsBcc[i][id.String()] = id
		}
		// Object audience
		objsAudience[i] = make(map[string]*url.URL)
		var oAudience vocab.ActivityStreamsAudienceProperty
		if tr, ok := iter.GetType().(audiencer); !ok {
			return fmt.Errorf("the Create object at %d has no 'audience' property", i)
		} else {
			oAudience = tr.GetActivityStreamsAudience()
			if oAudience == nil {
				oAudience = streams.NewActivityStreamsAudienceProperty()
				tr.SetActivityStreamsAudience(oAudience)
			}
		}
		for iter := oAudience.Begin(); iter != oAudience.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			objsAudience[i][id.String()] = id
		}
		// Phase 2: Apply missing recipients to the object from the
		// activity.
		//
		// Activity to -> Object to
		for k, v := range actorToMap {
			if _, ok := objsTo[i][k]; !ok {
				oTo.AppendIRI(v)
			}
		}
		// Activity bto -> Object bto
		for k, v := range actorBtoMap {
			if _, ok := objsBto[i][k]; !ok {
				oBto.AppendIRI(v)
			}
		}
		// Activity cc -> Object cc
		for k, v := range actorCcMap {
			if _, ok := objsCc[i][k]; !ok {
				oCc.AppendIRI(v)
			}
		}
		// Activity bcc -> Object bcc
		for k, v := range actorBccMap {
			if _, ok := objsBcc[i][k]; !ok {
				oBcc.AppendIRI(v)
			}
		}
		// Activity audience -> Object audience
		for k, v := range actorAudienceMap {
			if _, ok := objsAudience[i][k]; !ok {
				oAudience.AppendIRI(v)
			}
		}
	}
	// Phase 3: Apply missing recipients to the activity from the objects.
	//
	// Object to -> Activity to
	for i := 0; i < len(objsTo); i++ {
		for k, v := range objsTo[i] {
			if _, ok := actorToMap[k]; !ok {
				actorTo.AppendIRI(v)
			}
		}
	}
	// Object bto -> Activity bto
	for i := 0; i < len(objsBto); i++ {
		for k, v := range objsBto[i] {
			if _, ok := actorBtoMap[k]; !ok {
				actorBto.AppendIRI(v)
			}
		}
	}
	// Object cc -> Activity cc
	for i := 0; i < len(objsCc); i++ {
		for k, v := range objsCc[i] {
			if _, ok := actorCcMap[k]; !ok {
				actorCc.AppendIRI(v)
			}
		}
	}
	// Object bcc -> Activity bcc
	for i := 0; i < len(objsBcc); i++ {
		for k, v := range objsBcc[i] {
			if _, ok := actorBccMap[k]; !ok {
				actorBcc.AppendIRI(v)
			}
		}
	}
	// Object audience -> Activity audience
	for i := 0; i < len(objsAudience); i++ {
		for k, v := range objsAudience[i] {
			if _, ok := actorAudienceMap[k]; !ok {
				actorAudience.AppendIRI(v)
			}
		}
	}
	return nil
}

// toTombstone creates a Tombstone object for the given ActivityStreams value.
func toTombstone(obj vocab.Type, id *url.URL, now time.Time) vocab.ActivityStreamsTombstone {
	tomb := streams.NewActivityStreamsTombstone()
	// id property
	idProp := streams.NewJSONLDIdProperty()
	idProp.Set(id)
	tomb.SetJSONLDId(idProp)
	// formerType property
	former := streams.NewActivityStreamsFormerTypeProperty()
	tomb.SetActivityStreamsFormerType(former)
	// Populate Former Type
	former.AppendXMLSchemaString(obj.GetTypeName())
	// Copy over the published property if it existed
	if pubber, ok := obj.(publisheder); ok {
		if pub := pubber.GetActivityStreamsPublished(); pub != nil {
			tomb.SetActivityStreamsPublished(pub)
		}
	}
	// Copy over the updated property if it existed
	if upder, ok := obj.(updateder); ok {
		if upd := upder.GetActivityStreamsUpdated(); upd != nil {
			tomb.SetActivityStreamsUpdated(upd)
		}
	}
	// Set deleted time to now.
	deleted := streams.NewActivityStreamsDeletedProperty()
	deleted.Set(now)
	tomb.SetActivityStreamsDeleted(deleted)
	return tomb
}

// mustHaveActivityActorsMatchObjectActors ensures that the actors on types in
// the 'object' property are all listed in the 'actor' property.
func mustHaveActivityActorsMatchObjectActors(c context.Context,
	actors vocab.ActivityStreamsActorProperty,
	op vocab.ActivityStreamsObjectProperty,
	newTransport func(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t Transport, err error),
	boxIRI *url.URL,
) error {
	activityActorMap := make(map[string]bool, actors.Len())
	for iter := actors.Begin(); iter != actors.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		activityActorMap[id.String()] = true
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		iri, err := ToId(iter)
		if err != nil {
			return err
		}
		// Attempt to dereference the IRI, regardless whether it is a
		// type or IRI
		tport, err := newTransport(c, boxIRI, goFedUserAgent())
		if err != nil {
			return err
		}
		resp, err := tport.Dereference(c, iri)
		if err != nil {
			return err
		}
		m, err := readActivityPubResp(resp)
		if err != nil {
			return err
		}
		t, err := streams.ToType(c, m)
		if err != nil {
			return err
		}
		ac, ok := t.(actorer)
		if !ok {
			return fmt.Errorf("cannot verify actors: object value has no 'actor' property")
		}
		objActors := ac.GetActivityStreamsActor()
		for iter := objActors.Begin(); iter != objActors.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			if !activityActorMap[id.String()] {
				return fmt.Errorf("activity does not have all actors from its object's actors")
			}
		}
	}
	return nil
}

// add implements the logic of adding object ids to a target Collection or
// OrderedCollection. This logic is shared by both the C2S and S2S protocols.
func add(c context.Context,
	op vocab.ActivityStreamsObjectProperty,
	target vocab.ActivityStreamsTargetProperty,
	db Database,
) error {
	opIds := make([]*url.URL, 0, op.Len())
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		opIds = append(opIds, id)
	}
	targetIds := make([]*url.URL, 0, op.Len())
	for iter := target.Begin(); iter != target.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		targetIds = append(targetIds, id)
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(t *url.URL) error {
		unlock, err := db.Lock(c, t)
		if err != nil {
			return err
		}
		defer unlock()
		if owns, err := db.Owns(c, t); err != nil {
			return err
		} else if !owns {
			return nil
		}
		tp, err := db.Get(c, t)
		if err != nil {
			return err
		}
		if streams.IsOrExtendsActivityStreamsOrderedCollection(tp) {
			oi, ok := tp.(orderedItemser)
			if !ok {
				return fmt.Errorf("type extending from OrderedCollection cannot convert to orderedItemser interface")
			}
			oiProp := oi.GetActivityStreamsOrderedItems()
			if oiProp == nil {
				oiProp = streams.NewActivityStreamsOrderedItemsProperty()
				oi.SetActivityStreamsOrderedItems(oiProp)
			}
			for _, objId := range opIds {
				oiProp.AppendIRI(objId)
			}
		} else if streams.IsOrExtendsActivityStreamsCollection(tp) {
			i, ok := tp.(itemser)
			if !ok {
				return fmt.Errorf("type extending from Collection cannot convert to itemser interface")
			}
			iProp := i.GetActivityStreamsItems()
			if iProp == nil {
				iProp = streams.NewActivityStreamsItemsProperty()
				i.SetActivityStreamsItems(iProp)
			}
			for _, objId := range opIds {
				iProp.AppendIRI(objId)
			}
		} else {
			return fmt.Errorf("target in Add is neither a Collection nor an OrderedCollection")
		}
		err = db.Update(c, tp)
		if err != nil {
			return err
		}
		return nil
	}
	for _, t := range targetIds {
		if err := loopFn(t); err != nil {
			return err
		}
	}
	return nil
}

// remove implements the logic of removing object ids to a target Collection or
// OrderedCollection. This logic is shared by both the C2S and S2S protocols.
func remove(c context.Context,
	op vocab.ActivityStreamsObjectProperty,
	target vocab.ActivityStreamsTargetProperty,
	db Database,
) error {
	opIds := make(map[string]bool, op.Len())
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		opIds[id.String()] = true
	}
	targetIds := make([]*url.URL, 0, op.Len())
	for iter := target.Begin(); iter != target.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		targetIds = append(targetIds, id)
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(t *url.URL) error {
		unlock, err := db.Lock(c, t)
		if err != nil {
			return err
		}
		defer unlock()
		if owns, err := db.Owns(c, t); err != nil {
			return err
		} else if !owns {
			return nil
		}
		tp, err := db.Get(c, t)
		if err != nil {
			return err
		}
		if streams.IsOrExtendsActivityStreamsOrderedCollection(tp) {
			oi, ok := tp.(orderedItemser)
			if !ok {
				return fmt.Errorf("type extending from OrderedCollection cannot convert to orderedItemser interface")
			}
			oiProp := oi.GetActivityStreamsOrderedItems()
			if oiProp != nil {
				for i := 0; i < oiProp.Len(); /*Conditional*/ {
					id, err := ToId(oiProp.At(i))
					if err != nil {
						return err
					}
					if opIds[id.String()] {
						oiProp.Remove(i)
					} else {
						i++
					}
				}
			}
		} else if streams.IsOrExtendsActivityStreamsCollection(tp) {
			i, ok := tp.(itemser)
			if !ok {
				return fmt.Errorf("type extending from Collection cannot convert to itemser interface")
			}
			iProp := i.GetActivityStreamsItems()
			if iProp != nil {
				for i := 0; i < iProp.Len(); /*Conditional*/ {
					id, err := ToId(iProp.At(i))
					if err != nil {
						return err
					}
					if opIds[id.String()] {
						iProp.Remove(i)
					} else {
						i++
					}
				}
			}
		} else {
			return fmt.Errorf("target in Remove is neither a Collection nor an OrderedCollection")
		}
		err = db.Update(c, tp)
		if err != nil {
			return err
		}
		return nil
	}
	for _, t := range targetIds {
		if err := loopFn(t); err != nil {
			return err
		}
	}
	return nil
}

// clearSensitiveFields removes the 'bto' and 'bcc' entries on the given value
// and recursively on every 'object' property value.
func clearSensitiveFields(obj vocab.Type) {
	if t, ok := obj.(btoer); ok {
		t.SetActivityStreamsBto(nil)
	}
	if t, ok := obj.(bccer); ok {
		t.SetActivityStreamsBcc(nil)
	}
	if t, ok := obj.(objecter); ok {
		op := t.GetActivityStreamsObject()
		if op != nil {
			for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
				clearSensitiveFields(iter.GetType())
			}
		}
	}
}

// requestId forms an ActivityPub id based on the HTTP request. Always assumes
// that the id is HTTPS.
func requestId(r *http.Request, scheme string) *url.URL {
	id := r.URL
	id.Host = r.Host
	id.Scheme = scheme
	return id
}
