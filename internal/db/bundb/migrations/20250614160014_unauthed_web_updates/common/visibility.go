package common

// Visibility represents the
// visibility granularity of a status.
type Visibility int16

const (
	// VisibilityNone means nobody can see this.
	// It's only used for web status visibility.
	VisibilityNone Visibility = 1

	// VisibilityPublic means this status will
	// be visible to everyone on all timelines.
	VisibilityPublic Visibility = 2

	// VisibilityUnlocked means this status will be visible to everyone,
	// but will only show on home timeline to followers, and in lists.
	VisibilityUnlocked Visibility = 3

	// VisibilityFollowersOnly means this status is viewable to followers only.
	VisibilityFollowersOnly Visibility = 4

	// VisibilityMutualsOnly means this status
	// is visible to mutual followers only.
	VisibilityMutualsOnly Visibility = 5

	// VisibilityDirect means this status is
	// visible only to mentioned recipients.
	VisibilityDirect Visibility = 6

	// VisibilityDefault is used when no other setting can be found.
	VisibilityDefault Visibility = VisibilityUnlocked
)
