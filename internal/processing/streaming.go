package processing

import (
	"github.com/gorilla/websocket"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) AuthorizeStreamingRequest(accessToken string) (*gtsmodel.Account, error) {
	return p.streamingProcessor.AuthorizeStreamingRequest(accessToken)
}

func (p *processor) OpenStreamForAccount(c *websocket.Conn, account *gtsmodel.Account, streamType string) gtserror.WithCode {
	return p.streamingProcessor.OpenStreamForAccount(c, account, streamType)
}
