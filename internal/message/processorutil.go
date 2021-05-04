package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) processVisibility(form *model.AdvancedStatusCreateForm, accountDefaultVis gtsmodel.Visibility, status *gtsmodel.Status) error {
	// by default all flags are set to true
	gtsAdvancedVis := &gtsmodel.VisibilityAdvanced{
		Federated: true,
		Boostable: true,
		Replyable: true,
		Likeable:  true,
	}

	var gtsBasicVis gtsmodel.Visibility
	// Advanced takes priority if it's set.
	// If it's not set, take whatever masto visibility is set.
	// If *that's* not set either, then just take the account default.
	// If that's also not set, take the default for the whole instance.
	if form.VisibilityAdvanced != nil {
		gtsBasicVis = gtsmodel.Visibility(*form.VisibilityAdvanced)
	} else if form.Visibility != "" {
		gtsBasicVis = p.tc.MastoVisToVis(form.Visibility)
	} else if accountDefaultVis != "" {
		gtsBasicVis = accountDefaultVis
	} else {
		gtsBasicVis = gtsmodel.VisibilityDefault
	}

	switch gtsBasicVis {
	case gtsmodel.VisibilityPublic:
		// for public, there's no need to change any of the advanced flags from true regardless of what the user filled out
		break
	case gtsmodel.VisibilityUnlocked:
		// for unlocked the user can set any combination of flags they like so look at them all to see if they're set and then apply them
		if form.Federated != nil {
			gtsAdvancedVis.Federated = *form.Federated
		}

		if form.Boostable != nil {
			gtsAdvancedVis.Boostable = *form.Boostable
		}

		if form.Replyable != nil {
			gtsAdvancedVis.Replyable = *form.Replyable
		}

		if form.Likeable != nil {
			gtsAdvancedVis.Likeable = *form.Likeable
		}

	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		// for followers or mutuals only, boostable will *always* be false, but the other fields can be set so check and apply them
		gtsAdvancedVis.Boostable = false

		if form.Federated != nil {
			gtsAdvancedVis.Federated = *form.Federated
		}

		if form.Replyable != nil {
			gtsAdvancedVis.Replyable = *form.Replyable
		}

		if form.Likeable != nil {
			gtsAdvancedVis.Likeable = *form.Likeable
		}

	case gtsmodel.VisibilityDirect:
		// direct is pretty easy: there's only one possible setting so return it
		gtsAdvancedVis.Federated = true
		gtsAdvancedVis.Boostable = false
		gtsAdvancedVis.Federated = true
		gtsAdvancedVis.Likeable = true
	}

	status.Visibility = gtsBasicVis
	status.VisibilityAdvanced = gtsAdvancedVis
	return nil
}

func (p *processor) processReplyToID(form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) error {
	if form.InReplyToID == "" {
		return nil
	}

	// If this status is a reply to another status, we need to do a bit of work to establish whether or not this status can be posted:
	//
	// 1. Does the replied status exist in the database?
	// 2. Is the replied status marked as replyable?
	// 3. Does a block exist between either the current account or the account that posted the status it's replying to?
	//
	// If this is all OK, then we fetch the repliedStatus and the repliedAccount for later processing.
	repliedStatus := &gtsmodel.Status{}
	repliedAccount := &gtsmodel.Account{}
	// check replied status exists + is replyable
	if err := p.db.GetByID(form.InReplyToID, repliedStatus); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return fmt.Errorf("status with id %s not replyable because it doesn't exist", form.InReplyToID)
		}
		return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
	}

	if !repliedStatus.VisibilityAdvanced.Replyable {
		return fmt.Errorf("status with id %s is marked as not replyable", form.InReplyToID)
	}

	// check replied account is known to us
	if err := p.db.GetByID(repliedStatus.AccountID, repliedAccount); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return fmt.Errorf("status with id %s not replyable because account id %s is not known", form.InReplyToID, repliedStatus.AccountID)
		}
		return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
	}
	// check if a block exists
	if blocked, err := p.db.Blocked(thisAccountID, repliedAccount.ID); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
		}
	} else if blocked {
		return fmt.Errorf("status with id %s not replyable", form.InReplyToID)
	}
	status.InReplyToID = repliedStatus.ID
	status.InReplyToAccountID = repliedAccount.ID

	return nil
}

func (p *processor) processMediaIDs(form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) error {
	if form.MediaIDs == nil {
		return nil
	}

	gtsMediaAttachments := []*gtsmodel.MediaAttachment{}
	attachments := []string{}
	for _, mediaID := range form.MediaIDs {
		// check these attachments exist
		a := &gtsmodel.MediaAttachment{}
		if err := p.db.GetByID(mediaID, a); err != nil {
			return fmt.Errorf("invalid media type or media not found for media id %s", mediaID)
		}
		// check they belong to the requesting account id
		if a.AccountID != thisAccountID {
			return fmt.Errorf("media with id %s does not belong to account %s", mediaID, thisAccountID)
		}
		// check they're not already used in a status
		if a.StatusID != "" || a.ScheduledStatusID != "" {
			return fmt.Errorf("media with id %s is already attached to a status", mediaID)
		}
		gtsMediaAttachments = append(gtsMediaAttachments, a)
		attachments = append(attachments, a.ID)
	}
	status.GTSMediaAttachments = gtsMediaAttachments
	status.Attachments = attachments
	return nil
}

