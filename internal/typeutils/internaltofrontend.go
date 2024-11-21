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

package typeutils

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"codeberg.org/gruf/go-debug"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/language"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

const (
	instanceStatusesCharactersReservedPerURL = 25
	instancePollsMinExpiration               = 300     // seconds
	instancePollsMaxExpiration               = 2629746 // seconds
	instanceAccountsMaxFeaturedTags          = 10
	instanceAccountsMaxProfileFields         = 6 // FIXME: https://github.com/superseriousbusiness/gotosocial/issues/1876
	instanceSourceURL                        = "https://github.com/superseriousbusiness/gotosocial"
	instanceMastodonVersion                  = "3.5.3"
)

var instanceStatusesSupportedMimeTypes = []string{
	string(apimodel.StatusContentTypePlain),
	string(apimodel.StatusContentTypeMarkdown),
}

func toMastodonVersion(in string) string {
	return instanceMastodonVersion + "+" + strings.ReplaceAll(in, " ", "-")
}

// UserToAPIUser converts a *gtsmodel.User to an API
// representation suitable for serving to that user.
//
// Contains sensitive info so should only
// ever be served to the user themself.
func (c *Converter) UserToAPIUser(ctx context.Context, u *gtsmodel.User) *apimodel.User {
	user := &apimodel.User{
		ID:               u.ID,
		CreatedAt:        util.FormatISO8601(u.CreatedAt),
		Email:            u.Email,
		UnconfirmedEmail: u.UnconfirmedEmail,
		Reason:           u.Reason,
		Moderator:        *u.Moderator,
		Admin:            *u.Admin,
		Disabled:         *u.Disabled,
		Approved:         *u.Approved,
	}

	// Zero-able dates.
	if !u.LastEmailedAt.IsZero() {
		user.LastEmailedAt = util.FormatISO8601(u.LastEmailedAt)
	}

	if !u.ConfirmedAt.IsZero() {
		user.ConfirmedAt = util.FormatISO8601(u.ConfirmedAt)
	}

	if !u.ConfirmationSentAt.IsZero() {
		user.ConfirmationSentAt = util.FormatISO8601(u.ConfirmationSentAt)
	}

	if !u.ResetPasswordSentAt.IsZero() {
		user.ResetPasswordSentAt = util.FormatISO8601(u.ResetPasswordSentAt)
	}

	return user
}

// AccountToAPIAccountSensitive takes a db model application as a param, and returns a populated apitype application, or an error
// if something goes wrong. The returned application should be ready to serialize on an API level, and may have sensitive fields
// (such as client id and client secret), so serve it only to an authorized user who should have permission to see it.
func (c *Converter) AccountToAPIAccountSensitive(ctx context.Context, a *gtsmodel.Account) (*apimodel.Account, error) {
	// We can build this sensitive account model
	// by first getting the public account, and
	// then adding the Source object and role permissions bitmap to it.
	apiAccount, err := c.AccountToAPIAccountPublic(ctx, a)
	if err != nil {
		return nil, err
	}

	// Ensure account stats populated.
	if err := c.state.DB.PopulateAccountStats(ctx, a); err != nil {
		return nil, gtserror.Newf(
			"error getting stats for account %s: %w",
			a.ID, err,
		)
	}

	// Populate the account's role permissions bitmap and highlightedness from its public role.
	if len(apiAccount.Roles) > 0 {
		apiAccount.Role = c.APIAccountDisplayRoleToAPIAccountRoleSensitive(&apiAccount.Roles[0])
	} else {
		apiAccount.Role = c.APIAccountDisplayRoleToAPIAccountRoleSensitive(nil)
	}

	statusContentType := string(apimodel.StatusContentTypeDefault)
	if a.Settings.StatusContentType != "" {
		statusContentType = a.Settings.StatusContentType
	}

	apiAccount.Source = &apimodel.Source{
		Privacy:             c.VisToAPIVis(ctx, a.Settings.Privacy),
		WebVisibility:       c.VisToAPIVis(ctx, a.Settings.WebVisibility),
		Sensitive:           *a.Settings.Sensitive,
		Language:            a.Settings.Language,
		StatusContentType:   statusContentType,
		Note:                a.NoteRaw,
		Fields:              c.fieldsToAPIFields(a.FieldsRaw),
		FollowRequestsCount: *a.Stats.FollowRequestsCount,
		AlsoKnownAsURIs:     a.AlsoKnownAsURIs,
	}

	return apiAccount, nil
}

// AccountToAPIAccountPublic takes a db model account as a param, and returns a populated apitype account, or an error
// if something goes wrong. The returned account should be ready to serialize on an API level, and may NOT have sensitive fields.
// In other words, this is the public record that the server has of an account.
func (c *Converter) AccountToAPIAccountPublic(ctx context.Context, a *gtsmodel.Account) (*apimodel.Account, error) {
	account, err := c.accountToAPIAccountPublic(ctx, a)
	if err != nil {
		return nil, err
	}

	if a.MovedTo != nil {
		account.Moved, err = c.accountToAPIAccountPublic(ctx, a.MovedTo)
		if err != nil {
			log.Errorf(ctx, "error converting account movedTo: %v", err)
		}
	}

	return account, nil
}

// AccountToWebAccount converts a gts model account into an
// api representation suitable for serving into a web template.
//
// Should only be used when preparing to template an account,
// callers looking to serialize an account into a model for
// serving over the client API should always use one of the
// AccountToAPIAccount functions instead.
func (c *Converter) AccountToWebAccount(
	ctx context.Context,
	a *gtsmodel.Account,
) (*apimodel.WebAccount, error) {
	apiAccount, err := c.AccountToAPIAccountPublic(ctx, a)
	if err != nil {
		return nil, err
	}

	webAccount := &apimodel.WebAccount{
		Account: apiAccount,
	}

	// Set additional avatar information for
	// serving the avatar in a nice <picture>.
	if ogAvi := a.AvatarMediaAttachment; ogAvi != nil {
		avatarAttachment, err := c.AttachmentToAPIAttachment(ctx, ogAvi)
		if err != nil {
			// This is just extra data so just
			// log but don't return any error.
			log.Errorf(ctx, "error converting account avatar attachment: %v", err)
		} else {
			webAccount.AvatarAttachment = &apimodel.WebAttachment{
				Attachment:      &avatarAttachment,
				MIMEType:        ogAvi.File.ContentType,
				PreviewMIMEType: ogAvi.Thumbnail.ContentType,
			}
		}
	}

	// Set additional header information for
	// serving the header in a nice <picture>.
	if ogHeader := a.HeaderMediaAttachment; ogHeader != nil {
		headerAttachment, err := c.AttachmentToAPIAttachment(ctx, ogHeader)
		if err != nil {
			// This is just extra data so just
			// log but don't return any error.
			log.Errorf(ctx, "error converting account header attachment: %v", err)
		} else {
			webAccount.HeaderAttachment = &apimodel.WebAttachment{
				Attachment:      &headerAttachment,
				MIMEType:        ogHeader.File.ContentType,
				PreviewMIMEType: ogHeader.Thumbnail.ContentType,
			}
		}
	}

	return webAccount, nil
}

// accountToAPIAccountPublic provides all the logic for AccountToAPIAccount, MINUS fetching moved account, to prevent possible recursion.
func (c *Converter) accountToAPIAccountPublic(ctx context.Context, a *gtsmodel.Account) (*apimodel.Account, error) {

	// Populate account struct fields.
	err := c.state.DB.PopulateAccount(ctx, a)

	switch {
	case err == nil:
		// No problem.

	case a.Stats != nil:
		// We have stats so that's
		// *maybe* OK, try to continue.
		log.Errorf(ctx, "error(s) populating account, will continue: %s", err)

	default:
		// There was an error and we don't
		// have stats, we can't continue.
		return nil, gtserror.Newf("account stats not populated, could not continue: %w", err)
	}

	// Basic account stats:
	//   - Followers count
	//   - Following count
	//   - Statuses count
	//   - Last status time

	var (
		followersCount = *a.Stats.FollowersCount
		followingCount = *a.Stats.FollowingCount
		statusesCount  = *a.Stats.StatusesCount
		lastStatusAt   = func() *string {
			if a.Stats.LastStatusAt.IsZero() {
				return nil
			}
			return util.Ptr(util.FormatISO8601Date(a.Stats.LastStatusAt))
		}()
	)

	// Profile media + nice extras:
	//   - Avatar
	//   - Header
	//   - Fields
	//   - Emojis

	var (
		aviID           string
		aviURL          string
		aviURLStatic    string
		aviDesc         string
		headerID        string
		headerURL       string
		headerURLStatic string
		headerDesc      string
	)

	if a.AvatarMediaAttachment != nil {
		aviID = a.AvatarMediaAttachmentID
		aviURL = a.AvatarMediaAttachment.URL
		aviURLStatic = a.AvatarMediaAttachment.Thumbnail.URL
		aviDesc = a.AvatarMediaAttachment.Description
	}

	if a.HeaderMediaAttachment != nil {
		headerID = a.HeaderMediaAttachmentID
		headerURL = a.HeaderMediaAttachment.URL
		headerURLStatic = a.HeaderMediaAttachment.Thumbnail.URL
		headerDesc = a.HeaderMediaAttachment.Description
	}

	// convert account gts model fields to front api model fields
	fields := c.fieldsToAPIFields(a.Fields)

	// GTS model emojis -> frontend.
	apiEmojis, err := c.convertEmojisToAPIEmojis(ctx, a.Emojis, a.EmojiIDs)
	if err != nil {
		log.Errorf(ctx, "error converting account emojis: %v", err)
	}

	// Bits that vary between remote + local accounts:
	//   - Account (acct) string.
	//   - Role.
	//   - Settings things (enableRSS, theme, customCSS, hideCollections).

	var (
		acct            string
		roles           []apimodel.AccountDisplayRole
		enableRSS       bool
		theme           string
		customCSS       string
		hideCollections bool
	)

	if a.IsRemote() {
		// Domain may be in Punycode,
		// de-punify it just in case.
		d, err := util.DePunify(a.Domain)
		if err != nil {
			return nil, gtserror.Newf("error de-punifying domain %s for account id %s: %w", a.Domain, a.ID, err)
		}

		acct = a.Username + "@" + d
	} else {
		// This is a local account, try to
		// fetch more info. Skip for instance
		// accounts since they have no user.
		if !a.IsInstance() {
			user, err := c.state.DB.GetUserByAccountID(ctx, a.ID)
			if err != nil {
				return nil, gtserror.Newf("error getting user from database for account id %s: %w", a.ID, err)
			}
			if role := c.UserToAPIAccountDisplayRole(user); role != nil {
				roles = append(roles, *role)
			}

			enableRSS = *a.Settings.EnableRSS
			theme = a.Settings.Theme
			customCSS = a.Settings.CustomCSS
			hideCollections = *a.Settings.HideCollections
		}

		acct = a.Username // omit domain
	}

	var (
		locked       = util.PtrOrValue(a.Locked, true)
		discoverable = util.PtrOrValue(a.Discoverable, false)
		bot          = util.PtrOrValue(a.Bot, false)
	)

	// Remaining properties are simple and
	// can be populated directly below.

	accountFrontend := &apimodel.Account{
		ID:                a.ID,
		Username:          a.Username,
		Acct:              acct,
		DisplayName:       a.DisplayName,
		Locked:            locked,
		Discoverable:      discoverable,
		Bot:               bot,
		CreatedAt:         util.FormatISO8601(a.CreatedAt),
		Note:              a.Note,
		URL:               a.URL,
		Avatar:            aviURL,
		AvatarStatic:      aviURLStatic,
		AvatarDescription: aviDesc,
		AvatarMediaID:     aviID,
		Header:            headerURL,
		HeaderStatic:      headerURLStatic,
		HeaderDescription: headerDesc,
		HeaderMediaID:     headerID,
		FollowersCount:    followersCount,
		FollowingCount:    followingCount,
		StatusesCount:     statusesCount,
		LastStatusAt:      lastStatusAt,
		Emojis:            apiEmojis,
		Fields:            fields,
		Suspended:         !a.SuspendedAt.IsZero(),
		Theme:             theme,
		CustomCSS:         customCSS,
		EnableRSS:         enableRSS,
		HideCollections:   hideCollections,
		Roles:             roles,
	}

	// Bodge default avatar + header in,
	// if we didn't have one already.
	c.ensureAvatar(accountFrontend)
	c.ensureHeader(accountFrontend)

	return accountFrontend, nil
}

