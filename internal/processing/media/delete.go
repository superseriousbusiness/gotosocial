package media

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *processor) Delete(ctx context.Context, mediaAttachmentID string) gtserror.WithCode {
	attachment, err := p.db.GetAttachmentByID(ctx, mediaAttachmentID)
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
		if err := p.storage.Delete(ctx, attachment.Thumbnail.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			errs = append(errs, fmt.Sprintf("remove thumbnail at path %s: %s", attachment.Thumbnail.Path, err))
		}
	}

	// delete the file from storage
	if attachment.File.Path != "" {
		if err := p.storage.Delete(ctx, attachment.File.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			errs = append(errs, fmt.Sprintf("remove file at path %s: %s", attachment.File.Path, err))
		}
	}

	// delete the attachment
	if err := p.db.DeleteByID(ctx, mediaAttachmentID, attachment); err != nil {
		if err != db.ErrNoEntries {
			errs = append(errs, fmt.Sprintf("remove attachment: %s", err))
		}
	}

	if len(errs) != 0 {
		return gtserror.NewErrorInternalError(fmt.Errorf("Delete: one or more errors removing attachment with id %s: %s", mediaAttachmentID, strings.Join(errs, "; ")))
	}

	return nil
}
