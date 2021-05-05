/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package message

import (
	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Processor should be passed to api modules (see internal/apimodule/...). It is used for
// passing messages back and forth from the client API and the federating interface, via channels.
// It also contains logic for filtering which messages should end up where.
// It is designed to be used asynchronously: the client API and the federating API should just be able to
// fire messages into the processor and not wait for a reply before proceeding with other work. This allows
// for clean distribution of messages without slowing down the client API and harming the user experience.
type Processor interface {
	// ToClientAPI returns a channel for putting in messages that need to go to the gts client API.
	ToClientAPI() chan ToClientAPI
	// ToFederator returns a channel for putting in messages that need to go to the federator (activitypub).
	ToFederator() chan ToFederator

	/*
		API-FACING PROCESSING FUNCTIONS
		These functions are intended to be called when the API client needs an immediate (ie., synchronous) reply
		to an HTTP request. As such, they will only do the bare-minimum of work necessary to give a properly
		formed reply. For more intensive (and time-consuming) calls, where you don't require an immediate
		response, pass work to the processor using a channel instead.
	*/

	// AccountCreate processes the given form for creating a new account, returning an oauth token for that account if successful.
	AccountCreate(authed *oauth.Auth, form *apimodel.AccountCreateRequest) (*apimodel.Token, error)
	// AccountGet processes the given request for account information.
	AccountGet(authed *oauth.Auth, targetAccountID string) (*apimodel.Account, error)

	// AppCreate processes the creation of a new API application
	AppCreate(authed *oauth.Auth, form *apimodel.ApplicationCreateRequest) (*apimodel.Application, error)

	// StatusCreate processes the given form to create a new status, returning the api model representation of that status if it's OK.
	StatusCreate(authed *oauth.Auth, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, error)
	// StatusDelete processes the delete of a given status, returning the deleted status if the delete goes through.
	StatusDelete(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error)
	// StatusFave processes the faving of a given status, returning the updated status if the fave goes through.
	StatusFave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error)
	// StatusFavedBy returns a slice of accounts that have liked the given status, filtered according to privacy settings.
	StatusFavedBy(authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, error)
	// StatusGet gets the given status, taking account of privacy settings and blocks etc.
	StatusGet(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error)
	// StatusUnfave processes the unfaving of a given status, returning the updated status if the fave goes through.
	StatusUnfave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error)

	// MediaCreate handles the creation of a media attachment, using the given form.
	MediaCreate(authed *oauth.Auth, form *apimodel.AttachmentRequest) (*apimodel.Attachment, error)
	MediaGet(authed *oauth.Auth, form *apimodel.GetContentRequestForm) (*apimodel.Content, error)
	// AdminEmojiCreate handles the creation of a new instance emoji by an admin, using the given form.
	AdminEmojiCreate(authed *oauth.Auth, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, error)

	// Start starts the Processor, reading from its channels and passing messages back and forth.
	Start() error
	// Stop stops the processor cleanly, finishing handling any remaining messages before closing down.
	Stop() error
}

// processor just implements the Processor interface
type processor struct {
	// federator     pub.FederatingActor
	toClientAPI  chan ToClientAPI
	toFederator  chan ToFederator
	stop         chan interface{}
	log          *logrus.Logger
	config       *config.Config
	tc           typeutils.TypeConverter
	oauthServer  oauth.Server
	mediaHandler media.Handler
	storage      storage.Storage
	db           db.DB
}

// NewProcessor returns a new Processor that uses the given federator and logger
func NewProcessor(config *config.Config, tc typeutils.TypeConverter, oauthServer oauth.Server, mediaHandler media.Handler, storage storage.Storage, db db.DB, log *logrus.Logger) Processor {
	return &processor{
		toClientAPI:  make(chan ToClientAPI, 100),
		toFederator:  make(chan ToFederator, 100),
		stop:         make(chan interface{}),
		log:          log,
		config:       config,
		tc:           tc,
		oauthServer:  oauthServer,
		mediaHandler: mediaHandler,
		storage:      storage,
		db:           db,
	}
}

func (d *processor) ToClientAPI() chan ToClientAPI {
	return d.toClientAPI
}

func (d *processor) ToFederator() chan ToFederator {
	return d.toFederator
}

// Start starts the Processor, reading from its channels and passing messages back and forth.
func (d *processor) Start() error {
	go func() {
	DistLoop:
		for {
			select {
			case clientMsg := <-d.toClientAPI:
				d.log.Infof("received message TO client API: %+v", clientMsg)
			case federatorMsg := <-d.toFederator:
				d.log.Infof("received message TO federator: %+v", federatorMsg)
			case <-d.stop:
				break DistLoop
			}
		}
	}()
	return nil
}

// Stop stops the processor cleanly, finishing handling any remaining messages before closing down.
// TODO: empty message buffer properly before stopping otherwise we'll lose federating messages.
func (d *processor) Stop() error {
	close(d.stop)
	return nil
}

// ToClientAPI wraps a message that travels from the processor into the client API
type ToClientAPI struct {
	APObjectType   gtsmodel.ActivityStreamsObject
	APActivityType gtsmodel.ActivityStreamsActivity
	Activity       interface{}
}

// ToFederator wraps a message that travels from the processor into the federator
type ToFederator struct {
	APObjectType   gtsmodel.ActivityStreamsObject
	APActivityType gtsmodel.ActivityStreamsActivity
	Activity       interface{}
}
