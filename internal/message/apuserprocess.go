package message

import (
	"net/http"
)

func (p *processor) GetAPUser(requestHeaders http.Header, username string) (interface{}, error) {

	// // get the account the request is referring to
	// requestedAccount := &gtsmodel.Account{}
	// if err := m.db.GetLocalAccountByUsername(username, requestedAccount); err != nil {
	// 	return nil, NewErrorNotAuthorized(fmt.Errorf("database error getting account with username %s: %s", username, err))
	// }

	// // and create a transport for it
	// transport, err := p.federator.TransportController().NewTransport(requestedAccount.PublicKeyURI, requestedAccount.PrivateKey)
	// if err != nil {
	// 	l.Errorf("error creating transport for username %s: %s", requestedUsername, err)
	// 	// we'll just return not authorized here to avoid giving anything away
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
	// 	return
	// }

	// // authenticate the request
	// authentication, err := federation.AuthenticateFederatedRequest(transport, c.Request)
	// if err != nil {
	// 	l.Errorf("error authenticating GET user request: %s", err)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
	// 	return
	// }

	// if !authentication.Authenticated {
	// 	l.Debug("request not authorized")
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
	// 	return
	// }

	// requestingAccount := &gtsmodel.Account{}
	// if authentication.RequestingPublicKeyID != nil {
	// 	if err := m.db.GetWhere("public_key_uri", authentication.RequestingPublicKeyID.String(), requestingAccount); err != nil {

	// 	}
	// }

	// authorization, err := federation.AuthorizeFederatedRequest

	// person, err := m.tc.AccountToAS(requestedAccount)
	// if err != nil {
	// 	l.Errorf("error converting account to ap person: %s", err)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
	// 	return
	// }

	// data, err := person.Serialize()
	// if err != nil {
	// 	l.Errorf("error serializing user: %s", err)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
	// 	return
	// }

	// c.JSON(http.StatusOK, data)
	return nil, nil
}
