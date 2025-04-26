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

package media

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
)

// Delete deletes the media attachment with the given ID, including all files pertaining to that attachment.
func (p *Processor) Delete(ctx context.Context, mediaAttachmentID string) gtserror.WithCode {
	attachment, err := p.state.DB.GetAttachmentByID(ctx, mediaAttachmentID)
	if err != nil {
		if err == db.ErrNoEntries {
			// attachment already gone
			return nil
		}
		// actual error
		return gtserror.NewErrorInternalError(err)
	}

	errs := []string{}

	// delete the thumbnail from storage
	if attachment.Thumbnail.Path != "" {
		if err := p.state.Storage.Delete(ctx, attachment.Thumbnail.Path); err != nil && !storage.IsNotFound(err) {
			errs = append(errs, fmt.Sprintf("remove thumbnail at path %s: %s", attachment.Thumbnail.Path, err))
		}
	}

	// delete the file from storage
	if attachment.File.Path != "" {
		if err := p.state.Storage.Delete(ctx, attachment.File.Path); err != nil && !storage.IsNotFound(err) {
			errs = append(errs, fmt.Sprintf("remove file at path %s: %s", attachment.File.Path, err))
		}
	}

	// delete the attachment
	if err := p.state.DB.DeleteAttachment(ctx, mediaAttachmentID); err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs = append(errs, fmt.Sprintf("remove attachment: %s", err))
	}

	if len(errs) != 0 {
		return gtserror.NewErrorInternalError(fmt.Errorf("Delete: one or more errors removing attachment with id %s: %s", mediaAttachmentID, strings.Join(errs, "; ")))
	}

	return nil
}
