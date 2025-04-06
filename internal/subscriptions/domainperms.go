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

package subscriptions

import (
	"bufio"
	"cmp"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"codeberg.org/gruf/go-kv"

	"github.com/superseriousbusiness/gotosocial/internal/admin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ScheduleJobs schedules domain permission subscription
// fetching + updating using configured parameters.
//
// Returns an error if `MediaCleanupFrom`
// is not a valid format (hh:mm:ss).
func (s *Subscriptions) ScheduleJobs() error {
	const hourMinute = "15:04"

	var (
		now            = time.Now()
		processEvery   = config.GetInstanceSubscriptionsProcessEvery()
		processFromStr = config.GetInstanceSubscriptionsProcessFrom()
	)

	// Parse processFromStr as hh:mm.
	// Resulting time will be on 1 Jan year zero.
	processFrom, err := time.Parse(hourMinute, processFromStr)
	if err != nil {
		return gtserror.Newf(
			"error parsing '%s' in time format 'hh:mm': %w",
			processFromStr, err,
		)
	}

	// Time travel from
	// year zero, groovy.
	firstProcessAt := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		processFrom.Hour(),
		processFrom.Minute(),
		0,
		0,
		now.Location(),
	)

	// Ensure first processing is in the future.
	for firstProcessAt.Before(now) {
		firstProcessAt = firstProcessAt.Add(processEvery)
	}

	fn := func(ctx context.Context, start time.Time) {
		log.Info(ctx, "starting instance subscriptions processing")

		// In blocklist (default) mode, process allows
		// first to provide immunity to block side effects.
		//
		// In allowlist mode, process blocks first to
		// ensure allowlist doesn't override blocks.
		var order [2]gtsmodel.DomainPermissionType
		if config.GetInstanceFederationMode() == config.InstanceFederationModeBlocklist {
			order = [2]gtsmodel.DomainPermissionType{
				gtsmodel.DomainPermissionAllow,
				gtsmodel.DomainPermissionBlock,
			}
		} else {
			order = [2]gtsmodel.DomainPermissionType{
				gtsmodel.DomainPermissionBlock,
				gtsmodel.DomainPermissionAllow,
			}
		}

		// Fetch + process subscribed perms in order.
		for _, permType := range order {
			s.ProcessDomainPermissionSubscriptions(ctx, permType)
		}

		log.Infof(ctx, "finished instance subscriptions processing after %s", time.Since(start))
	}

	log.Infof(nil,
		"scheduling instance subscriptions processing to run every %s, starting from %s; next processing will run at %s",
		processEvery, processFromStr, firstProcessAt,
	)

	// Schedule processing to execute according to schedule.
	if !s.state.Workers.Scheduler.AddRecurring(
		"@subsprocessing",
		firstProcessAt,
		processEvery,
		fn,
	) {
		panic("failed to schedule @subsprocessing")
	}

	return nil
}

// ProcessDomainPermissionSubscriptions processes all domain permission
// subscriptions of the given permission type by, in turn, calling the
// URI of each subscription, parsing the result into a list of domain
// permissions, and creating (or skipping) each permission as appropriate.
func (s *Subscriptions) ProcessDomainPermissionSubscriptions(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
) {
	log.Info(ctx, "start")
	defer log.Info(ctx, "finished")

	// Get permission subscriptions in priority order (highest -> lowest).
	permSubs, err := s.state.DB.GetDomainPermissionSubscriptionsByPriority(ctx, permType)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real db error.
		log.Errorf(ctx, "db error getting domain perm subs by priority: %v", err)
		return
	}

	if len(permSubs) == 0 {
		// No subscriptions of this
		// type, so nothing to do.
		return
	}

	// Get a transport using the instance account,
	// we can reuse this for each HTTP call.
	tsport, err := s.transportController.NewTransportForUsername(ctx, "")
	if err != nil {
		log.Errorf(ctx, "error getting transport for instance account: %v", err)
		return
	}

	for i, permSub := range permSubs {
		// Higher priority permission subs = everything
		// above this permission sub in the slice.
		higherPrios := permSubs[:i]

		_, err := s.ProcessDomainPermissionSubscription(
			ctx,
			permSub,
			tsport,
			higherPrios,
			false, // Not dry. Wet, if you will.
		)
		if err != nil {
			// Real db error.
			log.Errorf(ctx,
				"error processing domain permission subscription %s: %v",
				permSub.URI, err,
			)
			return
		}

		// Update this perm sub.
		err = s.state.DB.UpdateDomainPermissionSubscription(ctx, permSub)
		if err != nil {
			// Real db error.
			log.Errorf(ctx, "db error updating domain perm sub: %v", err)
			return
		}
	}
}

