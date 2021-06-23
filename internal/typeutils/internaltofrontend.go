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
	"fmt"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) AccountToMastoSensitive(a *gtsmodel.Account) (*model.Account, error) {
	// we can build this sensitive account easily by first getting the public account....
	mastoAccount, err := c.AccountToMastoPublic(a)
	if err != nil {
		return nil, err
	}

	// then adding the Source object to it...

	// check pending follow requests aimed at this account
	fr := []gtsmodel.FollowRequest{}
	if err := c.db.GetFollowRequestsForAccountID(a.ID, &fr); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting follow requests: %s", err)
		}
	}
	var frc int
	if fr != nil {
		frc = len(fr)
	}

	mastoAccount.Source = &model.Source{
		Privacy:             c.VisToMasto(a.Privacy),
		Sensitive:           a.Sensitive,
		Language:            a.Language,
		Note:                a.Note,
		Fields:              mastoAccount.Fields,
		FollowRequestsCount: frc,
	}

	return mastoAccount, nil
}

func (c *converter) AccountToMastoPublic(a *gtsmodel.Account) (*model.Account, error) {
	// count followers
	followers := []gtsmodel.Follow{}
	if err := c.db.GetFollowersByAccountID(a.ID, &followers, false); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting followers: %s", err)
		}
	}
	var followersCount int
	if followers != nil {
		followersCount = len(followers)
	}

	// count following
	following := []gtsmodel.Follow{}
	if err := c.db.GetFollowingByAccountID(a.ID, &following); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting following: %s", err)
		}
	}
	var followingCount int
	if following != nil {
		followingCount = len(following)
	}

	// count statuses
	statusesCount, err := c.db.CountStatusesByAccountID(a.ID)
	if err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting last statuses: %s", err)
		}
	}

	// check when the last status was
	lastStatus := &gtsmodel.Status{}
	if err := c.db.GetLastStatusForAccountID(a.ID, lastStatus); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting last status: %s", err)
		}
	}
	var lastStatusAt string
	if lastStatus != nil {
		lastStatusAt = lastStatus.CreatedAt.Format(time.RFC3339)
	}

	// build the avatar and header URLs
	avi := &gtsmodel.MediaAttachment{}
	if err := c.db.GetAvatarForAccountID(avi, a.ID); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting avatar: %s", err)
		}
	}
	aviURL := avi.URL
	aviURLStatic := avi.Thumbnail.URL

	header := &gtsmodel.MediaAttachment{}
	if err := c.db.GetHeaderForAccountID(header, a.ID); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting header: %s", err)
		}
	}
	headerURL := header.URL
	headerURLStatic := header.Thumbnail.URL

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

	return &model.Account{
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
	}, nil
}

func (c *converter) AppToMastoSensitive(a *gtsmodel.Application) (*model.Application, error) {
	return &model.Application{
		ID:           a.ID,
		Name:         a.Name,
		Website:      a.Website,
		RedirectURI:  a.RedirectURI,
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		VapidKey:     a.VapidKey,
	}, nil
}

func (c *converter) AppToMastoPublic(a *gtsmodel.Application) (*model.Application, error) {
	return &model.Application{
		Name:    a.Name,
		Website: a.Website,
	}, nil
}

