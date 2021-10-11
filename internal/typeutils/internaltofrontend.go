/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package typeutils

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) AccountToAPIAccountSensitive(ctx context.Context, a *gtsmodel.Account) (*model.Account, error) {
	// we can build this sensitive account easily by first getting the public account....
	apiAccount, err := c.AccountToAPIAccountPublic(ctx, a)
	if err != nil {
		return nil, err
	}

	// then adding the Source object to it...

	// check pending follow requests aimed at this account
	frs, err := c.db.GetAccountFollowRequests(ctx, a.ID)
	if err != nil {
		if err != db.ErrNoEntries {
			return nil, fmt.Errorf("error getting follow requests: %s", err)
		}
	}
	var frc int
	if frs != nil {
		frc = len(frs)
	}

	apiAccount.Source = &model.Source{
		Privacy:             c.VisToAPIVis(ctx, a.Privacy),
		Sensitive:           a.Sensitive,
		Language:            a.Language,
		Note:                a.Note,
		Fields:              apiAccount.Fields,
		FollowRequestsCount: frc,
	}

	return apiAccount, nil
}

func (c *converter) AccountToAPIAccountPublic(ctx context.Context, a *gtsmodel.Account) (*model.Account, error) {
	if a == nil {
		return nil, fmt.Errorf("given account was nil")
	}

	// count followers
	followersCount, err := c.db.CountAccountFollowedBy(ctx, a.ID, false)
	if err != nil {
		return nil, fmt.Errorf("error counting followers: %s", err)
	}

	// count following
	followingCount, err := c.db.CountAccountFollows(ctx, a.ID, false)
	if err != nil {
		return nil, fmt.Errorf("error counting following: %s", err)
	}

	// count statuses
	statusesCount, err := c.db.CountAccountStatuses(ctx, a.ID)
	if err != nil {
		return nil, fmt.Errorf("error counting statuses: %s", err)
	}

	// check when the last status was
	var lastStatusAt string
	lastPosted, err := c.db.GetAccountLastPosted(ctx, a.ID)
	if err == nil && !lastPosted.IsZero() {
		lastStatusAt = lastPosted.Format(time.RFC3339)
	}

	// build the avatar and header URLs
	var aviURL string
	var aviURLStatic string
	if a.AvatarMediaAttachmentID != "" {
		// make sure avi is pinned to this account
		if a.AvatarMediaAttachment == nil {
			avi, err := c.db.GetAttachmentByID(ctx, a.AvatarMediaAttachmentID)
			if err != nil {
				return nil, fmt.Errorf("error retrieving avatar: %s", err)
			}
			a.AvatarMediaAttachment = avi
		}
		aviURL = a.AvatarMediaAttachment.URL
		aviURLStatic = a.AvatarMediaAttachment.Thumbnail.URL
	}

	var headerURL string
	var headerURLStatic string
	if a.HeaderMediaAttachmentID != "" {
		// make sure header is pinned to this account
		if a.HeaderMediaAttachment == nil {
			avi, err := c.db.GetAttachmentByID(ctx, a.HeaderMediaAttachmentID)
			if err != nil {
				return nil, fmt.Errorf("error retrieving avatar: %s", err)
			}
			a.HeaderMediaAttachment = avi
		}
		headerURL = a.HeaderMediaAttachment.URL
		headerURLStatic = a.HeaderMediaAttachment.Thumbnail.URL
	}

	// get the fields set on this account
	fields := []model.Field{}
	for _, f := range a.Fields {
		mField := model.Field{
			Name:  f.Name,
			Value: f.Value,
		}
		if !f.VerifiedAt.IsZero() {
			mField.VerifiedAt = f.VerifiedAt.Format(time.RFC3339)
		}
		fields = append(fields, mField)
	}

	emojis := []model.Emoji{}
	// TODO: account emojis

	var acct string
	if a.Domain != "" {
		// this is a remote user
		acct = fmt.Sprintf("%s@%s", a.Username, a.Domain)
	} else {
		// this is a local user
		acct = a.Username
	}

	var suspended bool
	if !a.SuspendedAt.IsZero() {
		suspended = true
	}

	accountFrontend := &model.Account{
		ID:             a.ID,
		Username:       a.Username,
		Acct:           acct,
		DisplayName:    a.DisplayName,
		Locked:         a.Locked,
		Bot:            a.Bot,
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
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
		Emojis:         emojis, // TODO: implement this
		Fields:         fields,
		Suspended:      suspended,
	}

	return accountFrontend, nil
}

