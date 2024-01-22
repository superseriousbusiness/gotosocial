# Federating with GoToSocial

Information on the various (ActivityPub) elements needed to federate with GoToSocial.

## HTTP Signatures

GoToSocial requires all `GET` and `POST` requests to ActivityPub s2s endpoints to be accompanied by a valid http signature.

GoToSocial will also sign all outgoing `GET` and `POST` requests that it makes to other servers.

This behavior is the equivalent of Mastodon's [AUTHORIZED_FETCH / "secure mode"](https://docs.joinmastodon.org/admin/config/#authorized_fetch).

GoToSocial uses the [go-fed/httpsig](https://github.com/go-fed/httpsig) library for signing outgoing requests, and for parsing and validating the signatures of incoming requests. This library strictly follows the [Cavage http signature RFC](https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures), which is the same RFC used by other implementations like Mastodon, Pixelfed, Akkoma/Pleroma, etc. (This RFC has since been superceded by the [httpbis http signature RFC](https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures), but this is not yet widely implemented.)

### Incoming Requests

GoToSocial request signature validation is implemented in [internal/federation](https://github.com/superseriousbusiness/gotosocial/blob/main/internal/federation/authenticate.go).

GoToSocial will attempt to parse the signature using the following algorithms (in order), stopping at the first success:

```text
RSA_SHA256
RSA_SHA512
ED25519
```

### Outgoing Requests

GoToSocial request signing is implemented in [internal/transport](https://github.com/superseriousbusiness/gotosocial/blob/main/internal/transport/signing.go).

When assembling signatures:

- outgoing `GET` requests use `(request-target) host date`
- outgoing `POST` requests use `(request-target) host date digest` 

GoToSocial uses the `RSA_SHA256` algorithm for signing requests, which is in line with other ActivityPub implementations.

### Quirks

The `keyId` used by GoToSocial in the `Signature` header will look something like the following:

```text
https://example.org/users/example_user/main-key
```

This is different from most other implementations, which usually use a fragment (`#`) in the `keyId` uri. For example, on Mastodon the user's key would instead be found at:

```text
https://example.org/users/example_user#main-key
```

For Mastodon, the public key of a user is served as part of that user's Actor representation. GoToSocial mimics this behavior when serving the public key of a user, but instead of returning the entire Actor at the `main-key` endpoint (which may contain sensitive fields), will return only a partial stub of the actor. This looks like the following:

```json
{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams"
  ],
  "id": "https://example.org/users/example_user",
  "preferredUsername": "example_user",
  "publicKey": {
    "id": "https://example.org/users/example_user/main-key",
    "owner": "https://example.org/users/example_user",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzGB3yDvMl+8p+ViutVRG\nVDl9FO7ZURYXnwB3TedSfG13jyskoiMDNvsbLoUQM9ajZPB0zxJPZUlB/W3BWHRC\nNFQglE5DkB30GjTClNZoOrx64vLRT5wAEwIOjklKVNk9GJi1hFFxrgj931WtxyML\nBvo+TdEblBcoru6MKAov8IU4JjQj5KUmjnW12Rox8dj/rfGtdaH8uJ14vLgvlrAb\neQbN5Ghaxh9DGTo1337O9a9qOsir8YQqazl8ahzS2gvYleV+ou09RDhS75q9hdF2\nLI+1IvFEQ2ZO2tLk3umUP1ioa+5CWKsWD0GAXbQu9uunAV0VoExP4+/9WYOuP0ei\nKwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "type": "Person"
}
```

Remote servers federating with GoToSocial should extract the public key from the `publicKey` field. Then, they should use the `owner` field of the public key to further dereference the full version of the Actor, using a signed `GET` request.

This behavior was introduced as a way of avoiding having remote servers make unsigned `GET` requests to the full Actor endpoint. However, this may change in future as it is not compliant and causes issues. Tracked in [this issue](https://github.com/superseriousbusiness/gotosocial/issues/1186).

## Access Control

GoToSocial uses access control restrictions to protect users and resources from unwanted interactions with remote accounts and instances.

As shown in the [HTTP Signatures](#http-signatures) section, GoToSocial requires all incoming `GET` and `POST` requests from remote servers to be signed. Unsigned requests will be denied with http code `401 Unauthorized`.

Access control restrictions are implemented by checking the `keyId` of the signature (who owns the public/private key pair making the request).

First, the host value of the `keyId` uri is checked against the GoToSocial instance's list of blocked (defederated) domains. If the host is recognized as a blocked domain, then the http request will immediately be aborted with http code `403 Forbidden`.

Next, GoToSocial will check for the existence of a block (in either direction) between the owner of the public key making the http request, and the owner of the resource that the request is targeting. If the GoToSocial user blocks the remote account making the request, then the request will be aborted with http code `403 Forbidden`.

## Request Throttling & Rate Limiting

GoToSocial applies http request throttling and rate limiting to the ActivityPub API endpoints (inboxes, user endpoints, emojis, etc).

This ensures that remote servers cannot flood a GoToSocial instance with spurious requests. Instead, remote servers making GET or POST requests to the ActivityPub API endpoints should respect 429 and 503 http codes, and take account of the `retry-after` http response header.

For more details on request throttling and rate limiting behavior, please see the [throttling](../api/throttling.md) and [rate limiting](../api/ratelimiting.md) documents.

## Inbox

GoToSocial implements Inboxes for Actors following the ActivityPub specification [here](https://www.w3.org/TR/activitypub/#inbox).

Remote servers should deliver Activities to a GoToSocial server by making an HTTP POST request to each Inbox of the desired audience of an Activity, as described [here](https://www.w3.org/TR/activitypub/#delivery).

GoToSocial accounts do not currently implement a [sharedInbox](https://www.w3.org/TR/activitypub/#shared-inbox-delivery) endpoint, though this is subject to change. Deduplication of delivered Activities, in case more than one Actor on a GoToSocial server is in the audience for an Activity, is handled on GoToSocial's side.

POSTs to a GoToSocial Actor's inbox must be appropriately [http-signed](#http-signatures) by the delivering Actor.

Accepted Inbox POST `Content-Type` headers are:

- `application/activity+json`
- `application/activity+json; charset=utf-8`
- `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`

Inbox POST requests that do not use one of the above `Content-Type` headers will be rejected with HTTP status code [406 - Not Acceptable](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/406).

For more information on acceptable content types, see [the server-to-server interactions](https://www.w3.org/TR/activitypub/#server-to-server-interactions) section of the ActivityPub protocol.

GoToSocial will return HTTP status code [202 - Accepted](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/202) in response to validly-formed and signed Inbox POST requests.

Invalidly-formed Inbox POST requests will receive a [400 - Bad Request](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400) HTTP status code in response. The response body may contain more information on why the GoToSocial server considered the request content to be badly formed. Other servers should not retry delivery of the Activity in case of a code `400` response.

Even if GoToSocial returns a `202` status code, it may not continue processing the Activity delivered, depending on the originator(s), target(s) and type of the Activity. ActivityPub is an extensive protocol, and GoToSocial does not cover every combination of Activity and Object.

## Outbox

GoToSocial implements Outboxes for Actors (ie., instance accounts) following the ActivityPub specification [here](https://www.w3.org/TR/activitypub/#outbox).

To get an [OrderedCollection](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollection) of Activities that an Actor has published recently, remote servers can do a `GET` request to a user's outbox. The address of this will be something like `https://example.org/users/whatever/outbox`.

The server will return an OrderedCollection of the following structure:

```json
{
    "@context": "https://www.w3.org/ns/activitystreams",
    "id": "https://example.org/users/whatever/outbox",
    "type": "OrderedCollection",
    "first": "https://example.org/users/whatever/outbox?page=true"
}
```

Note that the `OrderedCollection` itself contains no items. Callers must dereference the `first` page to start getting items. For example, a `GET` to `https://example.org/users/whatever/outbox?page=true` will produce something like the following:

```json
{
    "id": "https://example.org/users/whatever/outbox?page=true",
    "type": "OrderedCollectionPage",
    "next": "https://example.org/users/whatever/outbox?max_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
    "prev": "https://example.org/users/whatever/outbox?min_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
    "partOf": "https://example.org/users/whatever/outbox",
    "orderedItems": [
        {
            "id": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7/activity",
            "type": "Create",
            "actor": "https://example.org/users/whatever",
            "published": "2021-10-18T20:06:18Z",
            "to": [
                "https://www.w3.org/ns/activitystreams#Public"
            ],
            "cc": [
                "https://example.org/users/whatever/followers"
            ],
            "object": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7"
        }
    ]
}
```

The `orderedItems` array will contain up to 30 entries. To get more entries beyond that, the caller can use the `next` link provided in the response.

Note that in the returned `orderedItems`, all activity types will be `Create`. On each activity, the `object` field will be the AP URI of an original public status created by the Actor who owns the Outbox (ie., a `Note` with `https://www.w3.org/ns/activitystreams#Public` in the `to` field, which is not a reply to another status). Callers can use the returned AP URIs to dereference the content of the notes.

## Conversation Threads

Due to the nature of decentralization and federation, it is practically impossible for any one server on the fediverse to be aware of every post in a given conversation thread.

With that said, it is possible to do 'best effort' dereferencing of threads, whereby remote replies are fetched from one server onto another, to try to more fully flesh out a conversation.

GoToSocial does this by iterating up and down the thread of a conversation, pulling in remote statuses where possible.

Let's say we have two accounts: `local_account` on `our.server`, and `remote_1` on `remote.1`.

In this scenario, `local_account` follows `remote_1`, so posts from `remote_1` show up in the home timeline of `local_account`.

Now, `remote_1` boosts/reblogs a post from a third account, `remote_2`, residing on server `remote.2`.

`local_account` does not follow `remote_2`, and neither does anybody else on `our.server`, which means that `our.server` has not seen this post by `remote_2` before.

![A diagram of the conversation thread, showing the post from remote_2, and possible ancestor and descendant posts](../assets/diagrams/conversation_thread.png)

What GoToSocial will do now, is 'dereference' the post by `remote_2` to check if it is part of a thread and, if so, whether any other parts of the thread can be obtained.

GtS begins by checking the `inReplyTo` property of the post, which is set when a post is a reply to another post. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-inreplyto). If `inReplyTo` is set, GoToSocial derefences the replied-to post. If *this* post also has an `inReplyTo` set, then GoToSocial dereferences that too, and so on.

Once all of these **ancestors** of a status have been retrieved, GtS will begin working down through the **descendants** of posts.

It does this by checking the `replies` property of a derefenced post, and working through replies, and replies of replies. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-replies).

This process of thread dereferencing will likely involve making multiple HTTP calls to different servers, especially if the thread is long and complicated.

The end result of this dereferencing is that, assuming the reblogged post by `remote_2` was part of a thread, then `local_account` should now be able to see posts in the thread when they open the status on their home timeline. In other words, they will see replies from accounts on other servers (who they may not have come across yet), in addition to any previous and next posts in the thread as posted by `remote_2`.

This gives `local_account` a more complete view on the conversation, as opposed to just seeing the reblogged post in isolation and out of context. It also gives `local_account` the opportunity to discover new accounts to follow, based on replies to `remote_2`.

## Reports / Flags

Like other microblogging ActivityPub implementations, GoToSocial uses the [Flag](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-flag) Activity type to communicate user moderation reports to other servers.

### Outgoing

The json of an outgoing GoToSocial `Flag` looks like the following:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://example.org/users/example.org",
  "content": "dark souls sucks, please yeet this nerd",
  "id": "http://example.org/reports/01GP3AWY4CRDVRNZKW0TEAMB5R",
  "object": [
    "http://fossbros-anonymous.io/users/foss_satan",
    "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M"
  ],
  "type": "Flag"
}
```

The `actor` of the `Flag` will always be the instance actor of the GoToSocial instance on which the `Flag` was created. This is done to preserve partial anonymity of the user who created the report, in order to prevent them becoming a target for harassment.

The `content` of the `Flag` is a piece of text submitted by the user who created the `Flag`, which should give remote instance admins a reason why the report was created. This may be an empty string, or may not be present on the json, if no reason was submitted by the user.

The value of the `object` field of the `Flag` will either be a string (the ActivityPub `id` of the user being reported), or it will be an array of strings, where the first entry in the array is the `id` of the reported user, and subsequent entries are the `id`s of one or more reported `Note`s / statuses.

The `Flag` activity is delivered as-is to the `inbox` (or shared inbox) of the reported user. It is not wrapped in a `Create` activity.

### Incoming

GoToSocial assumes incoming reports will be delivered as a `Flag` Activity to the `inbox` of the account being reported.  It will parse the incoming `Flag` following the same formula that it uses for creating outgoing `Flag`s, with one difference: it will attempt to parse status URLs from both the `object` field, and from a Misskey/Calckey-formatted `content` value, which includes in-line status URLs.

GoToSocial will not assume that the `to` field will be set on an incoming `Flag` activity. Instead, it assumes that remote instances use `bto` to direct the `Flag` to its recipient.

A valid incoming `Flag` Activity will be made available as a report to the admin(s) of the GoToSocial instance that received the report, so that they can take any necessary moderation action against the reported user.

The reported user themself will not see the report, or be notified that they have been reported, unless the GtS admin chooses to share this information with them via some other channel.

## Featured (aka pinned) Posts

GoToSocial allows users to feature (or 'pin') posts on their profile.

In ActivityPub terms, GoToSocial serves these pinned posts as an [OrderedCollection](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollection) at the endpoint indicated in an Actor's [featured](https://docs.joinmastodon.org/spec/activitypub/#featured) field. The value of this field will be set to something like `https://example.org/users/some_user/collections/featured`.

By making a signed GET request to this endpoint, remote instances can dereference the featured posts collection, which will return an `OrderedCollection` with a list of post URIs in the `orderedItems` field.

Example of a featured collection of a user who has pinned multiple `Note`s:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/some_user/collections/featured",
  "orderedItems": [
    "https://example.org/users/some_user/statuses/01GS7VTYH0S77NNXTP6W4G9EAG",
    "https://example.org/users/some_user/statuses/01GSFY2SZK9TPCJFQ1WCCPGDRT",
    "https://example.org/users/some_user/statuses/01GSCXY70MZCBFMH5EKJW9ENC8"
  ],
  "totalItems": 3,
  "type": "OrderedCollection"
}
```

Example of a user who has pinned one `Note`:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/some_user/collections/featured",
  "orderedItems": [
    "https://example.org/users/some_user/statuses/01GS7VTYH0S77NNXTP6W4G9EAG"
  ],
  "totalItems": 1,
  "type": "OrderedCollection"
}
```

Example with no pinned `Note`s:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/some_user/collections/featured",
  "orderedItems": [],
  "totalItems": 0,
  "type": "OrderedCollection"
}
```

Unlike Mastodon and some other implementations, GoToSocial does *not* serve full `Note` representations as `orderedItems` values. Instead, it provides just the URI of each `Note`, which the remote server can then dereference (or not, if they already have the `Note` cached locally).

Some of the URIs served as part of the collection may point to followers-only posts which the requesting `Actor` won't necessarily have permission to view. Remote servers should make sure to do their own filtering (as with any other post type) to ensure that these posts are only shown to users who are permitted to view them.

Another difference between GoToSocial and other server implementations is that GoToSocial does not send updates to remote servers when a post is pinned or unpinned by a user. Mastodon does this by sending [Add](https://www.w3.org/TR/activitypub/#add-activity-inbox) and [Remove](https://www.w3.org/TR/activitypub/#remove-activity-inbox) Activity types where the `object` is the post being pinned or unpinned, and the `target` is the sending `Actor`'s `featured` collection. While this conceptually makes sense, it is not in line with what the ActivityPub protocol recommends, since the `target` of the Activity "is not owned by the receiving server, and thus they can't update it".

Instead, to build a view of a GoToSocial user's pinned posts, it is recommended that remote instances simply poll a GoToSocial Actor's `featured` collection every so often, and add/remove posts in their cached representation as appropriate.

## Post Deletes

GoToSocial allows users to delete posts that they have created. These deletes will be federated out to other instances, which are expected to also delete their local cache of the post.

### Outgoing

When a post is deleted by a GoToSocial user, the server will send a `Delete` activity out to other instances.

The `Delete` will have the ActivityPub URI of the post set as the value of the `Object` entry.

`to` and `cc` will be set according to the visibility of the original post, and any users mentioned/replied to by the original post.

If the original post was not a direct message, the ActivityPub `Public` URI will be addressed in `to`. Otherwise, only mentioned and replied to users will be addressed.

In the following example, the 'admin' user deletes a public post of theirs in which the 'foss_satan' user was mentioned:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://example.org/users/admin",
  "cc": [
    "http://example.org/users/admin/followers",
    "http://fossbros-anonymous.io/users/foss_satan"
  ],
  "object": "http://example.org/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}
```

In the next example, the '1happyturtle' user deletes a direct message which was originally addressed to the 'the_mighty_zork' user.

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://example.org/users/1happyturtle",
  "cc": [],
  "object": "http://example.org/users/1happyturtle/statuses/01FN3VJGFH10KR7S2PB0GFJZYG",
  "to": "http://somewhere.com/users/the_mighty_zork",
  "type": "Delete"
}
```

To process a `Delete` activity coming from a GoToSocial instance, remote instances should check if they have the `Object` stored according to the provided URI. If they do, they should remove it from their local cache. If not, then no action is required, since they never had the now-deleted post stored in the first place.

### Incoming

GoToSocial processes `Delete` activities coming in from remote instances as thoroughly as possible in order to respect the privacy of other users.

When a GoToSocial instance receives a `Delete`, it will attempt to derive the deleted post URI from the `Object` field. If the `Object` is just a URI, then this URI will be taken. If the `Object` is a `Note` or another type commonly used to represent a post, then the URI will be extracted from it.

Then, GoToSocial will check if it has a post stored with the given URI. If it does, it will be completely deleted from the database and all user timelines.

GoToSocial will only delete a post if it can be sure that the original post was owned by the `actor` that the `Delete` is attributed to.

## Profile Fields

Like Mastodon and other fediverse softwares, GoToSocial lets users set key/value pairs on their profile; useful for conveying short pieces of information like links, pronouns, age, etc.

For the sake of compatibility with other implementations, GoToSocial uses the same schema.org PropertyValue extension that Mastodon uses, present as an `attachment` array value on `actor`s that have fields set. For example, the below JSON shows an account with two PropertyValue fields:

```json
{
  "@context": [
    "http://joinmastodon.org/ns",
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    "http://schema.org"
  ],
  "attachment": [
    {
      "name": "should you follow me?",
      "type": "PropertyValue",
      "value": "maybe!"
    },
    {
      "name": "age",
      "type": "PropertyValue",
      "value": "120"
    }
  ],
  "discoverable": false,
  "featured": "http://example.org/users/1happyturtle/collections/featured",
  "followers": "http://example.org/users/1happyturtle/followers",
  "following": "http://example.org/users/1happyturtle/following",
  "id": "http://example.org/users/1happyturtle",
  "inbox": "http://example.org/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://example.org/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {
    "id": "http://example.org/users/1happyturtle#main-key",
    "owner": "http://example.org/users/1happyturtle",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtTc6Jpg6LrRPhVQG4KLz\n2+YqEUUtZPd4YR+TKXuCnwEG9ZNGhgP046xa9h3EWzrZXaOhXvkUQgJuRqPrAcfN\nvc8jBHV2xrUeD8pu/MWKEabAsA/tgCv3nUC47HQ3/c12aHfYoPz3ufWsGGnrkhci\nv8PaveJ3LohO5vjCn1yZ00v6osMJMViEZvZQaazyE9A8FwraIexXabDpoy7tkHRg\nA1fvSkg4FeSG1XMcIz2NN7xyUuFACD+XkuOk7UqzRd4cjPUPLxiDwIsTlcgGOd3E\nUFMWVlPxSGjY2hIKa3lEHytaYK9IMYdSuyCsJshd3/yYC9LqxZY2KdlKJ80VOVyh\nyQIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://example.org/@1happyturtle"
}
```

For `actor`s that have no `PropertyValue` fields set, the `attachment` property will not be set at all. That is, the `attachment` key value will not be present on the `actor` (not even as an empty array or null value).

While `attachment` is not technically an ordered collection, GoToSocial--again, in line with what other implementations do--does present `attachment` `PropertyValue` fields in the order in which they should to be displayed.

GoToSocial will also parse PropertyValue fields from remote `actor`s discovered by the GoToSocial instance, to allow them to be displayed to users on the GoToSocial instance.

GoToSocial allows up to 6 `PropertyValue` fields by default, as opposed to Mastodon's default 4.

## Hashtags

GoToSocial users can include hashtags in their posts, which indicate to other instances that that user wishes their post to be grouped together with other posts using the same hashtag, for discovery purposes.

In line with other ActivityPub server implementations, GoToSocial implicitly expects that only public-addressed posts will be grouped by hashtag.

To federate hashtags in and out, GoToSocial uses the widely-adopted [ActivityStreams `Hashtag` type extension](https://www.w3.org/wiki/Activity_Streams_extensions#as:Hashtag_type) in the `tag` property of objects.

Here's what the `tag` property might look like on an outgoing message that uses one custom emoji, and one tag:

```json
"tag": [
  {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
    },
    "id": "https://example.org/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
    "name": ":rainbow:",
    "type": "Emoji",
    "updated": "2021-09-20T10:40:37Z"
  },
  {
    "href": "https://example.org/tags/welcome",
    "name": "#welcome",
    "type": "Hashtag"
  }
]
```

With just one tag, the `tag` property will be an object rather than an array, which will look like this:

```json
"tag": {
  "href": "https://example.org/tags/welcome",
  "name": "#welcome",
  "type": "Hashtag"
}
```

### Hashtag `href` property

The `href` URL provided by GoToSocial in outgoing tags points to a web URL that serves `text/html`.

GoToSocial makes no guarantees whatsoever about what the content of the given `text/html` will be, and remote servers should not interpret the URL as a canonical ActivityPub ID/URI property. The `href` URL is provided merely as an endpoint which *might* contain more information about the given hashtag.

## Mentions

GoToSocial users can Mention other users in their posts, using the common `@[username]@[domain]` format. For example, if a GoToSocial user wanted to mention user `someone` on instance `example.org`, they could do this by including `@someone@example.org` in their post somewhere.

!!! info "Mentions and activity addressing"
    
    Mentions are not just aesthetic, they affect addressing of Activities as well.
    
    If a GoToSocial user explicitly mentions another user in a Note, the URI of that user will always be included in the `To` or `Cc` property of the Note's Create activity.
    
    If the Note is direct (ie., not `To` public or followers), each mention target URI will be in the `To` property of the wrapping Activity
    
    In all other cases, mentions will be included in the `Cc` property of the wrapping Activity.

### Outgoing

When a GoToSocial user Mentions another account, the Mention is included in outgoing federated messages as an entry in the `tag` property.

For example, say a user on a GoToSocial instance Mentions `@someone@example.org`, the `tag` property of the outgoing Note might look like the following:

```json
"tag": {
  "href": "http://example.org/users/someone",
  "name": "@someone@example.org",
  "type": "Mention"
}
```

If a user Mentions a local user they share an instance with, the full `name` of the local user will still be included.

For example, a GoToSocial user on domain `some.gotosocial.instance` mentions another user on the same instance called `user2`. They also mention `@someone@example.org` as above. The `tag` property of the outgoing Note would look like the following:

```json
"tag": [
  {
    "href": "http://example.org/users/someone",
    "name": "@someone@example.org",
    "type": "Mention"
  },
  {
    "href": "http://some.gotosocial.instance/users/user2",
    "name": "@user2@some.gotosocial.instance",
    "type": "Mention"
  }
]
```

For the convenience of remote servers, GoToSocial will always provide both the `href` and the `name` properties on outgoing Mentions. The `href` property used by GoToSocial will always be the ActivityPub ID/URI of the target account, not the web URL.

### Incoming

GoToSocial tries to parse incoming Mentions in the same way it sends them out: as a `Mention` type entry in the `tag` property. However, when parsing incoming Mentions it's a bit more relaxed with regards to which properties must be set.

GoToSocial will prefer the `href` property, which can be either the ActivityPub ID/URI or the web URL of the target; if `href` is not present, it will fall back to using the `name` property. If neither property is present, the mention will be considered invalid and discarded.

## Content, ContentMap, and Language

In line with other ActivityPub implementations, GoToSocial uses `content` and `contentMap` fields on `Objects` to infer content and language of incoming posts, and to set content and language on outgoing posts.

### Outgoing

If an outgoing `Object` (usually a `Note`) has content, it will be set as stringified HTML on the `content` field.

If the `content` is in a specific user-selected language, then the `Object` will also have the `contentMap` property set to a single-entry key/value map, where the key is a BCP47 language tag, and the value is the same content from the `content` field.

For example, a post written in English (`en`) will look something like this:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Note",
  "attributedTo": "http://example.org/users/i_p_freely",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "cc": "http://example.org/users/i_p_freely/followers",
  "id": "http://example.org/users/i_p_freely/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "url": "http://example.org/@i_p_freely/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "published": "2021-11-20T13:32:16Z",
  "content": "<p>This is an example note.</p>",
  "contentMap": {
    "en": "<p>This is an example note.</p>"
  },
  "attachment": [],
  "replies": {...},
  "sensitive": false,
  "summary": "",
  "tag": {...}
}
```

GoToSocial will always set the `content` field if the post has content, but it may not always set the `contentMap` field, if an old version of GoToSocial is in use, or the language used by a user is not set or not a recognized BCP47 language tag.

### Incoming

GoToSocial uses both the `content` and the `contentMap` properties on incoming `Object`s to determine the content and infer the intended "primary" language for that content. It uses the following algorithm:

#### Only `content` is set

Take that content only and mark language as unknown.

#### Both `content` and `contentMap` are set

Look for a language tag as key in the `contentMap`, with a value that matches the stringified HTML set in `content`.

If a match is found, use this as the post's language.

If a match is not found, keep content from `content` and mark language as unknown.

#### Only `contentMap` is set

If `contentMap` has only one entry, take the language tag and content value as the "primary" language and content.

If `contentMap` has multiple entries, we have no way of determining the intended preferred content and language of the post, since map order is not deterministic. In this case, try to pick a language and content entry that matches one of the languages configured in the GoToSocial instance's [configured languages](../configuration/instance.md). If no language can be matched this way, pick a language and content entry from the `contentMap` at random as the "primary" language and content.

!!! Note
    In all of the above cases, if the inferred language cannot be parsed as a valid BCP47 language tag, language will fall back to unknown.

## Polls

To federate polls in and out, GoToSocial uses the widely-adopted [ActivityStreams `Question` type](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-question). This however, as first introduced and popularised by Mastodon, does slightly vary from the ActivityStreams specification. In the specification the Question type is marked as an extension of "IntransitiveActivity", an "Activity" extension that should be passed without an "Object" and all further details contained implicitly. But in implementation it is passed as an "Object", as part of "Create" or "Update" activities.

It is also worth noting that while GoToSocial internally may treat a poll as a type of status attachment, the ActivityStreams representation treats statuses and statuses-with-polls as 2 different "Object" types. Statuses are federated as "Note" types, and polls as "Question" types.

The "Question" type that GoToSocial transmits (and expects to receive) contain all the typical expected "Note" properties, with a few additions. They expect the following additional (pseudo-)JSON:

```json
{
  "@context":[
    {
      // toot:votersCount extension which is
      // used to add the votersCount property.
      "toot":"http://joinmastodon.org/ns#",
      "votersCount":"toot:votersCount"
    },
  ],

  // oneOf / anyOf contains the the poll options
  // themselves. Only one of the two will be set,
  // where "oneOf" indicates a single-choice poll
  // and "anyOf" indicates multiple-choice.
  //
  // Either property contains an array of "Notes",
  // special in that they contain a "name" and unset
  // "content", where the "name" represents the actual
  // poll option string. Additionally they contain
  // a "replies" property as a "Collection" type,
  // which represents currently known vote counts
  // for each poll option via "totalItems".
  "oneOf": [ // or "anyOf"
    {
      "type": "Note",
      "name": "option 1",
      "replies": {
        "type": "Collection",
        "totalItems": 0
      }
    },
    {
      "type": "Note",
      "name": "option 2",
      "replies": {
        "type": "Collection",
        "totalItems": 0
      }
    }
  ],

  // endTime indicates the date at which this
  // poll will close. Some server implementations
  // support endless polls, or use "closed" to
  // imply "endTime", so may not always be set.
  "endTime": "2023-01-01T20:04:45Z",

  // closed indicates the date at which this
  // poll closed. Will be unset until this time.
  "closed": "2023-01-01T20:04:45Z",

  // votersCount indicates the total number of
  // participants, which is useful in the case
  // of multiple choice polls.
  "votersCount": 10
}
```

### Outgoing

You can expect to receive a poll from GoToSocial in the form of a "Question", passed as the object property in either a "Create" or "Update" activity. In the case of an "Update" activity, if anything in the poll but the "votersCount", "replies.totalItems" or "closed" has changed, this indicates that the wrapping status was edited in a way that requires the attached poll to be recreated, and thus, reset. You can expect to receive these activities at the following times:

- "Create": the status with attached poll was just created

- "Update": the poll vote / voter counts have changed, or the poll has just ended

The JSON you can expect from a GoToSocial generated "Question" can be seen in the section above's pseudo-JSON. Following from this the "endTime" field will always be set, (as we do not support creating endless polls), and the "closed" field will only be set when the poll has closed.

### Incoming

GoToSocial expects to receive polls in largely the same manner that it sends them out, with a little more leniency when it comes to parsing the "Question" object.

- if "closed" is provided without an "endTime", this will also be taken as the value for "endTime"

- if neither "closed" nor "endTime" is provided, the poll is assumed to be endless

- any time an "Update" activity with "Question" provides a "closed" time, when there was previously none, the poll will be assumed to have just closed. this triggers client notifications to our local voting users

## Poll Votes

To federate poll votes in and out, GoToSocial uses a specifically formatted version of the [ActivityStreams "Note" type](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-note). This is widely accepted by ActivityPub servers as the way to federate poll votes, only ever attached as an "Object" to "Create" activities.

The "Note" type that GoToSocial transmits (and expects to receive) specifically contains:
- "name": [exact poll option text]
- "content": [left unset]
- "inReplyTo": [IRI of AS Question]
- "attributedTo": [IRI of vote author]
- "to": [IRI of poll author]

For example:

```json
{
  "type": "Note",
  "name": "Option 1",
  "inReplyTo": "https://example.org/users/bobby_tables/statuses/123456",
  "attributedTo": "https://sample.com/users/willy_nilly",
  "to": "https://example.org/users/bobby_tables"
}
```

### Outgoing

You can expect to receive poll votes from GoToSocial in the form of "Note" objects, as specifically described in the section above. These will only ever be sent out as the object(s) attached to a "Create" activity.

In particular, as described in the section above, GoToSocial will provide the option text in the "name" field, the "content" field unset, and the "inReplyTo" field being an IRI pointing toward a status with poll authored on your instance.

Here's an example of a "Create", in which user "https://sample.com/users/willy_nilly" votes on a multiple-choice poll created by user "https://example.org/users/bobby_tables":

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://sample.com/users/willy_nilly",
  "id": "https://sample.com/users/willy_nilly/activity#vote/https://example.org/users/bobby_tables/statuses/123456",
  "object": [
    {
      "attributedTo": "https://sample.com/users/willy_nilly",
      "id": "https://sample.com/users/willy_nilly#01HEN2R65468ZG657C4ZPHJ4EX/votes/1",
      "inReplyTo": "https://example.org/users/bobby_tables/statuses/123456",
      "name": "tissues",
      "to": "https://example.org/users/bobby_tables",
      "type": "Note"
    },
    {
      "attributedTo": "https://sample.com/users/willy_nilly",
      "id": "https://sample.com/users/willy_nilly#01HEN2R65468ZG657C4ZPHJ4EX/votes/2",
      "inReplyTo": "https://example.org/users/bobby_tables/statuses/123456",
      "name": "financial times",
      "to": "https://example.org/users/bobby_tables",
      "type": "Note"
    }
  ],
  "published": "2021-09-11T11:45:37+02:00",
  "to": "https://example.org/users/bobby_tables",
  "type": "Create"
}
```

### Incoming

GoToSocial expects to receive poll votes in much the same manner that it sends them out. They will only ever expect to be received as part of a "Create" activity.

In particular, GoToSocial recognizes votes as different to other "Note" objects by the inclusion of a "name" field, missing "content" field, and the "inReplyTo" field being an IRI pointing to a status with attached poll. If any of these conditions are not met, GoToSocial will consider the provided "Note" to be a malformed status object.
