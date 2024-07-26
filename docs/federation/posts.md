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

GoToSocial uses the property `interactionPolicy` on posts in order to indicate to remote instances what sort of interactions will be (conditionally) permitted for any given post.

!!! danger
    
    Interaction policy is an attempt to limit the harmful effects of unwanted replies and other interactions on a user's posts (eg., "reply guys").
    
    However, it is far from being sufficient for this purpose, as there are still many "out-of-band" ways that posts can be distributed or replied to beyond a user's initial wishes or intentions.
    
    For example, a user might create a post with a very strict interaction policy attached to it, only to find that other server softwares do not respect that policy, and users on other instances are having discussions and replying to the post *from their instance's perspective*. The original poster's instance will automatically drop these unwanted interactions from their view, but remote instances may still show them.
    
    Another example: someone might see a post that specifies nobody can reply to it, but screenshot the post, post the screenshot in their own new post, and tag the original post author in as a mention. Alternatively, they might just link to the URL of the post and tag the author in as a mention. In this case, they effectively "reply" to the post by creating a new thread.
    
    For better or worse, GoToSocial can offer only a best-effort, partial, technological solution to what is more or less an issue of social behavior and boundaries.

### Overview

`interactionPolicy` is a property attached to the status-like `Object`s `Note`, `Article`, `Question`, etc, with the following format:

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [ "zero_or_more_uris_that_can_always_do_this" ],
      "approvalRequired": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    },
    "canReply": {
      "always": [ "zero_or_more_uris_that_can_always_do_this" ],
      "approvalRequired": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    },
    "canAnnounce": {
      "always": [ "zero_or_more_uris_that_can_always_do_this" ],
      "approvalRequired": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    }
  },
  [...]
}
```

In this object:

- `canLike` indicates who can create a `Like` with the post URI as the `Object` of the `Like`.
- `canReply` indicates who can create a post with `inReplyTo` set to the URI of the post.
- `canAnnounce` indicates who can create an `Announce` with the post URI as the `Object` of the `Announce`. 

And:

- `always` is an array of ActivityPub URIs/IDs of `Actor`s or `Collection`s of `Actor`s who do not require an `Accept` in order to distribute an interaction to their followers (more on this below).
- `approvalRequired` is an array of ActivityPub URIs/IDs of `Actor`s or `Collection`s of `Actor`s who can interact, but should wait for an `Accept` before distributing an interaction to their followers.

Valid URI entries in `always` and `approvalRequired` include the magic ActivityStreams Public URI `https://www.w3.org/ns/activitystreams#Public`, the URIs of the post creator's `Following` and/or `Followers` collections, and individual Actor URIs. For example:

```json
[
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/someone/followers",
    "https://example.org/users/someone/following",
    "https://example.org/users/someone_else",
    "https://somewhere.else.example.org/users/someone_on_a_different_instance"
]
```

### Specifying Nobody

!!! note
    GoToSocial makes implicit assumptions about who can/can't interact, even if a policy specifies nobody. See [implicit assumptions](#implicit-assumptions).

An empty array, or a missing or null key, indicates that nobody can do the interaction.

For example, the following `canLike` value indicates that nobody can `Like` the post:

```json
"canLike": {
  "always": [],
  "approvalRequired": []
},
```

Likewise, a `canLike` value of `null` also indicates that nobody can `Like` the post:

```json
"canLike": null
```

or

```json
"canLike": {
  "always": null,
  "approvalRequired": null
}
```

And a missing `canLike` value does the same thing:

```json
{
  [...],
  "interactionPolicy": {
    "canReply": {
      "always": [ "zero_or_more_uris_that_can_always_do_this" ],
      "approvalRequired": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    },
    "canAnnounce": {
      "always": [ "zero_or_more_uris_that_can_always_do_this" ],
      "approvalRequired": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    }
  },
  [...]
}
```

### Conflicting / Duplicate Values

In cases where a user is present in a Collection URI, and is *also* targeted explicitly by URI, the **more specific value** takes precedence.

For example:

```json
[...],
"canReply": {
  "always": [
    "https://example.org/users/someone"
  ],
  "approvalRequired": [
    "https://www.w3.org/ns/activitystreams#Public"
  ]
},
[...]
```

Here, `@someone@example.org` is present in the `always` array, and is also implicitly present in the magic ActivityStreams Public collection in the `approvalRequired` array. In this case, they can always reply, as the `always` value is more explicit.

