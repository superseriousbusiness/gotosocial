package model

// StatusTimelineResponse wraps a slice of statuses, ready to be serialized, along with the Link
// header for the previous and next queries, to be returned to the client.
type StatusTimelineResponse struct {
	Statuses   []*Status
	LinkHeader string
}
