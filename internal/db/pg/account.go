package pg

import (
	"errors"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) GetAccountHeader(header *gtsmodel.MediaAttachment, accountID string) error {
	acct := &gtsmodel.Account{}
	if err := ps.conn.Model(acct).Where("id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}

	if acct.HeaderMediaAttachmentID == "" {
		return db.ErrNoEntries{}
	}

	if err := ps.conn.Model(header).Where("id = ?", acct.HeaderMediaAttachmentID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountAvatar(avatar *gtsmodel.MediaAttachment, accountID string) error {
	acct := &gtsmodel.Account{}
	if err := ps.conn.Model(acct).Where("id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}

	if acct.AvatarMediaAttachmentID == "" {
		return db.ErrNoEntries{}
	}

	if err := ps.conn.Model(avatar).Where("id = ?", acct.AvatarMediaAttachmentID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountLastStatus(accountID string, status *gtsmodel.Status) error {
	if err := ps.conn.Model(status).Order("created_at DESC").Limit(1).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil

}

func (ps *postgresService) SetAccountHeaderOrAvatar(mediaAttachment *gtsmodel.MediaAttachment, accountID string) error {
	if mediaAttachment.Avatar && mediaAttachment.Header {
		return errors.New("one media attachment cannot be both header and avatar")
	}

	var headerOrAVI string
	if mediaAttachment.Avatar {
		headerOrAVI = "avatar"
	} else if mediaAttachment.Header {
		headerOrAVI = "header"
	} else {
		return errors.New("given media attachment was neither a header nor an avatar")
	}

	// TODO: there are probably more side effects here that need to be handled
	if _, err := ps.conn.Model(mediaAttachment).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		return err
	}

	if _, err := ps.conn.Model(&gtsmodel.Account{}).Set(fmt.Sprintf("%s_media_attachment_id = ?", headerOrAVI), mediaAttachment.ID).Where("id = ?", accountID).Update(); err != nil {
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountByUserID(userID string, account *gtsmodel.Account) error {
	user := &gtsmodel.User{
		ID: userID,
	}
	if err := ps.conn.Model(user).Where("id = ?", userID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	if err := ps.conn.Model(account).Where("id = ?", user.AccountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetLocalAccountByUsername(username string, account *gtsmodel.Account) error {
	if err := ps.conn.Model(account).Where("username = ?", username).Where("? IS NULL", pg.Ident("domain")).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountFollowRequests(accountID string, followRequests *[]gtsmodel.FollowRequest) error {
	if err := ps.conn.Model(followRequests).Where("target_account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountFollowing(accountID string, following *[]gtsmodel.Follow) error {
	if err := ps.conn.Model(following).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountFollowers(accountID string, followers *[]gtsmodel.Follow, localOnly bool) error {

	q := ps.conn.Model(followers)

	if localOnly {
		// for local accounts let's get where domain is null OR where domain is an empty string, just to be safe
		whereGroup := func(q *pg.Query) (*pg.Query, error) {
			q = q.
				WhereOr("? IS NULL", pg.Ident("a.domain")).
				WhereOr("a.domain = ?", "")
			return q, nil
		}

		q = q.ColumnExpr("follow.*").
			Join("JOIN accounts AS a ON follow.account_id = TEXT(a.id)").
			Where("follow.target_account_id = ?", accountID).
			WhereGroup(whereGroup)
	} else {
		q = q.Where("target_account_id = ?", accountID)
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountFaves(accountID string, faves *[]gtsmodel.StatusFave) error {
	if err := ps.conn.Model(faves).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAccountStatusesCount(accountID string) (int, error) {
	count, err := ps.conn.Model(&gtsmodel.Status{}).Where("account_id = ?", accountID).Count()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (ps *postgresService) GetAccountStatuses(accountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]*gtsmodel.Status, error) {
	ps.log.Debugf("getting statuses for account %s", accountID)
	statuses := []*gtsmodel.Status{}

	q := ps.conn.Model(&statuses).Order("id DESC")
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if excludeReplies {
		q = q.Where("? IS NULL", pg.Ident("in_reply_to_id"))
	}

	if pinnedOnly {
		q = q.Where("pinned = ?", true)
	}

	if mediaOnly {
		q = q.WhereGroup(func(q *pg.Query) (*pg.Query, error) {
			return q.Where("? IS NOT NULL", pg.Ident("attachments")).Where("attachments != '{}'"), nil
		})
	}

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, db.ErrNoEntries{}
	}

	ps.log.Debugf("returning statuses for account %s", accountID)
	return statuses, nil
}