// ProcessDomainPermissionSubscription processes one domain permission
// subscription by dereferencing the URI, parsing the response into a list
// of permissions, and for each discovered permission either creating an
// entry in the database, or ignoring it if it's excluded or already
// covered by a higher-priority subscription.
//
// On success, the slice of discovered DomainPermissions will be returned.
// In case of parsing error, or error on the remote side, permSub.Error
// will be updated with the calling/parsing error, and `nil, nil` will be
// returned. In case of an actual db error, `nil, err` will be returned and
// the caller should handle it.
//
// getHigherPrios should be a function for returning a slice of domain
// permission subscriptions with a higher priority than the given permSub.
//
// If dry == true, then the URI will still be called, and permissions
// will be parsed, but they will not actually be created.
//
// Note that while this function modifies fields on the given permSub,
// it's up to the caller to update it in the database (if desired).
func (s *Subscriptions) ProcessDomainPermissionSubscription(
	ctx context.Context,
	permSub *gtsmodel.DomainPermissionSubscription,
	tsport transport.Transport,
	higherPrios []*gtsmodel.DomainPermissionSubscription,
	dry bool,
) ([]gtsmodel.DomainPermission, error) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"permType", permSub.PermissionType.String()},
			{"permSubURI", permSub.URI},
		}...)

	// Set FetchedAt as we're
	// going to attempt this now.
	permSub.FetchedAt = time.Now()

	// Call the URI, and only skip
	// cache if we're doing a dry run.
	resp, err := tsport.DereferenceDomainPermissions(
		ctx, permSub, dry,
	)
	if err != nil {
		// Couldn't get this one,
		// set error + return.
		errStr := err.Error()
		l.Warnf("couldn't dereference permSubURI: %+v", err)
		permSub.Error = errStr
		return nil, nil
	}

	// If the permissions at URI weren't modified
	// since last time, just update some metadata
	// to indicate a successful fetch, and return.
	if resp.Unmodified {
		l.Debug("received 304 Not Modified from remote")
		permSub.ETag = resp.ETag
		permSub.LastModified = resp.LastModified
		permSub.SuccessfullyFetchedAt = permSub.FetchedAt
		return nil, nil
	}

	// At this point we know we got a 200 OK
	// from the URI, so we've got a live body!
	// Try to parse the body as a list of wantedPerms
	// that the subscription wants to create.
	var wantedPerms []gtsmodel.DomainPermission

	switch permSub.ContentType {

	// text/csv
	case gtsmodel.DomainPermSubContentTypeCSV:
		wantedPerms, err = permsFromCSV(l, permSub.PermissionType, resp.Body)

	// application/json
	case gtsmodel.DomainPermSubContentTypeJSON:
		wantedPerms, err = permsFromJSON(l, permSub.PermissionType, resp.Body)

	// text/plain
	case gtsmodel.DomainPermSubContentTypePlain:
		wantedPerms, err = permsFromPlain(l, permSub.PermissionType, resp.Body)
	}

	if err != nil {
		// We retrieved the permissions from remote but
		// the connection died halfway through transfer,
		// or we couldn't parse the results, or something.
		// Just set error and return.
		errStr := err.Error()
		l.Warnf("couldn't parse results: %+v", err)
		permSub.Error = errStr
		return nil, nil
	}

	if len(wantedPerms) == 0 {
		// Fetch was OK, and parsing was, on the surface at
		// least, OK, but we didn't get any perms. Consider
		// this an error as users will probably want to know.
		const errStr = "fetch successful but parsed zero usable results"
		l.Warn(errStr)
		permSub.Error = errStr
		return nil, nil
	}

	// This can now be considered a successful fetch.
	permSub.SuccessfullyFetchedAt = permSub.FetchedAt
	permSub.ETag = resp.ETag
	permSub.LastModified = resp.LastModified
	permSub.Error = ""

	// Keep track of which domain perms are
	// created (or would be, if dry == true).
	createdPerms := make([]gtsmodel.DomainPermission, 0, len(wantedPerms))

	// Iterate through wantedPerms and
	// create (or dry create) each one.
	for _, wantedPerm := range wantedPerms {
		l := l.WithField("domain", wantedPerm.GetDomain())
		created, err := s.processDomainPermission(
			ctx, l,
			wantedPerm,
			permSub,
			higherPrios,
			dry,
		)
		if err != nil {
			// Proper db error.
			return nil, err
		}

		if !created {
			continue
		}

		createdPerms = append(createdPerms, wantedPerm)
	}

	return createdPerms, nil
}