func (p *processor) processLanguage(form *apimodel.AdvancedStatusCreateForm, accountDefaultLanguage string, status *gtsmodel.Status) error {
	if form.Language != "" {
		status.Language = form.Language
	} else {
		status.Language = accountDefaultLanguage
	}
	if status.Language == "" {
		return errors.New("no language given either in status create form or account default")
	}
	return nil
}

func (p *processor) processMentions(form *apimodel.AdvancedStatusCreateForm, accountID string, status *gtsmodel.Status) error {
	menchies := []string{}
	gtsMenchies, err := p.db.MentionStringsToMentions(util.DeriveMentions(form.Status), accountID, status.ID)
	if err != nil {
		return fmt.Errorf("error generating mentions from status: %s", err)
	}
	for _, menchie := range gtsMenchies {
		if err := p.db.Put(menchie); err != nil {
			return fmt.Errorf("error putting mentions in db: %s", err)
		}
		menchies = append(menchies, menchie.TargetAccountID)
	}
	// add full populated gts menchies to the status for passing them around conveniently
	status.GTSMentions = gtsMenchies
	// add just the ids of the mentioned accounts to the status for putting in the db
	status.Mentions = menchies
	return nil
}

func (p *processor) processTags(form *apimodel.AdvancedStatusCreateForm, accountID string, status *gtsmodel.Status) error {
	tags := []string{}
	gtsTags, err := p.db.TagStringsToTags(util.DeriveHashtags(form.Status), accountID, status.ID)
	if err != nil {
		return fmt.Errorf("error generating hashtags from status: %s", err)
	}
	for _, tag := range gtsTags {
		if err := p.db.Upsert(tag, "name"); err != nil {
			return fmt.Errorf("error putting tags in db: %s", err)
		}
		tags = append(tags, tag.ID)
	}
	// add full populated gts tags to the status for passing them around conveniently
	status.GTSTags = gtsTags
	// add just the ids of the used tags to the status for putting in the db
	status.Tags = tags
	return nil
}

func (p *processor) processEmojis(form *apimodel.AdvancedStatusCreateForm, accountID string, status *gtsmodel.Status) error {
	emojis := []string{}
	gtsEmojis, err := p.db.EmojiStringsToEmojis(util.DeriveEmojis(form.Status), accountID, status.ID)
	if err != nil {
		return fmt.Errorf("error generating emojis from status: %s", err)
	}
	for _, e := range gtsEmojis {
		emojis = append(emojis, e.ID)
	}
	// add full populated gts emojis to the status for passing them around conveniently
	status.GTSEmojis = gtsEmojis
	// add just the ids of the used emojis to the status for putting in the db
	status.Emojis = emojis
	return nil
}

/*
	HELPER FUNCTIONS
*/

// TODO: try to combine the below two functions because this is a lot of code repetition.

// updateAccountAvatar does the dirty work of checking the avatar part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new avatar image.
func (p *processor) updateAccountAvatar(avatar *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	var err error
	if int(avatar.Size) > p.config.MediaConfig.MaxImageSize {
		err = fmt.Errorf("avatar with size %d exceeded max image size of %d bytes", avatar.Size, p.config.MediaConfig.MaxImageSize)
		return nil, err
	}
	f, err := avatar.Open()
	if err != nil {
		return nil, fmt.Errorf("could not read provided avatar: %s", err)
	}

	// extract the bytes
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("could not read provided avatar: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided avatar: size 0 bytes")
	}

	// do the setting
	avatarInfo, err := p.mediaHandler.ProcessHeaderOrAvatar(buf.Bytes(), accountID, media.MediaAvatar)
	if err != nil {
		return nil, fmt.Errorf("error processing avatar: %s", err)
	}

	return avatarInfo, f.Close()
}

// updateAccountHeader does the dirty work of checking the header part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new header image.
func (p *processor) updateAccountHeader(header *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	var err error
	if int(header.Size) > p.config.MediaConfig.MaxImageSize {
		err = fmt.Errorf("header with size %d exceeded max image size of %d bytes", header.Size, p.config.MediaConfig.MaxImageSize)
		return nil, err
	}
	f, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("could not read provided header: %s", err)
	}

	// extract the bytes
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("could not read provided header: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided header: size 0 bytes")
	}

	// do the setting
	headerInfo, err := p.mediaHandler.ProcessHeaderOrAvatar(buf.Bytes(), accountID, media.MediaHeader)
	if err != nil {
		return nil, fmt.Errorf("error processing header: %s", err)
	}

	return headerInfo, f.Close()
}
