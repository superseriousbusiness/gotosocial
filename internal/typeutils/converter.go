/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"net/url"
	"sync"

	"github.com/gorilla/feeds"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// TypeConverter is an interface for the common action of converting between apimodule (frontend, serializable) models,
// internal gts models used in the database, and activitypub models used in federation.
//
// It requires access to the database because many of the conversions require pulling out database entries and counting them etc.
// That said, it *absolutely should not* manipulate database entries in any way, only examine them.
type TypeConverter interface {
	/*
		INTERNAL (gts) MODEL TO FRONTEND (api) MODEL
	*/

	// AccountToAPIAccountSensitive takes a db model account as a param, and returns a populated apitype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may have sensitive fields,
	// so serve it only to an authorized user who should have permission to see it.
	AccountToAPIAccountSensitive(ctx context.Context, account *gtsmodel.Account) (*model.Account, error)
	// AccountToAPIAccountPublic takes a db model account as a param, and returns a populated apitype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may NOT have sensitive fields.
	// In other words, this is the public record that the server has of an account.
	AccountToAPIAccountPublic(ctx context.Context, account *gtsmodel.Account) (*model.Account, error)
	// AccountToAPIAccountBlocked takes a db model account as a param, and returns a apitype account, or an error if
	// something goes wrong. The returned account will be a bare minimum representation of the account. This function should be used
	// when someone wants to view an account they've blocked.
	AccountToAPIAccountBlocked(ctx context.Context, account *gtsmodel.Account) (*model.Account, error)
	// AppToAPIAppSensitive takes a db model application as a param, and returns a populated apitype application, or an error
	// if something goes wrong. The returned application should be ready to serialize on an API level, and may have sensitive fields
	// (such as client id and client secret), so serve it only to an authorized user who should have permission to see it.
	AppToAPIAppSensitive(ctx context.Context, application *gtsmodel.Application) (*model.Application, error)
	// AppToAPIAppPublic takes a db model application as a param, and returns a populated apitype application, or an error
	// if something goes wrong. The returned application should be ready to serialize on an API level, and has sensitive
	// fields sanitized so that it can be served to non-authorized accounts without revealing any private information.
	AppToAPIAppPublic(ctx context.Context, application *gtsmodel.Application) (*model.Application, error)
	// AttachmentToAPIAttachment converts a gts model media attacahment into its api representation for serialization on the API.
	AttachmentToAPIAttachment(ctx context.Context, attachment *gtsmodel.MediaAttachment) (model.Attachment, error)
	// MentionToAPIMention converts a gts model mention into its api (frontend) representation for serialization on the API.
	MentionToAPIMention(ctx context.Context, m *gtsmodel.Mention) (model.Mention, error)
	// EmojiToAPIEmoji converts a gts model emoji into its api (frontend) representation for serialization on the API.
	EmojiToAPIEmoji(ctx context.Context, e *gtsmodel.Emoji) (model.Emoji, error)
	// EmojiToAdminAPIEmoji converts a gts model emoji into an API representation with extra admin information.
	EmojiToAdminAPIEmoji(ctx context.Context, e *gtsmodel.Emoji) (*model.AdminEmoji, error)
	// EmojiCategoryToAPIEmojiCategory converts a gts model emoji category into its api (frontend) representation.
	EmojiCategoryToAPIEmojiCategory(ctx context.Context, category *gtsmodel.EmojiCategory) (*model.EmojiCategory, error)
	// TagToAPITag converts a gts model tag into its api (frontend) representation for serialization on the API.
	TagToAPITag(ctx context.Context, t *gtsmodel.Tag) (model.Tag, error)
	// StatusToAPIStatus converts a gts model status into its api (frontend) representation for serialization on the API.
	//
	// Requesting account can be nil.
	StatusToAPIStatus(ctx context.Context, s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*model.Status, error)
	// VisToAPIVis converts a gts visibility into its api equivalent
	VisToAPIVis(ctx context.Context, m gtsmodel.Visibility) model.Visibility
	// InstanceToAPIInstance converts a gts instance into its api equivalent for serving at /api/v1/instance
	InstanceToAPIInstance(ctx context.Context, i *gtsmodel.Instance) (*model.Instance, error)
	// RelationshipToAPIRelationship converts a gts relationship into its api equivalent for serving in various places
	RelationshipToAPIRelationship(ctx context.Context, r *gtsmodel.Relationship) (*model.Relationship, error)
	// NotificationToAPINotification converts a gts notification into a api notification
	NotificationToAPINotification(ctx context.Context, n *gtsmodel.Notification) (*model.Notification, error)
	// DomainBlockToAPIDomainBlock converts a gts model domin block into a api domain block, for serving at /api/v1/admin/domain_blocks
	DomainBlockToAPIDomainBlock(ctx context.Context, b *gtsmodel.DomainBlock, export bool) (*model.DomainBlock, error)

	/*
		INTERNAL (gts) MODEL TO FRONTEND (rss) MODEL
	*/

	StatusToRSSItem(ctx context.Context, s *gtsmodel.Status) (*feeds.Item, error)

	/*
		FRONTEND (api) MODEL TO INTERNAL (gts) MODEL
	*/

	// APIVisToVis converts an API model visibility into its internal gts equivalent.
	APIVisToVis(m model.Visibility) gtsmodel.Visibility

	/*
		ACTIVITYSTREAMS MODEL TO INTERNAL (gts) MODEL
	*/

	// ASPersonToAccount converts a remote account/person/application representation into a gts model account.
	//
	// If update is false, and the account is already known in the database, then the existing account entry will be returned.
	// If update is true, then even if the account is already known, all fields in the accountable will be parsed and a new *gtsmodel.Account
	// will be generated. This is useful when one needs to force refresh of an account, eg., during an Update of a Profile.
	//
	// If accountDomain is set (not an empty string) then this value will be used as the account's Domain. If not set,
	// then the Host of the accountable's AP ID will be used instead.
	ASRepresentationToAccount(ctx context.Context, accountable ap.Accountable, accountDomain string, update bool) (*gtsmodel.Account, error)
	// ASStatus converts a remote activitystreams 'status' representation into a gts model status.
	ASStatusToStatus(ctx context.Context, statusable ap.Statusable) (*gtsmodel.Status, error)
	// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow request.
	ASFollowToFollowRequest(ctx context.Context, followable ap.Followable) (*gtsmodel.FollowRequest, error)
	// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow.
	ASFollowToFollow(ctx context.Context, followable ap.Followable) (*gtsmodel.Follow, error)
	// ASLikeToFave converts a remote activitystreams 'like' representation into a gts model status fave.
	ASLikeToFave(ctx context.Context, likeable ap.Likeable) (*gtsmodel.StatusFave, error)
	// ASBlockToBlock converts a remote activity streams 'block' representation into a gts model block.
	ASBlockToBlock(ctx context.Context, blockable ap.Blockable) (*gtsmodel.Block, error)
	// ASAnnounceToStatus converts an activitystreams 'announce' into a status.
	//
	// The returned bool indicates whether this status is new (true) or not new (false).
	//
	// In other words, if the status is already in the database with the ID set on the announceable, then that will be returned,
	// the returned bool will be false, and no further processing is necessary. If the returned bool is true, indicating
	// that this is a new announce, then further processing will be necessary, because the returned status will be bareboned and
	// require further dereferencing.
	//
	// This is useful when multiple users on an instance might receive the same boost, and we only want to process the boost once.
	//
	// NOTE -- this is different from one status being boosted multiple times! In this case, new boosts should indeed be created.
	ASAnnounceToStatus(ctx context.Context, announceable ap.Announceable) (status *gtsmodel.Status, new bool, err error)

	/*
		INTERNAL (gts) MODEL TO ACTIVITYSTREAMS MODEL
	*/

	// AccountToAS converts a gts model account into an activity streams person, suitable for federation
	AccountToAS(ctx context.Context, a *gtsmodel.Account) (vocab.ActivityStreamsPerson, error)
	// AccountToASMinimal converts a gts model account into an activity streams person, suitable for federation.
	//
	// The returned account will just have the Type, Username, PublicKey, and ID properties set. This is
	// suitable for serving to requesters to whom we want to give as little information as possible because
	// we don't trust them (yet).
	AccountToASMinimal(ctx context.Context, a *gtsmodel.Account) (vocab.ActivityStreamsPerson, error)
	// StatusToAS converts a gts model status into an activity streams note, suitable for federation
	StatusToAS(ctx context.Context, s *gtsmodel.Status) (vocab.ActivityStreamsNote, error)
	// FollowToASFollow converts a gts model Follow into an activity streams Follow, suitable for federation
	FollowToAS(ctx context.Context, f *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (vocab.ActivityStreamsFollow, error)
	// MentionToAS converts a gts model mention into an activity streams Mention, suitable for federation
	MentionToAS(ctx context.Context, m *gtsmodel.Mention) (vocab.ActivityStreamsMention, error)
	// EmojiToAS converts a gts emoji into a mastodon ns Emoji, suitable for federation
	EmojiToAS(ctx context.Context, e *gtsmodel.Emoji) (vocab.TootEmoji, error)
	// AttachmentToAS converts a gts model media attachment into an activity streams Attachment, suitable for federation
	AttachmentToAS(ctx context.Context, a *gtsmodel.MediaAttachment) (vocab.ActivityStreamsDocument, error)
	// FaveToAS converts a gts model status fave into an activityStreams LIKE, suitable for federation.
	FaveToAS(ctx context.Context, f *gtsmodel.StatusFave) (vocab.ActivityStreamsLike, error)
	// BoostToAS converts a gts model boost into an activityStreams ANNOUNCE, suitable for federation
	BoostToAS(ctx context.Context, boostWrapperStatus *gtsmodel.Status, boostingAccount *gtsmodel.Account, boostedAccount *gtsmodel.Account) (vocab.ActivityStreamsAnnounce, error)
	// BlockToAS converts a gts model block into an activityStreams BLOCK, suitable for federation.
	BlockToAS(ctx context.Context, block *gtsmodel.Block) (vocab.ActivityStreamsBlock, error)
	// StatusToASRepliesCollection converts a gts model status into an activityStreams REPLIES collection.
	StatusToASRepliesCollection(ctx context.Context, status *gtsmodel.Status, onlyOtherAccounts bool) (vocab.ActivityStreamsCollection, error)
	// StatusURIsToASRepliesPage returns a collection page with appropriate next/part of pagination.
	StatusURIsToASRepliesPage(ctx context.Context, status *gtsmodel.Status, onlyOtherAccounts bool, minID string, replies map[string]*url.URL) (vocab.ActivityStreamsCollectionPage, error)
	// OutboxToASCollection returns an ordered collection with appropriate id, next, and last fields.
	// The returned collection won't have any actual entries; just links to where entries can be obtained.
	OutboxToASCollection(ctx context.Context, outboxID string) (vocab.ActivityStreamsOrderedCollection, error)
	// StatusesToASOutboxPage returns an ordered collection page using the given statuses and parameters as contents.
	//
	// The maxID and minID should be the parameters that were passed to the database to obtain the given statuses.
	// These will be used to create the 'id' field of the collection.
	//
	// OutboxID is used to create the 'partOf' field in the collection.
	//
	// Appropriate 'next' and 'prev' fields will be created based on the highest and lowest IDs present in the statuses slice.
	StatusesToASOutboxPage(ctx context.Context, outboxID string, maxID string, minID string, statuses []*gtsmodel.Status) (vocab.ActivityStreamsOrderedCollectionPage, error)

	/*
		INTERNAL (gts) MODEL TO INTERNAL MODEL
	*/

	// FollowRequestToFollow just converts a follow request into a follow, that's it! No bells and whistles.
	FollowRequestToFollow(ctx context.Context, f *gtsmodel.FollowRequest) *gtsmodel.Follow
	// StatusToBoost wraps the given status into a boosting status.
	StatusToBoost(ctx context.Context, s *gtsmodel.Status, boostingAccount *gtsmodel.Account) (*gtsmodel.Status, error)

	/*
		WRAPPER CONVENIENCE FUNCTIONS
	*/

	// WrapPersonInUpdate
	WrapPersonInUpdate(person vocab.ActivityStreamsPerson, originAccount *gtsmodel.Account) (vocab.ActivityStreamsUpdate, error)
	// WrapNoteInCreate wraps a Note with a Create activity.
	//
	// If objectIRIOnly is set to true, then the function won't put the *entire* note in the Object field of the Create,
	// but just the AP URI of the note. This is useful in cases where you want to give a remote server something to dereference,
	// and still have control over whether or not they're allowed to actually see the contents.
	WrapNoteInCreate(note vocab.ActivityStreamsNote, objectIRIOnly bool) (vocab.ActivityStreamsCreate, error)
}

type converter struct {
	db             db.DB
	defaultAvatars []string
	randAvatars    sync.Map
}

// NewConverter returns a new Converter
func NewConverter(db db.DB) TypeConverter {
	return &converter{
		db:             db,
		defaultAvatars: populateDefaultAvatars(),
	}
}
