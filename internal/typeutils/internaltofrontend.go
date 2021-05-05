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
	if err := c.db.GetFollowersByAccountID(a.ID, &followers); err != nil {
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
	statuses := []gtsmodel.Status{}
	if err := c.db.GetStatusesByAccountID(a.ID, &statuses); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting last statuses: %s", err)
		}
	}
	var statusesCount int
	if statuses != nil {
		statusesCount = len(statuses)
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
	if err := c.db.GetHeaderForAccountID(avi, a.ID); err != nil {
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
		Emojis:         nil, // TODO: implement this
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
		Type:             string(a.Type),
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
		acct = fmt.Sprintf("@%s", target.Username)
	} else {
		acct = fmt.Sprintf("@%s@%s", target.Username, target.Domain)
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

func (c *converter) StatusToMasto(
	s *gtsmodel.Status,
	targetAccount *gtsmodel.Account,
	requestingAccount *gtsmodel.Account,
	boostOfAccount *gtsmodel.Account,
	replyToAccount *gtsmodel.Account,
	reblogOfStatus *gtsmodel.Status) (*model.Status, error) {

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

	var faved bool
	var reblogged bool
	var bookmarked bool
	var pinned bool
	var muted bool

	// requestingAccount will be nil for public requests without auth
	// But if it's not nil, we can also get information about the requestingAccount's interaction with this status
	if requestingAccount != nil {
		faved, err = c.db.StatusFavedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has faved status: %s", err)
		}

		reblogged, err = c.db.StatusRebloggedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has reblogged status: %s", err)
		}

		muted, err = c.db.StatusMutedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has muted status: %s", err)
		}

		bookmarked, err = c.db.StatusBookmarkedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has bookmarked status: %s", err)
		}

		pinned, err = c.db.StatusPinnedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has pinned status: %s", err)
		}
	}

	var mastoRebloggedStatus *model.Status // TODO

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

	mastoTargetAccount, err := c.AccountToMastoPublic(targetAccount)
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
		Favourited:         faved,
		Reblogged:          reblogged,
		Muted:              muted,
		Bookmarked:         bookmarked,
		Pinned:             pinned,
		Content:            s.Content,
		Reblog:             mastoRebloggedStatus,
		Application:        mastoApplication,
		Account:            mastoTargetAccount,
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