// UserToAPIAccountDisplayRole returns the API representation of a user's display role.
// This will accept a nil user but does not always return a value:
// the default "user" role is considered uninteresting and not returned.
func (c *Converter) UserToAPIAccountDisplayRole(user *gtsmodel.User) *apimodel.AccountDisplayRole {
	switch {
	case user == nil:
		return nil
	case *user.Admin:
		return &apimodel.AccountDisplayRole{
			ID:   string(apimodel.AccountRoleAdmin),
			Name: apimodel.AccountRoleAdmin,
		}
	case *user.Moderator:
		return &apimodel.AccountDisplayRole{
			ID:   string(apimodel.AccountRoleModerator),
			Name: apimodel.AccountRoleModerator,
		}
	default:
		return nil
	}
}

// APIAccountDisplayRoleToAPIAccountRoleSensitive returns the API representation of a user's role,
// with permission bitmap. This will accept a nil display role and always returns a value.
func (c *Converter) APIAccountDisplayRoleToAPIAccountRoleSensitive(display *apimodel.AccountDisplayRole) *apimodel.AccountRole {
	// Default to user role.
	role := &apimodel.AccountRole{
		AccountDisplayRole: apimodel.AccountDisplayRole{
			ID:   string(apimodel.AccountRoleUser),
			Name: apimodel.AccountRoleUser,
		},
		Permissions: apimodel.AccountRolePermissionsNone,
		Highlighted: false,
	}

	// If there's a display role, use that instead.
	if display != nil {
		role.AccountDisplayRole = *display
		role.Highlighted = true
		switch display.Name {
		case apimodel.AccountRoleAdmin:
			role.Permissions = apimodel.AccountRolePermissionsForAdminRole
		case apimodel.AccountRoleModerator:
			role.Permissions = apimodel.AccountRolePermissionsForModeratorRole
		}
	}

	return role
}

func (c *Converter) fieldsToAPIFields(f []*gtsmodel.Field) []apimodel.Field {
	fields := make([]apimodel.Field, len(f))

	for i, field := range f {
		mField := apimodel.Field{
			Name:  field.Name,
			Value: field.Value,
		}

		if !field.VerifiedAt.IsZero() {
			verified := util.FormatISO8601(field.VerifiedAt)
			mField.VerifiedAt = util.Ptr(verified)
		}

		fields[i] = mField
	}

	return fields
}

// AccountToAPIAccountBlocked takes a db model account as a param, and returns a apitype account, or an error if
// something goes wrong. The returned account will be a bare minimum representation of the account. This function should be used
// when someone wants to view an account they've blocked.
func (c *Converter) AccountToAPIAccountBlocked(ctx context.Context, a *gtsmodel.Account) (*apimodel.Account, error) {
	var (
		acct  string
		roles []apimodel.AccountDisplayRole
	)

	if a.IsRemote() {
		// Domain may be in Punycode,
		// de-punify it just in case.
		d, err := util.DePunify(a.Domain)
		if err != nil {
			return nil, gtserror.Newf("error de-punifying domain %s for account id %s: %w", a.Domain, a.ID, err)
		}

		acct = a.Username + "@" + d
	} else {
		// This is a local account, try to
		// fetch more info. Skip for instance
		// accounts since they have no user.
		if !a.IsInstance() {
			user, err := c.state.DB.GetUserByAccountID(ctx, a.ID)
			if err != nil {
				return nil, gtserror.Newf("error getting user from database for account id %s: %w", a.ID, err)
			}
			if role := c.UserToAPIAccountDisplayRole(user); role != nil {
				roles = append(roles, *role)
			}
		}

		acct = a.Username // omit domain
	}

	account := &apimodel.Account{
		ID:        a.ID,
		Username:  a.Username,
		Acct:      acct,
		Bot:       *a.Bot,
		CreatedAt: util.FormatISO8601(a.CreatedAt),
		URL:       a.URL,
		// Empty array (not nillable).
		Emojis: make([]apimodel.Emoji, 0),
		// Empty array (not nillable).
		Fields:    make([]apimodel.Field, 0),
		Suspended: !a.SuspendedAt.IsZero(),
		Roles:     roles,
	}

	// Don't show the account's actual
	// avatar+header since it may be
	// upsetting to the blocker. Just
	// show generic avatar+header instead.
	c.ensureAvatar(account)
	c.ensureHeader(account)

	return account, nil
}

func (c *Converter) AccountToAdminAPIAccount(ctx context.Context, a *gtsmodel.Account) (*apimodel.AdminAccountInfo, error) {
	var (
		email                  string
		ip                     *string
		domain                 *string
		locale                 string
		confirmed              bool
		inviteRequest          *string
		approved               bool
		disabled               bool
		role                   = *c.APIAccountDisplayRoleToAPIAccountRoleSensitive(nil)
		createdByApplicationID string
	)

	if err := c.state.DB.PopulateAccount(ctx, a); err != nil {
		log.Errorf(ctx, "error(s) populating account, will continue: %s", err)
	}

	if a.IsRemote() {
		// Domain may be in Punycode,
		// de-punify it just in case.
		d, err := util.DePunify(a.Domain)
		if err != nil {
			return nil, fmt.Errorf("AccountToAdminAPIAccount: error de-punifying domain %s for account id %s: %w", a.Domain, a.ID, err)
		}

		domain = &d
	} else if !a.IsInstance() {
		// This is a local, non-instance
		// acct; we can fetch more info.
		user, err := c.state.DB.GetUserByAccountID(ctx, a.ID)
		if err != nil {
			return nil, fmt.Errorf("AccountToAdminAPIAccount: error getting user from database for account id %s: %w", a.ID, err)
		}

		if user.Email != "" {
			email = user.Email
		} else {
			email = user.UnconfirmedEmail
		}

		if i := user.SignUpIP.String(); i != "<nil>" {
			ip = &i
		}

		locale = user.Locale
		if user.Reason != "" {
			inviteRequest = &user.Reason
		}

		role = *c.APIAccountDisplayRoleToAPIAccountRoleSensitive(
			c.UserToAPIAccountDisplayRole(user),
		)

		confirmed = !user.ConfirmedAt.IsZero()
		approved = *user.Approved
		disabled = *user.Disabled
		createdByApplicationID = user.CreatedByApplicationID
	}

	apiAccount, err := c.AccountToAPIAccountPublic(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("AccountToAdminAPIAccount: error converting account to api account for account id %s: %w", a.ID, err)
	}

	return &apimodel.AdminAccountInfo{
		ID:                     a.ID,
		Username:               a.Username,
		Domain:                 domain,
		CreatedAt:              util.FormatISO8601(a.CreatedAt),
		Email:                  email,
		IP:                     ip,
		IPs:                    []interface{}{}, // not implemented,
		Locale:                 locale,
		InviteRequest:          inviteRequest,
		Role:                   role,
		Confirmed:              confirmed,
		Approved:               approved,
		Disabled:               disabled,
		Silenced:               !a.SilencedAt.IsZero(),
		Suspended:              !a.SuspendedAt.IsZero(),
		Account:                apiAccount,
		CreatedByApplicationID: createdByApplicationID,
		InvitedByAccountID:     "", // not implemented (yet)
	}, nil
}

func (c *Converter) AppToAPIAppSensitive(ctx context.Context, a *gtsmodel.Application) (*apimodel.Application, error) {
	return &apimodel.Application{
		ID:           a.ID,
		Name:         a.Name,
		Website:      a.Website,
		RedirectURI:  a.RedirectURI,
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
	}, nil
}

// AppToAPIAppPublic takes a db model application as a param, and returns a populated apitype application, or an error
// if something goes wrong. The returned application should be ready to serialize on an API level, and has sensitive
// fields sanitized so that it can be served to non-authorized accounts without revealing any private information.
func (c *Converter) AppToAPIAppPublic(ctx context.Context, a *gtsmodel.Application) (*apimodel.Application, error) {
	return &apimodel.Application{
		Name:    a.Name,
		Website: a.Website,
	}, nil
}

// AttachmentToAPIAttachment converts a gts model media attacahment into its api representation for serialization on the API.
func (c *Converter) AttachmentToAPIAttachment(ctx context.Context, media *gtsmodel.MediaAttachment) (apimodel.Attachment, error) {
	var api apimodel.Attachment
	api.Type = media.Type.String()
	api.ID = media.ID

	// Only add file details if
	// we have stored locally.
	if media.File.Path != "" {
		api.Meta = new(apimodel.MediaMeta)
		api.Meta.Original = apimodel.MediaDimensions{
			Width:     media.FileMeta.Original.Width,
			Height:    media.FileMeta.Original.Height,
			Aspect:    media.FileMeta.Original.Aspect,
			Size:      toAPISize(media.FileMeta.Original.Width, media.FileMeta.Original.Height),
			FrameRate: toAPIFrameRate(media.FileMeta.Original.Framerate),
			Duration:  util.PtrOrZero(media.FileMeta.Original.Duration),
			Bitrate:   util.PtrOrZero(media.FileMeta.Original.Bitrate),
		}

		// Copy over local file URL.
		api.URL = util.Ptr(media.URL)
		api.TextURL = util.Ptr(media.URL)

		// Set file focus details.
		// (this doesn't make much sense if media
		// has no image, but the API doesn't yet
		// distinguish between zero values vs. none).
		api.Meta.Focus = new(apimodel.MediaFocus)
		api.Meta.Focus.X = media.FileMeta.Focus.X
		api.Meta.Focus.Y = media.FileMeta.Focus.Y

		// Only add thumbnail details if
		// we have thumbnail stored locally.
		if media.Thumbnail.Path != "" {
			api.Meta.Small = apimodel.MediaDimensions{
				Width:  media.FileMeta.Small.Width,
				Height: media.FileMeta.Small.Height,
				Aspect: media.FileMeta.Small.Aspect,
				Size:   toAPISize(media.FileMeta.Small.Width, media.FileMeta.Small.Height),
			}

			// Copy over local thumbnail file URL.
			api.PreviewURL = util.Ptr(media.Thumbnail.URL)
		}
	}

	// Set remaining API attachment fields.
	api.Blurhash = util.PtrIf(media.Blurhash)
	api.RemoteURL = util.PtrIf(media.RemoteURL)
	api.PreviewRemoteURL = util.PtrIf(media.Thumbnail.RemoteURL)
	api.Description = util.PtrIf(media.Description)

	return api, nil
}

// MentionToAPIMention converts a gts model mention into its api (frontend) representation for serialization on the API.
func (c *Converter) MentionToAPIMention(ctx context.Context, m *gtsmodel.Mention) (apimodel.Mention, error) {
	if m.TargetAccount == nil {
		targetAccount, err := c.state.DB.GetAccountByID(ctx, m.TargetAccountID)
		if err != nil {
			return apimodel.Mention{}, err
		}
		m.TargetAccount = targetAccount
	}

	var acct string
	if m.TargetAccount.IsLocal() {
		acct = m.TargetAccount.Username
	} else {
		// Domain may be in Punycode,
		// de-punify it just in case.
		d, err := util.DePunify(m.TargetAccount.Domain)
		if err != nil {
			err = fmt.Errorf("MentionToAPIMention: error de-punifying domain %s for account id %s: %w", m.TargetAccount.Domain, m.TargetAccountID, err)
			return apimodel.Mention{}, err
		}

		acct = m.TargetAccount.Username + "@" + d
	}

	return apimodel.Mention{
		ID:       m.TargetAccount.ID,
		Username: m.TargetAccount.Username,
		URL:      m.TargetAccount.URL,
		Acct:     acct,
	}, nil
}

// EmojiToAPIEmoji converts a gts model emoji into its api (frontend) representation for serialization on the API.
func (c *Converter) EmojiToAPIEmoji(ctx context.Context, e *gtsmodel.Emoji) (apimodel.Emoji, error) {
	var category string

	if e.CategoryID != "" {
		if e.Category == nil {
			var err error
			e.Category, err = c.state.DB.GetEmojiCategory(ctx, e.CategoryID)
			if err != nil {
				return apimodel.Emoji{}, err
			}
		}
		category = e.Category.Name
	}

	return apimodel.Emoji{
		Shortcode:       e.Shortcode,
		URL:             e.ImageURL,
		StaticURL:       e.ImageStaticURL,
		VisibleInPicker: *e.VisibleInPicker,
		Category:        category,
	}, nil
}