Another example:

```json
[...],
"canReply": {
  "always": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "approvalRequired": [
    "https://example.org/users/someone"
  ]
},
[...]
```

Here, `@someone@example.org` is present in the `approvalRequired` array, but is also implicitly present in the magic ActivityStreams Public collection in the `always` array. In this case everyone can reply without approval, **except** for `@someone@example.org`, who requires approval.

In case the **exact same** URI is present in both `always` and `approvalRequired`, the **highest level of permission** takes precedence (ie., a URI in `always` takes precedence over the same URI in `approvalRequired`).

### Implicit Assumptions

GoToSocial makes several implicit assumptions about `interactionPolicy`s.

**Firstly**, users [mentioned](#mentions) in, or replied to by, a post should **ALWAYS** be able to reply to that post without requiring approval, regardless of the post visiblity and the `interactionPolicy`, **UNLESS** the post that mentioned or replied to them is itself currently pending approval.

This is to prevent a would-be harasser from mentioning someone in an abusive post, and leaving no recourse to the mentioned user to reply.

As such, when sending out interaction policies, GoToSocial will **ALWAYS** add the URIs of mentioned users to the `canReply.always` array, unless they are already covered by the ActivityStreams magic public URI.

Likewise, when enforcing received interaction policies, GoToSocial will **ALWAYS** behave as though the URIs of mentioned users were present in the `canReply.always` array, even if they weren't.

**Secondly**, a user should **ALWAYS** be able to reply to their own post, like their own post, and boost their own post without requiring approval, **UNLESS** that post is itself currently pending approval.

As such, when sending out interaction policies, GoToSocial will **ALWAYS** add the URI of the post author to the `canLike.always`, `canReply.always`, and `canAnnounce.always` arrays, unless they are already covered by the ActivityStreams magic public URI.

Likewise, when enforcing received interaction policies, GoToSocial will **ALWAYS** behave as though the URI of the post author is present in these `always` arrays, even if it wasn't.

### Defaults

When the `interactionPolicy` property is not present at all on a post, GoToSocial assumes a default `interactionPolicy` for that post appropriate to the visibility level of the post, and the post author.

For a **public** or **unlocked** post by `@someone@example.org`, the default `interactionPolicy` is:

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

For a **followers-only** post by `@someone@example.org`, the assumed `interactionPolicy` is:

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://example.org/users/someone",
        "https://example.org/users/someone/followers",
        [...URIs of any mentioned users...]
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/someone",
        "https://example.org/users/someone/followers",
        [...URIs of any mentioned users...]
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/someone"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

For a **direct** post by `@someone@example.org`, the assumed `interactionPolicy` is:

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://example.org/users/someone",
        [...URIs of any mentioned users...]
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/someone",
        [...URIs of any mentioned users...]
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/someone"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

### Example 1 - Limiting scope of a conversation

In this example, the user `@the_mighty_zork` wants to begin a conversation with the users `@booblover6969` and `@hodor`.

To avoid the discussion being derailed by others, they want replies to their post by users other than the three participants to be permitted only if they're approved by `@the_mighty_zork`.

Furthermore, they want to limit the boosting / `Announce`ing of their post to only their own followers, and to the three conversation participants.

However, anyone should be able to `Like` the post by `@the_mighty_zork`.

This can be achieved with the following `interactionPolicy`, which is attached to a post with visibility level public:

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ],
      "approvalRequired": [
        "https://www.w3.org/ns/activitystreams#Public"
      ]
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/the_mighty_zork/followers",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

### Example 2 - Long solo thread

In this example, the user `@the_mighty_zork` wants to write a long solo thread.

They don't mind if people boost and like posts in the thread, but they don't want to get any replies because they don't have the energy to moderate the discussion; they just want to vent by throwing their thoughts out there.

This can be achieved by setting the following `interactionPolicy` on every post in the thread:

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/the_mighty_zork"
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

Here, anyone is allowed to like or boost, but nobody is permitted to reply (except `@the_mighty_zork` themself).

### Example 3 - Completely open

In this example, `@the_mighty_zork` wants to write a completely open post that can be replied to, boosted, or liked by anyone who can see it (ie., the default behavior for unlocked and public posts):

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

### Requesting, Obtaining, and Validating Approval

When a user's URI is in the `approvalRequired` array for a type of interaction, and that user wishes to obtain approval to distribute an interaction, they should do the following:

1. Compose the interaction `Activity` (ie., `Like`, `Create` (reply), or `Announce`), as normal.
2. Address the `Activity` `to` and `cc` the expected recipients for that `Activity`, as normal.
3. `POST` the `Activity` only to the `Inbox` (or `sharedInbox`) of the author of the post being interacted with.
4. **DO NOT DISTRIBUTE THE ACTIVITY FURTHER THAN THIS AT THIS POINT**.

At this point, the interaction can be considered as pending approval, and should not be shown in the replies or likes collections, etc., of the post interacted with.

It may be shown to the user who sent the interaction as a sort of "interaction pending" modal, but ideally it should not be shown to other users who share an instance with that user.

From here, one of three things may happen:

#### Rejection

In this scenario, the author of the post being interacted with sends back a `Reject` `Activity` with the URI/ID of the interaction `Activity` as the `Object` property.

For example, the following json object `Reject`s the attempt of `@someone@somewhere.else.example.org` to reply to a post by `@post_author@example.org`:

```json
{
  "actor": "https://example.org/users/post_author",
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "type": "Reject"
}
```

If this happens, `@someone@somewhere.else.example.org` (and their instance) should consider the interaction as having been rejected. The instance should delete the activity from its internal storage (ie., database), or otherwise indicate that it's been rejected, and it should not distribute the `Activity` further, or retry the interaction.

#### Nothing

In this scenario, the author of the post being interacted with never sends back a `Reject` or an `Accept` `Activity`. In such a case, the interaction is considered "pending" in perpetuity. Instances may wish to implement some kind of cleanup feature, where sent and pending interactions that reach a certain age should be considered expired, and `Rejected` and then removed in the manner gestured towards above.

#### Acceptance

In this scenario, the author of the post being interacted with sends back an `Accept` `Activity` with the URI/ID of the interaction `Activity` as the `Object` property.

For example, the following json object `Accept`s the attempt of `@someone@somewhere.else.example.org` to reply to a post by `@post_author@example.org`:

```json
{
  "actor": "https://example.org/users/post_author",
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "type": "Accept"
}
```

