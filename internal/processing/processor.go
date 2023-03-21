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
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	mm "github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/processing/admin"
	"github.com/superseriousbusiness/gotosocial/internal/processing/fedi"
	"github.com/superseriousbusiness/gotosocial/internal/processing/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/report"
	"github.com/superseriousbusiness/gotosocial/internal/processing/status"
	"github.com/superseriousbusiness/gotosocial/internal/processing/stream"
	"github.com/superseriousbusiness/gotosocial/internal/processing/user"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

type Processor struct {
	federator       federation.Federator
	tc              typeutils.TypeConverter
	oauthServer     oauth.Server
	mediaManager    mm.Manager
	statusTimelines timeline.Manager
	state           *state.State
	emailSender     email.Sender
	filter          *visibility.Filter

	/*
		SUB-PROCESSORS
	*/

	account account.Processor
	admin   admin.Processor
	fedi    fedi.Processor
	media   media.Processor
	report  report.Processor
	status  status.Processor
	stream  stream.Processor
	user    user.Processor
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

func (p *Processor) Media() *media.Processor {
	return &p.media
}

func (p *Processor) Report() *report.Processor {
	return &p.report
}

func (p *Processor) Status() *status.Processor {
	return &p.status
}

func (p *Processor) Stream() *stream.Processor {
	return &p.stream
}

func (p *Processor) User() *user.Processor {
	return &p.user
}

// NewProcessor returns a new Processor.
func NewProcessor(
	tc typeutils.TypeConverter,
	federator federation.Federator,
	oauthServer oauth.Server,
	mediaManager mm.Manager,
	state *state.State,
	emailSender email.Sender,
) *Processor {
	parseMentionFunc := GetParseMentionFunc(state.DB, federator)

	filter := visibility.NewFilter(state)

	processor := &Processor{
		federator:    federator,
		tc:           tc,
		oauthServer:  oauthServer,
		mediaManager: mediaManager,
		statusTimelines: timeline.NewManager(
			StatusGrabFunction(state.DB),
			StatusFilterFunction(state.DB, filter),
			StatusPrepareFunction(state.DB, tc),
			StatusSkipInsertFunction(),
		),
		state:       state,
		filter:      filter,
		emailSender: emailSender,
	}

	// sub processors
	processor.account = account.New(state, tc, mediaManager, oauthServer, federator, filter, parseMentionFunc)
	processor.admin = admin.New(state, tc, mediaManager, federator.TransportController(), emailSender)
	processor.fedi = fedi.New(state, tc, federator, filter)
	processor.media = media.New(state, tc, mediaManager, federator.TransportController())
	processor.report = report.New(state, tc)
	processor.status = status.New(state, tc, filter, parseMentionFunc)
	processor.stream = stream.New(state, oauthServer)
	processor.user = user.New(state, emailSender)

	return processor
}

func (p *Processor) EnqueueClientAPI(ctx context.Context, msgs ...messages.FromClientAPI) {
	log.Trace(ctx, "enqueuing")
	_ = p.state.Workers.ClientAPI.MustEnqueueCtx(ctx, func(ctx context.Context) {
		for _, msg := range msgs {
			log.Trace(ctx, "processing: %+v", msg)
			if err := p.ProcessFromClientAPI(ctx, msg); err != nil {
				log.Errorf(ctx, "error processing client API message: %v", err)
			}
		}
	})
}

func (p *Processor) EnqueueFederator(ctx context.Context, msgs ...messages.FromFederator) {
	log.Trace(ctx, "enqueuing")
	_ = p.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
		for _, msg := range msgs {
			log.Trace(ctx, "processing: %+v", msg)
			if err := p.ProcessFromFederator(ctx, msg); err != nil {
				log.Errorf(ctx, "error processing federator message: %v", err)
			}
		}
	})
}

// Start starts the Processor.
func (p *Processor) Start() error {
	return p.statusTimelines.Start()
}

// Stop stops the processor cleanly.
func (p *Processor) Stop() error {
	return p.statusTimelines.Stop()
}
