# ActivityPub Outbox

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