func (c *converter) AccountToAPIAccountBlocked(ctx context.Context, a *gtsmodel.Account) (*model.Account, error) {
	var acct string
	if a.Domain != "" {
		// this is a remote user
		acct = fmt.Sprintf("%s@%s", a.Username, a.Domain)
	} else {
		// this is a local user
		acct = a.Username
	}

	var suspended bool
	if !a.SuspendedAt.IsZero() {
		suspended = true
	}

	return &model.Account{
		ID:          a.ID,
		Username:    a.Username,
		Acct:        acct,
		DisplayName: a.DisplayName,
		Bot:         a.Bot,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		URL:         a.URL,
		Suspended:   suspended,
	}, nil
}

func (c *converter) AppToAPIAppSensitive(ctx context.Context, a *gtsmodel.Application) (*model.Application, error) {
	return &model.Application{
		ID:           a.ID,
		Name:         a.Name,
		Website:      a.Website,
		RedirectURI:  a.RedirectURI,
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
	}, nil
}

func (c *converter) AppToAPIAppPublic(ctx context.Context, a *gtsmodel.Application) (*model.Application, error) {
	return &model.Application{
		Name:    a.Name,
		Website: a.Website,
	}, nil
}

func (c *converter) AttachmentToAPIAttachment(ctx context.Context, a *gtsmodel.MediaAttachment) (model.Attachment, error) {
	return model.Attachment{
		ID:               a.ID,
		Type:             strings.ToLower(string(a.Type)),
		URL:              a.URL,
		PreviewURL:       a.Thumbnail.URL,
		RemoteURL:        a.RemoteURL,
		PreviewRemoteURL: a.Thumbnail.RemoteURL,
		Meta: model.MediaMeta{
			Original: model.MediaDimensions{
				Width:  a.FileMeta.Original.Width,
				Height: a.FileMeta.Original.Height,
				Size:   fmt.Sprintf("%dx%d", a.FileMeta.Original.Width, a.FileMeta.Original.Height),
				Aspect: float32(a.FileMeta.Original.Aspect),
			},
			Small: model.MediaDimensions{
				Width:  a.FileMeta.Small.Width,
				Height: a.FileMeta.Small.Height,
				Size:   fmt.Sprintf("%dx%d", a.FileMeta.Small.Width, a.FileMeta.Small.Height),
				Aspect: float32(a.FileMeta.Small.Aspect),
			},
			Focus: model.MediaFocus{
				X: a.FileMeta.Focus.X,
				Y: a.FileMeta.Focus.Y,
			},
		},
		Description: a.Description,
		Blurhash:    a.Blurhash,
	}, nil
}

func (c *converter) MentionToAPIMention(ctx context.Context, m *gtsmodel.Mention) (model.Mention, error) {
	if m.TargetAccount == nil {
		targetAccount, err := c.db.GetAccountByID(ctx, m.TargetAccountID)
		if err != nil {
			return model.Mention{}, err
		}
		m.TargetAccount = targetAccount
	}

	var local bool
	if m.TargetAccount.Domain == "" {
		local = true
	}

	var acct string
	if local {
		acct = m.TargetAccount.Username
	} else {
		acct = fmt.Sprintf("%s@%s", m.TargetAccount.Username, m.TargetAccount.Domain)
	}

	return model.Mention{
		ID:       m.TargetAccount.ID,
		Username: m.TargetAccount.Username,
		URL:      m.TargetAccount.URL,
		Acct:     acct,
	}, nil
}

func (c *converter) EmojiToAPIEmoji(ctx context.Context, e *gtsmodel.Emoji) (model.Emoji, error) {
	return model.Emoji{
		Shortcode:       e.Shortcode,
		URL:             e.ImageURL,
		StaticURL:       e.ImageStaticURL,
		VisibleInPicker: e.VisibleInPicker,
		Category:        e.CategoryID,
	}, nil
}

func (c *converter) TagToAPITag(ctx context.Context, t *gtsmodel.Tag) (model.Tag, error) {
	return model.Tag{
		Name: t.Name,
		URL:  t.URL,
	}, nil
}

