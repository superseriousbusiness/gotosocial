package typeutils

import (
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (c *converter) FollowRequestToFollow(f *gtsmodel.FollowRequest) *gtsmodel.Follow {
	return &gtsmodel.Follow{
		ID:              f.ID,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
		AccountID:       f.AccountID,
		TargetAccountID: f.TargetAccountID,
		ShowReblogs:     f.ShowReblogs,
		URI:             f.URI,
		Notify:          f.Notify,
	}
}

func (c *converter) StatusToBoost(s *gtsmodel.Status, boostingAccount *gtsmodel.Account) (*gtsmodel.Status, error) {
	// the wrapper won't use the same ID as the boosted status so we generate some new UUIDs
	uris := util.GenerateURIsForAccount(boostingAccount.Username, c.config.Protocol, c.config.Host)
	boostWrapperStatusID, err := id.NewULID()
	if err != nil {
		return nil, err
	}
	boostWrapperStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, boostWrapperStatusID)
	boostWrapperStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, boostWrapperStatusID)

	local := true
	if boostingAccount.Domain != "" {
		local = false
	}

	boostWrapperStatus := &gtsmodel.Status{
		ID:  boostWrapperStatusID,
		URI: boostWrapperStatusURI,
		URL: boostWrapperStatusURL,

		// the boosted status is not created now, but the boost certainly is
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Local:     local,
		AccountID: boostingAccount.ID,

		// replies can be boosted, but boosts are never replies
		InReplyToID:        "",
		InReplyToAccountID: "",

		// these will all be wrapped in the boosted status so set them empty here
		Attachments: []string{},
		Tags:        []string{},
		Mentions:    []string{},
		Emojis:      []string{},

		// the below fields will be taken from the target status
		Content:             s.Content,
		ContentWarning:      s.ContentWarning,
		ActivityStreamsType: s.ActivityStreamsType,
		Sensitive:           s.Sensitive,
		Language:            s.Language,
		Text:                s.Text,
		BoostOfID:           s.ID,
		Visibility:          s.Visibility,
		VisibilityAdvanced:  s.VisibilityAdvanced,

		// attach these here for convenience -- the boosted status/account won't go in the DB
		// but they're needed in the processor and for the frontend. Since we have them, we can
		// attach them so we don't need to fetch them again later (save some DB calls)
		GTSBoostedStatus: s,
	}

	return boostWrapperStatus, nil
}
