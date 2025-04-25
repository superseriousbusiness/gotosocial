# pub

Implements the Social and Federating Protocols in the ActivityPub specification.

## Reference & Tutorial

The [go-fed website](https://go-fed.org/) contains tutorials and reference
materials, in addition to the rest of this README.

## How To Use

```
go get github.com/go-fed/activity
```

The root of all ActivityPub behavior is the `Actor`, which requires you to
implement a few interfaces:

```golang
import (
  "code.superseriousbusiness.org/activity/pub"
)

type myActivityPubApp struct { /* ... */ }
type myAppsDatabase struct { /* ... */ }
type myAppsClock struct { /* ... */ }

var (
  // Your app will implement pub.CommonBehavior, and either
  // pub.SocialProtocol, pub.FederatingProtocol, or both.
  myApp = &myActivityPubApp{}
  myCommonBehavior pub.CommonBehavior = myApp
  mySocialProtocol pub.SocialProtocol = myApp
  myFederatingProtocol pub.FederatingProtocol = myApp
  // Your app's database implementation.
  myDatabase pub.Database = &myAppsDatabase{}
  // Your app's clock.
  myClock pub.Clock = &myAppsClock{}
)

// Only support the C2S Social protocol
actor := pub.NewSocialActor(
  myCommonBehavior,
  mySocialProtocol,
  myDatabase,
  myClock)
// OR
//
// Only support S2S Federating protocol
actor = pub.NewFederatingActor(
  myCommonBehavior,
  myFederatingProtocol,
  myDatabase,
  myClock)
// OR
//
// Support both C2S Social and S2S Federating protocol.
actor = pub.NewActor(
  myCommonBehavior,
  mySocialProtocol,
  myFederatingProtocol,
  myDatabase,
  myClock)
```

Next, hook the `Actor` into your web server:

```golang
// The application's actor
var actor pub.Actor
var outboxHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
  c := context.Background()
  // Populate c with request-specific information
  if handled, err := actor.PostOutbox(c, w, r); err != nil {
    // Write to w
    return
  } else if handled {
    return
  } else if handled, err = actor.GetOutbox(c, w, r); err != nil {
    // Write to w
    return
  } else if handled {
    return
  }
  // else:
  //
  // Handle non-ActivityPub request, such as serving a webpage.
}
var inboxHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
  c := context.Background()
  // Populate c with request-specific information
  if handled, err := actor.PostInbox(c, w, r); err != nil {
    // Write to w
    return
  } else if handled {
    return
  } else if handled, err = actor.GetInbox(c, w, r); err != nil {
    // Write to w
    return
  } else if handled {
    return
  }
  // else:
  //
  // Handle non-ActivityPub request, such as serving a webpage.
}
// Add the handlers to a HTTP server
serveMux := http.NewServeMux()
serveMux.HandleFunc("/actor/outbox", outboxHandler)
serveMux.HandleFunc("/actor/inbox", inboxHandler)
var server http.Server
server.Handler = serveMux
```

To serve ActivityStreams data:

```golang
myHander := pub.NewActivityStreamsHandler(myDatabase, myClock)
var activityStreamsHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
  c := context.Background()
  // Populate c with request-specific information
  if handled, err := myHandler(c, w, r); err != nil {
    // Write to w
    return
  } else if handled {
    return
  }
  // else:
  //
  // Handle non-ActivityPub request, such as serving a webpage.
}
serveMux.HandleFunc("/some/data/like/a/note", activityStreamsHandler)
```

### Dependency Injection

Package `pub` relies on dependency injection to provide out-of-the-box support
for ActivityPub. The interfaces to be satisfied are:

* `CommonBehavior` - Behavior needed regardless of which Protocol is used.
* `SocialProtocol` - Behavior needed for the Social Protocol.
* `FederatingProtocol` - Behavior needed for the Federating Protocol.
* `Database` - The data store abstraction, not tied to the `database/sql`
package.
* `Clock` - The server's internal clock.
* `Transport` - Responsible for the network that serves requests and deliveries
of ActivityStreams data. A `HttpSigTransport` type is provided.

These implementations form the core of an application's behavior without
worrying about the particulars and pitfalls of the ActivityPub protocol.
Implementing these interfaces gives you greater assurance about being
ActivityPub compliant.

### Application Logic

The `SocialProtocol` and `FederatingProtocol` are responsible for returning
callback functions compatible with `streams.TypeResolver`. They also return
`SocialWrappedCallbacks` and `FederatingWrappedCallbacks`, which are nothing
more than a bundle of default behaviors for types like `Create`, `Update`, and
so on.

Applications will want to focus on implementing their specific behaviors in the
callbacks, and have fine-grained control over customization:

```golang
// Implements the FederatingProtocol interface.
//
// This illustration can also be applied for the Social Protocol.
func (m *myAppsFederatingProtocol) Callbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}) {
  // The context 'c' has request-specific logic and can be used to apply complex
  // logic building the right behaviors, if desired.
  //
  // 'c' will later be passed through to the callbacks created below.
  wrapped = pub.FederatingWrappedCallbacks{
    Create: func(ctx context.Context, create vocab.ActivityStreamsCreate) error {
      // This function is wrapped by default behavior.
      //
      // More application specific logic can be written here.
      //
      // 'ctx' will have request-specific information from the HTTP handler. It
      // is the same as the 'c' passed to the Callbacks method.
      // 'create' has, at this point, already triggered the recommended
      // ActivityPub side effect behavior. The application can process it
      // further as needed.
      return nil
    },
  }
  // The 'other' must contain functions that satisfy the signature pattern
  // required by streams.JSONResolver.
  //
  // If they are not, at runtime errors will be returned to indicate this.
  other = []interface{}{
    // The FederatingWrappedCallbacks has default behavior for an "Update" type,
    // but since we are providing this behavior in "other" and not in the
    // FederatingWrappedCallbacks.Update member, we will entirely replace the
    // default behavior provided by go-fed. Be careful that this still
    // implements ActivityPub properly.
    func(ctx context.Context, update vocab.ActivityStreamsUpdate) error {
      // This function is NOT wrapped by default behavior.
      //
      // Application specific logic can be written here.
      //
      // 'ctx' will have request-specific information from the HTTP handler. It
      // is the same as the 'c' passed to the Callbacks method.
      // 'update' will NOT trigger the recommended ActivityPub side effect
      // behavior. The application should do so in addition to any other custom
      // side effects required.
      return nil
    },
    // The "Listen" type has no default suggested behavior in ActivityPub, so
    // this just makes this application able to handle "Listen" activities.
    func(ctx context.Context, listen vocab.ActivityStreamsListen) error {
      // This function is NOT wrapped by default behavior. There's not a
      // FederatingWrappedCallbacks.Listen member to wrap.
      //
      // Application specific logic can be written here.
      //
      // 'ctx' will have request-specific information from the HTTP handler. It
      // is the same as the 'c' passed to the Callbacks method.
      // 'listen' can be processed with side effects as the application needs.
      return nil
    },
  }
  return
}
```

The `pub` package supports applications that grow into more custom solutions by
overriding the default behaviors as needed.

### ActivityStreams Extensions: Future-Proofing An Application

Package `pub` relies on the `streams.TypeResolver` and `streams.JSONResolver`
code generated types. As new ActivityStreams extensions are developed and their
code is generated, `pub` will automatically pick up support for these
extensions.

The steps to rapidly implement a new extension in a `pub` application are:

1. Generate an OWL definition of the ActivityStreams extension. This definition
could be the same one defining the vocabulary at the `@context` IRI.
2. Run `astool` to autogenerate the golang types in the `streams` package.
3. Implement the application's callbacks in the `FederatingProtocol.Callbacks`
or `SocialProtocol.Callbacks` for the new behaviors needed.
4. Build the application, which builds `pub`, with the newly generated `streams`
code. No code changes in `pub` are required.

Whether an author of an ActivityStreams extension or an application developer,
these quick steps should reduce the barrier to adopion in a statically-typed
environment.

### DelegateActor

For those that need a near-complete custom ActivityPub solution, or want to have
that possibility in the future after adopting go-fed, the `DelegateActor`
interface can be used to obtain an `Actor`:

```golang
// Use custom ActivityPub implementation
actor = pub.NewCustomActor(
  myDelegateActor,
  isSocialProtocolEnabled,
  isFederatedProtocolEnabled,
  myAppsClock)
```

It does not guarantee that an implementation adheres to the ActivityPub
specification. It acts as a stepping stone for applications that want to build
up to a fully custom solution and not be locked into the `pub` package
implementation.