If this happens, `@someone@somewhere.else.example.org` (and their instance) should consider the interaction as having been approved / accepted. The instance can then feel free to distribute the interaction `Activity` to all of the recipients targed by `to`, `cc`, etc, with the additional property `approvedBy` ([see below](#approvedby)).

### Validating presence in a Followers / Following collection

If an `Actor` interacting with an `Object` (via `Like`, `inReplyTo`, or `Announce`) is permitted to do that interaction based on their presence in a `Followers` or `Following` collection in the `always` field of an interaction policy, then their server should *still* wait for an `Accept` to be received from the server of the target account, before distributing the interaction more widely with the `approvedBy` property set to the URI of the `Accept`.

This is to prevent scenarios where third servers have to somehow verify the presence of the interacting `Actor` in the `Followers` or `Following` collection of the `Actor` being interacted with. It is simpler to allow the target server to do that verification, and to trust that their `Accept` implicitly agrees that the interacting `Actor` is present in the relevant collection.

Likewise, when receiving an interaction from an `Actor` whose permission to interact matches with one of the `Following` or `Followers` collections in the `always` property, the server of the interacted-with `Actor` should ensure that they *always* send out an `Accept` as soon as possible, so that the interacting `Actor` server can send out the `Activity` with the proper proof of acceptance.

This process should bypass the normal "pending approval" stage whereby the server of the `Actor` being interacted with notifies them of the pending interaction, and waits for them to accept or reject, since there is no point notifying an `Actor` of a pending approval that they have already explicitly agreed to. In the GoToSocial codebase in particular, this is called "preapproval".

### `approvedBy`

`approvedBy` is an additional property added to the `Like`, and `Announce` activities, and any `Object`s considered to be "posts" (`Note`, `Article`, etc).

The presence of `approvedBy` signals that the author of the post targeted by the `Activity` or replied-to by the `Object` has approved/accepted the interaction, and it can now be distributed to its intended audience.

The value of `approvedBy` should be the URI of the `Accept` `Activity` created by the author of the post being interacted with.

For example, the following `Announce` `Activity` indicates, by the presence of `approvedBy`, that it has been `Accept`ed by `@post_author@example.org`:

```json
{
  "actor": "https://somewhere.else.example.org/users/someone",
  "to": [
    "https://somewhere.else.example.org/users/someone/followers"
  ],
  "cc": [
    "https://example.org/users/post_author"
  ],
  "id": "https://somewhere.else.example.org/users/someone/activities/announce/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://example.org/users/post_author/statuses/01J17ZZFK6W82K9MJ9SYQ33Y3D",
  "approvedBy": "https://example.org/users/post_author/activities/accept/01J18043HGECBDZQPT09CP6F2X",
  "type": "Announce"
}
```

When receiving an `Activity` with an `approvedBy` value attached to it, remote instances should dereference the URI value of the field to get the `Accept` `Activity`.

They should then validate that the `Accept` `Activity` has an `object` value equal to the `id` of the interaction `Activity` or `Object`, and an `actor` value equal to the author of the post being interacted with.

Moreover, they should ensure that the URL host/domain of the dereferenced `Accept` is equal to the URL host/domain of the author of the post being interacted with.

If the `Accept` cannot be dereferenced, or does not pass validity checks, the interaction should be considered invalid and dropped.

As a consequence of this validadtion mechanism, instances should make sure that they serve a valid ActivityPub Object in response to dereferences of `Accept` URIs that pertain to an `interactionPolicy`. If they do not, they inadvertently risk restricting the ability of remote instances to distribute their posts.

### Subsequent Replies / Scope Widening

Each subsequent reply in a conversation will have its own interaction policy, chosen by the user who created the reply. In other words, the entire *conversation* or *thread* is not controlled by one `interactionPolicy`, but the policy can differ for each subsequent post in a thread, as set by the post author.

Unfortunately, this means that even with `interactionPolicy` in place, the scope of a thread can inadvertently widen beyond the intention of the author of the first post in the thread.

For instance, in [example 1](#example-1---limiting-scope-of-a-conversation) above, `@the_mighty_zork` specifies in the first post a `canReply.always` value of

```json
[
  "https://example.org/users/the_mighty_zork",
  "https://example.org/users/booblover6969",
  "https://example.org/users/hodor"
]
```

In a subsequent reply, either accidentally or on purpose `@booblover6969` sets the `canReply.always` value to:

```json
[
  "https://www.w3.org/ns/activitystreams#Public"
]
```

This widens the scope of the conversation, as now anyone can reply to `@booblover6969`'s post, and possibly also tag `@the_mighty_zork` in that reply.

To avoid this issue, it is recommended that remote instances prevent users from being able to widen scope (exact mechanism of doing this TBD).

It is also a good idea for instances to consider any interaction with a post- or status-like `Object` that is itself currently pending approval, as also pending approval. 

In other words, instances should mark all children interactions below a pending-approval parent as also pending approval, no matter what the interaction policy on the parent would ordinarily allow.

This avoids situations where someone could reply to a post, then, even if their reply is pending approval, they could reply *to their own reply* and have that marked as permitted (since as author, they would normally have [implicit permission to reply](#implicit-assumptions)).

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

![A diagram of the conversation thread, showing the post from remote_2, and possible ancestor and descendant posts](../assets/diagrams/conversation_thread.png)

What GoToSocial will do now, is 'dereference' the post by `remote_2` to check if it is part of a thread and, if so, whether any other parts of the thread can be obtained.

GtS begins by checking the `inReplyTo` property of the post, which is set when a post is a reply to another post. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-inreplyto). If `inReplyTo` is set, GoToSocial derefences the replied-to post. If *this* post also has an `inReplyTo` set, then GoToSocial dereferences that too, and so on.

Once all of these **ancestors** of a status have been retrieved, GtS will begin working down through the **descendants** of posts.

It does this by checking the `replies` property of a derefenced post, and working through replies, and replies of replies. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-replies).

This process of thread dereferencing will likely involve making multiple HTTP calls to different servers, especially if the thread is long and complicated.

The end result of this dereferencing is that, assuming the reblogged post by `remote_2` was part of a thread, then `local_account` should now be able to see posts in the thread when they open the status on their home timeline. In other words, they will see replies from accounts on other servers (who they may not have come across yet), in addition to any previous and next posts in the thread as posted by `remote_2`.

This gives `local_account` a more complete view on the conversation, as opposed to just seeing the reblogged post in isolation and out of context. It also gives `local_account` the opportunity to discover new accounts to follow, based on replies to `remote_2`.
