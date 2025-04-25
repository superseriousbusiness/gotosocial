package models

// Client client model
type Client interface {
	GetID() string
	GetSecret() string
	GetDomain() string
	GetUserID() string
}

func New(id string, secret string, domain string, userID string) Client {
	return &simpleClient{
		id:     id,
		secret: secret,
		domain: domain,
		userID: userID,
	}
}

// simpleClient is a very simple client model that satisfies the Client interface
type simpleClient struct {
	id     string
	secret string
	domain string
	userID string
}

// GetID client id
func (c *simpleClient) GetID() string {
	return c.id
}

// GetSecret client secret
func (c *simpleClient) GetSecret() string {
	return c.secret
}

// GetDomain client domain
func (c *simpleClient) GetDomain() string {
	return c.domain
}

// GetUserID user id
func (c *simpleClient) GetUserID() string {
	return c.userID
}
