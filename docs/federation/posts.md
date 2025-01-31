# Posts and Post Properties

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

## Emojis

GoToSocial uses the `http://joinmastodon.org/ns#Emoji` type to allow users to add custom emoji to their posts.

For example:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "type": "Note",
  "content": "<p>here's a stinky creature -> :shocked_pikachu:</p>",
  [...],
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png"
    },
    "id": "https://example.org/emoji/01AZY1Y5YQD6TREB5W50HGTCSZ",
    "name": ":shocked_pikachu:",
    "type": "Emoji",
    "updated": "2022-11-17T11:36:05Z"
  }
  [...]
}
```

The text `:shocked_pikachu:` in the `content` of the above `Note` should be replaced by clients with a small (inline) version of the emoji image, when rendering the `Note` and displaying it to users.

The `updated` and `icon.url` properties of the emoji can be used by remote instances to determine whether their representation of the GoToSocial emoji image is up to date.

The `Emoji` can also be dereferenced at its `id` URI if necessary, so that remotes can check whether their cached version of the emoji metadata is up to date.

By default, GoToSocial sets a 50kb limit on the size of emoji images that can be uploaded and sent out, and a 100kb limit on the size of emoji images that can be federated in, though both of these settings are configurable by users.

GoToSocial can send and receive emoji images of the type `image/png`, `image/jpeg`, `image/gif`, and `image/webp`.

!!! info
    Note that the `tag` property can be either an array of objects, or a single object.

### `null` / empty `id` property

Some server softwares like Akkoma include emojis as [anonymous objects](https://www.w3.org/TR/activitypub/#obj-id) on statuses. That is, they set the `id` property to the value `null` to indicate that the emoji cannot be dereferenced at any specific endpoint.

When receiving such emojis, GoToSocial will generate a dummy id for that emoji in its database in the form `https://[host]/dummy_emoji_path?shortcode=[shortcode]`, for example, `https://example.org/dummy_emoji_path?shortcode=shocked_pikachu`.

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

## Interaction Policy

GoToSocial uses the property `interactionPolicy` on posts, in order to indicate to remote instances what sort of interactions are (conditionally) permitted to be processed and stored by the origin server, for any given post.

For more details on this, see the separate [interaction policy](./interaction_policy.md) document.

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

## Conversation Threads

Due to the nature of decentralization and federation, it is practically impossible for any one server on the fediverse to be aware of every post in a given conversation thread.

With that said, it is possible to do 'best effort' dereferencing of threads, whereby remote replies are fetched from one server onto another, to try to more fully flesh out a conversation.

GoToSocial does this by iterating up and down the thread of a conversation, pulling in remote statuses where possible.

Let's say we have two accounts: `local_account` on `our.server`, and `remote_1` on `remote.1`.

In this scenario, `local_account` follows `remote_1`, so posts from `remote_1` show up in the home timeline of `local_account`.

Now, `remote_1` boosts/reblogs a post from a third account, `remote_2`, residing on server `remote.2`.

`local_account` does not follow `remote_2`, and neither does anybody else on `our.server`, which means that `our.server` has not seen this post by `remote_2` before.

![A diagram of the conversation thread, showing the post from remote_2, and possible ancestor and descendant posts](../public/diagrams/conversation_thread.png)

What GoToSocial will do now, is 'dereference' the post by `remote_2` to check if it is part of a thread and, if so, whether any other parts of the thread can be obtained.

GtS begins by checking the `inReplyTo` property of the post, which is set when a post is a reply to another post. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-inreplyto). If `inReplyTo` is set, GoToSocial derefences the replied-to post. If *this* post also has an `inReplyTo` set, then GoToSocial dereferences that too, and so on.

Once all of these **ancestors** of a status have been retrieved, GtS will begin working down through the **descendants** of posts.

It does this by checking the `replies` property of a derefenced post, and working through replies, and replies of replies. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-replies).

This process of thread dereferencing will likely involve making multiple HTTP calls to different servers, especially if the thread is long and complicated.

The end result of this dereferencing is that, assuming the reblogged post by `remote_2` was part of a thread, then `local_account` should now be able to see posts in the thread when they open the status on their home timeline. In other words, they will see replies from accounts on other servers (who they may not have come across yet), in addition to any previous and next posts in the thread as posted by `remote_2`.

This gives `local_account` a more complete view on the conversation, as opposed to just seeing the reblogged post in isolation and out of context. It also gives `local_account` the opportunity to discover new accounts to follow, based on replies to `remote_2`.
