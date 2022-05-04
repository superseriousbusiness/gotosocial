# Glossary

This document describes some commonly-used terms in discussions of federation.

## `ActivityPub`

A decentralized social networking protocol based on the ActivityStreams data format. See [here](https://www.w3.org/TR/activitypub/).

GoToSocial uses the ActivityPub protocol to communicate between GtS servers, and with other federated servers like Mastodon, Pixelfed, etc.

## `ActivityStreams`

A model/data format for representing potential and completed activities using JSON. See [here](https://www.w3.org/TR/activitystreams-core/).

GoToSocial uses the ActivityStreams data model to 'speak' ActivityPub with other servers.

## `Actor`

An actor is an ActivityStreams object that is capable of performing some Activity like following, liking, creating a post, reblogging, etc. See [here](https://www.w3.org/TR/activitypub/#actors).

In GoToSocial, each account/user is an actor.

## `Dereference`

To 'dereference' a post or a profile means to make an HTTP call to the server that hosts that post or profile, in order to obtain its ActivityStreams representation.

GoToSocial 'dereferences' posts and profiles on remote servers, in order to convert them to models that GoToSocial can understand and work with.

Here's a more detailed explanation with some examples:

Let's say that someone on an ActivityPub server searches for the username `@tobi@goblin.technology`.

Their server would then do a webfinger lookup at `goblin.technology` for the username `tobi`, at the following URL:

```text
https://goblin.technology/.well-known/webfinger?resource=acct:tobi@goblin.technology
```

The `goblin.technology` server would give back some JSON in response; something like this:

```json
{
  "subject": "acct:tobi@goblin.technology",
  "aliases": [
    "https://goblin.technology/users/tobi",
    "https://goblin.technology/@tobi"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "https://goblin.technology/@tobi"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "https://goblin.technology/users/tobi"
    }
  ]
}
```

Under the links section the requesting server would look for a link of type `application/activity+json`, which denotes the ActivityStreams representation of the user. In this case, the URL is:

```text
https://goblin.technology/users/tobi
```

The above URL is a *reference* to the activitypub representation of the user/Actor `tobi` on the `goblin.technology` instance. It's called a reference because it doesn't contain all of the information about that user, it's only a reference point for where that information can be found.

Now, the requesting server will make a request to that URL in order to obtain a fuller representation of `@tobi@goblin.technology`, which complies to the ActivityPub spec. In other words, the server now follows a *reference* to get to the thing it references. This makes it *not a reference anymore*, hence the term *dereferencing*.

For an analogy, consider what happens when you look something up in the index of a book: first you get the page number that the material you're interested in is on, which is a reference. Then you turn to the referenced page to see the content, which is dereferencing.