// processDomainPermission processes one wanted domain
// permission discovered via a domain permission sub's URI.
//
// If dry == true, then the returned boolean indicates whether
// the permission would actually be created. If dry == false,
// the bool indicates whether the permission was created or adopted.
//
// Error will only be returned in case of an actual database
// error, else the error will be logged and nil returned.
func (s *Subscriptions) processDomainPermission(
	ctx context.Context,
	l log.Entry,
	wantedPerm gtsmodel.DomainPermission,
	permSub *gtsmodel.DomainPermissionSubscription,
	higherPrios []*gtsmodel.DomainPermissionSubscription,
	dry bool,
) (bool, error) {
	// If domain is excluded from automatic
	// permission creation, don't process it.
	domain := wantedPerm.GetDomain()
	excluded, err := s.state.DB.IsDomainPermissionExcluded(ctx, domain)
	if err != nil {
		// Proper db error.
		return false, err
	}

	if excluded {
		l.Debug("domain is excluded, skipping")
		return false, err
	}

	// Check if a permission already exists for
	// this domain, and if it's covered already
	// by a higher-priority subscription.
	existingPerm, covered, err := s.existingCovered(
		ctx, permSub.PermissionType, domain, higherPrios,
	)
	if err != nil {
		// Proper db error.
		return false, err
	}

	if covered {
		l.Debug("domain is covered by a higher-priority subscription, skipping")
		return false, err
	}

	// True if a perm already exists.
	// Note: != nil doesn't work because
	// of Go interface idiosyncracies.
	existing := !util.IsNil(existingPerm)

	if dry {
		// If this is a dry run, return
		// now without doing any DB changes.
		return !existing, nil
	}

	// Handle perm creation differently depending
	// on whether or not a perm already existed.
	switch {

	case !existing && *permSub.AsDraft:
		// No existing perm, create as draft.
		err = s.state.DB.PutDomainPermissionDraft(
			ctx,
			&gtsmodel.DomainPermissionDraft{
				ID:                 id.NewULID(),
				PermissionType:     permSub.PermissionType,
				Domain:             domain,
				CreatedByAccountID: permSub.CreatedByAccount.ID,
				CreatedByAccount:   permSub.CreatedByAccount,
				PrivateComment:     permSub.URI,
				PublicComment:      wantedPerm.GetPublicComment(),
				Obfuscate:          wantedPerm.GetObfuscate(),
				SubscriptionID:     permSub.ID,
			},
		)

	case !existing && !*permSub.AsDraft:
		// No existing perm, create a new one of the
		// appropriate type, and process side effects.
		var (
			insertF func() error
			action  *gtsmodel.AdminAction
			actionF admin.ActionF
		)

		if permSub.PermissionType == gtsmodel.DomainPermissionBlock {
			// Prepare to insert + process a block.
			domainBlock := &gtsmodel.DomainBlock{
				ID:                 id.NewULID(),
				Domain:             domain,
				CreatedByAccountID: permSub.CreatedByAccount.ID,
				CreatedByAccount:   permSub.CreatedByAccount,
				PrivateComment:     permSub.URI,
				PublicComment:      wantedPerm.GetPublicComment(),
				Obfuscate:          wantedPerm.GetObfuscate(),
				SubscriptionID:     permSub.ID,
			}
			insertF = func() error { return s.state.DB.PutDomainBlock(ctx, domainBlock) }

			action = &gtsmodel.AdminAction{
				ID:             id.NewULID(),
				TargetCategory: gtsmodel.AdminActionCategoryDomain,
				TargetID:       domain,
				Type:           gtsmodel.AdminActionSuspend,
				AccountID:      permSub.CreatedByAccountID,
			}
			actionF = s.state.AdminActions.DomainBlockF(action.ID, domainBlock)

		} else {
			// Prepare to insert + process an allow.
			domainAllow := &gtsmodel.DomainAllow{
				ID:                 id.NewULID(),
				Domain:             domain,
				CreatedByAccountID: permSub.CreatedByAccount.ID,
				CreatedByAccount:   permSub.CreatedByAccount,
				PrivateComment:     permSub.URI,
				PublicComment:      wantedPerm.GetPublicComment(),
				Obfuscate:          wantedPerm.GetObfuscate(),
				SubscriptionID:     permSub.ID,
			}
			insertF = func() error { return s.state.DB.PutDomainAllow(ctx, domainAllow) }

			action = &gtsmodel.AdminAction{
				ID:             id.NewULID(),
				TargetCategory: gtsmodel.AdminActionCategoryDomain,
				TargetID:       domain,
				Type:           gtsmodel.AdminActionUnsuspend,
				AccountID:      permSub.CreatedByAccountID,
			}
			actionF = s.state.AdminActions.DomainAllowF(action.ID, domainAllow)
		}

		// Insert the new perm in the db.
		if err = insertF(); err != nil {
			// Couldn't insert wanted perm,
			// don't process side effects.
			break
		}

		// Run admin action to process
		// side effects of permission.
		err = s.state.AdminActions.Run(ctx, action, actionF)

	case existingPerm.IsOrphan():
		// Perm already exists, but it's not managed
		// by a subscription, ie., it's an orphan.
		if !*permSub.AdoptOrphans {
			l.Debug("permission exists as an orphan that we shouldn't adopt, skipping")
			return false, nil
		}

		// Orphan is adoptable, so adopt
		// it by rewriting some fields.
		//
		// TODO: preserve previous private
		// + public comment in some way.
		l.Debug("adopting orphan permission")
		err = s.adoptPerm(
			ctx,
			existingPerm,
			permSub,
			wantedPerm.GetObfuscate(),
			permSub.URI,
			wantedPerm.GetPublicComment(),
		)

	case existingPerm.GetSubscriptionID() != permSub.ID:
		// Perm already exists, and is managed
		// by a lower-priority subscription.
		// Take it for ourselves.
		//
		// TODO: preserve previous private
		// + public comment in some way.
		l.Debug("taking over permission from lower-priority subscription")
		err = s.adoptPerm(
			ctx,
			existingPerm,
			permSub,
			wantedPerm.GetObfuscate(),
			permSub.URI,
			wantedPerm.GetPublicComment(),
		)

	default:
		// Perm exists and is managed by us.
		//
		// TODO: update public/private comment
		// from latest version if it's changed.
		l.Debug("permission already exists and is managed by this subscription, skipping")
		return false, nil
	}

	if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
		// Proper db error.
		return false, err
	}

	return true, nil
}