func (c *converter) StatusToAPIStatus(ctx context.Context, s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*model.Status, error) {
	repliesCount, err := c.db.CountStatusReplies(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("error counting replies: %s", err)
	}

	reblogsCount, err := c.db.CountStatusReblogs(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("error counting reblogs: %s", err)
	}

	favesCount, err := c.db.CountStatusFaves(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("error counting faves: %s", err)
	}

	var apiRebloggedStatus *model.Status
	if s.BoostOfID != "" {
		// the boosted status might have been set on this struct already so check first before doing db calls
		if s.BoostOf == nil {
			// it's not set so fetch it from the db
			bs, err := c.db.GetStatusByID(ctx, s.BoostOfID)
			if err != nil {
				return nil, fmt.Errorf("error getting boosted status with id %s: %s", s.BoostOfID, err)
			}
			s.BoostOf = bs
		}

		// the boosted account might have been set on this struct already or passed as a param so check first before doing db calls
		if s.BoostOfAccount == nil {
			// it's not set so fetch it from the db
			ba, err := c.db.GetAccountByID(ctx, s.BoostOf.AccountID)
			if err != nil {
				return nil, fmt.Errorf("error getting boosted account %s from status with id %s: %s", s.BoostOf.AccountID, s.BoostOfID, err)
			}
			s.BoostOfAccount = ba
			s.BoostOf.Account = ba
		}

		apiRebloggedStatus, err = c.StatusToAPIStatus(ctx, s.BoostOf, requestingAccount)
		if err != nil {
			return nil, fmt.Errorf("error converting boosted status to apitype: %s", err)
		}
	}

	var apiApplication *model.Application
	if s.CreatedWithApplicationID != "" {
		gtsApplication := &gtsmodel.Application{}
		if err := c.db.GetByID(ctx, s.CreatedWithApplicationID, gtsApplication); err != nil {
			return nil, fmt.Errorf("error fetching application used to create status: %s", err)
		}
		apiApplication, err = c.AppToAPIAppPublic(ctx, gtsApplication)
		if err != nil {
			return nil, fmt.Errorf("error parsing application used to create status: %s", err)
		}
	}

	if s.Account == nil {
		a, err := c.db.GetAccountByID(ctx, s.AccountID)
		if err != nil {
			return nil, fmt.Errorf("error getting status author: %s", err)
		}
		s.Account = a
	}

	apiAuthorAccount, err := c.AccountToAPIAccountPublic(ctx, s.Account)
	if err != nil {
		return nil, fmt.Errorf("error parsing account of status author: %s", err)
	}

	apiAttachments := []model.Attachment{}
	// the status might already have some gts attachments on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts attachments into api ones
	if s.Attachments != nil {
		for _, gtsAttachment := range s.Attachments {
			apiAttachment, err := c.AttachmentToAPIAttachment(ctx, gtsAttachment)
			if err != nil {
				logrus.Errorf("error converting attachment with id %s: %s", gtsAttachment.ID, err)
				continue
			}
			apiAttachments = append(apiAttachments, apiAttachment)
		}
		// the status doesn't have gts attachments on it, but it does have attachment IDs
		// in this case, we need to pull the gts attachments from the db to convert them into api ones
	} else {
		for _, aID := range s.AttachmentIDs {
			gtsAttachment, err := c.db.GetAttachmentByID(ctx, aID)
			if err != nil {
				logrus.Errorf("error getting attachment with id %s: %s", aID, err)
				continue
			}
			apiAttachment, err := c.AttachmentToAPIAttachment(ctx, gtsAttachment)
			if err != nil {
				logrus.Errorf("error converting attachment with id %s: %s", aID, err)
				continue
			}
			apiAttachments = append(apiAttachments, apiAttachment)
		}
	}

	apiMentions := []model.Mention{}
	// the status might already have some gts mentions on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts mentions into api ones
	if s.Mentions != nil {
		for _, gtsMention := range s.Mentions {
			apiMention, err := c.MentionToAPIMention(ctx, gtsMention)
			if err != nil {
				logrus.Errorf("error converting mention with id %s: %s", gtsMention.ID, err)
				continue
			}
			apiMentions = append(apiMentions, apiMention)
		}
		// the status doesn't have gts mentions on it, but it does have mention IDs
		// in this case, we need to pull the gts mentions from the db to convert them into api ones
	} else {
		for _, mID := range s.MentionIDs {
			gtsMention, err := c.db.GetMention(ctx, mID)
			if err != nil {
				logrus.Errorf("error getting mention with id %s: %s", mID, err)
				continue
			}
			apiMention, err := c.MentionToAPIMention(ctx, gtsMention)
			if err != nil {
				logrus.Errorf("error converting mention with id %s: %s", gtsMention.ID, err)
				continue
			}
			apiMentions = append(apiMentions, apiMention)
		}
	}

	apiTags := []model.Tag{}
	// the status might already have some gts tags on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts tags into api ones
	if s.Tags != nil {
		for _, gtsTag := range s.Tags {
			apiTag, err := c.TagToAPITag(ctx, gtsTag)
			if err != nil {
				logrus.Errorf("error converting tag with id %s: %s", gtsTag.ID, err)
				continue
			}
			apiTags = append(apiTags, apiTag)
		}
		// the status doesn't have gts tags on it, but it does have tag IDs
		// in this case, we need to pull the gts tags from the db to convert them into api ones
	} else {
		for _, t := range s.TagIDs {
			gtsTag := &gtsmodel.Tag{}
			if err := c.db.GetByID(ctx, t, gtsTag); err != nil {
				logrus.Errorf("error getting tag with id %s: %s", t, err)
				continue
			}
			apiTag, err := c.TagToAPITag(ctx, gtsTag)
			if err != nil {
				logrus.Errorf("error converting tag with id %s: %s", gtsTag.ID, err)
				continue
			}
			apiTags = append(apiTags, apiTag)
		}
	}

	apiEmojis := []model.Emoji{}
	// the status might already have some gts emojis on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts emojis into api ones
	if s.Emojis != nil {
		for _, gtsEmoji := range s.Emojis {
			apiEmoji, err := c.EmojiToAPIEmoji(ctx, gtsEmoji)
			if err != nil {
				logrus.Errorf("error converting emoji with id %s: %s", gtsEmoji.ID, err)
				continue
			}
			apiEmojis = append(apiEmojis, apiEmoji)
		}
		// the status doesn't have gts emojis on it, but it does have emoji IDs
		// in this case, we need to pull the gts emojis from the db to convert them into api ones
	} else {
		for _, e := range s.EmojiIDs {
			gtsEmoji := &gtsmodel.Emoji{}
			if err := c.db.GetByID(ctx, e, gtsEmoji); err != nil {
				logrus.Errorf("error getting emoji with id %s: %s", e, err)
				continue
			}
			apiEmoji, err := c.EmojiToAPIEmoji(ctx, gtsEmoji)
			if err != nil {
				logrus.Errorf("error converting emoji with id %s: %s", gtsEmoji.ID, err)
				continue
			}
			apiEmojis = append(apiEmojis, apiEmoji)
		}
	}

	var apiCard *model.Card
	var apiPoll *model.Poll

	statusInteractions := &statusInteractions{}
	si, err := c.interactionsWithStatusForAccount(ctx, s, requestingAccount)
	if err == nil {
		statusInteractions = si
	}

	apiStatus := &model.Status{
		ID:                 s.ID,
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
		InReplyToID:        s.InReplyToID,
		InReplyToAccountID: s.InReplyToAccountID,
		Sensitive:          s.Sensitive,
		SpoilerText:        s.ContentWarning,
		Visibility:         c.VisToAPIVis(ctx, s.Visibility),
		Language:           s.Language,
		URI:                s.URI,
		URL:                s.URL,
		RepliesCount:       repliesCount,
		ReblogsCount:       reblogsCount,
		FavouritesCount:    favesCount,
		Favourited:         statusInteractions.Faved,
		Bookmarked:         statusInteractions.Bookmarked,
		Muted:              statusInteractions.Muted,
		Reblogged:          statusInteractions.Reblogged,
		Pinned:             s.Pinned,
		Content:            s.Content,
		Application:        apiApplication,
		Account:            apiAuthorAccount,
		MediaAttachments:   apiAttachments,
		Mentions:           apiMentions,
		Tags:               apiTags,
		Emojis:             apiEmojis,
		Card:               apiCard, // TODO: implement cards
		Poll:               apiPoll, // TODO: implement polls
		Text:               s.Text,
	}

	if apiRebloggedStatus != nil {
		apiStatus.Reblog = &model.StatusReblogged{Status: apiRebloggedStatus}
	}

	return apiStatus, nil
}