// EmojiToAdminAPIEmoji converts a gts model emoji into an API representation with extra admin information.
func (c *Converter) EmojiToAdminAPIEmoji(ctx context.Context, e *gtsmodel.Emoji) (*apimodel.AdminEmoji, error) {
	emoji, err := c.EmojiToAPIEmoji(ctx, e)
	if err != nil {
		return nil, err
	}

	if !e.IsLocal() {
		// Domain may be in Punycode,
		// de-punify it just in case.
		var err error
		e.Domain, err = util.DePunify(e.Domain)
		if err != nil {
			err = fmt.Errorf("EmojiToAdminAPIEmoji: error de-punifying domain %s for emoji id %s: %w", e.Domain, e.ID, err)
			return nil, err
		}
	}

	return &apimodel.AdminEmoji{
		Emoji:         emoji,
		ID:            e.ID,
		Disabled:      *e.Disabled,
		Domain:        e.Domain,
		UpdatedAt:     util.FormatISO8601(e.UpdatedAt),
		TotalFileSize: e.ImageFileSize + e.ImageStaticFileSize,
		ContentType:   e.ImageContentType,
		URI:           e.URI,
	}, nil
}

// EmojiCategoryToAPIEmojiCategory converts a gts model emoji category into its api (frontend) representation.
func (c *Converter) EmojiCategoryToAPIEmojiCategory(ctx context.Context, category *gtsmodel.EmojiCategory) (*apimodel.EmojiCategory, error) {
	return &apimodel.EmojiCategory{
		ID:   category.ID,
		Name: category.Name,
	}, nil
}

// TagToAPITag converts a gts model tag into its api (frontend) representation for serialization on the API.
// If stubHistory is set to 'true', then the 'history' field of the tag will be populated with a pointer to an empty slice, for API compatibility reasons.
// following is an optional flag marking whether the currently authenticated user (if there is one) is following the tag.
func (c *Converter) TagToAPITag(ctx context.Context, t *gtsmodel.Tag, stubHistory bool, following *bool) (apimodel.Tag, error) {
	return apimodel.Tag{
		Name: strings.ToLower(t.Name),
		URL:  uris.URIForTag(t.Name),
		History: func() *[]any {
			if !stubHistory {
				return nil
			}

			h := make([]any, 0)
			return &h
		}(),
		Following: following,
	}, nil
}

