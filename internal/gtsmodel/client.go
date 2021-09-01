package gtsmodel

// Client is a handy little wrapper for typical oauth client details
type Client struct {
	ID     string `bun:"type:CHAR(26),pk,notnull"`
	Secret string
	Domain string
	UserID string
}
