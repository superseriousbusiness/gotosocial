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
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

const (
	instanceStatusesCharactersReservedPerURL    = 25
	instanceMediaAttachmentsImageMatrixLimit    = 16777216 // width * height
	instanceMediaAttachmentsVideoMatrixLimit    = 16777216 // width * height
	instanceMediaAttachmentsVideoFrameRateLimit = 60
	instancePollsMinExpiration                  = 300     // seconds
	instancePollsMaxExpiration                  = 2629746 // seconds
	instanceAccountsMaxFeaturedTags             = 10
	instanceAccountsMaxProfileFields            = 6 // FIXME: https://github.com/superseriousbusiness/gotosocial/issues/1876
	instanceSourceURL                           = "https://github.com/superseriousbusiness/gotosocial"
	instanceMastodonVersion                     = "3.5.3"
)

var instanceStatusesSupportedMimeTypes = []string{
	string(apimodel.StatusContentTypePlain),
	string(apimodel.StatusContentTypeMarkdown),
}

func toMastodonVersion(in string) string {
	return instanceMastodonVersion + "+" + strings.ReplaceAll(in, " ", "-")
}

// AppToAPIAppSensitive takes a db model application as a param, and returns a populated apitype application, or an error
// if something goes wrong. The returned application should be ready to serialize on an API level, and may have sensitive fields
// (such as client id and client secret), so serve it only to an authorized user who should have permission to see it.
func (c *Converter) AccountToAPIAccountSensitive(ctx context.Context, a *gtsmodel.Account) (*apimodel.Account, error) {
	// we can build this sensitive account easily by first getting the public account....
	apiAccount, err := c.AccountToAPIAccountPublic(ctx, a)
	if err != nil {
		return nil, err
	}

	// then adding the Source object to it...

	// check pending follow requests aimed at this account
	frc, err := c.state.DB.CountAccountFollowRequests(ctx, a.ID)
	if err != nil {
		return nil, fmt.Errorf("error counting follow requests: %s", err)
	}

	statusContentType := string(apimodel.StatusContentTypeDefault)
	if a.StatusContentType != "" {
		statusContentType = a.StatusContentType
	}

	apiAccount.Source = &apimodel.Source{
		Privacy:             c.VisToAPIVis(ctx, a.Privacy),
		Sensitive:           *a.Sensitive,
		Language:            a.Language,
		StatusContentType:   statusContentType,
		Note:                a.NoteRaw,
		Fields:              c.fieldsToAPIFields(a.FieldsRaw),
		FollowRequestsCount: frc,
	}

	return apiAccount, nil
}

