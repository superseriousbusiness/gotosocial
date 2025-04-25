// Package pub implements the ActivityPub protocol.
//
// Note that every time the ActivityStreams types are changed (added, removed)
// due to code generation, the internal function toASType needs to be modified
// to know about these types.
//
// Note that every version change should also include a change in the version.go
// file.
package pub
