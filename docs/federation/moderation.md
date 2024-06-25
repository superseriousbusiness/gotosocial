# Moderation

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
