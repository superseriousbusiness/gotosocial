package typeutils

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (c *converter) interactionsWithStatusForAccount(ctx context.Context, s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*statusInteractions, error) {
	si := &statusInteractions{}

	if requestingAccount != nil {
		faved, err := c.db.IsStatusFavedBy(ctx, s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has faved status: %s", err)
		}
		si.Faved = faved

		reblogged, err := c.db.IsStatusRebloggedBy(ctx, s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has reblogged status: %s", err)
		}
		si.Reblogged = reblogged

		muted, err := c.db.IsStatusMutedBy(ctx, s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has muted status: %s", err)
		}
		si.Muted = muted

		bookmarked, err := c.db.IsStatusBookmarkedBy(ctx, s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has bookmarked status: %s", err)
		}
		si.Bookmarked = bookmarked
	}
	return si, nil
}

// StatusInteractions denotes interactions with a status on behalf of an account.
type statusInteractions struct {
	Faved      bool
	Muted      bool
	Bookmarked bool
	Reblogged  bool
}

func generateUnknownAttachmentHelperText(apiAttachments []model.Attachment) string {
	unknownAttachmentEntries := []string{}

	for _, a := range apiAttachments {
		if a.Type == "unknown" && a.RemoteURL != nil {
			buf := &bytes.Buffer{}
			buf.WriteString(`<li>`)
			buf.WriteString(`<a href="` + *a.RemoteURL + `">` + *a.RemoteURL + `</a>`)
			buf.WriteString(`</li>`)
			unknownAttachmentEntries = append(unknownAttachmentEntries, buf.String())
		}
	}

	var unknownAttachmentHelperText string
	if count := len(unknownAttachmentEntries); count != 0 {
		ac := config.GetAccountDomain()
		buf := &bytes.Buffer{}
		buf.WriteString(`<hr />`)
		buf.WriteString(`<p>`)
		buf.WriteString(`GoToSocial (` + ac + `): `)
		buf.WriteString(`This post contains ` + strconv.Itoa(count) + ` attachment(s) not recognized by the server. `)
		buf.WriteString(`Check external link(s) carefully and click at your own risk.`)
		buf.WriteString(`</p>`)
		buf.WriteString(`<ul>` + strings.Join(unknownAttachmentEntries, "") + `</ul>`)
		unknownAttachmentHelperText = buf.String()
	}

	return text.SanitizeHTML(unknownAttachmentHelperText)
}
