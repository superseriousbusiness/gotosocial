package streaming

import (
	"github.com/gorilla/websocket"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) StreamForAccount(c *websocket.Conn, account *gtsmodel.Account, streamType string) gtserror.WithCode {

	v, loaded := p.streamMap.LoadOrStore(account.ID, sync.Slice)
	if loaded {

	}

	return nil
}

type streams struct {
	accountID string
	
}
