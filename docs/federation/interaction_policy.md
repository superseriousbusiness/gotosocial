# Interaction Policy

!!! warning "Feature still in development"
    Much like GoToSocial, this feature is still in pre-v1 beta development. Naming or schemas or methods of approval / rejection may change. We aim to finalize the document by GoToSocial v0.21.0, at which point it can be considered "ready" for other ActivityPub implementers to use.

GoToSocial uses the property `interactionPolicy` on posts, in order to indicate to remote instances what sort of interactions are (conditionally) permitted to be processed and stored by the origin server, for any given post.

The `@context` document for `interactionPolicy` and related objects and properties is at `https://gotosocial.org/ns`.

!!! danger
    Interaction policy is an attempt to limit the harmful effects of unwanted replies and other interactions on a user's posts (eg., "reply guys").
    
    However, it is far from being sufficient for this purpose, as there are still many "out-of-band" ways that posts can be distributed or replied to beyond a user's initial wishes or intentions.
    
    For example, a user might create a post with a very strict interaction policy attached to it, only to find that other server softwares do not respect that policy, and users on other instances are having discussions and replying to the post *from their instance's perspective*. The original poster's instance will automatically drop these unwanted interactions from their view, but remote instances may still show them.
    
    Another example: someone might see a post that specifies nobody can reply to it, but screenshot the post, post the screenshot in their own new post, and tag the original post author in as a mention. Alternatively, they might just link to the URL of the post and tag the author in as a mention. In this case, they effectively "reply" to the post by creating a new thread.
    
    For better or worse, GoToSocial can offer only a best-effort, partial, technological solution to what is more or less an issue of social behavior and boundaries.

!!! info "Deprecated `always` and `approvalRequired` properties"
    Previous versions of this document used the properties `always` and `approvalRequired`. These are now deprecated in favor of `automaticApproval` and `manualApproval`. GoToSocial versions before v0.20.0 send and receive only these deprecated properties. GoToSocial v0.20.0 sends and receives both the deprecated and the new properties. GoToSocial v0.21.0 onwards uses only the new properties. 

## Overview

`interactionPolicy` is an object property attached to the post-like `Object`s `Note`, `Article`, `Question`, etc, with the following format:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "automaticApproval": [ "zero_or_more_uris_that_can_always_do_this" ],
      "manualApproval": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    },
    "canReply": {
      "automaticApproval": [ "zero_or_more_uris_that_can_always_do_this" ],
      "manualApproval": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    },
    "canAnnounce": {
      "automaticApproval": [ "zero_or_more_uris_that_can_always_do_this" ],
      "manualApproval": [ "zero_or_more_uris_that_require_approval_to_do_this" ]
    }
  },
  [...]
}
```

In the `interactionPolicy` object:

- `canLike` is a sub-policy which indicates who is permitted to create a `Like` with the post URI as the `object` of the `Like`.
- `canReply` is a sub-policy which indicates who is permitted to create a post with `inReplyTo` set to the URI/ID of the post.
- `canAnnounce` is a sub-policy which indicates who is permitted to create an `Announce` with the post URI/ID as the `object` of the `Announce`. 

And:

- `automaticApproval` denotes ActivityPub URIs/IDs of `Actor`s or `Collection`s of `Actor`s who will receive automated approval from the post author to create + distribute an interaction targeting a post.
- `manualApproval` denotes ActivityPub URIs/IDs of `Actor`s or `Collection`s of `Actor`s who will receive approval from the post author at the author's own discretion, and may not receive approval at all, or may be rejected (see [Requesting, Obtaining, and Validating Approval](#requesting-obtaining-and-validating-approval)).

Valid URI entries in `automaticApproval` and `manualApproval` include:

- the magic ActivityStreams Public URI `https://www.w3.org/ns/activitystreams#Public`
- the URIs of the post creator's `Following` and/or `Followers` collections
- individual Actor URIs

For example:

```json
[
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/someone/followers",
    "https://example.org/users/someone/following",
    "https://example.org/users/someone_else",
    "https://somewhere.else.example.org/users/someone_on_a_different_instance"
]
```

