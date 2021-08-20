package media

import (
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Delete(mediaAttachmentID string) gtserror.WithCode {
	a := &gtsmodel.MediaAttachment{}
	if err := p.db.GetByID(mediaAttachmentID, a); err != nil {
		if err == db.ErrNoEntries {
			// attachment already gone
			return nil
		}
		// actual error
		return gtserror.NewErrorInternalError(err)
	}

	errs := []string{}

	// delete the thumbnail from storage
	if a.Thumbnail.Path != "" {
		if err := p.storage.RemoveFileAt(a.Thumbnail.Path); err != nil {
			errs = append(errs, fmt.Sprintf("remove thumbnail at path %s: %s", a.Thumbnail.Path, err))
		}
	}

	// delete the file from storage
	if a.File.Path != "" {
		if err := p.storage.RemoveFileAt(a.File.Path); err != nil {
			errs = append(errs, fmt.Sprintf("remove file at path %s: %s", a.File.Path, err))
		}
	}

	// delete the attachment
	if err := p.db.DeleteByID(mediaAttachmentID, a); err != nil {
		if err != db.ErrNoEntries {
			errs = append(errs, fmt.Sprintf("remove attachment: %s", err))
		}
	}

	if len(errs) != 0 {
		return gtserror.NewErrorInternalError(fmt.Errorf("Delete: one or more errors removing attachment with id %s: %s", mediaAttachmentID, strings.Join(errs, "; ")))
	}

	return nil
}