func permsFromCSV(
	l log.Entry,
	permType gtsmodel.DomainPermissionType,
	body io.ReadCloser,
) ([]gtsmodel.DomainPermission, error) {
	csvReader := csv.NewReader(body)

	// Read and validate column headers.
	columnHeaders, err := csvReader.Read()
	if err != nil {
		body.Close()
		return nil, gtserror.NewfAt(3, "error decoding csv column headers: %w", err)
	}

	var (
		domainI        *int
		severityI      *int
		publicCommentI *int
		obfuscateI     *int
	)

	for i, columnHeader := range columnHeaders {
		// Remove leading # if present.
		columnHeader = strings.TrimLeft(columnHeader, "#")

		// Find index of each column header we
		// care about, ensuring no duplicates.
		switch {

		case columnHeader == "domain":
			if domainI != nil {
				body.Close()
				err := gtserror.NewfAt(3, "duplicate domain column header in csv: %+v", columnHeaders)
				return nil, err
			}
			domainI = &i

		case columnHeader == "severity":
			if severityI != nil {
				body.Close()
				err := gtserror.NewfAt(3, "duplicate severity column header in csv: %+v", columnHeaders)
				return nil, err
			}
			severityI = &i

		case columnHeader == "public_comment" || columnHeader == "comment":
			if publicCommentI != nil {
				body.Close()
				err := gtserror.NewfAt(3, "duplicate public_comment or comment column header in csv: %+v", columnHeaders)
				return nil, err
			}
			publicCommentI = &i

		case columnHeader == "obfuscate":
			if obfuscateI != nil {
				body.Close()
				err := gtserror.NewfAt(3, "duplicate obfuscate column header in csv: %+v", columnHeaders)
				return nil, err
			}
			obfuscateI = &i
		}
	}

	// Ensure we have at least a domain
	// index, as that's the bare minimum.
	if domainI == nil {
		body.Close()
		err := gtserror.NewfAt(3, "no domain column header in csv: %+v", columnHeaders)
		return nil, err
	}

	// Read remaining CSV records.
	records, err := csvReader.ReadAll()

	// Totally done
	// with body now.
	body.Close()

	// Check for decode error.
	if err != nil {
		err := gtserror.NewfAt(3, "error decoding body into csv: %w", err)
		return nil, err
	}

	// Make sure we actually
	// have some records.
	if len(records) == 0 {
		return nil, nil
	}

	// Convert records to permissions slice.
	perms := make([]gtsmodel.DomainPermission, 0, len(records))
	for _, record := range records {
		if len(record) != 6 {
			l.Warnf("skipping invalid-length record: %+v", record)
			continue
		}

		// Skip records that specify severity
		// that's not "suspend" (we don't support
		// "silence" or "limit" or whatever yet).
		if severityI != nil {
			severity := record[*severityI]
			if severity != "suspend" {
				l.Warnf("skipping non-suspend record: %+v", record)
				continue
			}
		}

		// Normalize + validate domain.
		domainRaw := record[*domainI]
		domain, err := util.PunifySafely(domainRaw)
		if err != nil {
			l.Warnf("skipping invalid domain %s: %+v", domainRaw, err)
			continue
		}

		// Instantiate the permission
		// as either block or allow.
		var perm gtsmodel.DomainPermission
		switch permType {
		case gtsmodel.DomainPermissionBlock:
			perm = &gtsmodel.DomainBlock{Domain: domain}
		case gtsmodel.DomainPermissionAllow:
			perm = &gtsmodel.DomainAllow{Domain: domain}
		}

		// Set remaining optional fields
		// if they're present in the CSV.
		if publicCommentI != nil {
			perm.SetPublicComment(record[*publicCommentI])
		}

		var obfuscate bool
		if obfuscateI != nil {
			obfuscate, err = strconv.ParseBool(record[*obfuscateI])
			if err != nil {
				l.Warnf("couldn't parse obfuscate field of record: %+v", record)
				continue
			}
		}
		perm.SetObfuscate(&obfuscate)

		// We're done.
		perms = append(perms, perm)
	}

	return perms, nil
}