!!! info
    Be aware that according to JSON-LD the values of `automaticApproval` and `manualApproval` can be either single strings or arrays of strings. That is, the following are all valid:
    
    - Single string: `"automaticApproval": "https://example.org/users/someone"`
    - Single-entry array: `"automaticApproval": [ "https://example.org/users/someone" ]`
    - Multiple-entry array: `"automaticApproval": [ "https://example.org/users/someone", "https://example.org/users/someone_else" ]`

## Specifying Nobody

To specify that **nobody** can perform an interaction on a post **except** for its author (who is always permitted), implementations should set the `automaticApproval` array to **just the URI of the post author**, and `manualApproval` can be unset, `null`, or empty.

For example, the following `canLike` value indicates that nobody can `Like` the post it is attached to except for the post author:

```json
"canLike": {
  "automaticApproval": "the_activitypub_uri_of_the_post_author"
},
```

Another example. The following `interactionPolicy` on a post by `https://example.org/users/someone` indicates that anyone can like the post, but nobody but the author can reply or announce:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "automaticApproval": "https://example.org/users/someone"
    },
    "canAnnounce": {
      "automaticApproval": "https://example.org/users/someone"
    }
  },
  [...]
}
```

!!! note
    To avoid mischief, GoToSocial makes implicit assumptions about who can/can't interact, even if a policy specifies nobody. See [implicit assumptions](#implicit-assumptions).

## Conflicting / Duplicate Values

In cases where a user is present in a Collection URI, and is *also* targeted explicitly by URI, the **more specific value** takes precedence.

For example:

```json
[...],
"canReply": {
  "automaticApproval": "https://example.org/users/someone",
  "manualApproval": "https://www.w3.org/ns/activitystreams#Public"
},
[...]
```

Here, `@someone@example.org` is present in `automaticApproval`, and is also implicitly present in the magic ActivityStreams Public collection in `manualApproval`. In this case, they can always reply, as the `automaticApproval` value is more explicit.

Another example:

```json
[...],
"canReply": {
  "automaticApproval": "https://www.w3.org/ns/activitystreams#Public",
  "manualApproval": "https://example.org/users/someone"
},
[...]
```

Here, `@someone@example.org` is present in `manualApproval`, but is also implicitly present in the magic ActivityStreams Public collection in `automaticApproval`. In this case everyone can reply without approval, **except** for `@someone@example.org`, who requires approval.

In case the **exact same** URI is present in both `automaticApproval` and `manualApproval`, the **highest level of permission** takes precedence (ie., a URI in `automaticApproval` takes precedence over the same URI in `manualApproval`).

## Default / fallback `interactionPolicy`

When the `interactionPolicy` property is not present at all on a post, or the `interactionPolicy` key is set but its value resolves to `null` or `{}`, implementations can assume the following implicit, default `interactionPolicy` for that post:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canAnnounce": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    }
  },
  [...]
}
```

As implied by the lack of any `manualApproval` property in any of the sub-policies, the default value for `manualApproval` is an empty array.

This default `interactionPolicy` was designed to reflect the de facto interaction policy of all posts from pre-v0.17.0 GoToSocial, and other ActivityPub server softwares, at the time of writing. That is to say, it is exactly what servers that are not interaction policy aware *already assume* about interaction permissions.

!!! info "Actors can only ever interact with a post they are permitted to see"
    Note that even when assuming a default `interactionPolicy` for a post, the **visibility** of a post must still be accounted for by looking at the `to`, `cc`, and/or `audience` properties, to ensure that actors who cannot *see* a post also cannot *interact* with it. Eg., if a post is addressed to followers-only, and the default `interactionPolicy` is assumed, then someone who does not follow the post creator should still *not* be able to see *or* interact with it.

!!! tip
    As is standard across AP implementations, implementations will likely still wish to limit `Announce` actities targeting the post to only the author themself if the post is addressed to followers-only.

## Defaults per sub-policy

When an interaction policy is only *partially* defined (eg., only `canReply` is set, `canLike` or `canAnnounce` keys are not set), then implementations should make the following assumptions for each sub-policy in the `interactionPolicy` object that is *undefined*.

!!! tip "Future extensions with different defaults"
    Note that **the below list is not exhaustive**, and extensions to `interactionPolicy` may wish to define **different defaults** for other types of interaction.

### `canLike`

If `canLike` is missing on an `interactionPolicy`, or the value of `canLike` is `null` or `{}`, then implementations should assume:

```json
"canLike": {
  "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
}
```

In other words, the default is **anyone who can see the post can like it**.

### `canReply`

If `canReply` is missing on an `interactionPolicy`, or the value of `canReply` is `null` or `{}`, then implementations should assume:

```json
"canReply": {
  "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
}
```

In other words, the default is **anyone who can see the post can reply to it**.

### `canAnnounce`

If `canAnnounce` is missing on an `interactionPolicy`, or the value of `canAnnounce` is `null` or `{}`, then implementations should assume:

```json
"canAnnounce": {
  "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
}
```

In other words, the default is **anyone who can see the post can announce it**.

## Indicating that verification is required / not required per sub-policy

Because not all servers have implemented interaction policies at the time of writing, it is necessary to provide a method by which implementing servers can indicate that they are both **aware of** and **will enforce** interaction policies as described below in the [Interaction Verification](#interaction-verification) section.

This indication of interaction policy participation is done via a server explicitly setting `interactionPolicy` and its sub-policies on outgoing posts, instead of relying on the defaults described above.

That is, **by setting `interactionPolicy.*` on a post, an instance indicates to other instances that they will enforce validation of interactions for each sub-policy that is explicitly set.**

This means that implementations should always explicitly set all sub-policies on an `interactionPolicy` for which they have implemented interaction controls themselves, and with which they would like other servers to comply.

For example, if a server understands and wishes to enforce the `canLike`, `canReply`, and `canAnnounce` sub-policies (as is the case with GoToSocial), then they should explicitly set those sub-policies on an outgoing post *even when the values do not differ from the implicit defaults*. This allows remote servers to know that the origin server does enforcement, and knows how to handle appropriate `Reject` / `Accept` messages for each sub-policy.

Another example: if a server only implements the `canReply` interaction sub-policy, but not `canLike` or `canAnnounce`, then they should always set `interactionPolicy.canReply`, and leave the other two sub-policies out of the `interactionPolicy` to indicate that they cannot understand or enforce them.

This means of indicating participation in interaction policies through the absence of presence of keys was designed so that the large majority of servers that *do not* set `interactionPolicy` at all, because they have not (yet) implemented it, do not need to change their behavior. Servers that do implement `interactionPolicy` can understand, by the absence of the `interactionPolicy` key on a post, that the origin server is not `interactionPolicy` aware, and behave accordingly.

## Implicit Assumptions

For common-sense safety reasons, GoToSocial makes, and will always apply, two implicit assumptions about interaction policies.

### 1. Mentioned + replied-to actors can always reply

Actors mentioned in, or replied to by, a post should **ALWAYS** be able to reply to that post without requiring approval, regardless of the post visiblity and the `interactionPolicy`, **UNLESS** the post that mentioned or replied to them is itself currently pending approval.

This is to prevent a would-be harasser from mentioning someone in an abusive post, and leaving no recourse to the mentioned user to reply.

As such, when sending out interaction policies, GoToSocial will **ALWAYS** add the URIs of mentioned users to the `canReply.always` array, unless they are already covered by the ActivityStreams magic public URI.

Likewise, when enforcing received interaction policies, GoToSocial will **ALWAYS** behave as though the URIs of mentioned users were present in the `canReply.always` array, even if they weren't.

### 2. An actor can always interact in any way with their own post

**Secondly**, an actor should **ALWAYS** be able to reply to their own post, like their own post, and boost their own post without requiring approval, **UNLESS** that post is itself currently pending approval.

As such, when sending out interaction policies, GoToSocial will **ALWAYS** add the URI of the post author to the `canLike.always`, `canReply.always`, and `canAnnounce.always` arrays, **UNLESS** they are already covered by the ActivityStreams magic public URI.

Likewise, when enforcing received interaction policies, GoToSocial will **ALWAYS** behave as though the URI of the post author themself is present in each `automaticApproval` field, even if it wasn't.

## Examples

Here's some examples of what interaction policies allow users to do.

### 1. Limiting scope of a conversation

In this example, the user `@the_mighty_zork` wants to begin a conversation with the users `@booblover6969` and `@hodor`.

To avoid the discussion being derailed by others, they want replies to their post by users other than the three participants to be permitted only if they're approved by `@the_mighty_zork`.

Furthermore, they want to limit the boosting / `Announce`ing of their post to only their own followers, and to the three conversation participants.

However, anyone should be able to `Like` the post by `@the_mighty_zork`.

This can be achieved with the following `interactionPolicy`, which is attached to a post with visibility level public:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "automaticApproval": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ],
      "manualApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canAnnounce": {
      "automaticApproval": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/the_mighty_zork/followers",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ]
    }
  },
  [...]
}
```

### 2. Long solo thread

In this example, the user `@the_mighty_zork` wants to write a long solo thread.

They don't mind if people boost and like posts in the thread, but they don't want to get any replies because they don't have the energy to moderate the discussion; they just want to vent by throwing their thoughts out there.

This can be achieved by setting the following `interactionPolicy` on every post in the thread:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "automaticApproval": "https://example.org/users/the_mighty_zork"
    },
    "canAnnounce": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    }
  },
  [...]
}
```