// AccountToAPIAccountPublic takes a db model account as a param, and returns a populated apitype account, or an error
// if something goes wrong. The returned account should be ready to serialize on an API level, and may NOT have sensitive fields.
// In other words, this is the public record that the server has of an account.
func (c *Converter) AccountToAPIAccountPublic(ctx context.Context, a *gtsmodel.Account) (*apimodel.Account, error) {
	if err := c.state.DB.PopulateAccount(ctx, a); err != nil {
		log.Errorf(ctx, "error(s) populating account, will continue: %s", err)
	}

	// Basic account stats:
	//   - Followers count
	//   - Following count
	//   - Statuses count
	//   - Last status time

	followersCount, err := c.state.DB.CountAccountFollowers(ctx, a.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("AccountToAPIAccountPublic: error counting followers: %w", err)
	}

	followingCount, err := c.state.DB.CountAccountFollows(ctx, a.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("AccountToAPIAccountPublic: error counting following: %w", err)
	}

	statusesCount, err := c.state.DB.CountAccountStatuses(ctx, a.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("AccountToAPIAccountPublic: error counting statuses: %w", err)
	}

	var lastStatusAt *string
	lastPosted, err := c.state.DB.GetAccountLastPosted(ctx, a.ID, false)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("AccountToAPIAccountPublic: error counting statuses: %w", err)
	}

	if !lastPosted.IsZero() {
		lastStatusAt = func() *string { t := util.FormatISO8601(lastPosted); return &t }()
	}

	// Profile media + nice extras:
	//   - Avatar
	//   - Header
	//   - Fields
	//   - Emojis

	var (
		aviURL          string
		aviURLStatic    string
		headerURL       string
		headerURLStatic string
	)

	if a.AvatarMediaAttachment != nil {
		aviURL = a.AvatarMediaAttachment.URL
		aviURLStatic = a.AvatarMediaAttachment.Thumbnail.URL
	}

	if a.HeaderMediaAttachment != nil {
		headerURL = a.HeaderMediaAttachment.URL
		headerURLStatic = a.HeaderMediaAttachment.Thumbnail.URL
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

	var (
		acct string
		role *apimodel.AccountRole
	)

	if a.IsRemote() {
		// Domain may be in Punycode,
		// de-punify it just in case.
		d, err := util.DePunify(a.Domain)
		if err != nil {
			return nil, fmt.Errorf("AccountToAPIAccountPublic: error de-punifying domain %s for account id %s: %w", a.Domain, a.ID, err)
		}

		acct = a.Username + "@" + d
	} else {
		// This is a local account, try to
		// fetch more info. Skip for instance
		// accounts since they have no user.
		if !a.IsInstance() {
			user, err := c.state.DB.GetUserByAccountID(ctx, a.ID)
			if err != nil {
				return nil, fmt.Errorf("AccountToAPIAccountPublic: error getting user from database for account id %s: %w", a.ID, err)
			}

			switch {
			case *user.Admin:
				role = &apimodel.AccountRole{Name: apimodel.AccountRoleAdmin}
			case *user.Moderator:
				role = &apimodel.AccountRole{Name: apimodel.AccountRoleModerator}
			default:
				role = &apimodel.AccountRole{Name: apimodel.AccountRoleUser}
			}
		}

		acct = a.Username // omit domain
	}

	// Remaining properties are simple and
	// can be populated directly below.

	accountFrontend := &apimodel.Account{
		ID:             a.ID,
		Username:       a.Username,
		Acct:           acct,
		DisplayName:    a.DisplayName,
		Locked:         *a.Locked,
		Discoverable:   *a.Discoverable,
		Bot:            *a.Bot,
		CreatedAt:      util.FormatISO8601(a.CreatedAt),
		Note:           a.Note,
		URL:            a.URL,
		Avatar:         aviURL,
		AvatarStatic:   aviURLStatic,
		Header:         headerURL,
		HeaderStatic:   headerURLStatic,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		StatusesCount:  statusesCount,
		LastStatusAt:   lastStatusAt,
		Emojis:         apiEmojis,
		Fields:         fields,
		Suspended:      !a.SuspendedAt.IsZero(),
		CustomCSS:      a.CustomCSS,
		EnableRSS:      *a.EnableRSS,
		Role:           role,
	}

	// Bodge default avatar + header in,
	// if we didn't have one already.
	c.ensureAvatar(accountFrontend)
	c.ensureHeader(accountFrontend)

	return accountFrontend, nil
}

