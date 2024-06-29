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

package workers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// emailUserReportClosed emails the user who created the
// given report, to inform them the report has been closed.
func (s *Surface) emailUserReportClosed(ctx context.Context, report *gtsmodel.Report) error {
	user, err := s.State.DB.GetUserByAccountID(ctx, report.Account.ID)
	if err != nil {
		return gtserror.Newf("db error getting user: %w", err)
	}

	if user.ConfirmedAt.IsZero() ||
		!*user.Approved ||
		*user.Disabled ||
		user.Email == "" {
		// Only email users who:
		// - are confirmed
		// - are approved
		// - are not disabled
		// - have an email address
		return nil
	}

	instance, err := s.State.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return gtserror.Newf("db error getting instance: %w", err)
	}

	if err := s.State.DB.PopulateReport(ctx, report); err != nil {
		return gtserror.Newf("error populating report: %w", err)
	}

	reportClosedData := email.ReportClosedData{
		Username:             report.Account.Username,
		InstanceURL:          instance.URI,
		InstanceName:         instance.Title,
		ReportTargetUsername: report.TargetAccount.Username,
		ReportTargetDomain:   report.TargetAccount.Domain,
		ActionTakenComment:   report.ActionTaken,
	}

	return s.EmailSender.SendReportClosedEmail(user.Email, reportClosedData)
}

// emailUserPleaseConfirm emails the given user
// to ask them to confirm their email address.
//
// If newSignup is true, template will be geared
// towards someone who just created an account.
func (s *Surface) emailUserPleaseConfirm(ctx context.Context, user *gtsmodel.User, newSignup bool) error {
	if user.UnconfirmedEmail == "" ||
		user.UnconfirmedEmail == user.Email {
		// User has already confirmed this
		// email address; nothing to do.
		return nil
	}

	instance, err := s.State.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return gtserror.Newf("db error getting instance: %w", err)
	}

	// We need a token and a link for the
	// user to click on. We'll use a uuid
	// as our token since it's secure enough
	// for this purpose.
	var (
		confirmToken = uuid.NewString()
		confirmLink  = uris.GenerateURIForEmailConfirm(confirmToken)
	)

	// Assemble email contents and send the email.
	if err := s.EmailSender.SendConfirmEmail(
		user.UnconfirmedEmail,
		email.ConfirmData{
			Username:     user.Account.Username,
			InstanceURL:  instance.URI,
			InstanceName: instance.Title,
			ConfirmLink:  confirmLink,
			NewSignup:    newSignup,
		},
	); err != nil {
		return err
	}

	// Email sent, update the user entry
	// with the new confirmation token.
	now := time.Now()
	user.ConfirmationToken = confirmToken
	user.ConfirmationSentAt = now
	user.LastEmailedAt = now

	if err := s.State.DB.UpdateUser(
		ctx,
		user,
		"confirmation_token",
		"confirmation_sent_at",
		"last_emailed_at",
	); err != nil {
		return gtserror.Newf("error updating user entry after email sent: %w", err)
	}

	return nil
}

// emailUserSignupApproved emails the given user
// to inform them their sign-up has been approved.
func (s *Surface) emailUserSignupApproved(ctx context.Context, user *gtsmodel.User) error {
	// User may have been approved without
	// their email address being confirmed
	// yet. Just send to whatever we have.
	emailAddr := user.Email
	if emailAddr == "" {
		emailAddr = user.UnconfirmedEmail
	}

	instance, err := s.State.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return gtserror.Newf("db error getting instance: %w", err)
	}

	// Assemble email contents and send the email.
	if err := s.EmailSender.SendSignupApprovedEmail(
		emailAddr,
		email.SignupApprovedData{
			Username:     user.Account.Username,
			InstanceURL:  instance.URI,
			InstanceName: instance.Title,
		},
	); err != nil {
		return err
	}

	// Email sent, update the user
	// entry with the emailed time.
	now := time.Now()
	user.LastEmailedAt = now

	if err := s.State.DB.UpdateUser(
		ctx,
		user,
		"last_emailed_at",
	); err != nil {
		return gtserror.Newf("error updating user entry after email sent: %w", err)
	}

	return nil
}

// emailUserSignupApproved emails the given user
// to inform them their sign-up has been approved.
func (s *Surface) emailUserSignupRejected(ctx context.Context, deniedUser *gtsmodel.DeniedUser) error {
	instance, err := s.State.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return gtserror.Newf("db error getting instance: %w", err)
	}

	// Assemble email contents and send the email.
	return s.EmailSender.SendSignupRejectedEmail(
		deniedUser.Email,
		email.SignupRejectedData{
			Message:      deniedUser.Message,
			InstanceURL:  instance.URI,
			InstanceName: instance.Title,
		},
	)
}

// emailAdminReportOpened emails all active moderators/admins
// of this instance that a new report has been created.
func (s *Surface) emailAdminReportOpened(ctx context.Context, report *gtsmodel.Report) error {
	instance, err := s.State.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return gtserror.Newf("error getting instance: %w", err)
	}

	toAddresses, err := s.State.DB.GetInstanceModeratorAddresses(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No registered moderator addresses.
			return nil
		}
		return gtserror.Newf("error getting instance moderator addresses: %w", err)
	}

	if err := s.State.DB.PopulateReport(ctx, report); err != nil {
		return gtserror.Newf("error populating report: %w", err)
	}

	reportData := email.NewReportData{
		InstanceURL:        instance.URI,
		InstanceName:       instance.Title,
		ReportURL:          instance.URI + "/settings/moderation/reports/" + report.ID,
		ReportDomain:       report.Account.Domain,
		ReportTargetDomain: report.TargetAccount.Domain,
	}

	if err := s.EmailSender.SendNewReportEmail(toAddresses, reportData); err != nil {
		return gtserror.Newf("error emailing instance moderators: %w", err)
	}

	return nil
}

// emailAdminNewSignup emails all active moderators/admins of this
// instance that a new account sign-up has been submitted to the instance.
func (s *Surface) emailAdminNewSignup(ctx context.Context, newUser *gtsmodel.User) error {
	instance, err := s.State.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return gtserror.Newf("error getting instance: %w", err)
	}

	toAddresses, err := s.State.DB.GetInstanceModeratorAddresses(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No registered moderator addresses.
			return nil
		}
		return gtserror.Newf("error getting instance moderator addresses: %w", err)
	}

	// Ensure user populated.
	if err := s.State.DB.PopulateUser(ctx, newUser); err != nil {
		return gtserror.Newf("error populating user: %w", err)
	}

	newSignupData := email.NewSignupData{
		InstanceURL:    instance.URI,
		InstanceName:   instance.Title,
		SignupEmail:    newUser.UnconfirmedEmail,
		SignupUsername: newUser.Account.Username,
		SignupReason:   newUser.Reason,
		SignupURL:      instance.URI + "/settings/moderation/accounts/" + newUser.AccountID,
	}

	if err := s.EmailSender.SendNewSignupEmail(toAddresses, newSignupData); err != nil {
		return gtserror.Newf("error emailing instance moderators: %w", err)
	}

	return nil
}