Here, anyone is allowed to like or boost, but nobody is permitted to reply (except `@the_mighty_zork` themself).

### 3. Completely open

In this example, `@the_mighty_zork` wants to write a completely open post that can be replied to, boosted, or liked by anyone who can see it:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canAnnounce": {
      "automaticApproval": "https://www.w3.org/ns/activitystreams#Public"
    }
  },
  [...]
}
```

## Subsequent Replies / Scope Widening

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

It is also a good idea for instances to consider any interaction with a post-like `Object` that is itself currently pending approval, as also pending approval. 

In other words, instances should mark all children interactions below a pending-approval parent as also pending approval, no matter what the interaction policy on the parent would ordinarily allow.

This avoids situations where someone could reply to a post, then, even if their reply is pending approval, they could reply *to their own reply* and have that marked as permitted (since as author, they would normally have [implicit permission to reply](#implicit-assumptions)).

## Interaction Verification

The [interaction policy](#interaction-policy) section described the shape of interaction policies, assumed defaults, and assumptions.

This section describes the enforcement and verification of interaction policies, ie., how servers that set interaction policies should send approval or rejection of a requested/pending interaction, and how other servers can prove that approval to interact with a post has been obtained by the interacter from the interactee.

### Requesting, Obtaining, and Validating Approval

When an actor's URI is in the `manualApproval` array for a type of interaction, **or** their presence in a collection needs to be validated (see [Validating presence in a Followers / Following collection](#validating-presence-in-a-followers--following-collection)), implementations wishing to obtain approval for that actor to interact with a policied post should do the following:

1. Compose the interaction `Activity` (ie., `Like`, `Create` (reply), or `Announce`), as normal.
2. Address the `Activity` `to` and `cc` the expected recipients for that `Activity`, as normal.
3. `POST` the `Activity` only to the `Inbox` (or `sharedInbox`) of the author of the post being interacted with.
4. **DO NOT DISTRIBUTE THE ACTIVITY FURTHER THAN THIS AT THIS POINT**.

At this point, the interaction can be considered as *pending approval*, and should not be shown in the replies or likes collections, etc., of the post interacted with.

It may be shown to the user who sent the interaction as a sort of "interaction pending" modal, but ideally it should not be shown to other users who share an instance with that user.

From here, one of three things may happen:

#### Rejection

In this scenario, the server of the author of the post being interacted with sends back a `Reject` `Activity` with the interaction URI/ID as the `object` property.

For example, the following json object `Reject`s the attempt of `@someone@somewhere.else.example.org` to reply to a post by `@post_author@example.org`:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "type": "Reject"
}
```

If this happens, `@someone@somewhere.else.example.org` (and their instance) should consider the interaction as having been rejected. The instance should delete the activity from its internal storage (ie., database), or otherwise indicate that it's been rejected, and it should not distribute the `Activity` further, or retry the interaction. The server may wish to indicate to the interacter that their interaction was rejected.

#### Nothing

In this scenario, the author of the post being interacted with never sends back a `Reject` or an `Accept` `Activity`. In such a case, the interaction is considered "pending" in perpetuity. Instances may wish to implement some kind of cleanup feature, where sent and pending interactions that reach a certain age should be considered expired, and `Rejected` and then removed in the manner gestured towards above.

#### Acceptance

