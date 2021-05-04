package message

import (
	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AppCreate(authed *oauth.Auth, form *apimodel.ApplicationCreateRequest) (*apimodel.Application, error) {
	// set default 'read' for scopes if it's not set, this follows the default of the mastodon api https://docs.joinmastodon.org/methods/apps/
	var scopes string
	if form.Scopes == "" {
		scopes = "read"
	} else {
		scopes = form.Scopes
	}

	// generate new IDs for this application and its associated client
	clientID := uuid.NewString()
	clientSecret := uuid.NewString()
	vapidKey := uuid.NewString()

	// generate the application to put in the database
	app := &gtsmodel.Application{
		Name:         form.ClientName,
		Website:      form.Website,
		RedirectURI:  form.RedirectURIs,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		VapidKey:     vapidKey,
	}

	// chuck it in the db
	if err := p.db.Put(app); err != nil {
		return nil, err
	}

	// now we need to model an oauth client from the application that the oauth library can use
	oc := &oauth.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: form.RedirectURIs,
		UserID: "", // This client isn't yet associated with a specific user,  it's just an app client right now
	}

	// chuck it in the db
	if err := p.db.Put(oc); err != nil {
		return nil, err
	}

	mastoApp, err := p.tc.AppToMastoSensitive(app)
	if err != nil {
		return nil, err
	}

	return mastoApp, nil
}
