package processing

import (
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) AuthorizeStreamingRequest(accessToken string) (*gtsmodel.Account, error) {
	return p.streamingProcessor.AuthorizeStreamingRequest(accessToken)
}

func (p *processor) OpenStreamForAccount(account *gtsmodel.Account, streamType string) (*gtsmodel.Stream, gtserror.WithCode) {
	return p.streamingProcessor.OpenStreamForAccount(account, streamType)
}
