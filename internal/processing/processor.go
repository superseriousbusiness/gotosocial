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

package processing

import (
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	mm "github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/processing/admin"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/processing/fedi"
	"github.com/superseriousbusiness/gotosocial/internal/processing/list"
	"github.com/superseriousbusiness/gotosocial/internal/processing/markers"
	"github.com/superseriousbusiness/gotosocial/internal/processing/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/report"
	"github.com/superseriousbusiness/gotosocial/internal/processing/search"
	"github.com/superseriousbusiness/gotosocial/internal/processing/status"
	"github.com/superseriousbusiness/gotosocial/internal/processing/stream"
	"github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/processing/user"
	"github.com/superseriousbusiness/gotosocial/internal/processing/workers"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// Processor groups together processing functions and
// sub processors for handling actions + events coming
// from either the client or federating APIs.
//
// Many of the functions available through this struct
// or sub processors will trigger asynchronous processing
// via the workers contained in state.
type Processor struct {
	converter   *typeutils.Converter
	oauthServer oauth.Server
	state       *state.State

	/*
		SUB-PROCESSORS
	*/

	account  account.Processor
	admin    admin.Processor
	fedi     fedi.Processor
	list     list.Processor
	markers  markers.Processor
	media    media.Processor
	report   report.Processor
	search   search.Processor
	status   status.Processor
	stream   stream.Processor
	timeline timeline.Processor
	user     user.Processor
	workers  workers.Processor
}

func (p *Processor) Account() *account.Processor {
	return &p.account
}

func (p *Processor) Admin() *admin.Processor {
	return &p.admin
}

func (p *Processor) Fedi() *fedi.Processor {
	return &p.fedi
}

func (p *Processor) List() *list.Processor {
	return &p.list
}

func (p *Processor) Markers() *markers.Processor {
	return &p.markers
}

func (p *Processor) Media() *media.Processor {
	return &p.media
}

func (p *Processor) Report() *report.Processor {
	return &p.report
}

func (p *Processor) Search() *search.Processor {
	return &p.search
}

func (p *Processor) Status() *status.Processor {
	return &p.status
}

func (p *Processor) Stream() *stream.Processor {
	return &p.stream
}

func (p *Processor) Timeline() *timeline.Processor {
	return &p.timeline
}

func (p *Processor) User() *user.Processor {
	return &p.user
}

func (p *Processor) Workers() *workers.Processor {
	return &p.workers
}

// NewProcessor returns a new Processor.
func NewProcessor(
	converter *typeutils.Converter,
	federator federation.Federator,
	oauthServer oauth.Server,
	mediaManager *mm.Manager,
	state *state.State,
	emailSender email.Sender,
) *Processor {
	var (
		parseMentionFunc = GetParseMentionFunc(state.DB, federator)
		filter           = visibility.NewFilter(state)
	)

	processor := &Processor{
		converter:   converter,
		oauthServer: oauthServer,
		state:       state,
	}

	// Instantiate sub processors.
	//
	// Start with sub processors that will
	// be required by the workers processor.
	commonProcessor := common.New(state, converter, federator, filter)
	accountProcessor := account.New(&commonProcessor, state, converter, mediaManager, oauthServer, federator, filter, parseMentionFunc)
	mediaProcessor := media.New(state, converter, mediaManager, federator.TransportController())
	streamProcessor := stream.New(state, oauthServer)

	// Instantiate the rest of the sub
	// processors + pin them to this struct.
	processor.account = accountProcessor
	processor.admin = admin.New(state, converter, mediaManager, federator.TransportController(), emailSender)
	processor.fedi = fedi.New(state, converter, federator, filter)
	processor.list = list.New(state, converter)
	processor.markers = markers.New(state, converter)
	processor.media = mediaProcessor
	processor.report = report.New(state, converter)
	processor.timeline = timeline.New(state, converter, filter)
	processor.search = search.New(state, federator, converter, filter)
	processor.status = status.New(state, federator, converter, filter, parseMentionFunc)
	processor.stream = streamProcessor
	processor.user = user.New(state, emailSender)

	// Workers processor handles asynchronous
	// worker jobs; instantiate it separately
	// and pass subset of sub processors it needs.
	processor.workers = workers.New(
		state,
		federator,
		converter,
		filter,
		emailSender,
		&accountProcessor,
		&mediaProcessor,
		&streamProcessor,
	)

	return processor
}
