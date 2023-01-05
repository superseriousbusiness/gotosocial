/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package processing

import (
	"context"
	"net/http"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *processor) GetFediUser(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetUser(ctx, requestedUsername, requestURL)
}

func (p *processor) GetFediFollowers(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetFollowers(ctx, requestedUsername, requestURL)
}

func (p *processor) GetFediFollowing(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetFollowing(ctx, requestedUsername, requestURL)
}

func (p *processor) GetFediStatus(ctx context.Context, requestedUsername string, requestedStatusID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetStatus(ctx, requestedUsername, requestedStatusID, requestURL)
}

func (p *processor) GetFediStatusReplies(ctx context.Context, requestedUsername string, requestedStatusID string, page bool, onlyOtherAccounts bool, minID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetStatusReplies(ctx, requestedUsername, requestedStatusID, page, onlyOtherAccounts, minID, requestURL)
}

func (p *processor) GetFediOutbox(ctx context.Context, requestedUsername string, page bool, maxID string, minID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetOutbox(ctx, requestedUsername, page, maxID, minID, requestURL)
}

func (p *processor) GetFediEmoji(ctx context.Context, requestedEmojiID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	return p.federationProcessor.GetEmoji(ctx, requestedEmojiID, requestURL)
}

func (p *processor) GetWebfingerAccount(ctx context.Context, requestedUsername string) (*apimodel.WellKnownResponse, gtserror.WithCode) {
	return p.federationProcessor.GetWebfingerAccount(ctx, requestedUsername)
}

func (p *processor) GetNodeInfoRel(ctx context.Context) (*apimodel.WellKnownResponse, gtserror.WithCode) {
	return p.federationProcessor.GetNodeInfoRel(ctx)
}

func (p *processor) GetNodeInfo(ctx context.Context) (*apimodel.Nodeinfo, gtserror.WithCode) {
	return p.federationProcessor.GetNodeInfo(ctx)
}

func (p *processor) InboxPost(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return p.federationProcessor.PostInbox(ctx, w, r)
}
