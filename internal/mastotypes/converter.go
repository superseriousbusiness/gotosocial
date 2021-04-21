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

package mastotypes

import (
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	mastotypes "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Converter is an interface for the common action of converting between mastotypes (frontend, serializable) models and internal gts models used in the database.
// It requires access to the database because many of the conversions require pulling out database entries and counting them etc.
type Converter interface {
	// AccountToMastoSensitive takes a db model account as a param, and returns a populated mastotype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may have sensitive fields,
	// so serve it only to an authorized user who should have permission to see it.
	AccountToMastoSensitive(account *gtsmodel.Account) (*mastotypes.Account, error)

	// AccountToMastoPublic takes a db model account as a param, and returns a populated mastotype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may NOT have sensitive fields.
	// In other words, this is the public record that the server has of an account.
	AccountToMastoPublic(account *gtsmodel.Account) (*mastotypes.Account, error)

	// AppToMastoSensitive takes a db model application as a param, and returns a populated mastotype application, or an error
	// if something goes wrong. The returned application should be ready to serialize on an API level, and may have sensitive fields
	// (such as client id and client secret), so serve it only to an authorized user who should have permission to see it.
	AppToMastoSensitive(application *gtsmodel.Application) (*mastotypes.Application, error)

	// AppToMastoPublic takes a db model application as a param, and returns a populated mastotype application, or an error
	// if something goes wrong. The returned application should be ready to serialize on an API level, and has sensitive
	// fields sanitized so that it can be served to non-authorized accounts without revealing any private information.
	AppToMastoPublic(application *gtsmodel.Application) (*mastotypes.Application, error)

	// AttachmentToMasto converts a gts model media attacahment into its mastodon representation for serialization on the API.
	AttachmentToMasto(attachment *gtsmodel.MediaAttachment) (mastotypes.Attachment, error)

	// MentionToMasto converts a gts model mention into its mastodon (frontend) representation for serialization on the API.
	MentionToMasto(m *gtsmodel.Mention) (mastotypes.Mention, error)

	// EmojiToMasto converts a gts model emoji into its mastodon (frontend) representation for serialization on the API.
	EmojiToMasto(e *gtsmodel.Emoji) (mastotypes.Emoji, error)

	// TagToMasto converts a gts model tag into its mastodon (frontend) representation for serialization on the API.
	TagToMasto(t *gtsmodel.Tag) (mastotypes.Tag, error)

	// StatusToMasto converts a gts model status into its mastodon (frontend) representation for serialization on the API.
	StatusToMasto(s *gtsmodel.Status, targetAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account, boostOfAccount *gtsmodel.Account, replyToAccount *gtsmodel.Account, reblogOfStatus *gtsmodel.Status) (*mastotypes.Status, error)
}

type converter struct {
	config *config.Config
	db     db.DB
}

// New returns a new Converter
func New(config *config.Config, db db.DB) Converter {
	return &converter{
		config: config,
		db:     db,
	}
}

func (c *converter) AccountToMastoSensitive(a *gtsmodel.Account) (*mastotypes.Account, error) {
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

	mastoAccount.Source = &mastotypes.Source{
		Privacy:             util.ParseMastoVisFromGTSVis(a.Privacy),
		Sensitive:           a.Sensitive,
		Language:            a.Language,
		Note:                a.Note,
		Fields:              mastoAccount.Fields,
		FollowRequestsCount: frc,
	}

	return mastoAccount, nil
}

func (c *converter) AccountToMastoPublic(a *gtsmodel.Account) (*mastotypes.Account, error) {
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
	fields := []mastotypes.Field{}
	for _, f := range a.Fields {
		mField := mastotypes.Field{
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

	return &mastotypes.Account{
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

func (c *converter) AppToMastoSensitive(a *gtsmodel.Application) (*mastotypes.Application, error) {
	return &mastotypes.Application{
		ID:           a.ID,
		Name:         a.Name,
		Website:      a.Website,
		RedirectURI:  a.RedirectURI,
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		VapidKey:     a.VapidKey,
	}, nil
}

func (c *converter) AppToMastoPublic(a *gtsmodel.Application) (*mastotypes.Application, error) {
	return &mastotypes.Application{
		Name:    a.Name,
		Website: a.Website,
	}, nil
}

func (c *converter) AttachmentToMasto(a *gtsmodel.MediaAttachment) (mastotypes.Attachment, error) {
	return mastotypes.Attachment{
		ID:               a.ID,
		Type:             string(a.Type),
		URL:              a.URL,
		PreviewURL:       a.Thumbnail.URL,
		RemoteURL:        a.RemoteURL,
		PreviewRemoteURL: a.Thumbnail.RemoteURL,
		Meta: mastotypes.MediaMeta{
			Original: mastotypes.MediaDimensions{
				Width:  a.FileMeta.Original.Width,
				Height: a.FileMeta.Original.Height,
				Size:   fmt.Sprintf("%dx%d", a.FileMeta.Original.Width, a.FileMeta.Original.Height),
				Aspect: float32(a.FileMeta.Original.Aspect),
			},
			Small: mastotypes.MediaDimensions{
				Width:  a.FileMeta.Small.Width,
				Height: a.FileMeta.Small.Height,
				Size:   fmt.Sprintf("%dx%d", a.FileMeta.Small.Width, a.FileMeta.Small.Height),
				Aspect: float32(a.FileMeta.Small.Aspect),
			},
			Focus: mastotypes.MediaFocus{
				X: a.FileMeta.Focus.X,
				Y: a.FileMeta.Focus.Y,
			},
		},
		Description: a.Description,
		Blurhash:    a.Blurhash,
	}, nil
}

func (c *converter) MentionToMasto(m *gtsmodel.Mention) (mastotypes.Mention, error) {
	target := &gtsmodel.Account{}
	if err := c.db.GetByID(m.TargetAccountID, target); err != nil {
		return mastotypes.Mention{}, err
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

	return mastotypes.Mention{
		ID:       target.ID,
		Username: target.Username,
		URL:      target.URL,
		Acct:     acct,
	}, nil
}

func (c *converter) EmojiToMasto(e *gtsmodel.Emoji) (mastotypes.Emoji, error) {
	return mastotypes.Emoji{
		Shortcode:       e.Shortcode,
		URL:             e.ImageURL,
		StaticURL:       e.ImageStaticURL,
		VisibleInPicker: e.VisibleInPicker,
		Category:        e.CategoryID,
	}, nil
}

func (c *converter) TagToMasto(t *gtsmodel.Tag) (mastotypes.Tag, error) {
	tagURL := fmt.Sprintf("%s://%s/tags/%s", c.config.Protocol, c.config.Host, t.Name)

	return mastotypes.Tag{
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
	reblogOfStatus *gtsmodel.Status) (*mastotypes.Status, error) {

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

	var mastoRebloggedStatus *mastotypes.Status
	if s.BoostOfID != "" {
		// the boosted status might have been set on this struct already so check first before doing db calls
		var gtsBoostedStatus *gtsmodel.Status
		if s.GTSBoostedStatus != nil {
			// it's set, great!
			gtsBoostedStatus = s.GTSBoostedStatus
		} else {
			// it's not set so fetch it from the db
			gtsBoostedStatus = &gtsmodel.Status{}
			if err := c.db.GetByID(s.BoostOfID, gtsBoostedStatus); err != nil {
				return nil, fmt.Errorf("error getting boosted status with id %s: %s", s.BoostOfID, err)
			}
		}

		// the boosted account might have been set on this struct already or passed as a param so check first before doing db calls
		var gtsBoostedAccount *gtsmodel.Account
		if s.GTSBoostedAccount != nil {
			// it's set, great!
			gtsBoostedAccount = s.GTSBoostedAccount
		} else if boostOfAccount != nil {
			// it's been given as a param, great!
			gtsBoostedAccount = boostOfAccount
		} else if boostOfAccount == nil && s.GTSBoostedAccount == nil {
			// it's not set so fetch it from the db
			gtsBoostedAccount = &gtsmodel.Account{}
			if err := c.db.GetByID(gtsBoostedStatus.AccountID, gtsBoostedAccount); err != nil {
				return nil, fmt.Errorf("error getting boosted account %s from status with id %s: %s", gtsBoostedStatus.AccountID, s.BoostOfID, err)
			}
		}

		// the boosted status might be a reply so check this
		var gtsBoostedReplyToAccount *gtsmodel.Account
		if gtsBoostedStatus.InReplyToAccountID != "" {
			gtsBoostedReplyToAccount = &gtsmodel.Account{}
			if err := c.db.GetByID(gtsBoostedStatus.InReplyToAccountID, gtsBoostedReplyToAccount); err != nil {
				return nil, fmt.Errorf("error getting account that boosted status was a reply to: %s", err)
			}
		}

		if gtsBoostedStatus != nil || gtsBoostedAccount != nil {
			mastoRebloggedStatus, err = c.StatusToMasto(gtsBoostedStatus, gtsBoostedAccount, requestingAccount, nil, gtsBoostedReplyToAccount, nil)
			if err != nil {
				return nil, fmt.Errorf("error converting boosted status to mastotype: %s", err)
			}
		} else {
			return nil, fmt.Errorf("boost of id was set to %s but that status or account was nil", s.BoostOfID)
		}
	}

	var mastoApplication *mastotypes.Application
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

	mastoAttachments := []mastotypes.Attachment{}
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

	mastoMentions := []mastotypes.Mention{}
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

	mastoTags := []mastotypes.Tag{}
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

	mastoEmojis := []mastotypes.Emoji{}
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

	var mastoCard *mastotypes.Card
	var mastoPoll *mastotypes.Poll

	return &mastotypes.Status{
		ID:                 s.ID,
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
		InReplyToID:        s.InReplyToID,
		InReplyToAccountID: s.InReplyToAccountID,
		Sensitive:          s.Sensitive,
		SpoilerText:        s.ContentWarning,
		Visibility:         util.ParseMastoVisFromGTSVis(s.Visibility),
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
