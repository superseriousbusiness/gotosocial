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
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	asPublicURI = "https://www.w3.org/ns/activitystreams#Public"
)

// TypeConverter is an interface for the common action of converting between apimodule (frontend, serializable) models,
// internal gts models used in the database, and activitypub models used in federation.
//
// It requires access to the database because many of the conversions require pulling out database entries and counting them etc.
// That said, it *absolutely should not* manipulate database entries in any way, only examine them.
type TypeConverter interface {
	/*
		INTERNAL (gts) MODEL TO FRONTEND (mastodon) MODEL
	*/

	// AccountToMastoSensitive takes a db model account as a param, and returns a populated mastotype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may have sensitive fields,
	// so serve it only to an authorized user who should have permission to see it.
	AccountToMastoSensitive(account *gtsmodel.Account) (*model.Account, error)

	// AccountToMastoPublic takes a db model account as a param, and returns a populated mastotype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may NOT have sensitive fields.
	// In other words, this is the public record that the server has of an account.
	AccountToMastoPublic(account *gtsmodel.Account) (*model.Account, error)

	// AppToMastoSensitive takes a db model application as a param, and returns a populated mastotype application, or an error
	// if something goes wrong. The returned application should be ready to serialize on an API level, and may have sensitive fields
	// (such as client id and client secret), so serve it only to an authorized user who should have permission to see it.
	AppToMastoSensitive(application *gtsmodel.Application) (*model.Application, error)

	// AppToMastoPublic takes a db model application as a param, and returns a populated mastotype application, or an error
	// if something goes wrong. The returned application should be ready to serialize on an API level, and has sensitive
	// fields sanitized so that it can be served to non-authorized accounts without revealing any private information.
	AppToMastoPublic(application *gtsmodel.Application) (*model.Application, error)

	// AttachmentToMasto converts a gts model media attacahment into its mastodon representation for serialization on the API.
	AttachmentToMasto(attachment *gtsmodel.MediaAttachment) (model.Attachment, error)

	// MentionToMasto converts a gts model mention into its mastodon (frontend) representation for serialization on the API.
	MentionToMasto(m *gtsmodel.Mention) (model.Mention, error)

	// EmojiToMasto converts a gts model emoji into its mastodon (frontend) representation for serialization on the API.
	EmojiToMasto(e *gtsmodel.Emoji) (model.Emoji, error)

	// TagToMasto converts a gts model tag into its mastodon (frontend) representation for serialization on the API.
	TagToMasto(t *gtsmodel.Tag) (model.Tag, error)

	// StatusToMasto converts a gts model status into its mastodon (frontend) representation for serialization on the API.
	StatusToMasto(s *gtsmodel.Status, targetAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account, boostOfAccount *gtsmodel.Account, replyToAccount *gtsmodel.Account, reblogOfStatus *gtsmodel.Status) (*model.Status, error)

	// VisToMasto converts a gts visibility into its mastodon equivalent
	VisToMasto(m gtsmodel.Visibility) model.Visibility

	// InstanceToMasto converts a gts instance into its mastodon equivalent for serving at /api/v1/instance
	InstanceToMasto(i *gtsmodel.Instance) (*model.Instance, error)

	// RelationshipToMasto converts a gts relationship into its mastodon equivalent for serving in various places
	RelationshipToMasto(r *gtsmodel.Relationship) (*model.Relationship, error)

	NotificationToMasto(n *gtsmodel.Notification) (*model.Notification, error)

	/*
		FRONTEND (mastodon) MODEL TO INTERNAL (gts) MODEL
	*/

	// MastoVisToVis converts a mastodon visibility into its gts equivalent.
	MastoVisToVis(m model.Visibility) gtsmodel.Visibility

	/*
		ACTIVITYSTREAMS MODEL TO INTERNAL (gts) MODEL
	*/

	// ASPersonToAccount converts a remote account/person/application representation into a gts model account.
	//
	// If update is false, and the account is already known in the database, then the existing account entry will be returned.
	// If update is true, then even if the account is already known, all fields in the accountable will be parsed and a new *gtsmodel.Account
	// will be generated. This is useful when one needs to force refresh of an account, eg., during an Update of a Profile.
	ASRepresentationToAccount(accountable Accountable, update bool) (*gtsmodel.Account, error)
	// ASStatus converts a remote activitystreams 'status' representation into a gts model status.
	ASStatusToStatus(statusable Statusable) (*gtsmodel.Status, error)
	// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow request.
	ASFollowToFollowRequest(followable Followable) (*gtsmodel.FollowRequest, error)
	// ASFollowToFollowRequest converts a remote activitystreams `follow` representation into gts model follow.
	ASFollowToFollow(followable Followable) (*gtsmodel.Follow, error)
	// ASLikeToFave converts a remote activitystreams 'like' representation into a gts model status fave.
	ASLikeToFave(likeable Likeable) (*gtsmodel.StatusFave, error)

	/*
		INTERNAL (gts) MODEL TO ACTIVITYSTREAMS MODEL
	*/

	// AccountToAS converts a gts model account into an activity streams person, suitable for federation
	AccountToAS(a *gtsmodel.Account) (vocab.ActivityStreamsPerson, error)

	// StatusToAS converts a gts model status into an activity streams note, suitable for federation
	StatusToAS(s *gtsmodel.Status) (vocab.ActivityStreamsNote, error)

	// FollowToASFollow converts a gts model Follow into an activity streams Follow, suitable for federation
	FollowToAS(f *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (vocab.ActivityStreamsFollow, error)

	// MentionToAS converts a gts model mention into an activity streams Mention, suitable for federation
	MentionToAS(m *gtsmodel.Mention) (vocab.ActivityStreamsMention, error)

	// AttachmentToAS converts a gts model media attachment into an activity streams Attachment, suitable for federation
	AttachmentToAS(a *gtsmodel.MediaAttachment) (vocab.ActivityStreamsDocument, error)

	// FaveToAS converts a gts model status fave into an activityStreams LIKE, suitable for federation.
	FaveToAS(f *gtsmodel.StatusFave) (vocab.ActivityStreamsLike, error)
}

type converter struct {
	config *config.Config
	db     db.DB
}

// NewConverter returns a new Converter
func NewConverter(config *config.Config, db db.DB) TypeConverter {
	return &converter{
		config: config,
		db:     db,
	}
}
