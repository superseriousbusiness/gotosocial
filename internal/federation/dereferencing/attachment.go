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

package dereferencing

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) GetRemoteAttachment(ctx context.Context, requestingUsername string, remoteAttachmentURI *url.URL, ownerAccountID string, statusID string, expectedContentType string) (*gtsmodel.MediaAttachment, error) {
	l := d.log.WithFields(logrus.Fields{
		"username":            requestingUsername,
		"remoteAttachmentURI": remoteAttachmentURI,
	})

	maybeAttachment := &gtsmodel.MediaAttachment{}
	where := []db.Where{
		{
			Key:   "remote_url",
			Value: remoteAttachmentURI.String(),
		},
	}

	if err := d.db.GetWhere(ctx, where, maybeAttachment); err == nil {
		// we already the attachment in the database
		l.Debugf("GetRemoteAttachment: attachment already exists with id %s", maybeAttachment.ID)
		return maybeAttachment, nil
	}

	a, err := d.RefreshAttachment(ctx, requestingUsername, remoteAttachmentURI, ownerAccountID, expectedContentType)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAttachment: error refreshing attachment: %s", err)
	}

	a.StatusID = statusID
	if err := d.db.Put(ctx, a); err != nil {
		if err != db.ErrAlreadyExists {
			return nil, fmt.Errorf("GetRemoteAttachment: error inserting attachment: %s", err)
		}
	}

	return a, nil
}

func (d *deref) RefreshAttachment(ctx context.Context, requestingUsername string, remoteAttachmentURI *url.URL, ownerAccountID string, expectedContentType string) (*gtsmodel.MediaAttachment, error) {
	// it just doesn't exist or we have to refresh
	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return nil, fmt.Errorf("RefreshAttachment: error creating transport: %s", err)
	}

	attachmentBytes, err := t.DereferenceMedia(ctx, remoteAttachmentURI, expectedContentType)
	if err != nil {
		return nil, fmt.Errorf("RefreshAttachment: error dereferencing media: %s", err)
	}

	a, err := d.mediaHandler.ProcessAttachment(ctx, attachmentBytes, ownerAccountID, remoteAttachmentURI.String())
	if err != nil {
		return nil, fmt.Errorf("RefreshAttachment: error processing attachment: %s", err)
	}

	return a, nil
}