// VisToapi converts a gts visibility into its api equivalent
func (c *converter) VisToAPIVis(ctx context.Context, m gtsmodel.Visibility) model.Visibility {
	switch m {
	case gtsmodel.VisibilityPublic:
		return model.VisibilityPublic
	case gtsmodel.VisibilityUnlocked:
		return model.VisibilityUnlisted
	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		return model.VisibilityPrivate
	case gtsmodel.VisibilityDirect:
		return model.VisibilityDirect
	}
	return ""
}

func (c *converter) InstanceToAPIInstance(ctx context.Context, i *gtsmodel.Instance) (*model.Instance, error) {
	mi := &model.Instance{
		URI:              i.URI,
		Title:            i.Title,
		Description:      i.Description,
		ShortDescription: i.ShortDescription,
		Email:            i.ContactEmail,
		Version:          i.Version,
		Stats:            make(map[string]int),
		ContactAccount:   &model.Account{},
	}

	// if the requested instance is *this* instance, we can add some extra information
	if i.Domain == c.config.Host {
		userCountKey := "user_count"
		statusCountKey := "status_count"
		domainCountKey := "domain_count"

		userCount, err := c.db.CountInstanceUsers(ctx, c.config.Host)
		if err == nil {
			mi.Stats[userCountKey] = userCount
		}

		statusCount, err := c.db.CountInstanceStatuses(ctx, c.config.Host)
		if err == nil {
			mi.Stats[statusCountKey] = statusCount
		}

		domainCount, err := c.db.CountInstanceDomains(ctx, c.config.Host)
		if err == nil {
			mi.Stats[domainCountKey] = domainCount
		}

		mi.Registrations = c.config.AccountsConfig.OpenRegistration
		mi.ApprovalRequired = c.config.AccountsConfig.RequireApproval
		mi.InvitesEnabled = false // TODO
		mi.MaxTootChars = uint(c.config.StatusesConfig.MaxChars)
		mi.URLS = &model.InstanceURLs{
			StreamingAPI: fmt.Sprintf("wss://%s", c.config.Host),
		}
		mi.Version = c.config.SoftwareVersion
	}

	// get the instance account if it exists and just skip if it doesn't
	ia, err := c.db.GetInstanceAccount(ctx, "")
	if err == nil {
		if ia.HeaderMediaAttachment != nil {
			mi.Thumbnail = ia.HeaderMediaAttachment.URL
		}
	}

	// contact account is optional but let's try to get it
	if i.ContactAccountID != "" {
		if i.ContactAccount == nil {
			contactAccount, err := c.db.GetAccountByID(ctx, i.ContactAccountID)
			if err == nil {
				i.ContactAccount = contactAccount
			}
		}
		ma, err := c.AccountToAPIAccountPublic(ctx, i.ContactAccount)
		if err == nil {
			mi.ContactAccount = ma
		}
	}

	return mi, nil
}

