package federatingdb

import (
	"context"
	"net/http"

	"github.com/miekg/dns"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Flag(ctx context.Context, flaggable vocab.ActivityStreamsFlag) error {
	log.DebugKV(ctx, "flag", serialize{flaggable})

	// Mark activity as handled.
	f.storeActivityID(flaggable)

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	// Convert received AS flag type to internal report model.
	report, err := f.converter.ASFlagToReport(ctx, flaggable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Requesting acc's domain must be at
	// least a subdomain of the reporting
	// account. i.e. if they're using a
	// different account domain to host.
	if dns.CompareDomainName(
		requesting.Domain,
		report.Account.Domain,
	) < 2 {
		return gtserror.NewfWithCode(http.StatusForbidden, "requester %s does not share a domain with Flag Actor account %s",
			requesting.URI, report.Account.URI)
	}

	// Ensure report received by correct account.
	if report.TargetAccountID != receiving.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "receiver %s is not expected object %s",
			receiving.URI, report.TargetAccount.URI)
	}

	// Generate new ID for report.
	report.ID = id.NewULID()

	// Insert the new validated reported into the database.
	if err := f.state.DB.PutReport(ctx, report); err != nil {
		return gtserror.Newf("error inserting %s into db: %w", report.URI, err)
	}

	// Push message to worker queue to handle report side-effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityFlag,
		APActivityType: ap.ActivityCreate,
		GTSModel:       report,
		Receiving:      receiving,
		Requesting:     requesting,
	})

	return nil
}
