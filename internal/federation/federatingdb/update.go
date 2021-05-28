package federatingdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Update sets an existing entry to the database based on the value's
// id.
//
// Note that Activity values received from federated peers may also be
// updated in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Update(ctx context.Context, asType vocab.Type) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "Update",
			"asType": asType.GetTypeName(),
		},
	)
	m, err := streams.Serialize(asType)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	l.Debugf("received UPDATE asType %s", string(b))

	receivingAcctI := ctx.Value(util.APAccount)
	if receivingAcctI == nil {
		l.Error("receiving account wasn't set on context")
	}
	receivingAcct, ok := receivingAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("receiving account was set on context but couldn't be parsed")
	}

	requestingAcctI := ctx.Value(util.APRequestingAccount)
	if receivingAcctI == nil {
		l.Error("requesting account wasn't set on context")
	}
	requestingAcct, ok := requestingAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("requesting account was set on context but couldn't be parsed")
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("from federator channel wasn't set on context")
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("from federator channel was set on context but couldn't be parsed")
	}

	typeName := asType.GetTypeName()
	if typeName == gtsmodel.ActivityStreamsApplication ||
		typeName == gtsmodel.ActivityStreamsGroup ||
		typeName == gtsmodel.ActivityStreamsOrganization ||
		typeName == gtsmodel.ActivityStreamsPerson ||
		typeName == gtsmodel.ActivityStreamsService {
		// it's an UPDATE to some kind of account
		var accountable typeutils.Accountable

		switch asType.GetTypeName() {
		case gtsmodel.ActivityStreamsApplication:
			l.Debug("got update for APPLICATION")
			i, ok := asType.(vocab.ActivityStreamsApplication)
			if !ok {
				return errors.New("could not convert type to application")
			}
			accountable = i
		case gtsmodel.ActivityStreamsGroup:
			l.Debug("got update for GROUP")
			i, ok := asType.(vocab.ActivityStreamsGroup)
			if !ok {
				return errors.New("could not convert type to group")
			}
			accountable = i
		case gtsmodel.ActivityStreamsOrganization:
			l.Debug("got update for ORGANIZATION")
			i, ok := asType.(vocab.ActivityStreamsOrganization)
			if !ok {
				return errors.New("could not convert type to organization")
			}
			accountable = i
		case gtsmodel.ActivityStreamsPerson:
			l.Debug("got update for PERSON")
			i, ok := asType.(vocab.ActivityStreamsPerson)
			if !ok {
				return errors.New("could not convert type to person")
			}
			accountable = i
		case gtsmodel.ActivityStreamsService:
			l.Debug("got update for SERVICE")
			i, ok := asType.(vocab.ActivityStreamsService)
			if !ok {
				return errors.New("could not convert type to service")
			}
			accountable = i
		}

		updatedAcct, err := f.typeConverter.ASRepresentationToAccount(accountable, true)
		if err != nil {
			return fmt.Errorf("error converting to account: %s", err)
		}

		if updatedAcct.Domain == f.config.Host {
			// no need to update local accounts
			// in fact, if we do this will break the shit out of things so do NOT
			return nil
		}

		if requestingAcct.URI != updatedAcct.URI {
			return fmt.Errorf("update for account %s was requested by account %s, this is not valid", updatedAcct.URI, requestingAcct.URI)
		}

		updatedAcct.ID = requestingAcct.ID // set this here so the db will update properly instead of trying to PUT this and getting constraint issues
		if err := f.db.UpdateByID(requestingAcct.ID, updatedAcct); err != nil {
			return fmt.Errorf("database error inserting updated account: %s", err)
		}

		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsProfile,
			APActivityType:   gtsmodel.ActivityStreamsUpdate,
			GTSModel:         updatedAcct,
			ReceivingAccount: receivingAcct,
		}

	}

	return nil
}
