# Why Federation?

GoToSocial uses the [ActivityPub](https://activitypub.rocks/) federation protocol.

Federation means that you can hang out not just with people on your home server, but with people all over the [Fediverse](https://en.wikipedia.org/wiki/Fediverse). Your home server is part of a network of servers all over the world that all communicate using the same protocol--they speak the same 'language'. Your data is no longer centralized on one company's servers, but resides on your own server and is shared -- as you see fit -- across a resilient web of servers run by other people.

Not all of the servers you 'federate' with will be running GoToSocial: popular implementations of ActivityPub include software like [Mastodon](https://joinmastodon.org/), [WriteFreely](https://writefreely.org/), and many others. GoToSocial communicates seamlessly with these other servers.

This federated approach also means that you aren't beholden to arbitrary rules from some gigantic corporation potentially thousands of miles away. Your server has its own rules and culture; your fellow server residents are your neighbors; you will likely get to know your server admins and moderators, or be an admin yourself.

GoToSocial advocates for many small, weird, specialist servers where people can feel at home, rather than a few big and generic ones where one person's voice can get lost in the crowd.

## Glossary

Some commonly-used terms in discussions of federation, and their meanings.

### `ActivityPub`

A decentralized social networking protocol based on the ActivityStreams data format. See [here](https://www.w3.org/TR/activitypub/).

GoToSocial uses the ActivityPub protocol to communicate between GtS servers, and with other federated servers like Mastodon, Pixelfed, etc.

### `ActivityStreams`

A model/data format for representing potential and completed activities using JSON. See [here](https://www.w3.org/TR/activitystreams-core/).

GoToSocial uses the ActivityStreams data model to 'speak' ActivityPub with other servers.

### `Actor`

An actor is an ActivityStreams object that is capable of performing some Activity like following, liking, creating a post, reblogging, etc. See [here](https://www.w3.org/TR/activitypub/#actors).

In GoToSocial, each account/user is an actor.

### `Dereference`

To 'dereference' a post or a profile means to make an HTTP call to the server that hosts that post or profile, in order to obtain its ActivityStreams representation.

GoToSocial 'dereferences' posts and profiles on remote servers, in order to convert them to models that GoToSocial can understand and work with.
