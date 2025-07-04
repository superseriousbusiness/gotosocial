// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package testrig

import (
	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/filter/interaction"
	"code.superseriousbusiness.org/gotosocial/internal/filter/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/subscriptions"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// TestStructs encapsulates structs needed to
// run one test independently. Each test should
// call SetupTestStructs to get a new TestStructs,
// and defer TearDownTestStructs to close it when
// the test is complete. The reason for doing things
// this way here is to prevent the tests in a
// package from overwriting one another's processors
// and worker queues, which was causing issues
// when running all tests at once.
type TestStructs struct {
	State               *state.State
	Common              *common.Processor
	Processor           *processing.Processor
	HTTPClient          *MockHTTPClient
	TypeConverter       *typeutils.Converter
	EmailSender         email.Sender
	WebPushSender       *WebPushMockSender
	TransportController transport.Controller
	InteractionFilter   *interaction.Filter
	StatusFilter        *status.Filter
}

func SetupTestStructs(
	rMediaPath string,
	rTemplatePath string,
) *TestStructs {
	state := state.State{}

	state.Caches.Init()

	db := NewTestDB(&state)
	state.DB = db
	state.AdminActions = admin.New(db, &state.Workers)

	storage := NewInMemoryStorage()
	state.Storage = storage
	typeconverter := typeutils.NewConverter(&state)
	visFilter := visibility.NewFilter(&state)
	muteFilter := mutes.NewFilter(&state)
	intFilter := interaction.NewFilter(&state)
	statusFilter := status.NewFilter(&state)

	httpClient := NewMockHTTPClient(nil, rMediaPath)
	httpClient.TestRemotePeople = NewTestFediPeople()
	httpClient.TestRemoteStatuses = NewTestFediStatuses()

	transportController := NewTestTransportController(&state, httpClient)
	mediaManager := NewTestMediaManager(&state)
	federator := NewTestFederator(&state, transportController, mediaManager)
	oauthServer := NewTestOauthServer(&state)
	emailSender := NewEmailSender(rTemplatePath, nil)
	webPushSender := NewWebPushMockSender()

	common := common.New(
		&state,
		mediaManager,
		typeconverter,
		federator,
		visFilter,
		muteFilter,
		statusFilter,
	)

	processor := processing.NewProcessor(
		cleaner.New(&state),
		subscriptions.New(&state, transportController, typeconverter),
		typeconverter,
		federator,
		oauthServer,
		mediaManager,
		&state,
		emailSender,
		webPushSender,
		visFilter,
		muteFilter,
		intFilter,
		statusFilter,
	)

	StartWorkers(&state, processor.Workers())

	StandardDBSetup(db, nil)
	StandardStorageSetup(storage, rMediaPath)

	return &TestStructs{
		State:               &state,
		Common:              &common,
		Processor:           processor,
		HTTPClient:          httpClient,
		TypeConverter:       typeconverter,
		EmailSender:         emailSender,
		WebPushSender:       webPushSender,
		TransportController: transportController,
		InteractionFilter:   intFilter,
		StatusFilter:        statusFilter,
	}
}

func TearDownTestStructs(testStructs *TestStructs) {
	StandardDBTeardown(testStructs.State.DB)
	StandardStorageTeardown(testStructs.State.Storage)
	StopWorkers(testStructs.State)
}