// StatusToAPIStatus converts a gts model
// status into its api (frontend) representation
// for serialization on the API.
//
// Requesting account can be nil.
//
// filterContext can be the empty string
// if these statuses are not being filtered.
//
// If there is a matching "hide" filter, the returned
// status will be nil with a ErrHideStatus error; callers
// need to handle that case by excluding it from results.
func (c *Converter) StatusToAPIStatus(
	ctx context.Context,
	status *gtsmodel.Status,
	requestingAccount *gtsmodel.Account,
	filterContext statusfilter.FilterContext,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (*apimodel.Status, error) {
	return c.statusToAPIStatus(
		ctx,
		status,
		requestingAccount,
		filterContext,
		filters,
		mutes,
		true,
		true,
	)
}

// statusToAPIStatus is the package-internal implementation
// of StatusToAPIStatus that lets the caller customize whether
// to placehold unknown attachment types, and/or add a note
// about the status being pending and requiring approval.
func (c *Converter) statusToAPIStatus(
	ctx context.Context,
	status *gtsmodel.Status,
	requestingAccount *gtsmodel.Account,
	filterContext statusfilter.FilterContext,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
	placeholdAttachments bool,
	addPendingNote bool,
) (*apimodel.Status, error) {
	apiStatus, err := c.statusToFrontend(
		ctx,
		status,
		requestingAccount, // Can be nil.
		filterContext,     // Can be empty.
		filters,
		mutes,
	)
	if err != nil {
		return nil, err
	}

	// Convert author to API model.
	acct, err := c.AccountToAPIAccountPublic(ctx, status.Account)
	if err != nil {
		return nil, gtserror.Newf("error converting status acct: %w", err)
	}
	apiStatus.Account = acct

	// Convert author of boosted
	// status (if set) to API model.
	if apiStatus.Reblog != nil {
		boostAcct, err := c.AccountToAPIAccountPublic(ctx, status.BoostOfAccount)
		if err != nil {
			return nil, gtserror.Newf("error converting boost acct: %w", err)
		}
		apiStatus.Reblog.Account = boostAcct
	}

	if placeholdAttachments {
		// Normalize status for API by pruning attachments
		// that were not able to be locally stored, and replacing
		// them with a helpful message + links to remote.
		var attachNote string
		attachNote, apiStatus.MediaAttachments = placeholderAttachments(apiStatus.MediaAttachments)
		apiStatus.Content += attachNote

		// Do the same for the reblogged status.
		if apiStatus.Reblog != nil {
			attachNote, apiStatus.Reblog.MediaAttachments = placeholderAttachments(apiStatus.Reblog.MediaAttachments)
			apiStatus.Reblog.Content += attachNote
		}
	}

	if addPendingNote {
		// If this status is pending approval and
		// replies to the requester, add a note
		// about how to approve or reject the reply.
		pendingApproval := util.PtrOrValue(status.PendingApproval, false)
		if pendingApproval &&
			requestingAccount != nil &&
			requestingAccount.ID == status.InReplyToAccountID {
			pendingNote, err := c.pendingReplyNote(ctx, status)
			if err != nil {
				return nil, gtserror.Newf("error deriving 'pending reply' note: %w", err)
			}

			apiStatus.Content += pendingNote
		}
	}

	return apiStatus, nil
}

// statusToAPIFilterResults applies filters and mutes to a status and returns an API filter result object.
// The result may be nil if no filters matched.
// If the status should not be returned at all, it returns the ErrHideStatus error.
func (c *Converter) statusToAPIFilterResults(
	ctx context.Context,
	s *gtsmodel.Status,
	requestingAccount *gtsmodel.Account,
	filterContext statusfilter.FilterContext,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) ([]apimodel.FilterResult, error) {
	// If there are no filters or mutes, we're done.
	// We never hide statuses authored by the requesting account,
	// since not being able to see your own posts is confusing.
	if filterContext == "" || (len(filters) == 0 && mutes.Len() == 0) || s.AccountID == requestingAccount.ID {
		return nil, nil
	}

	// Both mutes and filters can expire.
	now := time.Now()

	// If the requesting account mutes the account that created this status, hide the status.
	if mutes.Matches(s.AccountID, filterContext, now) {
		return nil, statusfilter.ErrHideStatus
	}

	// If this status is part of a multi-account discussion,
	// and all of the accounts replied to or mentioned are invisible to the requesting account
	// (due to blocks, domain blocks, moderation, etc.),
	// or are muted, hide the status.
	// First, collect the accounts we have to check.
	otherAccounts := make([]*gtsmodel.Account, 0, 1+len(s.Mentions))
	if s.InReplyToAccount != nil {
		otherAccounts = append(otherAccounts, s.InReplyToAccount)
	}
	for _, mention := range s.Mentions {
		otherAccounts = append(otherAccounts, mention.TargetAccount)
	}

	// If there are no other accounts, skip this check.
	if len(otherAccounts) > 0 {
		// Start by assuming that they're all invisible or muted.
		allOtherAccountsInvisibleOrMuted := true

		for _, account := range otherAccounts {
			// Is this account visible?
			visible, err := c.visFilter.AccountVisible(ctx, requestingAccount, account)
			if err != nil {
				return nil, err
			}
			if !visible {
				// It's invisible. Check the next account.
				continue
			}

			// If visible, is it muted?
			if mutes.Matches(account.ID, filterContext, now) {
				// It's muted. Check the next account.
				continue
			}

			// If we get here, the account is visible and not muted.
			// We should show this status, and don't have to check any more accounts.
			allOtherAccountsInvisibleOrMuted = false
			break
		}

		// If we didn't find any visible non-muted accounts, hide the status.
		if allOtherAccountsInvisibleOrMuted {
			return nil, statusfilter.ErrHideStatus
		}
	}

	// At this point, the status isn't muted, but might still be filtered.
	if len(filters) == 0 {
		// If it can't be filtered because there are no filters, we're done.
		return nil, nil
	}

	// Key this status based on ID + last updated time,
	// to ensure we always filter on latest version.
	statusKey := s.ID + strconv.FormatInt(s.UpdatedAt.Unix(), 10)

	// Check if we have filterable fields cached for this status.
	cache := c.state.Caches.StatusesFilterableFields
	fields, stored := cache.Get(statusKey)
	if !stored {
		// We don't have filterable fields
		// cached, calculate + cache now.
		fields = filterableFields(s)
		cache.Set(statusKey, fields)
	}

	// Record all matching warn filters and the reasons they matched.
	filterResults := make([]apimodel.FilterResult, 0, len(filters))
	for _, filter := range filters {
		if !filterAppliesInContext(filter, filterContext) {
			// Filter doesn't apply
			// to this context.
			continue
		}

		if filter.Expired(now) {
			// Filter doesn't
			// apply anymore.
			continue
		}

		// Assemble matching keywords (if any) from this filter.
		keywordMatches := make([]string, 0, len(filter.Keywords))
		for _, keyword := range filter.Keywords {
			// Check if at least one filterable field
			// in the status matches on this filter.
			if slices.ContainsFunc(
				fields,
				func(field string) bool {
					return keyword.Regexp.MatchString(field)
				},
			) {
				// At least one field matched on this filter.
				keywordMatches = append(keywordMatches, keyword.Keyword)
			}
		}

		// A status has only one ID. Not clear why this is a list in the Mastodon API.
		statusMatches := make([]string, 0, 1)
		for _, filterStatus := range filter.Statuses {
			if s.ID == filterStatus.StatusID {
				statusMatches = append(statusMatches, filterStatus.StatusID)
				break
			}
		}

		if len(keywordMatches) > 0 || len(statusMatches) > 0 {
			switch filter.Action {
			case gtsmodel.FilterActionWarn:
				// Record what matched.
				apiFilter, err := c.FilterToAPIFilterV2(ctx, filter)
				if err != nil {
					return nil, err
				}
				filterResults = append(filterResults, apimodel.FilterResult{
					Filter:         *apiFilter,
					KeywordMatches: keywordMatches,
					StatusMatches:  statusMatches,
				})

			case gtsmodel.FilterActionHide:
				// Don't show this status. Immediate return.
				return nil, statusfilter.ErrHideStatus
			}
		}
	}

	return filterResults, nil
}

// filterAppliesInContext returns whether a given filter applies in a given context.
func filterAppliesInContext(filter *gtsmodel.Filter, filterContext statusfilter.FilterContext) bool {
	switch filterContext {
	case statusfilter.FilterContextHome:
		return util.PtrOrValue(filter.ContextHome, false)
	case statusfilter.FilterContextNotifications:
		return util.PtrOrValue(filter.ContextNotifications, false)
	case statusfilter.FilterContextPublic:
		return util.PtrOrValue(filter.ContextPublic, false)
	case statusfilter.FilterContextThread:
		return util.PtrOrValue(filter.ContextThread, false)
	case statusfilter.FilterContextAccount:
		return util.PtrOrValue(filter.ContextAccount, false)
	}
	return false
}

// StatusToWebStatus converts a gts model status into an
// api representation suitable for serving into a web template.
//
// Requesting account can be nil.
func (c *Converter) StatusToWebStatus(
	ctx context.Context,
	s *gtsmodel.Status,
) (*apimodel.WebStatus, error) {
	apiStatus, err := c.statusToFrontend(ctx, s,
		nil,                            // No authed requester.
		statusfilter.FilterContextNone, // No filters.
		nil,                            // No filters.
		nil,                            // No mutes.
	)
	if err != nil {
		return nil, err
	}

	// Convert status author to web model.
	acct, err := c.AccountToWebAccount(ctx, s.Account)
	if err != nil {
		return nil, err
	}

	webStatus := &apimodel.WebStatus{
		Status:  apiStatus,
		Account: acct,
	}

	// Whack a newline before and after each "pre" to make it easier to outdent it.
	webStatus.Content = strings.ReplaceAll(webStatus.Content, "<pre>", "\n<pre>")
	webStatus.Content = strings.ReplaceAll(webStatus.Content, "</pre>", "</pre>\n")

	// Add additional information for template.
	// Assume empty langs, hope for not empty language.
	webStatus.LanguageTag = new(language.Language)
	if lang := webStatus.Language; lang != nil {
		langTag, err := language.Parse(*lang)
		if err != nil {
			log.Warnf(
				ctx,
				"error parsing %s as language tag: %v",
				*lang, err,
			)
		} else {
			webStatus.LanguageTag = langTag
		}
	}

	if poll := webStatus.Poll; poll != nil {
		// Calculate vote share of each poll option and
		// format them for easier template consumption.
		totalVotes := poll.VotesCount

		PollOptions := make([]apimodel.WebPollOption, len(poll.Options))
		for i, option := range poll.Options {
			var voteShare float32

			if totalVotes != 0 && option.VotesCount != nil {
				voteShare = float32(*option.VotesCount) / float32(totalVotes) * 100
			}

			// Format to two decimal points and ditch any
			// trailing zeroes.
			//
			// We want to be precise enough that eg., "1.54%"
			// is distinct from "1.68%" in polls with loads
			// of votes.
			//
			// However, if we've got eg., a two-option poll
			// in which each option has half the votes, then
			// "50%" looks better than "50.00%".
			//
			// By the same token, it's pointless to show
			// "0.00%" or "100.00%".
			voteShareStr := fmt.Sprintf("%.2f", voteShare)
			voteShareStr = strings.TrimSuffix(voteShareStr, ".00")

			webPollOption := apimodel.WebPollOption{
				PollOption:   option,
				PollID:       poll.ID,
				Emojis:       webStatus.Emojis,
				LanguageTag:  webStatus.LanguageTag,
				VoteShare:    voteShare,
				VoteShareStr: voteShareStr,
			}
			PollOptions[i] = webPollOption
		}

		webStatus.PollOptions = PollOptions
	}

	// Mark local.
	webStatus.Local = *s.Local

	// Set additional templating
	// variables on media attachments.

	// Get gtsmodel attachments
	// into a convenient map.
	ogAttachments := make(
		map[string]*gtsmodel.MediaAttachment,
		len(s.Attachments),
	)
	for _, a := range s.Attachments {
		ogAttachments[a.ID] = a
	}

	// Convert each API attachment
	// into a web attachment.
	webStatus.MediaAttachments = make(
		[]*apimodel.WebAttachment,
		len(apiStatus.MediaAttachments),
	)
	for i, apiAttachment := range apiStatus.MediaAttachments {
		ogAttachment := ogAttachments[apiAttachment.ID]
		webStatus.MediaAttachments[i] = &apimodel.WebAttachment{
			Attachment:      apiAttachment,
			Sensitive:       apiStatus.Sensitive,
			MIMEType:        ogAttachment.File.ContentType,
			PreviewMIMEType: ogAttachment.Thumbnail.ContentType,
		}
	}

	return webStatus, nil
}

// StatusToAPIStatusSource returns the *apimodel.StatusSource of the given status.
// Callers should check beforehand whether a requester has permission to view the
// source of the status, and ensure they're passing only a local status into this function.
func (c *Converter) StatusToAPIStatusSource(ctx context.Context, s *gtsmodel.Status) (*apimodel.StatusSource, error) {
	// TODO: remove this when edit support is added.
	text := "**STATUS EDITS ARE NOT CURRENTLY SUPPORTED IN GOTOSOCIAL (coming in 2024)**\n" +
		"You can review the original text of your status below, but you will not be able to submit this edit.\n\n---\n\n" + s.Text

	return &apimodel.StatusSource{
		ID:          s.ID,
		Text:        text,
		SpoilerText: s.ContentWarning,
	}, nil
}

// statusToFrontend is a package internal function for
// parsing a status into its initial frontend representation.
//
// Requesting account can be nil.
//
// This function also doesn't handle converting the
// account to api/web model -- the caller must do that.
func (c *Converter) statusToFrontend(
	ctx context.Context,
	status *gtsmodel.Status,
	requestingAccount *gtsmodel.Account,
	filterContext statusfilter.FilterContext,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (
	*apimodel.Status,
	error,
) {
	apiStatus, err := c.baseStatusToFrontend(ctx,
		status,
		requestingAccount,
		filterContext,
		filters,
		mutes,
	)
	if err != nil {
		return nil, err
	}

	if status.BoostOf != nil {
		reblog, err := c.baseStatusToFrontend(ctx,
			status.BoostOf,
			requestingAccount,
			filterContext,
			filters,
			mutes,
		)
		if errors.Is(err, statusfilter.ErrHideStatus) {
			// If we'd hide the original status, hide the boost.
			return nil, err
		} else if err != nil {
			return nil, gtserror.Newf("error converting boosted status: %w", err)
		}

		// Set boosted status and set interactions and filter results from original.
		apiStatus.Reblog = &apimodel.StatusReblogged{reblog}
		apiStatus.Favourited = apiStatus.Reblog.Favourited
		apiStatus.Bookmarked = apiStatus.Reblog.Bookmarked
		apiStatus.Muted = apiStatus.Reblog.Muted
		apiStatus.Reblogged = apiStatus.Reblog.Reblogged
		apiStatus.Pinned = apiStatus.Reblog.Pinned
		apiStatus.Filtered = apiStatus.Reblog.Filtered
	}

	return apiStatus, nil
}

// baseStatusToFrontend performs the main logic
// of statusToFrontend() without handling of boost
// logic, to prevent *possible* recursion issues.
//
// This function also doesn't handle converting the
// account to api/web model -- the caller must do that.
func (c *Converter) baseStatusToFrontend(
	ctx context.Context,
	s *gtsmodel.Status,
	requestingAccount *gtsmodel.Account,
	filterContext statusfilter.FilterContext,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (
	*apimodel.Status,
	error,
) {
	// Try to populate status struct pointer fields.
	// We can continue in many cases of partial failure,
	// but there are some fields we actually need.
	if err := c.state.DB.PopulateStatus(ctx, s); err != nil {
		switch {
		case s.Account == nil:
			return nil, gtserror.Newf("error(s) populating status, required account not set: %w", err)

		case s.BoostOfID != "" && s.BoostOf == nil:
			return nil, gtserror.Newf("error(s) populating status, required boost not set: %w", err)

		default:
			log.Errorf(ctx, "error(s) populating status, will continue: %v", err)
		}
	}

	repliesCount, err := c.state.DB.CountStatusReplies(ctx, s.ID)
	if err != nil {
		return nil, gtserror.Newf("error counting replies: %w", err)
	}

	reblogsCount, err := c.state.DB.CountStatusBoosts(ctx, s.ID)
	if err != nil {
		return nil, gtserror.Newf("error counting reblogs: %w", err)
	}

	favesCount, err := c.state.DB.CountStatusFaves(ctx, s.ID)
	if err != nil {
		return nil, gtserror.Newf("error counting faves: %w", err)
	}

	apiAttachments, err := c.convertAttachmentsToAPIAttachments(ctx, s.Attachments, s.AttachmentIDs)
	if err != nil {
		log.Errorf(ctx, "error converting status attachments: %v", err)
	}

	apiMentions, err := c.convertMentionsToAPIMentions(ctx, s.Mentions, s.MentionIDs)
	if err != nil {
		log.Errorf(ctx, "error converting status mentions: %v", err)
	}

	apiTags, err := c.convertTagsToAPITags(ctx, s.Tags, s.TagIDs)
	if err != nil {
		log.Errorf(ctx, "error converting status tags: %v", err)
	}

	apiEmojis, err := c.convertEmojisToAPIEmojis(ctx, s.Emojis, s.EmojiIDs)
	if err != nil {
		log.Errorf(ctx, "error converting status emojis: %v", err)
	}

	// Take status's interaction policy, or
	// fall back to default for its visibility.
	var p *gtsmodel.InteractionPolicy
	if s.InteractionPolicy != nil {
		p = s.InteractionPolicy
	} else {
		p = gtsmodel.DefaultInteractionPolicyFor(s.Visibility)
	}

	apiInteractionPolicy, err := c.InteractionPolicyToAPIInteractionPolicy(ctx, p, s, requestingAccount)
	if err != nil {
		return nil, gtserror.Newf("error converting interaction policy: %w", err)
	}

	apiStatus := &apimodel.Status{
		ID:                 s.ID,
		CreatedAt:          util.FormatISO8601(s.CreatedAt),
		InReplyToID:        nil, // Set below.
		InReplyToAccountID: nil, // Set below.
		Sensitive:          *s.Sensitive,
		SpoilerText:        s.ContentWarning,
		Visibility:         c.VisToAPIVis(ctx, s.Visibility),
		LocalOnly:          s.IsLocalOnly(),
		Language:           nil, // Set below.
		URI:                s.URI,
		URL:                s.URL,
		RepliesCount:       repliesCount,
		ReblogsCount:       reblogsCount,
		FavouritesCount:    favesCount,
		Content:            s.Content,
		Reblog:             nil, // Set below.
		Application:        nil, // Set below.
		Account:            nil, // Caller must do this.
		MediaAttachments:   apiAttachments,
		Mentions:           apiMentions,
		Tags:               apiTags,
		Emojis:             apiEmojis,
		Card:               nil, // TODO: implement cards
		Text:               s.Text,
		InteractionPolicy:  *apiInteractionPolicy,
	}

	// Nullable fields.
	if s.InReplyToID != "" {
		apiStatus.InReplyToID = util.Ptr(s.InReplyToID)
	}

	if s.InReplyToAccountID != "" {
		apiStatus.InReplyToAccountID = util.Ptr(s.InReplyToAccountID)
	}

	if s.Language != "" {
		apiStatus.Language = util.Ptr(s.Language)
	}

	if app := s.CreatedWithApplication; app != nil {
		apiStatus.Application, err = c.AppToAPIAppPublic(ctx, app)
		if err != nil {
			return nil, gtserror.Newf(
				"error converting application %s: %w",
				s.CreatedWithApplicationID, err,
			)
		}
	}

	if s.Poll != nil {
		// Set originating
		// status on the poll.
		poll := s.Poll
		poll.Status = s

		apiStatus.Poll, err = c.PollToAPIPoll(ctx, requestingAccount, poll)
		if err != nil {
			return nil, fmt.Errorf("error converting poll: %w", err)
		}
	}

	// Status interactions.
	//
	if s.BoostOf != nil { //nolint
		// populated *outside* this
		// function to prevent recursion.
	} else {
		interacts, err := c.interactionsWithStatusForAccount(ctx, s, requestingAccount)
		if err != nil {
			log.Errorf(ctx,
				"error getting interactions for status %s for account %s: %v",
				s.ID, requestingAccount.ID, err,
			)

			// Ensure non-nil object.
			interacts = new(statusInteractions)
		}
		apiStatus.Favourited = interacts.Favourited
		apiStatus.Bookmarked = interacts.Bookmarked
		apiStatus.Muted = interacts.Muted
		apiStatus.Reblogged = interacts.Reblogged
		apiStatus.Pinned = interacts.Pinned
	}

	// If web URL is empty for whatever
	// reason, provide AP URI as fallback.
	if s.URL == "" {
		s.URL = s.URI
	}

	// Apply filters.
	filterResults, err := c.statusToAPIFilterResults(ctx, s, requestingAccount, filterContext, filters, mutes)
	if err != nil {
		if errors.Is(err, statusfilter.ErrHideStatus) {
			return nil, err
		}
		return nil, fmt.Errorf("error applying filters: %w", err)
	}

	apiStatus.Filtered = filterResults

	return apiStatus, nil
}

// VisToAPIVis converts a gts visibility into its api equivalent
func (c *Converter) VisToAPIVis(ctx context.Context, m gtsmodel.Visibility) apimodel.Visibility {
	switch m {
	case gtsmodel.VisibilityPublic:
		return apimodel.VisibilityPublic
	case gtsmodel.VisibilityUnlocked:
		return apimodel.VisibilityUnlisted
	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		return apimodel.VisibilityPrivate
	case gtsmodel.VisibilityDirect:
		return apimodel.VisibilityDirect
	}
	return ""
}

// InstanceRuleToAdminAPIRule converts a local instance rule into its api equivalent for serving at /api/v1/admin/instance/rules/:id
func (c *Converter) InstanceRuleToAPIRule(r gtsmodel.Rule) apimodel.InstanceRule {
	return apimodel.InstanceRule{
		ID:   r.ID,
		Text: r.Text,
	}
}

// InstanceRulesToAPIRules converts all local instance rules into their api equivalent for serving at /api/v1/instance/rules
func (c *Converter) InstanceRulesToAPIRules(r []gtsmodel.Rule) []apimodel.InstanceRule {
	rules := make([]apimodel.InstanceRule, len(r))

	for i, v := range r {
		rules[i] = c.InstanceRuleToAPIRule(v)
	}

	return rules
}

// InstanceRuleToAdminAPIRule converts a local instance rule into its api equivalent for serving at /api/v1/admin/instance/rules/:id
func (c *Converter) InstanceRuleToAdminAPIRule(r *gtsmodel.Rule) *apimodel.AdminInstanceRule {
	return &apimodel.AdminInstanceRule{
		ID:        r.ID,
		CreatedAt: util.FormatISO8601(r.CreatedAt),
		UpdatedAt: util.FormatISO8601(r.UpdatedAt),
		Text:      r.Text,
	}
}

// InstanceToAPIV1Instance converts a gts instance into its api equivalent for serving at /api/v1/instance
func (c *Converter) InstanceToAPIV1Instance(ctx context.Context, i *gtsmodel.Instance) (*apimodel.InstanceV1, error) {
	domain := i.Domain
	accDomain := config.GetAccountDomain()
	if accDomain != "" {
		domain = accDomain
	}

	instance := &apimodel.InstanceV1{
		URI:                  domain,
		AccountDomain:        accDomain,
		Title:                i.Title,
		Description:          i.Description,
		DescriptionText:      i.DescriptionText,
		ShortDescription:     i.ShortDescription,
		ShortDescriptionText: i.ShortDescriptionText,
		Email:                i.ContactEmail,
		Version:              config.GetSoftwareVersion(),
		Languages:            config.GetInstanceLanguages().TagStrs(),
		Registrations:        config.GetAccountsRegistrationOpen(),
		ApprovalRequired:     true,                               // approval always required
		InvitesEnabled:       false,                              // todo: not supported yet
		MaxTootChars:         uint(config.GetStatusesMaxChars()), // #nosec G115 -- Already validated.
		Rules:                c.InstanceRulesToAPIRules(i.Rules),
		Terms:                i.Terms,
		TermsRaw:             i.TermsText,
	}

	if config.GetInstanceInjectMastodonVersion() {
		instance.Version = toMastodonVersion(instance.Version)
	}

	if debug.DEBUG {
		instance.Debug = util.Ptr(true)
	}

	// configuration
	instance.Configuration.Statuses.MaxCharacters = config.GetStatusesMaxChars()
	instance.Configuration.Statuses.MaxMediaAttachments = config.GetStatusesMediaMaxFiles()
	instance.Configuration.Statuses.CharactersReservedPerURL = instanceStatusesCharactersReservedPerURL
	instance.Configuration.Statuses.SupportedMimeTypes = instanceStatusesSupportedMimeTypes
	instance.Configuration.MediaAttachments.SupportedMimeTypes = media.SupportedMIMETypes

	// NOTE: we use the local max sizes here
	// as it hints to apps like Tusky for image
	// compression of locally uploaded media.
	//
	// TODO: return local / remote depending
	// on authorized endpoint user (if any)?
	localMax := config.GetMediaLocalMaxSize()
	imageSz := cmp.Or(config.GetMediaImageSizeHint(), localMax)
	videoSz := cmp.Or(config.GetMediaVideoSizeHint(), localMax)
	instance.Configuration.MediaAttachments.ImageSizeLimit = int(imageSz) // #nosec G115 -- Already validated.
	instance.Configuration.MediaAttachments.VideoSizeLimit = int(videoSz) // #nosec G115 -- Already validated.

	// we don't actually set any limits on these. set to max possible.
	instance.Configuration.MediaAttachments.ImageMatrixLimit = math.MaxInt32
	instance.Configuration.MediaAttachments.VideoFrameRateLimit = math.MaxInt32
	instance.Configuration.MediaAttachments.VideoMatrixLimit = math.MaxInt32

	instance.Configuration.Polls.MaxOptions = config.GetStatusesPollMaxOptions()
	instance.Configuration.Polls.MaxCharactersPerOption = config.GetStatusesPollOptionMaxChars()
	instance.Configuration.Polls.MinExpiration = instancePollsMinExpiration
	instance.Configuration.Polls.MaxExpiration = instancePollsMaxExpiration
	instance.Configuration.Accounts.AllowCustomCSS = config.GetAccountsAllowCustomCSS()
	instance.Configuration.Accounts.MaxFeaturedTags = instanceAccountsMaxFeaturedTags
	instance.Configuration.Accounts.MaxProfileFields = instanceAccountsMaxProfileFields
	instance.Configuration.Emojis.EmojiSizeLimit = int(config.GetMediaEmojiLocalMaxSize()) // #nosec G115 -- Already validated.
	instance.Configuration.OIDCEnabled = config.GetOIDCEnabled()

	// URLs
	instance.URLs.StreamingAPI = "wss://" + i.Domain

	// statistics
	stats := make(map[string]*int, 3)
	userCount, err := c.state.DB.CountInstanceUsers(ctx, i.Domain)
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting counting instance users: %w", err)
	}
	stats["user_count"] = util.Ptr(userCount)

	statusCount, err := c.state.DB.CountInstanceStatuses(ctx, i.Domain)
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting counting instance statuses: %w", err)
	}
	stats["status_count"] = util.Ptr(statusCount)

	domainCount, err := c.state.DB.CountInstanceDomains(ctx, i.Domain)
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting counting instance domains: %w", err)
	}
	stats["domain_count"] = util.Ptr(domainCount)
	instance.Stats = stats

	// thumbnail
	iAccount, err := c.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting instance account: %w", err)
	}

	if iAccount.AvatarMediaAttachmentID != "" {
		if iAccount.AvatarMediaAttachment == nil {
			avi, err := c.state.DB.GetAttachmentByID(ctx, iAccount.AvatarMediaAttachmentID)
			if err != nil {
				return nil, fmt.Errorf("InstanceToAPIInstance: error getting instance avatar attachment with id %s: %w", iAccount.AvatarMediaAttachmentID, err)
			}
			iAccount.AvatarMediaAttachment = avi
		}

		instance.Thumbnail = iAccount.AvatarMediaAttachment.URL
		instance.ThumbnailType = iAccount.AvatarMediaAttachment.File.ContentType
		instance.ThumbnailStatic = iAccount.AvatarMediaAttachment.Thumbnail.URL
		instance.ThumbnailStaticType = iAccount.AvatarMediaAttachment.Thumbnail.ContentType
		instance.ThumbnailDescription = iAccount.AvatarMediaAttachment.Description
	} else {
		instance.Thumbnail = config.GetProtocol() + "://" + i.Domain + "/assets/logo.webp" // default thumb
	}

	// contact account
	if i.ContactAccountID != "" {
		if i.ContactAccount == nil {
			contactAccount, err := c.state.DB.GetAccountByID(ctx, i.ContactAccountID)
			if err != nil {
				return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting instance contact account %s: %w", i.ContactAccountID, err)
			}
			i.ContactAccount = contactAccount
		}

		account, err := c.AccountToAPIAccountPublic(ctx, i.ContactAccount)
		if err != nil {
			return nil, fmt.Errorf("InstanceToAPIV1Instance: error converting instance contact account %s: %w", i.ContactAccountID, err)
		}
		instance.ContactAccount = account
	}

	return instance, nil
}

