package media

import gtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20211113114307_init"

type Media struct {
	Attachment *gtsmodel.MediaAttachment
}