func (c *Converter) fieldsToAPIFields(f []*gtsmodel.Field) []apimodel.Field {
	fields := make([]apimodel.Field, len(f))

	for i, field := range f {
		mField := apimodel.Field{
			Name:  field.Name,
			Value: field.Value,
		}

		if !field.VerifiedAt.IsZero() {
			mField.VerifiedAt = func() *string { s := util.FormatISO8601(field.VerifiedAt); return &s }()
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
		acct string
		role *apimodel.AccountRole
	)

	if a.IsRemote() {
		// Domain may be in Punycode,
		// de-punify it just in case.
		d, err := util.DePunify(a.Domain)
		if err != nil {
			return nil, fmt.Errorf("AccountToAPIAccountBlocked: error de-punifying domain %s for account id %s: %w", a.Domain, a.ID, err)
		}

		acct = a.Username + "@" + d
	} else {
		// This is a local account, try to
		// fetch more info. Skip for instance
		// accounts since they have no user.
		if !a.IsInstance() {
			user, err := c.state.DB.GetUserByAccountID(ctx, a.ID)
			if err != nil {
				return nil, fmt.Errorf("AccountToAPIAccountPublic: error getting user from database for account id %s: %w", a.ID, err)
			}

			switch {
			case *user.Admin:
				role = &apimodel.AccountRole{Name: apimodel.AccountRoleAdmin}
			case *user.Moderator:
				role = &apimodel.AccountRole{Name: apimodel.AccountRoleModerator}
			default:
				role = &apimodel.AccountRole{Name: apimodel.AccountRoleUser}
			}
		}

		acct = a.Username // omit domain
	}

	return &apimodel.Account{
		ID:          a.ID,
		Username:    a.Username,
		Acct:        acct,
		DisplayName: a.DisplayName,
		Bot:         *a.Bot,
		CreatedAt:   util.FormatISO8601(a.CreatedAt),
		URL:         a.URL,
		Suspended:   !a.SuspendedAt.IsZero(),
		Role:        role,
	}, nil
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
		role                   = apimodel.AccountRole{Name: apimodel.AccountRoleUser} // assume user by default
		createdByApplicationID string
	)

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

		if i := user.CurrentSignInIP.String(); i != "<nil>" {
			ip = &i
		}

		locale = user.Locale
		if user.Account.Reason != "" {
			inviteRequest = &user.Account.Reason
		}

		if *user.Admin {
			role.Name = apimodel.AccountRoleAdmin
		} else if *user.Moderator {
			role.Name = apimodel.AccountRoleModerator
		}

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
func (c *Converter) AttachmentToAPIAttachment(ctx context.Context, a *gtsmodel.MediaAttachment) (apimodel.Attachment, error) {
	apiAttachment := apimodel.Attachment{
		ID:         a.ID,
		Type:       strings.ToLower(string(a.Type)),
		TextURL:    a.URL,
		PreviewURL: a.Thumbnail.URL,
		Meta: apimodel.MediaMeta{
			Original: apimodel.MediaDimensions{
				Width:  a.FileMeta.Original.Width,
				Height: a.FileMeta.Original.Height,
			},
			Small: apimodel.MediaDimensions{
				Width:  a.FileMeta.Small.Width,
				Height: a.FileMeta.Small.Height,
				Size:   strconv.Itoa(a.FileMeta.Small.Width) + "x" + strconv.Itoa(a.FileMeta.Small.Height),
				Aspect: float32(a.FileMeta.Small.Aspect),
			},
		},
		Blurhash: a.Blurhash,
	}

	// nullable fields
	if i := a.URL; i != "" {
		apiAttachment.URL = &i
	}

	if i := a.RemoteURL; i != "" {
		apiAttachment.RemoteURL = &i
	}

	if i := a.Thumbnail.RemoteURL; i != "" {
		apiAttachment.PreviewRemoteURL = &i
	}

	if i := a.Description; i != "" {
		apiAttachment.Description = &i
	}

	// type specific fields
	switch a.Type {
	case gtsmodel.FileTypeImage:
		apiAttachment.Meta.Original.Size = strconv.Itoa(a.FileMeta.Original.Width) + "x" + strconv.Itoa(a.FileMeta.Original.Height)
		apiAttachment.Meta.Original.Aspect = float32(a.FileMeta.Original.Aspect)
		apiAttachment.Meta.Focus = &apimodel.MediaFocus{
			X: a.FileMeta.Focus.X,
			Y: a.FileMeta.Focus.Y,
		}
	case gtsmodel.FileTypeVideo:
		if i := a.FileMeta.Original.Duration; i != nil {
			apiAttachment.Meta.Original.Duration = *i
		}

		if i := a.FileMeta.Original.Framerate; i != nil {
			// the masto api expects this as a string in
			// the format `integer/1`, so 30fps is `30/1`
			round := math.Round(float64(*i))
			fr := strconv.FormatInt(int64(round), 10)
			apiAttachment.Meta.Original.FrameRate = fr + "/1"
		}

		if i := a.FileMeta.Original.Bitrate; i != nil {
			apiAttachment.Meta.Original.Bitrate = int(*i)
		}
	}

	return apiAttachment, nil
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

	if e.Domain != "" {
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
func (c *Converter) TagToAPITag(ctx context.Context, t *gtsmodel.Tag, stubHistory bool) (apimodel.Tag, error) {
	return apimodel.Tag{
		Name: strings.ToLower(t.Name),
		URL:  uris.GenerateURIForTag(t.Name),
		History: func() *[]any {
			if !stubHistory {
				return nil
			}

			h := make([]any, 0)
			return &h
		}(),
	}, nil
}

// StatusToAPIStatus converts a gts model status into its api (frontend) representation for serialization on the API.
//
// Requesting account can be nil.
func (c *Converter) StatusToAPIStatus(ctx context.Context, s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*apimodel.Status, error) {
	if err := c.state.DB.PopulateStatus(ctx, s); err != nil {
		// Ensure author account present + correct;
		// can't really go further without this!
		if s.Account == nil {
			return nil, fmt.Errorf("error(s) populating status, cannot continue: %w", err)
		}

		log.Errorf(ctx, "error(s) populating status, will continue: %v", err)
	}

	apiAuthorAccount, err := c.AccountToAPIAccountPublic(ctx, s.Account)
	if err != nil {
		return nil, fmt.Errorf("error converting status author: %w", err)
	}

	repliesCount, err := c.state.DB.CountStatusReplies(ctx, s.ID)
	if err != nil {
		return nil, fmt.Errorf("error counting replies: %w", err)
	}

	reblogsCount, err := c.state.DB.CountStatusBoosts(ctx, s.ID)
	if err != nil {
		return nil, fmt.Errorf("error counting reblogs: %w", err)
	}

	favesCount, err := c.state.DB.CountStatusFaves(ctx, s.ID)
	if err != nil {
		return nil, fmt.Errorf("error counting faves: %w", err)
	}

	interacts, err := c.interactionsWithStatusForAccount(ctx, s, requestingAccount)
	if err != nil {
		log.Errorf(ctx, "error getting interactions for status %s for account %s: %v", s.ID, requestingAccount.ID, err)

		// Ensure a non nil object
		interacts = &statusInteractions{}
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

	apiStatus := &apimodel.Status{
		ID:                 s.ID,
		CreatedAt:          util.FormatISO8601(s.CreatedAt),
		InReplyToID:        nil,
		InReplyToAccountID: nil,
		Sensitive:          *s.Sensitive,
		SpoilerText:        s.ContentWarning,
		Visibility:         c.VisToAPIVis(ctx, s.Visibility),
		Language:           nil,
		URI:                s.URI,
		URL:                s.URL,
		RepliesCount:       repliesCount,
		ReblogsCount:       reblogsCount,
		FavouritesCount:    favesCount,
		Favourited:         interacts.Faved,
		Bookmarked:         interacts.Bookmarked,
		Muted:              interacts.Muted,
		Reblogged:          interacts.Reblogged,
		Pinned:             interacts.Pinned,
		Content:            s.Content,
		Reblog:             nil,
		Application:        nil,
		Account:            apiAuthorAccount,
		MediaAttachments:   apiAttachments,
		Mentions:           apiMentions,
		Tags:               apiTags,
		Emojis:             apiEmojis,
		Card:               nil, // TODO: implement cards
		Poll:               nil, // TODO: implement polls
		Text:               s.Text,
	}

	// Nullable fields.

	if s.InReplyToID != "" {
		apiStatus.InReplyToID = func() *string { i := s.InReplyToID; return &i }()
	}

	if s.InReplyToAccountID != "" {
		apiStatus.InReplyToAccountID = func() *string { i := s.InReplyToAccountID; return &i }()
	}

	if s.Language != "" {
		apiStatus.Language = func() *string { i := s.Language; return &i }()
	}

	if s.BoostOf != nil {
		apiBoostOf, err := c.StatusToAPIStatus(ctx, s.BoostOf, requestingAccount)
		if err != nil {
			return nil, fmt.Errorf("error converting boosted status: %w", err)
		}

		apiStatus.Reblog = &apimodel.StatusReblogged{Status: apiBoostOf}
	}

	if appID := s.CreatedWithApplicationID; appID != "" {
		app, err := c.state.DB.GetApplicationByID(ctx, appID)
		if err != nil {
			return nil, fmt.Errorf("error getting application %s: %w", appID, err)
		}

		apiApp, err := c.AppToAPIAppPublic(ctx, app)
		if err != nil {
			return nil, fmt.Errorf("error converting application %s: %w", appID, err)
		}

		apiStatus.Application = apiApp
	}

	// Normalization.

	if s.URL == "" {
		// URL was empty for some reason;
		// provide AP URI as fallback.
		s.URL = s.URI
	}

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
	instance := &apimodel.InstanceV1{
		URI:              i.URI,
		AccountDomain:    config.GetAccountDomain(),
		Title:            i.Title,
		Description:      i.Description,
		ShortDescription: i.ShortDescription,
		Email:            i.ContactEmail,
		Version:          config.GetSoftwareVersion(),
		Languages:        []string{}, // todo: not supported yet
		Registrations:    config.GetAccountsRegistrationOpen(),
		ApprovalRequired: config.GetAccountsApprovalRequired(),
		InvitesEnabled:   false, // todo: not supported yet
		MaxTootChars:     uint(config.GetStatusesMaxChars()),
		Rules:            c.InstanceRulesToAPIRules(i.Rules),
	}

	if config.GetInstanceInjectMastodonVersion() {
		instance.Version = toMastodonVersion(instance.Version)
	}

	// configuration
	instance.Configuration.Statuses.MaxCharacters = config.GetStatusesMaxChars()
	instance.Configuration.Statuses.MaxMediaAttachments = config.GetStatusesMediaMaxFiles()
	instance.Configuration.Statuses.CharactersReservedPerURL = instanceStatusesCharactersReservedPerURL
	instance.Configuration.Statuses.SupportedMimeTypes = instanceStatusesSupportedMimeTypes
	instance.Configuration.MediaAttachments.SupportedMimeTypes = media.SupportedMIMETypes
	instance.Configuration.MediaAttachments.ImageSizeLimit = int(config.GetMediaImageMaxSize())
	instance.Configuration.MediaAttachments.ImageMatrixLimit = instanceMediaAttachmentsImageMatrixLimit
	instance.Configuration.MediaAttachments.VideoSizeLimit = int(config.GetMediaVideoMaxSize())
	instance.Configuration.MediaAttachments.VideoFrameRateLimit = instanceMediaAttachmentsVideoFrameRateLimit
	instance.Configuration.MediaAttachments.VideoMatrixLimit = instanceMediaAttachmentsVideoMatrixLimit
	instance.Configuration.Polls.MaxOptions = config.GetStatusesPollMaxOptions()
	instance.Configuration.Polls.MaxCharactersPerOption = config.GetStatusesPollOptionMaxChars()
	instance.Configuration.Polls.MinExpiration = instancePollsMinExpiration
	instance.Configuration.Polls.MaxExpiration = instancePollsMaxExpiration
	instance.Configuration.Accounts.AllowCustomCSS = config.GetAccountsAllowCustomCSS()
	instance.Configuration.Accounts.MaxFeaturedTags = instanceAccountsMaxFeaturedTags
	instance.Configuration.Accounts.MaxProfileFields = instanceAccountsMaxProfileFields
	instance.Configuration.Emojis.EmojiSizeLimit = int(config.GetMediaEmojiLocalMaxSize())

	// URLs
	instance.URLs.StreamingAPI = "wss://" + i.Domain

	// statistics
	stats := make(map[string]int, 3)
	userCount, err := c.state.DB.CountInstanceUsers(ctx, i.Domain)
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting counting instance users: %w", err)
	}
	stats["user_count"] = userCount

	statusCount, err := c.state.DB.CountInstanceStatuses(ctx, i.Domain)
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting counting instance statuses: %w", err)
	}
	stats["status_count"] = statusCount

	domainCount, err := c.state.DB.CountInstanceDomains(ctx, i.Domain)
	if err != nil {
		return nil, fmt.Errorf("InstanceToAPIV1Instance: db error getting counting instance domains: %w", err)
	}
	stats["domain_count"] = domainCount
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
		instance.ThumbnailDescription = iAccount.AvatarMediaAttachment.Description
	} else {
		instance.Thumbnail = config.GetProtocol() + "://" + i.Domain + "/assets/logo.png" // default thumb
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
	instance := &apimodel.InstanceV2{
		Domain:        i.Domain,
		AccountDomain: config.GetAccountDomain(),
		Title:         i.Title,
		Version:       config.GetSoftwareVersion(),
		SourceURL:     instanceSourceURL,
		Description:   i.Description,
		Usage:         apimodel.InstanceV2Usage{}, // todo: not implemented
		Languages:     []string{},                 // todo: not implemented
		Rules:         c.InstanceRulesToAPIRules(i.Rules),
	}

	if config.GetInstanceInjectMastodonVersion() {
		instance.Version = toMastodonVersion(instance.Version)
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
		thumbnail.Description = iAccount.AvatarMediaAttachment.Description
		thumbnail.Blurhash = iAccount.AvatarMediaAttachment.Blurhash
	} else {
		thumbnail.URL = config.GetProtocol() + "://" + i.Domain + "/assets/logo.png" // default thumb
	}

	instance.Thumbnail = thumbnail

	// configuration
	instance.Configuration.URLs.Streaming = "wss://" + i.Domain
	instance.Configuration.Statuses.MaxCharacters = config.GetStatusesMaxChars()
	instance.Configuration.Statuses.MaxMediaAttachments = config.GetStatusesMediaMaxFiles()
	instance.Configuration.Statuses.CharactersReservedPerURL = instanceStatusesCharactersReservedPerURL
	instance.Configuration.Statuses.SupportedMimeTypes = instanceStatusesSupportedMimeTypes
	instance.Configuration.MediaAttachments.SupportedMimeTypes = media.SupportedMIMETypes
	instance.Configuration.MediaAttachments.ImageSizeLimit = int(config.GetMediaImageMaxSize())
	instance.Configuration.MediaAttachments.ImageMatrixLimit = instanceMediaAttachmentsImageMatrixLimit
	instance.Configuration.MediaAttachments.VideoSizeLimit = int(config.GetMediaVideoMaxSize())
	instance.Configuration.MediaAttachments.VideoFrameRateLimit = instanceMediaAttachmentsVideoFrameRateLimit
	instance.Configuration.MediaAttachments.VideoMatrixLimit = instanceMediaAttachmentsVideoMatrixLimit
	instance.Configuration.Polls.MaxOptions = config.GetStatusesPollMaxOptions()
	instance.Configuration.Polls.MaxCharactersPerOption = config.GetStatusesPollOptionMaxChars()
	instance.Configuration.Polls.MinExpiration = instancePollsMinExpiration
	instance.Configuration.Polls.MaxExpiration = instancePollsMaxExpiration
	instance.Configuration.Accounts.AllowCustomCSS = config.GetAccountsAllowCustomCSS()
	instance.Configuration.Accounts.MaxFeaturedTags = instanceAccountsMaxFeaturedTags
	instance.Configuration.Accounts.MaxProfileFields = instanceAccountsMaxProfileFields
	instance.Configuration.Emojis.EmojiSizeLimit = int(config.GetMediaEmojiLocalMaxSize())

	// registrations
	instance.Registrations.Enabled = config.GetAccountsRegistrationOpen()
	instance.Registrations.ApprovalRequired = config.GetAccountsApprovalRequired()
	instance.Registrations.Message = nil // todo: not implemented

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
		DomainBlocking:      r.DomainBlocking,
		Endorsed:            r.Endorsed,
		Note:                r.Note,
	}, nil
}

// NotificationToAPINotification converts a gts notification into a api notification
func (c *Converter) NotificationToAPINotification(ctx context.Context, n *gtsmodel.Notification) (*apimodel.Notification, error) {
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
		apiStatus, err = c.StatusToAPIStatus(ctx, n.Status, n.TargetAccount)
		if err != nil {
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

// DomainPermToAPIDomainPerm converts a gts model domin block or allow into an api domain permission.
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
		status, err := c.StatusToAPIStatus(ctx, s, requestingAccount)
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

// convertAttachmentsToAPIAttachments will convert a slice of GTS model attachments to frontend API model attachments, falling back to IDs if no GTS models supplied.
func (c *Converter) convertAttachmentsToAPIAttachments(ctx context.Context, attachments []*gtsmodel.MediaAttachment, attachmentIDs []string) ([]apimodel.Attachment, error) {
	var errs gtserror.MultiError

	if len(attachments) == 0 {
		// GTS model attachments were not populated

		// Preallocate expected GTS slice
		attachments = make([]*gtsmodel.MediaAttachment, 0, len(attachmentIDs))

		// Fetch GTS models for attachment IDs
		for _, id := range attachmentIDs {
			attachment, err := c.state.DB.GetAttachmentByID(ctx, id)
			if err != nil {
				errs.Appendf("error fetching attachment %s from database: %v", id, err)
				continue
			}
			attachments = append(attachments, attachment)
		}
	}

	// Preallocate expected frontend slice
	apiAttachments := make([]apimodel.Attachment, 0, len(attachments))

	// Convert GTS models to frontend models
	for _, attachment := range attachments {
		apiAttachment, err := c.AttachmentToAPIAttachment(ctx, attachment)
		if err != nil {
			errs.Appendf("error converting attchment %s to api attachment: %v", attachment.ID, err)
			continue
		}
		apiAttachments = append(apiAttachments, apiAttachment)
	}

	return apiAttachments, errs.Combine()
}

// convertEmojisToAPIEmojis will convert a slice of GTS model emojis to frontend API model emojis, falling back to IDs if no GTS models supplied.
func (c *Converter) convertEmojisToAPIEmojis(ctx context.Context, emojis []*gtsmodel.Emoji, emojiIDs []string) ([]apimodel.Emoji, error) {
	var errs gtserror.MultiError

	if len(emojis) == 0 {
		// GTS model attachments were not populated

		// Preallocate expected GTS slice
		emojis = make([]*gtsmodel.Emoji, 0, len(emojiIDs))

		// Fetch GTS models for emoji IDs
		for _, id := range emojiIDs {
			emoji, err := c.state.DB.GetEmojiByID(ctx, id)
			if err != nil {
				errs.Appendf("error fetching emoji %s from database: %v", id, err)
				continue
			}
			emojis = append(emojis, emoji)
		}
	}

	// Preallocate expected frontend slice
	apiEmojis := make([]apimodel.Emoji, 0, len(emojis))

	// Convert GTS models to frontend models
	for _, emoji := range emojis {
		apiEmoji, err := c.EmojiToAPIEmoji(ctx, emoji)
		if err != nil {
			errs.Appendf("error converting emoji %s to api emoji: %v", emoji.ID, err)
			continue
		}
		apiEmojis = append(apiEmojis, apiEmoji)
	}

	return apiEmojis, errs.Combine()
}

// convertMentionsToAPIMentions will convert a slice of GTS model mentions to frontend API model mentions, falling back to IDs if no GTS models supplied.
func (c *Converter) convertMentionsToAPIMentions(ctx context.Context, mentions []*gtsmodel.Mention, mentionIDs []string) ([]apimodel.Mention, error) {
	var errs gtserror.MultiError

	if len(mentions) == 0 {
		var err error

		// GTS model mentions were not populated
		//
		// Fetch GTS models for mention IDs
		mentions, err = c.state.DB.GetMentions(ctx, mentionIDs)
		if err != nil {
			errs.Appendf("error fetching mentions from database: %v", err)
		}
	}

	// Preallocate expected frontend slice
	apiMentions := make([]apimodel.Mention, 0, len(mentions))

	// Convert GTS models to frontend models
	for _, mention := range mentions {
		apiMention, err := c.MentionToAPIMention(ctx, mention)
		if err != nil {
			errs.Appendf("error converting mention %s to api mention: %v", mention.ID, err)
			continue
		}
		apiMentions = append(apiMentions, apiMention)
	}

	return apiMentions, errs.Combine()
}

// convertTagsToAPITags will convert a slice of GTS model tags to frontend API model tags, falling back to IDs if no GTS models supplied.
func (c *Converter) convertTagsToAPITags(ctx context.Context, tags []*gtsmodel.Tag, tagIDs []string) ([]apimodel.Tag, error) {
	var errs gtserror.MultiError

	if len(tags) == 0 {
		var err error

		tags, err = c.state.DB.GetTags(ctx, tagIDs)
		if err != nil {
			errs.Appendf("error fetching tags from database: %v", err)
		}
	}

	// Preallocate expected frontend slice
	apiTags := make([]apimodel.Tag, 0, len(tags))

	// Convert GTS models to frontend models
	for _, tag := range tags {
		apiTag, err := c.TagToAPITag(ctx, tag, false)
		if err != nil {
			errs.Appendf("error converting tag %s to api tag: %v", tag.ID, err)
			continue
		}
		apiTags = append(apiTags, apiTag)
	}

	return apiTags, errs.Combine()
}