// InstanceToAPIV2Instance converts a gts instance into its api equivalent for serving at /api/v2/instance
func (c *Converter) InstanceToAPIV2Instance(ctx context.Context, i *gtsmodel.Instance) (*apimodel.InstanceV2, error) {
	domain := i.Domain
	accDomain := config.GetAccountDomain()
	if accDomain != "" {
		domain = accDomain
	}

	instance := &apimodel.InstanceV2{
		Domain:          domain,
		AccountDomain:   accDomain,
		Title:           i.Title,
		Version:         config.GetSoftwareVersion(),
		SourceURL:       instanceSourceURL,
		Description:     i.Description,
		DescriptionText: i.DescriptionText,
		Usage:           apimodel.InstanceV2Usage{}, // todo: not implemented
		Languages:       config.GetInstanceLanguages().TagStrs(),
		Rules:           c.InstanceRulesToAPIRules(i.Rules),
		Terms:           i.Terms,
		TermsText:       i.TermsText,
	}

	if config.GetInstanceInjectMastodonVersion() {
		instance.Version = toMastodonVersion(instance.Version)
	}

	if debug.DEBUG {
		instance.Debug = util.Ptr(true)
	}

	// thumbnail
	thumbnail := apimodel.InstanceV2Thumbnail{}

	iAccount, err := c.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV2Instance: db error getting instance account: %w", err)
	}

	if iAccount.AvatarMediaAttachmentID != "" {
		if iAccount.AvatarMediaAttachment == nil {
			avi, err := c.state.DB.GetAttachmentByID(ctx, iAccount.AvatarMediaAttachmentID)
			if err != nil {
				return nil, fmt.Errorf("InstanceToAPIV2Instance: error getting instance avatar attachment with id %s: %w", iAccount.AvatarMediaAttachmentID, err)
			}
			iAccount.AvatarMediaAttachment = avi
		}

		thumbnail.URL = iAccount.AvatarMediaAttachment.URL
		thumbnail.Type = iAccount.AvatarMediaAttachment.File.ContentType
		thumbnail.StaticURL = iAccount.AvatarMediaAttachment.Thumbnail.URL
		thumbnail.StaticType = iAccount.AvatarMediaAttachment.Thumbnail.ContentType
		thumbnail.Description = iAccount.AvatarMediaAttachment.Description
		thumbnail.Blurhash = iAccount.AvatarMediaAttachment.Blurhash
	} else {
		thumbnail.URL = config.GetProtocol() + "://" + i.Domain + "/assets/logo.webp" // default thumb
	}

	instance.Thumbnail = thumbnail

	// configuration
	instance.Configuration.URLs.Streaming = "wss://" + i.Domain
	instance.Configuration.Statuses.MaxCharacters = config.GetStatusesMaxChars()
	instance.Configuration.Statuses.MaxMediaAttachments = config.GetStatusesMediaMaxFiles()
	instance.Configuration.Statuses.CharactersReservedPerURL = instanceStatusesCharactersReservedPerURL
	instance.Configuration.Statuses.SupportedMimeTypes = instanceStatusesSupportedMimeTypes
	instance.Configuration.MediaAttachments.SupportedMimeTypes = media.SupportedMIMETypes

	// NOTE: we use the local max sizes here
	// as it hints to apps like Tusky for image
	// compression of locally uploaded media.
	//
	// TODO: return local / remote depending
	// on authorized endpoint user (if any)?
	localMax := config.GetMediaLocalMaxSize()
	imageSz := cmp.Or(config.GetMediaImageSizeHint(), localMax)
	videoSz := cmp.Or(config.GetMediaVideoSizeHint(), localMax)
	instance.Configuration.MediaAttachments.ImageSizeLimit = int(imageSz) // #nosec G115 -- Already validated.
	instance.Configuration.MediaAttachments.VideoSizeLimit = int(videoSz) // #nosec G115 -- Already validated.

	// we don't actually set any limits on these. set to max possible.
	instance.Configuration.MediaAttachments.ImageMatrixLimit = math.MaxInt32
	instance.Configuration.MediaAttachments.VideoFrameRateLimit = math.MaxInt32
	instance.Configuration.MediaAttachments.VideoMatrixLimit = math.MaxInt32

	instance.Configuration.Polls.MaxOptions = config.GetStatusesPollMaxOptions()
	instance.Configuration.Polls.MaxCharactersPerOption = config.GetStatusesPollOptionMaxChars()
	instance.Configuration.Polls.MinExpiration = instancePollsMinExpiration
	instance.Configuration.Polls.MaxExpiration = instancePollsMaxExpiration
	instance.Configuration.Accounts.AllowCustomCSS = config.GetAccountsAllowCustomCSS()
	instance.Configuration.Accounts.MaxFeaturedTags = instanceAccountsMaxFeaturedTags
	instance.Configuration.Accounts.MaxProfileFields = instanceAccountsMaxProfileFields
	instance.Configuration.Emojis.EmojiSizeLimit = int(config.GetMediaEmojiLocalMaxSize()) // #nosec G115 -- Already validated.
	instance.Configuration.OIDCEnabled = config.GetOIDCEnabled()

	// registrations
	instance.Registrations.Enabled = config.GetAccountsRegistrationOpen()
	instance.Registrations.ApprovalRequired = true // always required
	instance.Registrations.Message = nil           // todo: not implemented

	// contact
	instance.Contact.Email = i.ContactEmail
	if i.ContactAccountID != "" {
		if i.ContactAccount == nil {
			contactAccount, err := c.state.DB.GetAccountByID(ctx, i.ContactAccountID)
			if err != nil {
				return nil, fmt.Errorf("InstanceToAPIV2Instance: db error getting instance contact account %s: %w", i.ContactAccountID, err)
			}
			i.ContactAccount = contactAccount
		}

		account, err := c.AccountToAPIAccountPublic(ctx, i.ContactAccount)
		if err != nil {
			return nil, fmt.Errorf("InstanceToAPIV2Instance: error converting instance contact account %s: %w", i.ContactAccountID, err)
		}
		instance.Contact.Account = account
	}

	return instance, nil
}

