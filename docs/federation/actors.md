# Actors and Actor Properties

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

## Followers / Following Collections

GoToSocial implements followers and following collections as `OrderedCollection`s. A properly-signed `GET` request to an Actor's Following collection, for example, will return something like:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "first": "https://example.org/users/someone/following?limit=40",
  "id": "https://example.org/users/someone/following",
  "totalItems": 397,
  "type": "OrderedCollection"
}
```

From there, you can use the `first` page to start getting items. For example, a `GET` request to `https://example.org/users/someone/following?limit=40` will produce something like:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/someone/following?limit=40",
  "next": "https://example.org/users/someone/following?limit=40&max_id=01V1AY4ZJT4JK1NT271SH2WMGH",
  "orderedItems": [
    "https://example.org/users/someone_else",
    "https://somewhere.else.example.org/users/another_account",
    [... 38 more entries here ...]
  ],
  "partOf": "https://example.org/users/someone/following",
  "prev": "https://example.org/users/someone/following?limit=40&since_id=021HKBY346X7BPFYANPPJN493P",
  "totalItems": 397,
  "type": "OrderedCollectionPage"
}
```

You can then use the `next` and `prev` endpoints to page down and up through the OrderedCollection.

!!! Info "Hidden Followers / Following Collections"
    
    GoToSocial allows users to hide their followers/following collections if they wish.
    
    If a user has chosen to hide their collections, then only a stub collection with `totalItems` will be returned, and you will not be able to page through the Actor's followers/following collections.
    
    A `GET` to the following collection of an Actor with hidden collections will look like:
    
    ```json
    {
      "@context": "https://www.w3.org/ns/activitystreams",
      "id": "https://example.org/users/someone/following",
      "type": "OrderedCollection",
      "totalItems": 397
    }
    ```

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

## Actor Migration / Aliasing

GoToSocial supports account migration from one instance/server to another through a combination of the `Move` activity, and the Actor Object properties `alsoKnownAs` and `movedTo`.

### `alsoKnownAs`

GoToSocial supports account aliasing using the `alsoKnownAs` Actor property, which is an [accepted ActivityPub extension](https://www.w3.org/wiki/Activity_Streams_extensions#as:alsoKnownAs_property).

#### Incoming

On incoming AP messages, GoToSocial looks for the `alsoKnownAs` property on an Actor to be an array of ActivityPub IDs/URIs of other Actors by which the Actor is also known.

For example:

```json
{
  "@context": [
    "http://joinmastodon.org/ns",
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    "http://schema.org"
  ],
  "featured": "http://example.org/users/1happyturtle/collections/featured",
  "followers": "http://example.org/users/1happyturtle/followers",
  "following": "http://example.org/users/1happyturtle/following",
  "id": "http://example.org/users/1happyturtle",
  "inbox": "http://example.org/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://example.org/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {...},
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "type": "Person",
  "url": "http://example.org/@1happyturtle",
  "alsoKnownAs": [
    "https://another-server.com/users/1happyturtle",
    "https://somewhere-else.org/users/originalTurtle"
  ]
}
```

In the above AP JSON, the Actor `http://example.org/users/1happyturtle` is aliased to the other Actors `https://another-server.com/users/1happyturtle` and `https://somewhere-else.org/users/originalTurtle`.

GoToSocial will store incoming `alsoKnownAs` URIs in the database, but does not (currently) use them for anything except verifying a `Move` Activity (see below).

#### Outgoing

GoToSocial users can set multiple `alsoKnownAs` URIs on their account via the GoToSocial client API. GoToSocial will verify that these `alsoKnownAs` aliases are valid Actor URIs before storing them in the database and before serializing them in outgoing AP messages.

However, GoToSocial does not verify *ownership* of those `alsoKnownAs` URIs by the user setting the aliases before serializing them in outgoing messages; it expects remote servers to do their own verification before trusting any transmitted `alsoKnownAs` values.

As an example, the user `http://example.org/users/1happyturtle`, from their GoToSocial instance, might set `alsoKnownAs: [ "https://unrelated-server.com/users/someone_else" ]` on their account, and GoToSocial will duly transmit this alias to other servers.

In this case, though, `https://unrelated-server.com/users/someone_else` may not be the same person as `1happyturtle`. `1happyturtle` may have set this alias by mistake, or maliciously. To properly verify ownership of `someone_else` by `1happyturtle`, a remote server should check that the `alsoKnownAs` property of the Actor `https://unrelated-server.com/users/someone_else` contains an entry `http://example.org/users/1happyturtle`.

In other words, remote servers should not trust `alsoKnownAs` aliases by default, and should instead ensure that a **two-way alias** exists between Actors before treating the alias as valid.

