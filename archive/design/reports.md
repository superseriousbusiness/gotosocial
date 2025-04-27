# Reports / Flag Activity Federation

This document contains design notes for GoToSocial's federated (s2s) Flag functionality.

## What do existing implementations do?

Information below is true as of Jan 2023. If you're reading this much later, the below things may no longer apply.

### Mastodon

Mastodon uses the Flag activity to federate reports to other AP servers.

The Activity is wrapped inside a Create, which is addressed To the Inbox of the offending account.

To preserve anonymity of the reporter, the instance Actor is used as the Actor of the Activity.

Examples of the unwrapped Flag:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/actor",
  "content": "misinfo: it's not a good morning",
  "id": "https://example.org/341e866f-93f8-4755-9cf4-f8fb17f434fd",
  "object": [
    "https://bad.instance/users/tobi",
    "https://bad.instance/users/tobi/statuses/01GP388K19DGXSV3SW2RXWM533"
  ],
  "type": "Flag"
}
```

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/actor",
  "content": "smellyyyyyyyyyyyyy",
  "id": "https://example.org/3088184f-81b2-4545-8ce2-4cee4895449f",
  "object": "https://bad.instance/users/tobi",
  "type": "Flag"
}
```

The `content` field contains the report description.

The `object` field contains the reported Account URI and optionally one or more Note/Article/etc URIs. `object` value will be a string if just the Account URI is reported, or an array if the Account and one or more posts are being reported.

The `id` field is a generic URI that doesn't reveal any metadata. Trying to GET this URI gives a 404 Not Found error, which is OK since all the info needed to process the report is included in the Activity already.

### Misskey

Misskey uses the Flag activity to federate reports to other AP servers.

The Activity is wrapped inside a Create, which is addressed To the Inbox of the offending account.

To preserve anonymity of the reporter, the instance Actor is used as the Actor of the Activity.

Example of the unwrapped Flag:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/909i45meeo",
  "content": "Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ\n-----\nincites anti-police behaviour while being cute! ⛔",
  "id": "https://example.org/db22128d-884e-4358-9935-6a7c3940535d",
  "object": "https://bad.instance/users/tobi",
  "type": "Flag"
}
```

The `content` field contains the report description. Unlike with Mastodon, `content` also seems to include one or more statuses, as opposed to including statuses in the `object` field.

The `object` field contains the reported Account URI.

Trying to dereference the `id` field for a Misskey report with `Accept: application/activity+json` gives a 200 OK, but the returned content is some HTML unrelated to the report, so functionally equivalent to Mastodon's 404 behavior. Again, this is not really a problem.

### Calckey

Same as Misskey. Example:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/97wsu4gkns",
  "content": "Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ\n-----\nTest report from Calckey",
  "id": "https://example.org/b9a02404-d007-4b31-8dd6-bfc53387ad85",
  "object": "https://bad.instance/users/tobi",
  "type": "Flag"
}
```

### Pleroma / Akkoma

todo

### Friendica

Unsure: Friendica and GoToSocial still don't federate properly with one another (https://codeberg.org/superseriousbusiness/gotosocial/issues/169) so it's hard to test this.

## What should GoToSocial do?

Since the above implementations of Flag seem fairly consistent, GoToSocial should do more or less the same thing when federating reports outwards. So GtS ought to adopt the Mastodon behavior:

- Wrap Flag Activity in a Create and deliver it to the offending account.
- Use the GtS instance Actor as the Actor of the Flag.
- Generate an ID that doesn't reveal who created the report.
- Include Actor and one or more Note / Article / etc URIs in the `object` field

For incoming reports, all the above fields should be handled in order to generate a report for admins to look at.