func (c *converter) AttachmentToMasto(a *gtsmodel.MediaAttachment) (model.Attachment, error) {
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

func (c *converter) MentionToMasto(m *gtsmodel.Mention) (model.Mention, error) {
	target := &gtsmodel.Account{}
	if err := c.db.GetByID(m.TargetAccountID, target); err != nil {
		return model.Mention{}, err
	}

	var local bool
	if target.Domain == "" {
		local = true
	}

	var acct string
	if local {
		acct = target.Username
	} else {
		acct = fmt.Sprintf("%s@%s", target.Username, target.Domain)
	}

	return model.Mention{
		ID:       target.ID,
		Username: target.Username,
		URL:      target.URL,
		Acct:     acct,
	}, nil
}

func (c *converter) EmojiToMasto(e *gtsmodel.Emoji) (model.Emoji, error) {
	return model.Emoji{
		Shortcode:       e.Shortcode,
		URL:             e.ImageURL,
		StaticURL:       e.ImageStaticURL,
		VisibleInPicker: e.VisibleInPicker,
		Category:        e.CategoryID,
	}, nil
}

func (c *converter) TagToMasto(t *gtsmodel.Tag) (model.Tag, error) {
	tagURL := fmt.Sprintf("%s://%s/tags/%s", c.config.Protocol, c.config.Host, t.Name)

	return model.Tag{
		Name: t.Name,
		URL:  tagURL, // we don't serve URLs with collections of tagged statuses (FOR NOW) so this is purely for mastodon compatibility ¯\_(ツ)_/¯
	}, nil
}

func (c *converter) StatusToMasto(s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*model.Status, error) {
	repliesCount, err := c.db.GetReplyCountForStatus(s)
	if err != nil {
		return nil, fmt.Errorf("error counting replies: %s", err)
	}

	reblogsCount, err := c.db.GetReblogCountForStatus(s)
	if err != nil {
		return nil, fmt.Errorf("error counting reblogs: %s", err)
	}

	favesCount, err := c.db.GetFaveCountForStatus(s)
	if err != nil {
		return nil, fmt.Errorf("error counting faves: %s", err)
	}

	var mastoRebloggedStatus *model.Status
	if s.BoostOfID != "" {
		// the boosted status might have been set on this struct already so check first before doing db calls
		if s.GTSBoostedStatus == nil {
			// it's not set so fetch it from the db
			bs := &gtsmodel.Status{}
			if err := c.db.GetByID(s.BoostOfID, bs); err != nil {
				return nil, fmt.Errorf("error getting boosted status with id %s: %s", s.BoostOfID, err)
			}
			s.GTSBoostedStatus = bs
		}

		// the boosted account might have been set on this struct already or passed as a param so check first before doing db calls
		if s.GTSBoostedAccount == nil {
			// it's not set so fetch it from the db
			ba := &gtsmodel.Account{}
			if err := c.db.GetByID(s.GTSBoostedStatus.AccountID, ba); err != nil {
				return nil, fmt.Errorf("error getting boosted account %s from status with id %s: %s", s.GTSBoostedStatus.AccountID, s.BoostOfID, err)
			}
			s.GTSBoostedAccount = ba
			s.GTSBoostedStatus.GTSAuthorAccount = ba
		}

		mastoRebloggedStatus, err = c.StatusToMasto(s.GTSBoostedStatus, requestingAccount)
		if err != nil {
			return nil, fmt.Errorf("error converting boosted status to mastotype: %s", err)
		}
	}

	var mastoApplication *model.Application
	if s.CreatedWithApplicationID != "" {
		gtsApplication := &gtsmodel.Application{}
		if err := c.db.GetByID(s.CreatedWithApplicationID, gtsApplication); err != nil {
			return nil, fmt.Errorf("error fetching application used to create status: %s", err)
		}
		mastoApplication, err = c.AppToMastoPublic(gtsApplication)
		if err != nil {
			return nil, fmt.Errorf("error parsing application used to create status: %s", err)
		}
	}

	if s.GTSAuthorAccount == nil {
		a := &gtsmodel.Account{}
		if err := c.db.GetByID(s.AccountID, a); err != nil {
			return nil, fmt.Errorf("error getting status author: %s", err)
		}
		s.GTSAuthorAccount = a
	}

	mastoAuthorAccount, err := c.AccountToMastoPublic(s.GTSAuthorAccount)
	if err != nil {
		return nil, fmt.Errorf("error parsing account of status author: %s", err)
	}

	mastoAttachments := []model.Attachment{}
	// the status might already have some gts attachments on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts attachments into masto ones
	if s.GTSMediaAttachments != nil {
		for _, gtsAttachment := range s.GTSMediaAttachments {
			mastoAttachment, err := c.AttachmentToMasto(gtsAttachment)
			if err != nil {
				return nil, fmt.Errorf("error converting attachment with id %s: %s", gtsAttachment.ID, err)
			}
			mastoAttachments = append(mastoAttachments, mastoAttachment)
		}
		// the status doesn't have gts attachments on it, but it does have attachment IDs
		// in this case, we need to pull the gts attachments from the db to convert them into masto ones
	} else {
		for _, a := range s.Attachments {
			gtsAttachment := &gtsmodel.MediaAttachment{}
			if err := c.db.GetByID(a, gtsAttachment); err != nil {
				return nil, fmt.Errorf("error getting attachment with id %s: %s", a, err)
			}
			mastoAttachment, err := c.AttachmentToMasto(gtsAttachment)
			if err != nil {
				return nil, fmt.Errorf("error converting attachment with id %s: %s", a, err)
			}
			mastoAttachments = append(mastoAttachments, mastoAttachment)
		}
	}

	mastoMentions := []model.Mention{}
	// the status might already have some gts mentions on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts mentions into masto ones
	if s.GTSMentions != nil {
		for _, gtsMention := range s.GTSMentions {
			mastoMention, err := c.MentionToMasto(gtsMention)
			if err != nil {
				return nil, fmt.Errorf("error converting mention with id %s: %s", gtsMention.ID, err)
			}
			mastoMentions = append(mastoMentions, mastoMention)
		}
		// the status doesn't have gts mentions on it, but it does have mention IDs
		// in this case, we need to pull the gts mentions from the db to convert them into masto ones
	} else {
		for _, m := range s.Mentions {
			gtsMention := &gtsmodel.Mention{}
			if err := c.db.GetByID(m, gtsMention); err != nil {
				return nil, fmt.Errorf("error getting mention with id %s: %s", m, err)
			}
			mastoMention, err := c.MentionToMasto(gtsMention)
			if err != nil {
				return nil, fmt.Errorf("error converting mention with id %s: %s", gtsMention.ID, err)
			}
			mastoMentions = append(mastoMentions, mastoMention)
		}
	}

	mastoTags := []model.Tag{}
	// the status might already have some gts tags on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts tags into masto ones
	if s.GTSTags != nil {
		for _, gtsTag := range s.GTSTags {
			mastoTag, err := c.TagToMasto(gtsTag)
			if err != nil {
				return nil, fmt.Errorf("error converting tag with id %s: %s", gtsTag.ID, err)
			}
			mastoTags = append(mastoTags, mastoTag)
		}
		// the status doesn't have gts tags on it, but it does have tag IDs
		// in this case, we need to pull the gts tags from the db to convert them into masto ones
	} else {
		for _, t := range s.Tags {
			gtsTag := &gtsmodel.Tag{}
			if err := c.db.GetByID(t, gtsTag); err != nil {
				return nil, fmt.Errorf("error getting tag with id %s: %s", t, err)
			}
			mastoTag, err := c.TagToMasto(gtsTag)
			if err != nil {
				return nil, fmt.Errorf("error converting tag with id %s: %s", gtsTag.ID, err)
			}
			mastoTags = append(mastoTags, mastoTag)
		}
	}

	mastoEmojis := []model.Emoji{}
	// the status might already have some gts emojis on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts emojis into masto ones
	if s.GTSEmojis != nil {
		for _, gtsEmoji := range s.GTSEmojis {
			mastoEmoji, err := c.EmojiToMasto(gtsEmoji)
			if err != nil {
				return nil, fmt.Errorf("error converting emoji with id %s: %s", gtsEmoji.ID, err)
			}
			mastoEmojis = append(mastoEmojis, mastoEmoji)
		}
		// the status doesn't have gts emojis on it, but it does have emoji IDs
		// in this case, we need to pull the gts emojis from the db to convert them into masto ones
	} else {
		for _, e := range s.Emojis {
			gtsEmoji := &gtsmodel.Emoji{}
			if err := c.db.GetByID(e, gtsEmoji); err != nil {
				return nil, fmt.Errorf("error getting emoji with id %s: %s", e, err)
			}
			mastoEmoji, err := c.EmojiToMasto(gtsEmoji)
			if err != nil {
				return nil, fmt.Errorf("error converting emoji with id %s: %s", gtsEmoji.ID, err)
			}
			mastoEmojis = append(mastoEmojis, mastoEmoji)
		}
	}

	var mastoCard *model.Card
	var mastoPoll *model.Poll

	statusInteractions := &statusInteractions{}
	si, err := c.interactionsWithStatusForAccount(s, requestingAccount)
	if err == nil {
		statusInteractions = si
	}

	return &model.Status{
		ID:                 s.ID,
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
		InReplyToID:        s.InReplyToID,
		InReplyToAccountID: s.InReplyToAccountID,
		Sensitive:          s.Sensitive,
		SpoilerText:        s.ContentWarning,
		Visibility:         c.VisToMasto(s.Visibility),
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
		Reblog:             mastoRebloggedStatus,
		Application:        mastoApplication,
		Account:            mastoAuthorAccount,
		MediaAttachments:   mastoAttachments,
		Mentions:           mastoMentions,
		Tags:               mastoTags,
		Emojis:             mastoEmojis,
		Card:               mastoCard, // TODO: implement cards
		Poll:               mastoPoll, // TODO: implement polls
		Text:               s.Text,
	}, nil
}

// VisToMasto converts a gts visibility into its mastodon equivalent
func (c *converter) VisToMasto(m gtsmodel.Visibility) model.Visibility {
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

func (c *converter) InstanceToMasto(i *gtsmodel.Instance) (*model.Instance, error) {
	mi := &model.Instance{
		URI:              i.URI,
		Title:            i.Title,
		Description:      i.Description,
		ShortDescription: i.ShortDescription,
		Email:            i.ContactEmail,
		Version:          i.Version,
		Stats:            make(map[string]int),
	}

	// if the requested instance is *this* instance, we can add some extra information
	if i.Domain == c.config.Host {
		userCountKey := "user_count"
		statusCountKey := "status_count"
		domainCountKey := "domain_count"

		userCount, err := c.db.GetUserCountForInstance(c.config.Host)
		if err == nil {
			mi.Stats[userCountKey] = userCount
		}

		statusCount, err := c.db.GetStatusCountForInstance(c.config.Host)
		if err == nil {
			mi.Stats[statusCountKey] = statusCount
		}

		domainCount, err := c.db.GetDomainCountForInstance(c.config.Host)
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
	}aaaaaaaaaaaaaaaaaaaaaaaaaaaa

	// contact account is optional but let's try to get it
	if i.ContactAccountID != "" {
		ia := &gtsmodel.Account{}
		if err := c.db.GetByID(i.ContactAccountID, ia); err == nil {
			ma, err := c.AccountToMastoPublic(ia)
			if err == nil {
				mi.ContactAccount = ma
			}
		}
	}

	return mi, nil
}

func (c *converter) RelationshipToMasto(r *gtsmodel.Relationship) (*model.Relationship, error) {
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

func (c *converter) NotificationToMasto(n *gtsmodel.Notification) (*model.Notification, error) {

	if n.GTSTargetAccount == nil {
		tAccount := &gtsmodel.Account{}
		if err := c.db.GetByID(n.TargetAccountID, tAccount); err != nil {
			return nil, fmt.Errorf("NotificationToMasto: error getting target account with id %s from the db: %s", n.TargetAccountID, err)
		}
		n.GTSTargetAccount = tAccount
	}

	if n.GTSOriginAccount == nil {
		ogAccount := &gtsmodel.Account{}
		if err := c.db.GetByID(n.OriginAccountID, ogAccount); err != nil {
			return nil, fmt.Errorf("NotificationToMasto: error getting origin account with id %s from the db: %s", n.OriginAccountID, err)
		}
		n.GTSOriginAccount = ogAccount
	}
	mastoAccount, err := c.AccountToMastoPublic(n.GTSOriginAccount)
	if err != nil {
		return nil, fmt.Errorf("NotificationToMasto: error converting account to masto: %s", err)
	}

	var mastoStatus *model.Status
	if n.StatusID != "" {
		if n.GTSStatus == nil {
			status := &gtsmodel.Status{}
			if err := c.db.GetByID(n.StatusID, status); err != nil {
				return nil, fmt.Errorf("NotificationToMasto: error getting status with id %s from the db: %s", n.StatusID, err)
			}
			n.GTSStatus = status
		}

		if n.GTSStatus.GTSAuthorAccount == nil {
			if n.GTSStatus.AccountID == n.GTSTargetAccount.ID {
				n.GTSStatus.GTSAuthorAccount = n.GTSTargetAccount
			} else if n.GTSStatus.AccountID == n.GTSOriginAccount.ID {
				n.GTSStatus.GTSAuthorAccount = n.GTSOriginAccount
			}
		}

		var err error
		mastoStatus, err = c.StatusToMasto(n.GTSStatus, nil)
		if err != nil {
			return nil, fmt.Errorf("NotificationToMasto: error converting status to masto: %s", err)
		}
	}

	return &model.Notification{
		ID:        n.ID,
		Type:      string(n.NotificationType),
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		Account:   mastoAccount,
		Status:    mastoStatus,
	}, nil
}
