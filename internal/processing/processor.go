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
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	mm "github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/processing/admin"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/processing/fedi"
	filtersv1 "github.com/superseriousbusiness/gotosocial/internal/processing/filters/v1"
	filtersv2 "github.com/superseriousbusiness/gotosocial/internal/processing/filters/v2"
	"github.com/superseriousbusiness/gotosocial/internal/processing/list"
	"github.com/superseriousbusiness/gotosocial/internal/processing/markers"
	"github.com/superseriousbusiness/gotosocial/internal/processing/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/polls"
	"github.com/superseriousbusiness/gotosocial/internal/processing/report"
	"github.com/superseriousbusiness/gotosocial/internal/processing/search"
	"github.com/superseriousbusiness/gotosocial/internal/processing/status"
	"github.com/superseriousbusiness/gotosocial/internal/processing/stream"
	"github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/processing/user"
	"github.com/superseriousbusiness/gotosocial/internal/processing/workers"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
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
		Required for instance description / terms updating.
	*/

	formatter        *text.Formatter
	parseMentionFunc gtsmodel.ParseMentionFunc

	/*
		SUB-PROCESSORS
	*/

	account   account.Processor
	admin     admin.Processor
	fedi      fedi.Processor
	filtersv1 filtersv1.Processor
	filtersv2 filtersv2.Processor
	list      list.Processor
	markers   markers.Processor
	media     media.Processor
	polls     polls.Processor
	report    report.Processor
	search    search.Processor
	status    status.Processor
	stream    stream.Processor
	timeline  timeline.Processor
	user      user.Processor
	workers   workers.Processor
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

func (p *Processor) FiltersV1() *filtersv1.Processor {
	return &p.filtersv1
}

func (p *Processor) FiltersV2() *filtersv2.Processor {
	return &p.filtersv2
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

func (p *Processor) Polls() *polls.Processor {
	return &p.polls
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
	cleaner *cleaner.Cleaner,
	converter *typeutils.Converter,
	federator *federation.Federator,
	oauthServer oauth.Server,
	mediaManager *mm.Manager,
	state *state.State,
	emailSender email.Sender,
) *Processor {
	var (
		parseMentionFunc = GetParseMentionFunc(state, federator)
		filter           = visibility.NewFilter(state)
	)

	processor := &Processor{
		converter:        converter,
		oauthServer:      oauthServer,
		state:            state,
		formatter:        text.NewFormatter(state.DB),
		parseMentionFunc: parseMentionFunc,
	}

	// Instantiate sub processors.
	//
	// Start with sub processors that will
	// be required by the workers processor.
	common := common.New(state, mediaManager, converter, federator, filter)
	processor.account = account.New(&common, state, converter, mediaManager, federator, filter, parseMentionFunc)
	processor.media = media.New(state, converter, mediaManager, federator.TransportController())
	processor.stream = stream.New(state, oauthServer)

	// Instantiate the rest of the sub
	// processors + pin them to this struct.
	processor.account = account.New(&common, state, converter, mediaManager, federator, filter, parseMentionFunc)
	processor.admin = admin.New(state, cleaner, converter, mediaManager, federator.TransportController(), emailSender)
	processor.fedi = fedi.New(state, &common, converter, federator, filter)
	processor.filtersv1 = filtersv1.New(state, converter, &processor.stream)
	processor.filtersv2 = filtersv2.New(state, converter, &processor.stream)
	processor.list = list.New(state, converter)
	processor.markers = markers.New(state, converter)
	processor.polls = polls.New(&common, state, converter)
	processor.report = report.New(state, converter)
	processor.timeline = timeline.New(state, converter, filter)
	processor.search = search.New(state, federator, converter, filter)
	processor.status = status.New(state, &common, &processor.polls, federator, converter, filter, parseMentionFunc)
	processor.user = user.New(state, converter, oauthServer, emailSender)

	// Workers processor handles asynchronous
	// worker jobs; instantiate it separately
	// and pass subset of sub processors it needs.
	processor.workers = workers.New(
		state,
		federator,
		converter,
		filter,
		emailSender,
		&processor.account,
		&processor.media,
		&processor.stream,
	)

	return processor
}