// RelationshipToAPIRelationship converts a gts relationship into its api equivalent for serving in various places
func (c *Converter) RelationshipToAPIRelationship(ctx context.Context, r *gtsmodel.Relationship) (*apimodel.Relationship, error) {
	return &apimodel.Relationship{
		ID:                  r.ID,
		Following:           r.Following,
		ShowingReblogs:      r.ShowingReblogs,
		Notifying:           r.Notifying,
		FollowedBy:          r.FollowedBy,
		Blocking:            r.Blocking,
		BlockedBy:           r.BlockedBy,
		Muting:              r.Muting,
		MutingNotifications: r.MutingNotifications,
		Requested:           r.Requested,
		RequestedBy:         r.RequestedBy,
		DomainBlocking:      r.DomainBlocking,
		Endorsed:            r.Endorsed,
		Note:                r.Note,
	}, nil
}

// NotificationToAPINotification converts a gts notification into a api notification
func (c *Converter) NotificationToAPINotification(
	ctx context.Context,
	n *gtsmodel.Notification,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (*apimodel.Notification, error) {
	if n.TargetAccount == nil {
		tAccount, err := c.state.DB.GetAccountByID(ctx, n.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("NotificationToapi: error getting target account with id %s from the db: %s", n.TargetAccountID, err)
		}
		n.TargetAccount = tAccount
	}

	if n.OriginAccount == nil {
		ogAccount, err := c.state.DB.GetAccountByID(ctx, n.OriginAccountID)
		if err != nil {
			return nil, fmt.Errorf("NotificationToapi: error getting origin account with id %s from the db: %s", n.OriginAccountID, err)
		}
		n.OriginAccount = ogAccount
	}

	apiAccount, err := c.AccountToAPIAccountPublic(ctx, n.OriginAccount)
	if err != nil {
		return nil, fmt.Errorf("NotificationToapi: error converting account to api: %s", err)
	}

	var apiStatus *apimodel.Status
	if n.StatusID != "" {
		if n.Status == nil {
			status, err := c.state.DB.GetStatusByID(ctx, n.StatusID)
			if err != nil {
				return nil, fmt.Errorf("NotificationToapi: error getting status with id %s from the db: %s", n.StatusID, err)
			}
			n.Status = status
		}

		if n.Status.Account == nil {
			if n.Status.AccountID == n.TargetAccount.ID {
				n.Status.Account = n.TargetAccount
			} else if n.Status.AccountID == n.OriginAccount.ID {
				n.Status.Account = n.OriginAccount
			}
		}

		var err error
		apiStatus, err = c.StatusToAPIStatus(ctx, n.Status, n.TargetAccount, statusfilter.FilterContextNotifications, filters, mutes)
		if err != nil {
			if errors.Is(err, statusfilter.ErrHideStatus) {
				return nil, err
			}
			return nil, fmt.Errorf("NotificationToapi: error converting status to api: %s", err)
		}
	}

	if apiStatus != nil && apiStatus.Reblog != nil {
		// use the actual reblog status for the notifications endpoint
		apiStatus = apiStatus.Reblog.Status
	}

	return &apimodel.Notification{
		ID:        n.ID,
		Type:      string(n.NotificationType),
		CreatedAt: util.FormatISO8601(n.CreatedAt),
		Account:   apiAccount,
		Status:    apiStatus,
	}, nil
}

// ConversationToAPIConversation converts a conversation into its API representation.
// The conversation status will be filtered using the notification filter context,
// and may be nil if the status was hidden.
func (c *Converter) ConversationToAPIConversation(
	ctx context.Context,
	conversation *gtsmodel.Conversation,
	requester *gtsmodel.Account,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (*apimodel.Conversation, error) {
	apiConversation := &apimodel.Conversation{
		ID:     conversation.ID,
		Unread: !*conversation.Read,
	}

	// Populate most recent status in convo;
	// can be nil if this status is filtered.
	if conversation.LastStatus != nil {
		var err error
		apiConversation.LastStatus, err = c.StatusToAPIStatus(
			ctx,
			conversation.LastStatus,
			requester,
			statusfilter.FilterContextNotifications,
			filters,
			mutes,
		)
		if err != nil && !errors.Is(err, statusfilter.ErrHideStatus) {
			return nil, gtserror.Newf(
				"error converting status %s to API representation: %w",
				conversation.LastStatus.ID,
				err,
			)
		}
	}

	// If no other accounts are involved in this convo,
	// just include the requesting account and return.
	//
	// See: https://github.com/superseriousbusiness/gotosocial/issues/3385#issuecomment-2394033477
	otherAcctsLen := len(conversation.OtherAccounts)
	if otherAcctsLen == 0 {
		apiAcct, err := c.AccountToAPIAccountPublic(ctx, requester)
		if err != nil {
			err := gtserror.Newf(
				"error converting account %s to API representation: %w",
				requester.ID, err,
			)
			return nil, err
		}

		apiConversation.Accounts = []apimodel.Account{*apiAcct}
		return apiConversation, nil
	}

	// Other accounts are involved in the
	// convo. Convert each to API model.
	apiConversation.Accounts = make([]apimodel.Account, otherAcctsLen)
	for i, account := range conversation.OtherAccounts {
		blocked, err := c.state.DB.IsEitherBlocked(ctx,
			requester.ID, account.ID,
		)
		if err != nil {
			err := gtserror.Newf(
				"db error checking blocks between accounts %s and %s: %w",
				requester.ID, account.ID, err,
			)
			return nil, err
		}

		// API account model varies depending
		// on status of conversation participant.
		var apiAcct *apimodel.Account
		if blocked || account.IsSuspended() {
			apiAcct, err = c.AccountToAPIAccountBlocked(ctx, account)
		} else {
			apiAcct, err = c.AccountToAPIAccountPublic(ctx, account)
		}

		if err != nil {
			err := gtserror.Newf(
				"error converting account %s to API representation: %w",
				account.ID, err,
			)
			return nil, err
		}

		apiConversation.Accounts[i] = *apiAcct
	}

	return apiConversation, nil
}

// DomainPermToAPIDomainPerm converts a gtsmodel domain block,
// allow, draft, or ignore into an api domain permission.
func (c *Converter) DomainPermToAPIDomainPerm(
	ctx context.Context,
	d gtsmodel.DomainPermission,
	export bool,
) (*apimodel.DomainPermission, error) {
	// Domain may be in Punycode,
	// de-punify it just in case.
	domain, err := util.DePunify(d.GetDomain())
	if err != nil {
		return nil, gtserror.Newf("error de-punifying domain %s: %w", d.GetDomain(), err)
	}

	domainPerm := &apimodel.DomainPermission{
		Domain: apimodel.Domain{
			Domain:        domain,
			PublicComment: d.GetPublicComment(),
		},
	}

	// If we're exporting, provide
	// only bare minimum detail.
	if export {
		return domainPerm, nil
	}

	domainPerm.ID = d.GetID()
	domainPerm.Obfuscate = *d.GetObfuscate()
	domainPerm.PrivateComment = d.GetPrivateComment()
	domainPerm.SubscriptionID = d.GetSubscriptionID()
	domainPerm.CreatedBy = d.GetCreatedByAccountID()
	domainPerm.CreatedAt = util.FormatISO8601(d.GetCreatedAt())

	// If this is a draft, also add the permission type.
	if _, ok := d.(*gtsmodel.DomainPermissionDraft); ok {
		domainPerm.PermissionType = d.GetType().String()
	}

	return domainPerm, nil
}

// ReportToAPIReport converts a gts model report into an api model report, for serving at /api/v1/reports
func (c *Converter) ReportToAPIReport(ctx context.Context, r *gtsmodel.Report) (*apimodel.Report, error) {
	report := &apimodel.Report{
		ID:          r.ID,
		CreatedAt:   util.FormatISO8601(r.CreatedAt),
		ActionTaken: !r.ActionTakenAt.IsZero(),
		Category:    "other", // todo: only support default 'other' category right now
		Comment:     r.Comment,
		Forwarded:   *r.Forwarded,
		StatusIDs:   r.StatusIDs,
		RuleIDs:     r.RuleIDs,
	}

	if !r.ActionTakenAt.IsZero() {
		actionTakenAt := util.FormatISO8601(r.ActionTakenAt)
		report.ActionTakenAt = &actionTakenAt
	}

	if actionComment := r.ActionTaken; actionComment != "" {
		report.ActionTakenComment = &actionComment
	}

	if r.TargetAccount == nil {
		tAccount, err := c.state.DB.GetAccountByID(ctx, r.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("ReportToAPIReport: error getting target account with id %s from the db: %s", r.TargetAccountID, err)
		}
		r.TargetAccount = tAccount
	}

	apiAccount, err := c.AccountToAPIAccountPublic(ctx, r.TargetAccount)
	if err != nil {
		return nil, fmt.Errorf("ReportToAPIReport: error converting target account to api: %s", err)
	}
	report.TargetAccount = apiAccount

	return report, nil
}

// ReportToAdminAPIReport converts a gts model report into an admin view report, for serving at /api/v1/admin/reports
func (c *Converter) ReportToAdminAPIReport(ctx context.Context, r *gtsmodel.Report, requestingAccount *gtsmodel.Account) (*apimodel.AdminReport, error) {
	var (
		err                  error
		actionTakenAt        *string
		actionTakenComment   *string
		actionTakenByAccount *apimodel.AdminAccountInfo
	)

	if !r.ActionTakenAt.IsZero() {
		ata := util.FormatISO8601(r.ActionTakenAt)
		actionTakenAt = &ata
	}

	if r.Account == nil {
		r.Account, err = c.state.DB.GetAccountByID(ctx, r.AccountID)
		if err != nil {
			return nil, fmt.Errorf("ReportToAdminAPIReport: error getting account with id %s from the db: %w", r.AccountID, err)
		}
	}
	account, err := c.AccountToAdminAPIAccount(ctx, r.Account)
	if err != nil {
		return nil, fmt.Errorf("ReportToAdminAPIReport: error converting account with id %s to adminAPIAccount: %w", r.AccountID, err)
	}

	if r.TargetAccount == nil {
		r.TargetAccount, err = c.state.DB.GetAccountByID(ctx, r.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("ReportToAdminAPIReport: error getting target account with id %s from the db: %w", r.TargetAccountID, err)
		}
	}
	targetAccount, err := c.AccountToAdminAPIAccount(ctx, r.TargetAccount)
	if err != nil {
		return nil, fmt.Errorf("ReportToAdminAPIReport: error converting target account with id %s to adminAPIAccount: %w", r.TargetAccountID, err)
	}

	if r.ActionTakenByAccountID != "" {
		if r.ActionTakenByAccount == nil {
			r.ActionTakenByAccount, err = c.state.DB.GetAccountByID(ctx, r.ActionTakenByAccountID)
			if err != nil {
				return nil, fmt.Errorf("ReportToAdminAPIReport: error getting action taken by account with id %s from the db: %w", r.ActionTakenByAccountID, err)
			}
		}

		actionTakenByAccount, err = c.AccountToAdminAPIAccount(ctx, r.ActionTakenByAccount)
		if err != nil {
			return nil, fmt.Errorf("ReportToAdminAPIReport: error converting action taken by account with id %s to adminAPIAccount: %w", r.ActionTakenByAccountID, err)
		}
	}

	statuses := make([]*apimodel.Status, 0, len(r.StatusIDs))
	if len(r.StatusIDs) != 0 && len(r.Statuses) == 0 {
		r.Statuses, err = c.state.DB.GetStatusesByIDs(ctx, r.StatusIDs)
		if err != nil {
			return nil, fmt.Errorf("ReportToAdminAPIReport: error getting statuses from the db: %w", err)
		}
	}
	for _, s := range r.Statuses {
		status, err := c.statusToAPIStatus(
			ctx,
			s,
			requestingAccount,
			statusfilter.FilterContextNone,
			nil,  // No filters.
			nil,  // No mutes.
			true, // Placehold unknown attachments.

			// Don't add note about
			// pending, it's not
			// relevant here.
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("ReportToAdminAPIReport: error converting status with id %s to api status: %w", s.ID, err)
		}
		statuses = append(statuses, status)
	}

	rules := make([]*apimodel.InstanceRule, 0, len(r.RuleIDs))
	if len(r.RuleIDs) != 0 && len(r.Rules) == 0 {
		r.Rules, err = c.state.DB.GetRulesByIDs(ctx, r.RuleIDs)
		if err != nil {
			return nil, fmt.Errorf("ReportToAdminAPIReport: error getting rules from the db: %w", err)
		}
	}
	for _, v := range r.Rules {
		rules = append(rules, &apimodel.InstanceRule{
			ID:   v.ID,
			Text: v.Text,
		})
	}

	if ac := r.ActionTaken; ac != "" {
		actionTakenComment = &ac
	}

	return &apimodel.AdminReport{
		ID:                   r.ID,
		ActionTaken:          !r.ActionTakenAt.IsZero(),
		ActionTakenAt:        actionTakenAt,
		Category:             "other", // todo: only support default 'other' category right now
		Comment:              r.Comment,
		Forwarded:            *r.Forwarded,
		CreatedAt:            util.FormatISO8601(r.CreatedAt),
		UpdatedAt:            util.FormatISO8601(r.UpdatedAt),
		Account:              account,
		TargetAccount:        targetAccount,
		AssignedAccount:      actionTakenByAccount,
		ActionTakenByAccount: actionTakenByAccount,
		ActionTakenComment:   actionTakenComment,
		Statuses:             statuses,
		Rules:                rules,
	}, nil
}

// ListToAPIList converts one gts model list into an api model list, for serving at /api/v1/lists/{id}
func (c *Converter) ListToAPIList(ctx context.Context, l *gtsmodel.List) (*apimodel.List, error) {
	return &apimodel.List{
		ID:            l.ID,
		Title:         l.Title,
		RepliesPolicy: string(l.RepliesPolicy),
		Exclusive:     *l.Exclusive,
	}, nil
}

// MarkersToAPIMarker converts several gts model markers into an api marker, for serving at /api/v1/markers
func (c *Converter) MarkersToAPIMarker(ctx context.Context, markers []*gtsmodel.Marker) (*apimodel.Marker, error) {
	apiMarker := &apimodel.Marker{}
	for _, marker := range markers {
		apiTimelineMarker := &apimodel.TimelineMarker{
			LastReadID: marker.LastReadID,
			UpdatedAt:  util.FormatISO8601(marker.UpdatedAt),
			Version:    marker.Version,
		}
		switch apimodel.MarkerName(marker.Name) {
		case apimodel.MarkerNameHome:
			apiMarker.Home = apiTimelineMarker
		case apimodel.MarkerNameNotifications:
			apiMarker.Notifications = apiTimelineMarker
		default:
			return nil, fmt.Errorf("unknown marker timeline name: %s", marker.Name)
		}
	}
	return apiMarker, nil
}

// PollToAPIPoll converts a database (gtsmodel) Poll into an API model representation appropriate for the given requesting account.
func (c *Converter) PollToAPIPoll(ctx context.Context, requester *gtsmodel.Account, poll *gtsmodel.Poll) (*apimodel.Poll, error) {
	// Ensure the poll model is fully populated for src status.
	if err := c.state.DB.PopulatePoll(ctx, poll); err != nil {
		return nil, gtserror.Newf("error populating poll: %w", err)
	}

	var (
		options     []apimodel.PollOption
		totalVotes  int
		totalVoters *int
		hasVoted    *bool
		ownChoices  *[]int
		isAuthor    bool
		expiresAt   *string
		emojis      []apimodel.Emoji
	)

	// Preallocate a slice of frontend model poll choices.
	options = make([]apimodel.PollOption, len(poll.Options))

	// Add the titles to all of the options.
	for i, title := range poll.Options {
		options[i].Title = title
	}

	if requester != nil {
		// Get vote by requester in poll (if any).
		vote, err := c.state.DB.GetPollVoteBy(ctx,
			poll.ID,
			requester.ID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf("error getting vote for poll %s: %w", poll.ID, err)
		}

		if vote != nil {
			// Set choices by requester.
			ownChoices = &vote.Choices

			// Update default total in the
			// case that counts are hidden
			// (so we just show our own).
			totalVotes = len(vote.Choices)
		} else {
			// Requester hasn't yet voted, use
			// empty slice to serialize as `[]`.
			ownChoices = &[]int{}
		}

		// Check if requester is author of source status.
		isAuthor = (requester.ID == poll.Status.AccountID)

		// Set whether requester has voted in poll (or = author).
		hasVoted = util.Ptr((isAuthor || len(*ownChoices) > 0))
	}

	if isAuthor || !*poll.HideCounts {
		// Only in the case that hide counts is
		// disabled, or the requester is the author
		// do we actually populate the vote counts.

		// If we voted in this poll, we'll have set totalVotes
		// earlier. Reset here to avoid double counting.
		totalVotes = 0
		if *poll.Multiple {
			// The total number of voters are only
			// provided in the case of a multiple
			// choice poll. All else leaves it nil.
			totalVoters = poll.Voters
		}

		// Populate per-vote counts
		// and overall total vote count.
		for i, count := range poll.Votes {
			if options[i].VotesCount == nil {
				options[i].VotesCount = new(int)
			}
			(*options[i].VotesCount) += count
			totalVotes += count
		}
	}

	if !poll.ExpiresAt.IsZero() {
		// Calculate poll expiry string (if set).
		str := util.FormatISO8601(poll.ExpiresAt)
		expiresAt = &str
	}

	var err error

	// Try to inherit emojis from parent status.
	emojis, err = c.convertEmojisToAPIEmojis(ctx,
		poll.Status.Emojis,
		poll.Status.EmojiIDs,
	)
	if err != nil {
		log.Errorf(ctx, "error converting emojis from parent status: %v", err)
		emojis = []apimodel.Emoji{} // fallback to empty slice.
	}

	return &apimodel.Poll{
		ID:          poll.ID,
		ExpiresAt:   expiresAt,
		Expired:     poll.Closed(),
		Multiple:    (*poll.Multiple),
		VotesCount:  totalVotes,
		VotersCount: totalVoters,
		Voted:       hasVoted,
		OwnVotes:    ownChoices,
		Options:     options,
		Emojis:      emojis,
	}, nil
}

// convertAttachmentsToAPIAttachments will convert a slice of GTS model attachments to frontend API model attachments, falling back to IDs if no GTS models supplied.
func (c *Converter) convertAttachmentsToAPIAttachments(ctx context.Context, attachments []*gtsmodel.MediaAttachment, attachmentIDs []string) ([]*apimodel.Attachment, error) {
	var errs gtserror.MultiError

	if len(attachments) == 0 && len(attachmentIDs) > 0 {
		// GTS model attachments were not populated

		var err error

		// Fetch GTS models for attachment IDs
		attachments, err = c.state.DB.GetAttachmentsByIDs(ctx, attachmentIDs)
		if err != nil {
			errs.Appendf("error fetching attachments from database: %w", err)
		}
	}

	// Preallocate expected frontend slice
	apiAttachments := make([]*apimodel.Attachment, 0, len(attachments))

	// Convert GTS models to frontend models
	for _, attachment := range attachments {
		apiAttachment, err := c.AttachmentToAPIAttachment(ctx, attachment)
		if err != nil {
			errs.Appendf("error converting attchment %s to api attachment: %w", attachment.ID, err)
			continue
		}
		apiAttachments = append(apiAttachments, &apiAttachment)
	}

	return apiAttachments, errs.Combine()
}

// FilterToAPIFiltersV1 converts one GTS model filter into an API v1 filter list
func (c *Converter) FilterToAPIFiltersV1(ctx context.Context, filter *gtsmodel.Filter) ([]*apimodel.FilterV1, error) {
	apiFilters := make([]*apimodel.FilterV1, 0, len(filter.Keywords))
	for _, filterKeyword := range filter.Keywords {
		apiFilter, err := c.FilterKeywordToAPIFilterV1(ctx, filterKeyword)
		if err != nil {
			return nil, err
		}
		apiFilters = append(apiFilters, apiFilter)
	}
	return apiFilters, nil
}

// FilterKeywordToAPIFilterV1 converts one GTS model filter and filter keyword into an API v1 filter
func (c *Converter) FilterKeywordToAPIFilterV1(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) (*apimodel.FilterV1, error) {
	if filterKeyword.Filter == nil {
		return nil, gtserror.New("FilterKeyword model's Filter field isn't populated, but needs to be")
	}
	filter := filterKeyword.Filter

	return &apimodel.FilterV1{
		// v1 filters have a single keyword each, so we use the filter keyword ID as the v1 filter ID.
		ID:           filterKeyword.ID,
		Phrase:       filterKeyword.Keyword,
		Context:      filterToAPIFilterContexts(filter),
		WholeWord:    util.PtrOrValue(filterKeyword.WholeWord, false),
		ExpiresAt:    filterExpiresAtToAPIFilterExpiresAt(filter.ExpiresAt),
		Irreversible: filter.Action == gtsmodel.FilterActionHide,
	}, nil
}

// FilterToAPIFilterV2 converts one GTS model filter into an API v2 filter.
func (c *Converter) FilterToAPIFilterV2(ctx context.Context, filter *gtsmodel.Filter) (*apimodel.FilterV2, error) {
	apiFilterKeywords := make([]apimodel.FilterKeyword, 0, len(filter.Keywords))
	for _, filterKeyword := range filter.Keywords {
		apiFilterKeywords = append(apiFilterKeywords, *c.FilterKeywordToAPIFilterKeyword(ctx, filterKeyword))
	}

	apiFilterStatuses := make([]apimodel.FilterStatus, 0, len(filter.Keywords))
	for _, filterStatus := range filter.Statuses {
		apiFilterStatuses = append(apiFilterStatuses, *c.FilterStatusToAPIFilterStatus(ctx, filterStatus))
	}

	return &apimodel.FilterV2{
		ID:           filter.ID,
		Title:        filter.Title,
		Context:      filterToAPIFilterContexts(filter),
		ExpiresAt:    filterExpiresAtToAPIFilterExpiresAt(filter.ExpiresAt),
		FilterAction: filterActionToAPIFilterAction(filter.Action),
		Keywords:     apiFilterKeywords,
		Statuses:     apiFilterStatuses,
	}, nil
}

func filterExpiresAtToAPIFilterExpiresAt(expiresAt time.Time) *string {
	if expiresAt.IsZero() {
		return nil
	}
	return util.Ptr(util.FormatISO8601(expiresAt))
}

func filterToAPIFilterContexts(filter *gtsmodel.Filter) []apimodel.FilterContext {
	apiContexts := make([]apimodel.FilterContext, 0, apimodel.FilterContextNumValues)
	if util.PtrOrValue(filter.ContextHome, false) {
		apiContexts = append(apiContexts, apimodel.FilterContextHome)
	}
	if util.PtrOrValue(filter.ContextNotifications, false) {
		apiContexts = append(apiContexts, apimodel.FilterContextNotifications)
	}
	if util.PtrOrValue(filter.ContextPublic, false) {
		apiContexts = append(apiContexts, apimodel.FilterContextPublic)
	}
	if util.PtrOrValue(filter.ContextThread, false) {
		apiContexts = append(apiContexts, apimodel.FilterContextThread)
	}
	if util.PtrOrValue(filter.ContextAccount, false) {
		apiContexts = append(apiContexts, apimodel.FilterContextAccount)
	}
	return apiContexts
}

func filterActionToAPIFilterAction(m gtsmodel.FilterAction) apimodel.FilterAction {
	switch m {
	case gtsmodel.FilterActionWarn:
		return apimodel.FilterActionWarn
	case gtsmodel.FilterActionHide:
		return apimodel.FilterActionHide
	}
	return apimodel.FilterActionNone
}

// FilterKeywordToAPIFilterKeyword converts a GTS model filter status into an API filter status.
func (c *Converter) FilterKeywordToAPIFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) *apimodel.FilterKeyword {
	return &apimodel.FilterKeyword{
		ID:        filterKeyword.ID,
		Keyword:   filterKeyword.Keyword,
		WholeWord: util.PtrOrValue(filterKeyword.WholeWord, false),
	}
}

// FilterStatusToAPIFilterStatus converts a GTS model filter status into an API filter status.
func (c *Converter) FilterStatusToAPIFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus) *apimodel.FilterStatus {
	return &apimodel.FilterStatus{
		ID:       filterStatus.ID,
		StatusID: filterStatus.StatusID,
	}
}

