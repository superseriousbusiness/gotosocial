package federatingdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (f *federatingDB) Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Announce",
		},
	)
	m, err := streams.Serialize(announce)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	l.Debugf("received ANNOUNCE %s", string(b))

	targetAcctI := ctx.Value(util.APAccount)
	if targetAcctI == nil {
		l.Error("target account wasn't set on context")
		return nil
	}
	targetAcct, ok := targetAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("target account was set on context but couldn't be parsed")
		return nil
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("from federator channel wasn't set on context")
		return nil
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("from federator channel was set on context but couldn't be parsed")
		return nil
	}

	boost, isNew, err := f.typeConverter.ASAnnounceToStatus(announce)
	if err != nil {
		return fmt.Errorf("Announce: error converting announce to boost: %s", err)
	}

	if !isNew {
		// nothing to do here if this isn't a new announce
		return nil
	}

	// it's a new announce so pass it back to the processor async for dereferencing etc
	fromFederatorChan <- gtsmodel.FromFederator{
		APObjectType:     gtsmodel.ActivityStreamsAnnounce,
		APActivityType:   gtsmodel.ActivityStreamsCreate,
		GTSModel:         boost,
		ReceivingAccount: targetAcct,
	}

	return nil
}
