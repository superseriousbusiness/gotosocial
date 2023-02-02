package typeutils

import (
	"context"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (c *converter) FollowRequestToFollow(ctx context.Context, f *gtsmodel.FollowRequest) *gtsmodel.Follow {
	showReblogs := *f.ShowReblogs
	notify := *f.Notify
	return &gtsmodel.Follow{
		ID:              f.ID,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
		AccountID:       f.AccountID,
		TargetAccountID: f.TargetAccountID,
		ShowReblogs:     &showReblogs,
		URI:             f.URI,
		Notify:          &notify,
	}
}

func (c *converter) StatusToBoost(ctx context.Context, s *gtsmodel.Status, boostingAccount *gtsmodel.Account) (*gtsmodel.Status, error) {
	// the wrapper won't use the same ID as the boosted status so we generate some new UUIDs
	accountURIs := uris.GenerateURIsForAccount(boostingAccount.Username)
	boostWrapperStatusID := id.NewULID()
	boostWrapperStatusURI := accountURIs.StatusesURI + "/" + boostWrapperStatusID
	boostWrapperStatusURL := accountURIs.StatusesURL + "/" + boostWrapperStatusID

	local := true
	if boostingAccount.Domain != "" {
		local = false
	}

	sensitive := *s.Sensitive
	pinned := false // can't pin a boost
	federated := *s.Federated
	boostable := *s.Boostable
	replyable := *s.Replyable
	likeable := *s.Likeable

	boostWrapperStatus := &gtsmodel.Status{
		ID:  boostWrapperStatusID,
		URI: boostWrapperStatusURI,
		URL: boostWrapperStatusURL,

		// the boosted status is not created now, but the boost certainly is
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Local:      &local,
		AccountID:  boostingAccount.ID,
		AccountURI: boostingAccount.URI,

		// replies can be boosted, but boosts are never replies
		InReplyToID:        "",
		InReplyToAccountID: "",

		// these will all be wrapped in the boosted status so set them empty here
		AttachmentIDs: []string{},
		TagIDs:        []string{},
		MentionIDs:    []string{},
		EmojiIDs:      []string{},

		// the below fields will be taken from the target status
		Content:             s.Content,
		ContentWarning:      s.ContentWarning,
		ActivityStreamsType: s.ActivityStreamsType,
		Sensitive:           &sensitive,
		Language:            s.Language,
		Text:                s.Text,
		BoostOfID:           s.ID,
		BoostOfAccountID:    s.AccountID,
		Visibility:          s.Visibility,
		Pinned:              &pinned,
		Federated:           &federated,
		Boostable:           &boostable,
		Replyable:           &replyable,
		Likeable:            &likeable,

		// attach these here for convenience -- the boosted status/account won't go in the DB
		// but they're needed in the processor and for the frontend. Since we have them, we can
		// attach them so we don't need to fetch them again later (save some DB calls)
		BoostOf: s,
	}

	return boostWrapperStatus, nil
}