// convertEmojisToAPIEmojis will convert a slice of GTS model emojis to frontend API model emojis, falling back to IDs if no GTS models supplied.
func (c *Converter) convertEmojisToAPIEmojis(ctx context.Context, emojis []*gtsmodel.Emoji, emojiIDs []string) ([]apimodel.Emoji, error) {
	var errs gtserror.MultiError

	if len(emojis) == 0 && len(emojiIDs) > 0 {
		// GTS model attachments were not populated

		var err error

		// Fetch GTS models for emoji IDs
		emojis, err = c.state.DB.GetEmojisByIDs(ctx, emojiIDs)
		if err != nil {
			errs.Appendf("error fetching emojis from database: %w", err)
		}
	}

	// Preallocate expected frontend slice
	apiEmojis := make([]apimodel.Emoji, 0, len(emojis))

	// Convert GTS models to frontend models
	for _, emoji := range emojis {
		apiEmoji, err := c.EmojiToAPIEmoji(ctx, emoji)
		if err != nil {
			errs.Appendf("error converting emoji %s to api emoji: %w", emoji.ID, err)
			continue
		}
		apiEmojis = append(apiEmojis, apiEmoji)
	}

	return apiEmojis, errs.Combine()
}

// convertMentionsToAPIMentions will convert a slice of GTS model mentions to frontend API model mentions, falling back to IDs if no GTS models supplied.
func (c *Converter) convertMentionsToAPIMentions(ctx context.Context, mentions []*gtsmodel.Mention, mentionIDs []string) ([]apimodel.Mention, error) {
	var errs gtserror.MultiError

	if len(mentions) == 0 && len(mentionIDs) > 0 {
		var err error

		// GTS model mentions were not populated
		//
		// Fetch GTS models for mention IDs
		mentions, err = c.state.DB.GetMentions(ctx, mentionIDs)
		if err != nil {
			errs.Appendf("error fetching mentions from database: %w", err)
		}
	}

	// Preallocate expected frontend slice
	apiMentions := make([]apimodel.Mention, 0, len(mentions))

	// Convert GTS models to frontend models
	for _, mention := range mentions {
		apiMention, err := c.MentionToAPIMention(ctx, mention)
		if err != nil {
			errs.Appendf("error converting mention %s to api mention: %w", mention.ID, err)
			continue
		}
		apiMentions = append(apiMentions, apiMention)
	}

	return apiMentions, errs.Combine()
}