func (c *converter) RelationshipToAPIRelationship(ctx context.Context, r *gtsmodel.Relationship) (*model.Relationship, error) {
	return &model.Relationship{
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

func (c *converter) NotificationToAPINotification(ctx context.Context, n *gtsmodel.Notification) (*model.Notification, error) {
	if n.TargetAccount == nil {
		tAccount, err := c.db.GetAccountByID(ctx, n.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("NotificationToapi: error getting target account with id %s from the db: %s", n.TargetAccountID, err)
		}
		n.TargetAccount = tAccount
	}

	if n.OriginAccount == nil {
		ogAccount, err := c.db.GetAccountByID(ctx, n.OriginAccountID)
		if err != nil {
			return nil, fmt.Errorf("NotificationToapi: error getting origin account with id %s from the db: %s", n.OriginAccountID, err)
		}
		n.OriginAccount = ogAccount
	}

	apiAccount, err := c.AccountToAPIAccountPublic(ctx, n.OriginAccount)
	if err != nil {
		return nil, fmt.Errorf("NotificationToapi: error converting account to api: %s", err)
	}

	var apiStatus *model.Status
	if n.StatusID != "" {
		if n.Status == nil {
			status, err := c.db.GetStatusByID(ctx, n.StatusID)
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
		apiStatus, err = c.StatusToAPIStatus(ctx, n.Status, nil)
		if err != nil {
			return nil, fmt.Errorf("NotificationToapi: error converting status to api: %s", err)
		}
	}

	return &model.Notification{
		ID:        n.ID,
		Type:      string(n.NotificationType),
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		Account:   apiAccount,
		Status:    apiStatus,
	}, nil
}

func (c *converter) DomainBlockToAPIDomainBlock(ctx context.Context, b *gtsmodel.DomainBlock, export bool) (*model.DomainBlock, error) {

	domainBlock := &model.DomainBlock{
		Domain:        b.Domain,
		PublicComment: b.PublicComment,
	}

	// if we're exporting a domain block, return it with minimal information attached
	if !export {
		domainBlock.ID = b.ID
		domainBlock.Obfuscate = b.Obfuscate
		domainBlock.PrivateComment = b.PrivateComment
		domainBlock.SubscriptionID = b.SubscriptionID
		domainBlock.CreatedBy = b.CreatedByAccountID
		domainBlock.CreatedAt = b.CreatedAt.Format(time.RFC3339)
	}

	return domainBlock, nil
}