func permsFromJSON(
	l log.Entry,
	permType gtsmodel.DomainPermissionType,
	body io.ReadCloser,
) ([]gtsmodel.DomainPermission, error) {
	var (
		dec      = json.NewDecoder(body)
		apiPerms = make([]*apimodel.DomainPermission, 0)
	)

	// Read body into memory as
	// slice of domain permissions.
	if err := dec.Decode(&apiPerms); err != nil {
		_ = body.Close() // ensure closed.
		return nil, gtserror.NewfAt(3, "error decoding into json: %w", err)
	}

	// Perform a secondary decode just to ensure we drained the
	// entirety of the data source. Error indicates either extra
	// trailing garbage, or multiple JSON values (invalid data).
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		_ = body.Close() // ensure closed.
		return nil, gtserror.NewfAt(3, "data remaining after json")
	}

	// Done with body.
	_ = body.Close()

	// Convert apimodel perms to barebones internal perms.
	perms := make([]gtsmodel.DomainPermission, 0, len(apiPerms))
	for _, apiPerm := range apiPerms {

		// Normalize + validate domain.
		domainRaw := apiPerm.Domain.Domain
		domain, err := util.PunifySafely(domainRaw)
		if err != nil {
			l.Warnf("skipping invalid domain %s: %+v", domainRaw, err)
			continue
		}

		// Instantiate the permission
		// as either block or allow.
		var perm gtsmodel.DomainPermission
		switch permType {
		case gtsmodel.DomainPermissionBlock:
			perm = &gtsmodel.DomainBlock{Domain: domain}
		case gtsmodel.DomainPermissionAllow:
			perm = &gtsmodel.DomainAllow{Domain: domain}
		}

		// Set remaining fields.
		publicComment := cmp.Or(apiPerm.PublicComment, apiPerm.Comment)
		perm.SetPublicComment(util.PtrOrZero(publicComment))
		perm.SetObfuscate(util.Ptr(util.PtrOrZero(apiPerm.Obfuscate)))

		// We're done.
		perms = append(perms, perm)
	}

	return perms, nil
}

