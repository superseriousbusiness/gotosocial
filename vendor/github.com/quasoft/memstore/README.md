# memstore

[![GoDoc](https://godoc.org/github.com/quasoft/memstore?status.svg)](https://godoc.org/github.com/quasoft/memstore) [![Build Status](https://travis-ci.org/quasoft/memstore.png?branch=master)](https://travis-ci.org/quasoft/memstore) [![Coverage Status](https://coveralls.io/repos/github/quasoft/memstore/badge.svg?branch=master)](https://coveralls.io/github/quasoft/memstore?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/quasoft/memstore)](https://goreportcard.com/report/github.com/quasoft/memstore)

In-memory implementation of [gorilla/sessions](https://github.com/gorilla/sessions) for use in tests and dev environments

## How to install

    go get github.com/quasoft/memstore

## Documentation

Documentation, as usual, can be found at [godoc.org](http://www.godoc.org/github.com/quasoft/memstore).

The interface of [gorilla/sessions](https://github.com/gorilla/sessions) is described at http://www.gorillatoolkit.org/pkg/sessions.

### How to use
``` go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/quasoft/memstore"
)

func main() {
	// Create a memory store, providing authentication and
	// encryption key for securecookie
	store := memstore.NewMemStore(
		[]byte("authkey123"),
		[]byte("enckey12341234567890123456789012"),
	)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Get session by name.
		session, err := store.Get(r, "session1")
		if err != nil {
			log.Printf("Error retrieving session: %v", err)
		}

		// The name should be 'foobar' if home page was visited before that and 'Guest' otherwise.
		user, ok := session.Values["username"]
		if !ok {
			user = "Guest"
		}
		fmt.Fprintf(w, "Hello %s", user)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get session by name.
		session, err := store.Get(r, "session1")
		if err != nil {
			log.Printf("Error retrieving session: %v", err)
		}

		// Add values to the session object
		session.Values["username"] = "foobar"
		session.Values["email"] = "spam@eggs.com"

		// Save values
		err = session.Save(r, w)
		if err != nil {
			log.Fatalf("Error saving session: %v", err)
		}
	})

	log.Printf("listening on http://%s/", "127.0.0.1:9090")
	log.Fatal(http.ListenAndServe("127.0.0.1:9090", nil))
}

```
