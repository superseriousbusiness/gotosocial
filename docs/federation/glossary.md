# Glossary

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
