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

package cache

import (
	"crypto/rsa"
	"strings"
	"time"
	"unsafe"

	"codeberg.org/gruf/go-cache/v3/simple"
	"github.com/DmitriyVTitov/size"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

const (
	// example data values.
	exampleID   = id.Highest
	exampleURI  = "https://social.bbc/users/ItsMePrinceCharlesInit"
	exampleText = `
oh no me nan's gone and done it :shocked:
	
she fuckin killed the king :regicide:

nan what have you done :shocked:

no nan put down the knife, don't go after the landlords next! :knife:

you'll make society more equitable for all if you're not careful! :hammer_sickle:

#JustNanProblems #WhatWillSheDoNext #MaybeItWasntSuchABadThingAfterAll
`

	exampleTextSmall = "Small problem lads, me nan's gone on a bit of a rampage"
	exampleUsername  = "@SexHaver1969"

	// ID string size in memory (is always 26 char ULID).
	sizeofIDStr = unsafe.Sizeof(exampleID)

	// URI string size in memory (use some random example URI).
	sizeofURIStr = unsafe.Sizeof(exampleURI)

	// ID slice size in memory (using some estimate of length = 250).
	sizeofIDSlice = unsafe.Sizeof([]string{}) + 250*sizeofIDStr

	// result cache key size estimate which is tricky. it can
	// be a serialized string of almost any type, so we pick a
	// nice serialized key size on the upper end of normal.
	sizeofResultKey = 2 * sizeofIDStr
)

var (
	// Example time calculated at ~ 14th August, 2023. Because if
	// we use `time.Now()` in our structs below, it populates
	// them with locale data which throws-off size calculations.
	//
	// This is because the locale data is (relatively) very large
	// in-memory, but it's global "singletons" ptr'd to by the time
	// structs, so inconsequential to our calculated cache size.
	// Unfortunately the size.Of() function is not aware of this!
	exampleTime = time.Time{}.Add(1692010328 * time.Second)

	// stop trying to collapse this var
	// block, gofmt, you motherfucker.
	_ = interface{}(nil)
)

// calculateSliceCacheMax calculates the maximum capacity for a slice cache with given individual ratio.
func calculateSliceCacheMax(ratio float64) int {
	return calculateCacheMax(sizeofIDStr, sizeofIDSlice, ratio)
}

// calculateResultCacheMax calculates the maximum cache capacity for a result
// cache's individual ratio number, and the size of the struct model in memory.
func calculateResultCacheMax(structSz uintptr, ratio float64) int {
	// Estimate a worse-case scenario of extra lookup hash maps,
	// where lookups are the no. "keys" each result can be found under
	const lookups = 10

	// Calculate the extra cache lookup map overheads.
	totalLookupKeySz := uintptr(lookups) * sizeofResultKey
	totalLookupValSz := uintptr(lookups) * unsafe.Sizeof(uint64(0))

	// Primary cache sizes.
	pkeySz := unsafe.Sizeof(uint64(0))
	pvalSz := structSz

	// The result cache wraps each struct result in a wrapping
	// struct with further information, and possible error. This
	// also needs to be taken into account when calculating value.
	resultValueOverhead := uintptr(size.Of(&struct {
		_ int64
		_ []any
		_ any
		_ error
	}{}))

	return calculateCacheMax(
		pkeySz+totalLookupKeySz,
		pvalSz+totalLookupValSz+resultValueOverhead,
		ratio,
	)
}

// calculateCacheMax calculates the maximum cache capacity for a cache's
// individual ratio number, and key + value object sizes in memory.
func calculateCacheMax(keySz, valSz uintptr, ratio float64) int {
	if ratio < 0 {
		// Negative ratios are a secret little trick
		// to manually set the cache capacity sizes.
		return int(-1 * ratio)
	}

	// see: https://golang.org/src/runtime/map.go
	const emptyBucketOverhead = 10.79

	// This takes into account (roughly) that the underlying simple cache library wraps
	// elements within a simple.Entry{}, and the ordered map wraps each in a linked list elem.
	const cacheElemOverhead = unsafe.Sizeof(simple.Entry{}) + unsafe.Sizeof(struct {
		key, value interface{}
		next, prev uintptr
	}{})

	// The inputted memory ratio does not take into account the
	// total of all ratios, so divide it here to get perc. ratio.
	totalRatio := ratio / totalOfRatios()

	// TODO: we should also further weight this ratio depending
	// on the combined keySz + valSz as a ratio of all available
	// cache model memories. otherwise you can end up with a
	// low-ratio cache of tiny models with larger capacity than
	// a high-ratio cache of large models.

	// Get max available cache memory, calculating max for
	// this cache by multiplying by this cache's mem ratio.
	maxMem := config.GetCacheMemoryTarget()
	fMaxMem := float64(maxMem) * totalRatio

	// Cast to useable types.
	fKeySz := float64(keySz)
	fValSz := float64(valSz)

	// Calculated using the internal cache map size:
	// (($keysz + $valsz) * $len) + ($len * $allOverheads) = $memSz
	return int(fMaxMem / (fKeySz + fValSz + emptyBucketOverhead + float64(cacheElemOverhead)))
}

// totalOfRatios returns the total of all cache ratios added together.
func totalOfRatios() float64 {

	// NOTE: this is not performant calculating
	// this every damn time (mainly the mutex unlocks
	// required to access each config var). fortunately
	// we only do this on init so fuck it :D
	return 0 +
		config.GetCacheAccountMemRatio() +
		config.GetCacheAccountNoteMemRatio() +
		config.GetCacheAccountSettingsMemRatio() +
		config.GetCacheAccountStatsMemRatio() +
		config.GetCacheApplicationMemRatio() +
		config.GetCacheBlockMemRatio() +
		config.GetCacheBlockIDsMemRatio() +
		config.GetCacheBoostOfIDsMemRatio() +
		config.GetCacheClientMemRatio() +
		config.GetCacheEmojiMemRatio() +
		config.GetCacheEmojiCategoryMemRatio() +
		config.GetCacheFilterMemRatio() +
		config.GetCacheFilterKeywordMemRatio() +
		config.GetCacheFilterStatusMemRatio() +
		config.GetCacheFollowMemRatio() +
		config.GetCacheFollowIDsMemRatio() +
		config.GetCacheFollowRequestMemRatio() +
		config.GetCacheFollowRequestIDsMemRatio() +
		config.GetCacheFollowingTagIDsMemRatio() +
		config.GetCacheInReplyToIDsMemRatio() +
		config.GetCacheInstanceMemRatio() +
		config.GetCacheInteractionRequestMemRatio() +
		config.GetCacheListMemRatio() +
		config.GetCacheListIDsMemRatio() +
		config.GetCacheListedIDsMemRatio() +
		config.GetCacheMarkerMemRatio() +
		config.GetCacheMediaMemRatio() +
		config.GetCacheMentionMemRatio() +
		config.GetCacheMoveMemRatio() +
		config.GetCacheNotificationMemRatio() +
		config.GetCachePollMemRatio() +
		config.GetCachePollVoteMemRatio() +
		config.GetCachePollVoteIDsMemRatio() +
		config.GetCacheReportMemRatio() +
		config.GetCacheSinBinStatusMemRatio() +
		config.GetCacheStatusMemRatio() +
		config.GetCacheStatusBookmarkMemRatio() +
		config.GetCacheStatusBookmarkIDsMemRatio() +
		config.GetCacheStatusFaveMemRatio() +
		config.GetCacheStatusFaveIDsMemRatio() +
		config.GetCacheTagMemRatio() +
		config.GetCacheThreadMuteMemRatio() +
		config.GetCacheTokenMemRatio() +
		config.GetCacheTombstoneMemRatio() +
		config.GetCacheUserMemRatio() +
		config.GetCacheUserMuteMemRatio() +
		config.GetCacheUserMuteIDsMemRatio() +
		config.GetCacheWebfingerMemRatio() +
		config.GetCacheVisibilityMemRatio()
}

func sizeofAccount() uintptr {
	return uintptr(size.Of(&gtsmodel.Account{
		ID:                      exampleID,
		Username:                exampleUsername,
		AvatarMediaAttachmentID: exampleID,
		HeaderMediaAttachmentID: exampleID,
		DisplayName:             exampleUsername,
		Note:                    exampleText,
		NoteRaw:                 exampleText,
		Memorial:                func() *bool { ok := false; return &ok }(),
		CreatedAt:               exampleTime,
		UpdatedAt:               exampleTime,
		FetchedAt:               exampleTime,
		Bot:                     func() *bool { ok := true; return &ok }(),
		Locked:                  func() *bool { ok := true; return &ok }(),
		Discoverable:            func() *bool { ok := false; return &ok }(),
		URI:                     exampleURI,
		URL:                     exampleURI,
		InboxURI:                exampleURI,
		OutboxURI:               exampleURI,
		FollowersURI:            exampleURI,
		FollowingURI:            exampleURI,
		FeaturedCollectionURI:   exampleURI,
		ActorType:               ap.ActorPerson,
		PrivateKey:              &rsa.PrivateKey{},
		PublicKey:               &rsa.PublicKey{},
		PublicKeyURI:            exampleURI,
		SensitizedAt:            exampleTime,
		SilencedAt:              exampleTime,
		SuspendedAt:             exampleTime,
		SuspensionOrigin:        exampleID,
	}))
}

func sizeofAccountNote() uintptr {
	return uintptr(size.Of(&gtsmodel.AccountNote{
		ID:              exampleID,
		AccountID:       exampleID,
		TargetAccountID: exampleID,
		Comment:         exampleTextSmall,
	}))
}

func sizeofAccountSettings() uintptr {
	return uintptr(size.Of(&gtsmodel.AccountSettings{
		AccountID:         exampleID,
		CreatedAt:         exampleTime,
		UpdatedAt:         exampleTime,
		Privacy:           gtsmodel.VisibilityFollowersOnly,
		Sensitive:         util.Ptr(true),
		Language:          "fr",
		StatusContentType: "text/plain",
		CustomCSS:         exampleText,
		EnableRSS:         util.Ptr(true),
		HideCollections:   util.Ptr(false),
	}))
}

func sizeofAccountStats() uintptr {
	return uintptr(size.Of(&gtsmodel.AccountStats{
		AccountID:           exampleID,
		FollowersCount:      util.Ptr(100),
		FollowingCount:      util.Ptr(100),
		StatusesCount:       util.Ptr(100),
		StatusesPinnedCount: util.Ptr(100),
		LastStatusAt:        exampleTime,
	}))
}

func sizeofApplication() uintptr {
	return uintptr(size.Of(&gtsmodel.Application{
		ID:           exampleID,
		CreatedAt:    exampleTime,
		UpdatedAt:    exampleTime,
		Name:         exampleUsername,
		Website:      exampleURI,
		RedirectURI:  exampleURI,
		ClientID:     exampleID,
		ClientSecret: exampleID,
		Scopes:       exampleTextSmall,
	}))
}

func sizeofBlock() uintptr {
	return uintptr(size.Of(&gtsmodel.Block{
		ID:              exampleID,
		CreatedAt:       exampleTime,
		UpdatedAt:       exampleTime,
		URI:             exampleURI,
		AccountID:       exampleID,
		TargetAccountID: exampleID,
	}))
}

func sizeofClient() uintptr {
	return uintptr(size.Of(&gtsmodel.Client{
		ID:        exampleID,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		Secret:    exampleID,
		Domain:    exampleURI,
		UserID:    exampleID,
	}))
}

func sizeofConversation() uintptr {
	return uintptr(size.Of(&gtsmodel.Conversation{
		ID:               exampleID,
		CreatedAt:        exampleTime,
		UpdatedAt:        exampleTime,
		AccountID:        exampleID,
		OtherAccountIDs:  []string{exampleID, exampleID, exampleID},
		OtherAccountsKey: strings.Join([]string{exampleID, exampleID, exampleID}, ","),
		ThreadID:         exampleID,
		LastStatusID:     exampleID,
		Read:             util.Ptr(true),
	}))
}

func sizeofEmoji() uintptr {
	return uintptr(size.Of(&gtsmodel.Emoji{
		ID:                     exampleID,
		Shortcode:              exampleTextSmall,
		Domain:                 exampleURI,
		CreatedAt:              exampleTime,
		UpdatedAt:              exampleTime,
		ImageRemoteURL:         exampleURI,
		ImageStaticRemoteURL:   exampleURI,
		ImageURL:               exampleURI,
		ImagePath:              exampleURI,
		ImageStaticURL:         exampleURI,
		ImageStaticPath:        exampleURI,
		ImageContentType:       "image/png",
		ImageStaticContentType: "image/png",
		Disabled:               func() *bool { ok := false; return &ok }(),
		URI:                    "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
		VisibleInPicker:        func() *bool { ok := true; return &ok }(),
		CategoryID:             "01GGQ8V4993XK67B2JB396YFB7",
		Cached:                 func() *bool { ok := true; return &ok }(),
	}))
}

func sizeofEmojiCategory() uintptr {
	return uintptr(size.Of(&gtsmodel.EmojiCategory{
		ID:        exampleID,
		Name:      exampleUsername,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
	}))
}

func sizeofFilter() uintptr {
	return uintptr(size.Of(&gtsmodel.Filter{
		ID:        exampleID,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		ExpiresAt: exampleTime,
		AccountID: exampleID,
		Title:     exampleTextSmall,
		Action:    gtsmodel.FilterActionHide,
	}))
}

func sizeofFilterKeyword() uintptr {
	return uintptr(size.Of(&gtsmodel.FilterKeyword{
		ID:        exampleID,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		FilterID:  exampleID,
		Keyword:   exampleTextSmall,
	}))
}

func sizeofFilterStatus() uintptr {
	return uintptr(size.Of(&gtsmodel.FilterStatus{
		ID:        exampleID,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		FilterID:  exampleID,
		StatusID:  exampleID,
	}))
}

func sizeofFollow() uintptr {
	return uintptr(size.Of(&gtsmodel.Follow{
		ID:              exampleID,
		CreatedAt:       exampleTime,
		UpdatedAt:       exampleTime,
		AccountID:       exampleID,
		TargetAccountID: exampleID,
		ShowReblogs:     func() *bool { ok := true; return &ok }(),
		URI:             exampleURI,
		Notify:          func() *bool { ok := false; return &ok }(),
	}))
}

func sizeofFollowRequest() uintptr {
	return uintptr(size.Of(&gtsmodel.FollowRequest{
		ID:              exampleID,
		CreatedAt:       exampleTime,
		UpdatedAt:       exampleTime,
		AccountID:       exampleID,
		TargetAccountID: exampleID,
		ShowReblogs:     func() *bool { ok := true; return &ok }(),
		URI:             exampleURI,
		Notify:          func() *bool { ok := false; return &ok }(),
	}))
}

func sizeofInstance() uintptr {
	return uintptr(size.Of(&gtsmodel.Instance{
		ID:                     exampleID,
		CreatedAt:              exampleTime,
		UpdatedAt:              exampleTime,
		Domain:                 exampleURI,
		URI:                    exampleURI,
		Title:                  exampleTextSmall,
		ShortDescription:       exampleText,
		Description:            exampleText,
		ContactEmail:           exampleUsername,
		ContactAccountUsername: exampleUsername,
		ContactAccountID:       exampleID,
	}))
}

func sizeofInteractionRequest() uintptr {
	return uintptr(size.Of(&gtsmodel.InteractionRequest{
		ID:                   exampleID,
		CreatedAt:            exampleTime,
		StatusID:             exampleID,
		TargetAccountID:      exampleID,
		InteractingAccountID: exampleID,
		InteractionURI:       exampleURI,
		InteractionType:      gtsmodel.InteractionAnnounce,
		URI:                  exampleURI,
		AcceptedAt:           exampleTime,
	}))
}

func sizeofList() uintptr {
	return uintptr(size.Of(&gtsmodel.List{
		ID:            exampleID,
		CreatedAt:     exampleTime,
		UpdatedAt:     exampleTime,
		Title:         exampleTextSmall,
		AccountID:     exampleID,
		RepliesPolicy: gtsmodel.RepliesPolicyFollowed,
	}))
}

func sizeofMarker() uintptr {
	return uintptr(size.Of(&gtsmodel.Marker{
		AccountID:  exampleID,
		Name:       gtsmodel.MarkerNameHome,
		UpdatedAt:  exampleTime,
		Version:    0,
		LastReadID: exampleID,
	}))
}

func sizeofMedia() uintptr {
	return uintptr(size.Of(&gtsmodel.MediaAttachment{
		ID:                exampleID,
		StatusID:          exampleID,
		URL:               exampleURI,
		RemoteURL:         exampleURI,
		CreatedAt:         exampleTime,
		UpdatedAt:         exampleTime,
		Type:              gtsmodel.FileTypeImage,
		AccountID:         exampleID,
		Description:       exampleText,
		ScheduledStatusID: exampleID,
		Blurhash:          exampleTextSmall,
		File: gtsmodel.File{
			Path:        exampleURI,
			ContentType: "image/jpeg",
		},
		Thumbnail: gtsmodel.Thumbnail{
			Path:        exampleURI,
			ContentType: "image/jpeg",
			URL:         exampleURI,
			RemoteURL:   exampleURI,
		},
		Avatar: func() *bool { ok := false; return &ok }(),
		Header: func() *bool { ok := false; return &ok }(),
		Cached: func() *bool { ok := true; return &ok }(),
	}))
}

func sizeofMention() uintptr {
	return uintptr(size.Of(&gtsmodel.Mention{
		ID:               exampleURI,
		StatusID:         exampleURI,
		CreatedAt:        exampleTime,
		UpdatedAt:        exampleTime,
		OriginAccountID:  exampleURI,
		OriginAccountURI: exampleURI,
		TargetAccountID:  exampleID,
		NameString:       exampleUsername,
		TargetAccountURI: exampleURI,
		TargetAccountURL: exampleURI,
	}))
}

func sizeofMove() uintptr {
	return uintptr(size.Of(&gtsmodel.Move{
		ID:          exampleID,
		CreatedAt:   exampleTime,
		UpdatedAt:   exampleTime,
		AttemptedAt: exampleTime,
		SucceededAt: exampleTime,
		OriginURI:   exampleURI,
		TargetURI:   exampleURI,
		URI:         exampleURI,
	}))
}

func sizeofNotification() uintptr {
	return uintptr(size.Of(&gtsmodel.Notification{
		ID:               exampleID,
		NotificationType: gtsmodel.NotificationFave,
		CreatedAt:        exampleTime,
		TargetAccountID:  exampleID,
		OriginAccountID:  exampleID,
		StatusID:         exampleID,
		Read:             func() *bool { ok := false; return &ok }(),
	}))
}

func sizeofPoll() uintptr {
	return uintptr(size.Of(&gtsmodel.Poll{
		ID:         exampleID,
		Multiple:   func() *bool { ok := false; return &ok }(),
		HideCounts: func() *bool { ok := false; return &ok }(),
		Options:    []string{exampleTextSmall, exampleTextSmall, exampleTextSmall, exampleTextSmall},
		StatusID:   exampleID,
		ExpiresAt:  exampleTime,
	}))
}

func sizeofPollVote() uintptr {
	return uintptr(size.Of(&gtsmodel.PollVote{
		ID:        exampleID,
		Choices:   []int{69, 420, 1337},
		AccountID: exampleID,
		PollID:    exampleID,
		CreatedAt: exampleTime,
	}))
}

func sizeofReport() uintptr {
	return uintptr(size.Of(&gtsmodel.Report{
		ID:                     exampleID,
		CreatedAt:              exampleTime,
		UpdatedAt:              exampleTime,
		URI:                    exampleURI,
		AccountID:              exampleID,
		TargetAccountID:        exampleID,
		Comment:                exampleText,
		StatusIDs:              []string{exampleID, exampleID, exampleID},
		Forwarded:              func() *bool { ok := true; return &ok }(),
		ActionTaken:            exampleText,
		ActionTakenAt:          exampleTime,
		ActionTakenByAccountID: exampleID,
	}))
}

func sizeofSinBinStatus() uintptr {
	return uintptr(size.Of(&gtsmodel.SinBinStatus{
		ID:                  exampleID,
		CreatedAt:           exampleTime,
		UpdatedAt:           exampleTime,
		URI:                 exampleURI,
		URL:                 exampleURI,
		Domain:              exampleURI,
		AccountURI:          exampleURI,
		InReplyToURI:        exampleURI,
		Content:             exampleText,
		AttachmentLinks:     []string{exampleURI, exampleURI},
		MentionTargetURIs:   []string{exampleURI},
		EmojiLinks:          []string{exampleURI},
		PollOptions:         []string{exampleTextSmall, exampleTextSmall, exampleTextSmall, exampleTextSmall},
		ContentWarning:      exampleTextSmall,
		Visibility:          gtsmodel.VisibilityPublic,
		Sensitive:           util.Ptr(false),
		Language:            "en",
		ActivityStreamsType: ap.ObjectNote,
	}))
}

func sizeofStatus() uintptr {
	return uintptr(size.Of(&gtsmodel.Status{
		ID:                       exampleID,
		URI:                      exampleURI,
		URL:                      exampleURI,
		Content:                  exampleText,
		Text:                     exampleText,
		AttachmentIDs:            []string{exampleID, exampleID, exampleID},
		TagIDs:                   []string{exampleID, exampleID, exampleID},
		MentionIDs:               []string{},
		EmojiIDs:                 []string{exampleID, exampleID, exampleID},
		CreatedAt:                exampleTime,
		UpdatedAt:                exampleTime,
		FetchedAt:                exampleTime,
		Local:                    func() *bool { ok := false; return &ok }(),
		AccountURI:               exampleURI,
		AccountID:                exampleID,
		InReplyToID:              exampleID,
		InReplyToURI:             exampleURI,
		InReplyToAccountID:       exampleID,
		BoostOfID:                exampleID,
		BoostOfAccountID:         exampleID,
		ContentWarning:           exampleUsername, // similar length
		Visibility:               gtsmodel.VisibilityPublic,
		Sensitive:                func() *bool { ok := false; return &ok }(),
		Language:                 "en",
		CreatedWithApplicationID: exampleID,
		Federated:                func() *bool { ok := true; return &ok }(),
		ActivityStreamsType:      ap.ObjectNote,
	}))
}

func sizeofStatusBookmark() uintptr {
	return uintptr(size.Of(&gtsmodel.StatusBookmark{
		ID:              exampleID,
		AccountID:       exampleID,
		Account:         nil,
		TargetAccountID: exampleID,
		TargetAccount:   nil,
		StatusID:        exampleID,
		Status:          nil,
		CreatedAt:       exampleTime,
		UpdatedAt:       exampleTime,
	}))
}

func sizeofStatusFave() uintptr {
	return uintptr(size.Of(&gtsmodel.StatusFave{
		ID:              exampleID,
		CreatedAt:       exampleTime,
		AccountID:       exampleID,
		TargetAccountID: exampleID,
		StatusID:        exampleID,
		URI:             exampleURI,
	}))
}

func sizeofTag() uintptr {
	return uintptr(size.Of(&gtsmodel.Tag{
		ID:        exampleID,
		Name:      exampleUsername,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		Useable:   func() *bool { ok := true; return &ok }(),
		Listable:  func() *bool { ok := true; return &ok }(),
	}))
}

func sizeofThreadMute() uintptr {
	return uintptr(size.Of(&gtsmodel.ThreadMute{
		ID:        exampleID,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		ThreadID:  exampleID,
		AccountID: exampleID,
	}))
}

func sizeofToken() uintptr {
	return uintptr(size.Of(&gtsmodel.Token{
		ID:                  exampleID,
		CreatedAt:           exampleTime,
		UpdatedAt:           exampleTime,
		ClientID:            exampleID,
		UserID:              exampleID,
		RedirectURI:         exampleURI,
		Scope:               "r:w",
		Code:                "", // TODO
		CodeChallenge:       "", // TODO
		CodeChallengeMethod: "", // TODO
		CodeCreateAt:        exampleTime,
		CodeExpiresAt:       exampleTime,
		Access:              exampleID + exampleID,
		AccessCreateAt:      exampleTime,
		AccessExpiresAt:     exampleTime,
		Refresh:             "", // TODO: clients don't really support this very well yet
		RefreshCreateAt:     exampleTime,
		RefreshExpiresAt:    exampleTime,
	}))
}

func sizeofTombstone() uintptr {
	return uintptr(size.Of(&gtsmodel.Tombstone{
		ID:        exampleID,
		CreatedAt: exampleTime,
		UpdatedAt: exampleTime,
		Domain:    exampleUsername,
		URI:       exampleURI,
	}))
}

func sizeofVisibility() uintptr {
	return uintptr(size.Of(&CachedVisibility{
		ItemID:      exampleID,
		RequesterID: exampleID,
		Type:        VisibilityTypeAccount,
		Value:       false,
	}))
}

func sizeofUser() uintptr {
	return uintptr(size.Of(&gtsmodel.User{
		ID:                     exampleID,
		CreatedAt:              exampleTime,
		UpdatedAt:              exampleTime,
		Email:                  exampleURI,
		AccountID:              exampleID,
		EncryptedPassword:      exampleTextSmall,
		InviteID:               exampleID,
		Reason:                 exampleText,
		Locale:                 "en",
		CreatedByApplicationID: exampleID,
		LastEmailedAt:          exampleTime,
		ConfirmationToken:      exampleTextSmall,
		ConfirmationSentAt:     exampleTime,
		ConfirmedAt:            exampleTime,
		UnconfirmedEmail:       exampleURI,
		Moderator:              util.Ptr(false),
		Admin:                  util.Ptr(false),
		Disabled:               util.Ptr(false),
		Approved:               util.Ptr(false),
		ResetPasswordToken:     exampleTextSmall,
		ResetPasswordSentAt:    exampleTime,
		ExternalID:             exampleID,
	}))
}

func sizeofUserMute() uintptr {
	return uintptr(size.Of(&gtsmodel.UserMute{
		ID:              exampleID,
		CreatedAt:       exampleTime,
		UpdatedAt:       exampleTime,
		ExpiresAt:       exampleTime,
		AccountID:       exampleID,
		TargetAccountID: exampleID,
		Notifications:   util.Ptr(false),
	}))
}