!!! info
    The reason that GoToSocial does not perform verification of `alsoKnownAs` values before sending them out to other servers is to avoid a chicken and egg problem. Say that `1happyturtle` and `someone_else` *are* the same person, one of the two Actors must be able to set `alsoKnownAs` first, so that the instance of the other Actor can begin processing the alias. If both servers prevent an unverified alias from being serialized in the `alsoKnownAs` property, then it becomes impossible for either `1happyturtle` or `someone_else` to alias to one another.

### `movedTo`

GoToSocial marks accounts as moved using the `movedTo` property. Unlike `alsoKnownAs` this is not an accepted ActivityPub extension, but it has been widely popularized by Mastodon, which also uses it in connection with the `Move` activity. [See the Mastodon docs for more info](https://documentation.sig.gy/spec/activitypub/#namespaces).

#### Incoming

For incoming AP messages, GoToSocial looks for the `movedTo` property on an Actor to be set to a single ActivityPub Actor URI/ID.

For example:

```json
{
  "@context": [
    "http://joinmastodon.org/ns",
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    "http://schema.org"
  ],
  "featured": "http://example.org/users/1happyturtle/collections/featured",
  "followers": "http://example.org/users/1happyturtle/followers",
  "following": "http://example.org/users/1happyturtle/following",
  "id": "http://example.org/users/1happyturtle",
  "inbox": "http://example.org/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://example.org/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {...},
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "type": "Person",
  "url": "http://example.org/@1happyturtle",
  "alsoKnownAs": [
    "https://another-server.com/users/1happyturtle"
  ],
  "movedTo": "https://another-server.com/users/1happyturtle"
}
```

In the above JSON, the Actor `http://example.org/users/1happyturtle` has been aliased to the Actor `https://another-server.com/users/1happyturtle` and has also moved/migrated to that account.

GoToSocial stores incoming `movedTo` values in the database, but does not consider an account migration to have been processed unless the Actor doing the Move had previously transmitted a Move activity (see below).

#### Outgoing

GoToSocial will only set `movedTo` on outgoing Actors when an account `Move` has been verified and processed.

### `Move` Activity

To actually trigger account migrations, GoToSocial uses the `Move` Activity with Actor URI as Object and Target, for example:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/1happyturtle/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
  "actor": "https://example.org/users/1happyturtle",
  "type": "Move",
  "object": "https://example.org/users/1happyturtle",
  "target": "https://another-server.com/users/my_new_account_hurray",
  "to": "https://example.org/users/1happyturtle/followers"
}
```

In the above `Move`, Actor `https://example.org/users/1happyturtle` indicates that their account is moving to the URI `https://another-server.com/users/my_new_account_hurray`.

#### Incoming

On receiving a `Move` activity in an Actor's Inbox, GoToSocial will first validate the `Move` by making the following checks:

1. Request was signed by `actor`.
2. `actor` and `object` fields are the same (you can't `Move` someone else's account).
3. `actor` has not already moved somewhere else.
4. `target` is a valid Actor URI: retrievable, not suspended, not already moved, and on a domain that's not defederated by the GoToSocial instance that received the `Move`.
5. `target` has `alsoKnownAs` set to the `actor` that sent the `Move`. In this example, `https://another-server.com/users/my_new_account_hurray` must have an `alsoKnownAs` value that includes `https://example.org/users/1happyturtle`.

If checks pass, then GoToSocial will process the `Move` by redirecting followers to the new account:

1. Select all followers on this GtS instance of the `actor` doing the `Move`.
2. For each local follower selected in this way, send a follow request from that follower to the `target` of the `Move`.
3. Remove all follows targeting the "old" `actor`.

The end result of this is that all followers of `https://example.org/users/1happyturtle` on the receiving instance will now be following `https://another-server.com/users/my_new_account_hurray` instead.

GoToSocial will also remove all follow and pending follow requests owned by the `actor` doing the `Move`; it's up to the `target` account to send follow requests out again.

To prevent potential DoS vectors, GoToSocial enforces a 7-day cooldown on `Move`s. Once an account has successfully moved, GoToSocial will not process further moves from the new account until 7 days after the previous move.

#### Outgoing

Outgoing account migrations use the `Move` Activity in much the same way. When an Actor on a GoToSocial instance wants to `Move`, GtS will first check and validate the `Move` target, and ensure it has an `alsoKnownAs` entry equal to the Actor doing the `Move`. On successful validation, a `Move` message will be sent out to all of the moving Actor's followers, indicating the `target` of the Move. GoToSocial expects remote instances to transfer the `actor`'s followers to the `target`.
