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