func permsFromPlain(
	l log.Entry,
	permType gtsmodel.DomainPermissionType,
	body io.ReadCloser,
) ([]gtsmodel.DomainPermission, error) {
	// Scan + split by line.
	sc := bufio.NewScanner(body)

	// Read into domains
	// line by line.
	var domains []string
	for sc.Scan() {
		domains = append(domains, sc.Text())
	}

	// Whatever happened, we're
	// done with the body now.
	body.Close()

	// Check if error reading body.
	if err := sc.Err(); err != nil {
		return nil, gtserror.NewfAt(3, "error decoding into plain: %w", err)
	}

	// Convert raw domains to permissions.
	perms := make([]gtsmodel.DomainPermission, 0, len(domains))
	for _, domainRaw := range domains {

		// Normalize + validate domain as ASCII.
		domain, err := util.PunifySafely(domainRaw)
		if err != nil {
			l.Warnf("skipping invalid domain %s: %+v", domainRaw, err)
			continue
		}

		// Instantiate the permission
		// as either block or allow.
		var perm gtsmodel.DomainPermission
		switch permType {
		case gtsmodel.DomainPermissionBlock:
			perm = &gtsmodel.DomainBlock{
				Domain:    domain,
				Obfuscate: util.Ptr(false),
			}
		case gtsmodel.DomainPermissionAllow:
			perm = &gtsmodel.DomainAllow{
				Domain:    domain,
				Obfuscate: util.Ptr(false),
			}
		}

		// We're done.
		perms = append(perms, perm)
	}

	return perms, nil
}

func (s *Subscriptions) existingCovered(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
	domain string,
	higherPrios []*gtsmodel.DomainPermissionSubscription,
) (
	existingPerm gtsmodel.DomainPermission,
	covered bool,
	err error,
) {
	// Check for existing perm
	// of appropriate type.
	var dbErr error
	switch permType {
	case gtsmodel.DomainPermissionBlock:
		existingPerm, dbErr = s.state.DB.GetDomainBlock(ctx, domain)
	case gtsmodel.DomainPermissionAllow:
		existingPerm, dbErr = s.state.DB.GetDomainAllow(ctx, domain)
	}

	if dbErr != nil && !errors.Is(dbErr, db.ErrNoEntries) {
		// Real db error.
		err = dbErr
		return
	}

	if util.IsNil(existingPerm) {
		// Can't be covered if
		// no existing perm.
		return
	}

	subscriptionID := existingPerm.GetSubscriptionID()
	if subscriptionID == "" {
		// Can't be covered if
		// no subscription ID.
		return
	}

	// Covered if subscription ID is in the slice
	// of higher-priority permission subscriptions.
	covered = slices.ContainsFunc(
		higherPrios,
		func(permSub *gtsmodel.DomainPermissionSubscription) bool {
			return permSub.ID == subscriptionID
		},
	)

	return
}

func (s *Subscriptions) adoptPerm(
	ctx context.Context,
	perm gtsmodel.DomainPermission,
	permSub *gtsmodel.DomainPermissionSubscription,
	obfuscate *bool,
	privateComment string,
	publicComment string,
) error {
	// Set to our sub ID + this subs's
	// account as we're managing it now.
	perm.SetSubscriptionID(permSub.ID)
	perm.SetCreatedByAccountID(permSub.CreatedByAccount.ID)
	perm.SetCreatedByAccount(permSub.CreatedByAccount)

	// Set new metadata on the perm.
	perm.SetPrivateComment(privateComment)
	perm.SetPublicComment(publicComment)

	// Avoid trying to blat nil into the db directly by
	// defaulting to false if not set on wanted perm.
	perm.SetObfuscate(cmp.Or(obfuscate, util.Ptr(false)))

	// Update the perm in the db.
	var err error
	switch p := perm.(type) {
	case *gtsmodel.DomainBlock:
		err = s.state.DB.UpdateDomainBlock(ctx, p)
	case *gtsmodel.DomainAllow:
		err = s.state.DB.UpdateDomainAllow(ctx, p)
	}

	return err
}