// convertTagsToAPITags will convert a slice of GTS model tags to frontend API model tags, falling back to IDs if no GTS models supplied.
func (c *Converter) convertTagsToAPITags(ctx context.Context, tags []*gtsmodel.Tag, tagIDs []string) ([]apimodel.Tag, error) {
	var errs gtserror.MultiError

	if len(tags) == 0 && len(tagIDs) > 0 {
		var err error

		tags, err = c.state.DB.GetTags(ctx, tagIDs)
		if err != nil {
			errs.Appendf("error fetching tags from database: %w", err)
		}
	}

	// Preallocate expected frontend slice
	apiTags := make([]apimodel.Tag, 0, len(tags))

	// Convert GTS models to frontend models
	for _, tag := range tags {
		apiTag, err := c.TagToAPITag(ctx, tag, false, nil)
		if err != nil {
			errs.Appendf("error converting tag %s to api tag: %w", tag.ID, err)
			continue
		}
		apiTags = append(apiTags, apiTag)
	}

	return apiTags, errs.Combine()
}

// ThemesToAPIThemes converts a slice of gtsmodel Themes into apimodel Themes.
func (c *Converter) ThemesToAPIThemes(themes []*gtsmodel.Theme) []apimodel.Theme {
	apiThemes := make([]apimodel.Theme, len(themes))
	for i, theme := range themes {
		apiThemes[i] = apimodel.Theme{
			Title:       theme.Title,
			Description: theme.Description,
			FileName:    theme.FileName,
		}
	}
	return apiThemes
}

// Convert the given gtsmodel policy
// into an apimodel interaction policy.
//
// Provided status can be nil to convert a
// policy without a particular status in mind.
//
// RequestingAccount can also be nil for
// unauthorized requests (web, public api etc).
func (c *Converter) InteractionPolicyToAPIInteractionPolicy(
	ctx context.Context,
	policy *gtsmodel.InteractionPolicy,
	status *gtsmodel.Status,
	requester *gtsmodel.Account,
) (*apimodel.InteractionPolicy, error) {
	apiPolicy := &apimodel.InteractionPolicy{
		CanFavourite: apimodel.PolicyRules{
			Always:       policyValsToAPIPolicyVals(policy.CanLike.Always),
			WithApproval: policyValsToAPIPolicyVals(policy.CanLike.WithApproval),
		},
		CanReply: apimodel.PolicyRules{
			Always:       policyValsToAPIPolicyVals(policy.CanReply.Always),
			WithApproval: policyValsToAPIPolicyVals(policy.CanReply.WithApproval),
		},
		CanReblog: apimodel.PolicyRules{
			Always:       policyValsToAPIPolicyVals(policy.CanAnnounce.Always),
			WithApproval: policyValsToAPIPolicyVals(policy.CanAnnounce.WithApproval),
		},
	}

	if status == nil || requester == nil {
		// We're done here!
		return apiPolicy, nil
	}

	// Status and requester are both defined,
	// so we can add the "me" Value to the policy
	// for each interaction type, if applicable.

	likeable, err := c.intFilter.StatusLikeable(ctx, requester, status)
	if err != nil {
		err := gtserror.Newf("error checking status likeable by requester: %w", err)
		return nil, err
	}

	if likeable.Permission == gtsmodel.PolicyPermissionPermitted {
		// We can do this!
		apiPolicy.CanFavourite.Always = append(
			apiPolicy.CanFavourite.Always,
			apimodel.PolicyValueMe,
		)
	} else if likeable.Permission == gtsmodel.PolicyPermissionWithApproval {
		// We can do this with approval.
		apiPolicy.CanFavourite.WithApproval = append(
			apiPolicy.CanFavourite.WithApproval,
			apimodel.PolicyValueMe,
		)
	}

	replyable, err := c.intFilter.StatusReplyable(ctx, requester, status)
	if err != nil {
		err := gtserror.Newf("error checking status replyable by requester: %w", err)
		return nil, err
	}

	if replyable.Permission == gtsmodel.PolicyPermissionPermitted {
		// We can do this!
		apiPolicy.CanReply.Always = append(
			apiPolicy.CanReply.Always,
			apimodel.PolicyValueMe,
		)
	} else if replyable.Permission == gtsmodel.PolicyPermissionWithApproval {
		// We can do this with approval.
		apiPolicy.CanReply.WithApproval = append(
			apiPolicy.CanReply.WithApproval,
			apimodel.PolicyValueMe,
		)
	}

	boostable, err := c.intFilter.StatusBoostable(ctx, requester, status)
	if err != nil {
		err := gtserror.Newf("error checking status boostable by requester: %w", err)
		return nil, err
	}

	if boostable.Permission == gtsmodel.PolicyPermissionPermitted {
		// We can do this!
		apiPolicy.CanReblog.Always = append(
			apiPolicy.CanReblog.Always,
			apimodel.PolicyValueMe,
		)
	} else if boostable.Permission == gtsmodel.PolicyPermissionWithApproval {
		// We can do this with approval.
		apiPolicy.CanReblog.WithApproval = append(
			apiPolicy.CanReblog.WithApproval,
			apimodel.PolicyValueMe,
		)
	}

	return apiPolicy, nil
}

func policyValsToAPIPolicyVals(vals gtsmodel.PolicyValues) []apimodel.PolicyValue {

	var (
		valsLen = len(vals)

		// Use a map to deduplicate added vals as we go.
		addedVals = make(map[apimodel.PolicyValue]struct{}, valsLen)

		// Vals we'll be returning.
		apiVals = make([]apimodel.PolicyValue, 0, valsLen)
	)

	for _, policyVal := range vals {
		switch policyVal {

		case gtsmodel.PolicyValueAuthor:
			// Author can do this.
			newVal := apimodel.PolicyValueAuthor
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}

		case gtsmodel.PolicyValueMentioned:
			// Mentioned can do this.
			newVal := apimodel.PolicyValueMentioned
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}

		case gtsmodel.PolicyValueMutuals:
			// Mutuals can do this.
			newVal := apimodel.PolicyValueMutuals
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}

		case gtsmodel.PolicyValueFollowing:
			// Following can do this.
			newVal := apimodel.PolicyValueFollowing
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}

		case gtsmodel.PolicyValueFollowers:
			// Followers can do this.
			newVal := apimodel.PolicyValueFollowers
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}

		case gtsmodel.PolicyValuePublic:
			// Public can do this.
			newVal := apimodel.PolicyValuePublic
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}

		default:
			// Specific URI of ActivityPub Actor.
			newVal := apimodel.PolicyValue(policyVal)
			if _, added := addedVals[newVal]; !added {
				apiVals = append(apiVals, newVal)
				addedVals[newVal] = struct{}{}
			}
		}
	}

	return apiVals
}

// InteractionReqToAPIInteractionReq converts the given *gtsmodel.InteractionRequest
// to an *apimodel.InteractionRequest, from the perspective of requestingAcct.
func (c *Converter) InteractionReqToAPIInteractionReq(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
	requestingAcct *gtsmodel.Account,
) (*apimodel.InteractionRequest, error) {
	// Ensure interaction request is populated.
	if err := c.state.DB.PopulateInteractionRequest(ctx, req); err != nil {
		err := gtserror.Newf("error populating: %w", err)
		return nil, err
	}

	interactingAcct, err := c.AccountToAPIAccountPublic(ctx, req.InteractingAccount)
	if err != nil {
		err := gtserror.Newf("error converting interacting acct: %w", err)
		return nil, err
	}

	interactedStatus, err := c.StatusToAPIStatus(
		ctx,
		req.Status,
		requestingAcct,
		statusfilter.FilterContextNone,
		nil, // No filters.
		nil, // No mutes.
	)
	if err != nil {
		err := gtserror.Newf("error converting interacted status: %w", err)
		return nil, err
	}

	var reply *apimodel.Status
	if req.InteractionType == gtsmodel.InteractionReply && req.Reply != nil {
		reply, err = c.statusToAPIStatus(
			ctx,
			req.Reply,
			requestingAcct,
			statusfilter.FilterContextNone,
			nil,  // No filters.
			nil,  // No mutes.
			true, // Placehold unknown attachments.

			// Don't add note about pending;
			// requester already knows it's
			// pending because they're looking
			// at the request right now.
			false,
		)
		if err != nil {
			err := gtserror.Newf("error converting reply: %w", err)
			return nil, err
		}
	}

	var acceptedAt string
	if req.IsAccepted() {
		acceptedAt = util.FormatISO8601(req.AcceptedAt)
	}

	var rejectedAt string
	if req.IsRejected() {
		rejectedAt = util.FormatISO8601(req.RejectedAt)
	}

	return &apimodel.InteractionRequest{
		ID:         req.ID,
		Type:       req.InteractionType.String(),
		CreatedAt:  util.FormatISO8601(req.CreatedAt),
		Account:    interactingAcct,
		Status:     interactedStatus,
		Reply:      reply,
		AcceptedAt: acceptedAt,
		RejectedAt: rejectedAt,
		URI:        req.URI,
	}, nil
}