In this scenario, the author of the post being interacted with sends back an `Accept` `Activity` with the interaction URI/ID as the `object` property, and a dereferenceable URI of an approval object as the `result` property (see [Approval Objects](#approval-objects)).

For example, the following json object `Accept`s the attempt of `@someone@somewhere.else.example.org` to reply to a post by `@post_author@example.org`:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/accept/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "result": "https://example.org/users/post_author/reply_approvals/01JMMGABRDNA9G9BDNYJR7TC8D",
  "type": "Accept"
}
```

If this happens, `@someone@somewhere.else.example.org` (and their instance) should consider the interaction as having been approved/accepted by the interactee.

At this point, `somewhere.else.example.org` should once again send out the interaction, with the following differences:

1. This time, include the `result` URI/ID from the `accept` in the `approvedBy` field of the post contained in the `Create` (see [`approvedBy`](#approvedby)).
2. This time, distribute the interaction to **all** of the recipients targed by `to`, `cc`, etc.

!!! note
    While it is not strictly necessary, in the above example, actor `https://example.org/users/post_author` addresses the `Accept` activity not just to the interacting actor `https://somewhere.else.example.org/users/someone`, but to their followers collection as well (and, implicitly, to the public). This allows followers of `https://example.org/users/post_author` on other servers to also mark the interaction as accepted, and to show the interaction alongside the interacted-with post, without having to dereference + verify the URI in `approvedBy`.

### Approval Objects

An approval is an extension of a basic ActivityStreams Object, with the type `LikeApproval`, `ReplyApproval`, or `AnnounceApproval`. Each type corresponds to the type of interaction that the particular approval approves.

`LikeApproval`:

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "attributedTo": "https://example.org/users/post_author",
  "id": "https://example.org/users/post_author/approvals/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/likes/01JMPKG79EAH0NB04BHEM9D20N",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "type": "LikeApproval"
}
```

`ReplyApproval`:

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "attributedTo": "https://example.org/users/post_author",
  "id": "https://example.org/users/post_author/approvals/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "type": "ReplyApproval"
}
```

`AnnounceApproval`:

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "attributedTo": "https://example.org/users/post_author",
  "id": "https://example.org/users/post_author/approvals/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/boosts/01JMPKG79EAH0NB04BHEM9D20N",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "type": "AnnounceApproval"
}
```

In an approval object:

- `attributedTo`: should be the `Actor` who `Accept`ed an interaction, ie., the *interactee*.
- `object`: should be the *interacting* `Like`, `Announce`, or post-like `Object`.
- `target` (optional): if included, should be the post-like `Object` that was *interacted with*.

!!! info "Approval objects should be dereferenceable"
    As a consequence of the validation mechanism (see [Validating `approvedBy`](#validating-approvedby)), instances should make sure that they serve a valid ActivityPub response to dereferences of approval object URIs. If they do not, they inadvertently risk restricting the ability of remote instances to distribute their posts.

### `approvedBy`

`approvedBy` is an additional property added to the `Like`, and `Announce` activities, and any `Object`s considered to be "posts" (`Note`, `Article`, etc).

The presence of `approvedBy` signals that the author of the post targeted by the `Activity` or replied-to by the `Object` has approved/accepted the interaction, and it can now be distributed to its intended audience.

The value of `approvedBy` should be the `result` URI/ID that was sent along in an `Accept` from the interactee, which points towards a dereferenceable approval object.

For example, the following `Announce` `Activity` claims, by the presence of `approvedBy`, that it has been `Accept`ed by `@post_author@example.org`:

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
  "approvedBy": "https://example.org/users/post_author/reply_approvals/01JMMGABRDNA9G9BDNYJR7TC8D",
  "type": "Announce"
}
```

#### Validating `approvedBy`

When receiving an `Activity` or a post-like `Object` with an `approvedBy` value attached to it, remote instances should:

1. Validate that the host/domain of the `approvedBy` URI is equal to the host/domain of the author of the post being interacted with.
2. Dereference the `approvedBy` URI/ID to get the approval object (see [Approval Objects](#approval-objects)).
3. Check the type of the approval to ensure it's the correct one, eg., an `Announce` should have an `approvedBy` URI/ID set on it that points to an `AnnounceApproval`, not a `ReplyApproval` or a `LikeApproval`.
4. Check that the approval has an `attributedTo` value equal to the URI/ID of the actor being interacted with.
5. Check that the approval has an `object` value equal to the `id` of the interaction `Activity` or `Object`.

If the approval cannot be dereferenced, or does not pass validity checks, the interaction should be considered invalid and dropped.

!!! warning
    Versions 0.17.x and 0.18.x of GoToSocial did not include a `result` pointing towards an approval type. Instead, the URI/ID of the `Accept` was sent in the `approvedBy` field. 
    
    Version 0.18.x of GoToSocial is partially forward-compatible with approval types, since it can validate approval using either a dereferenced `Accept` or a dereferenced approval type, while still sending out an `Accept` URI itself in the `approvedBy` field.
    
    Versions of GoToSocial from 0.19.x upwards will send out an `approvedBy` pointing to an approval type, as described in this document, not an `Accept`.

### Validating presence in a Followers / Following collection

If an `Actor` interacting with an `object` (via `Like`, `inReplyTo`, or `Announce`) is permitted to do that interaction based on their presence in a `Followers` or `Following` collection in the `automaticApproval` field of an interaction policy, then their server **should still wait** for an `Accept` to be received from the server of the target actor, before distributing the interaction more widely with the `approvedBy` property set to the URI/ID of the approval.

This is to prevent scenarios where third servers have to somehow verify the presence of the interacting `Actor` in the `Followers` or `Following` collection of the `Actor` being interacted with. It is simpler to allow the target server to do that verification, and to trust that their approval implicitly agrees that the interacting `Actor` is present in the relevant collection.

Likewise, when receiving an interaction from an `Actor` whose permission to interact matches with one of the `Following` or `Followers` collections in the `automaticApproval` property, the server of the interacted-with `Actor` should ensure that they *always* send out an `Accept` as soon as possible, so that the interacting `Actor` server can send out the `Activity` with the proper proof of acceptance.

This process should bypass the normal "pending approval" stage whereby the server of the `Actor` being interacted with notifies them of the pending interaction, and waits for them to accept or reject, since there is no point notifying an `Actor` of a pending approval that they have already explicitly agreed to. In the GoToSocial codebase in particular, this is called "preapproval".

### Optional behaviors

This section describes optional behaviors that implementers *may* use when sending `Accept` and `Reject` messages, and *should* account for when receiving `Accept` and `Reject` messages. 

#### Always send out `Accept`s

Implementers may wish to *always* send out an `Accept` to remote interacters, even when the interaction is implicitly or explicitly permitted by their presence in the `automaticApproval` array. When receiving such an `Accept`, implementations may still want to update their representation of the interaction to include an `approvedBy` URI pointing at an approval. This may be useful later on when handling revocations (TODO).

#### Type hinting: inline `object` property on `Accept` and `Reject`

If desired, implementers may partially expand/inline the `object` property of an `Accept` or `Reject` to hint to remote servers about the type of interaction being `Accept`ed or `Reject`ed. When inlining in this way, the `object`'s `type` and `id` must be defined at a minimum. For example:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": {
    "type": "Note",
    "id": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
    [...]
  },
  "result": "https://example.org/users/post_author/approvals/01JMPS01E54DG9JCF2ZK3JDMXE",
  "type": "Accept"
}
```

#### Set `target` property on `Accept` and `Reject`

If desired, implementers may set the `target` property on outgoing `Accept` or `Reject` activities to the `id` of the post being interacted with, to make it easier for remote servers to understand the shape and relevance of the interaction that's being `Accept`ed or `Reject`ed.

For example, the following json object `Accept`s the attempt of `@someone@somewhere.else.example.org` to reply to a post by `@post_author@example.org` that has the id `https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT`:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "result": "https://example.org/users/post_author/approvals/01JMPS01E54DG9JCF2ZK3JDMXE",
  "type": "Accept"
}
```

If desired, the `target` property can also be partially expanded/inlined to type hint about the post that was interacted with. When inlining in this way, the `target`'s `type` and `id` must be defined at a minimum. For example:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "target": {
    "type": "Note",
    "id": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT"
    [ ... ]
  },
  "result": "https://example.org/users/post_author/approvals/01JMPS01E54DG9JCF2ZK3JDMXE",
  "type": "Accept"
}
```
