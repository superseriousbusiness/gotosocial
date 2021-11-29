package bundb

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type tagDB struct {
	config *config.Config
	conn   *DBConn
	// TODO
	cache *cache.TagCache
}

func (t *tagDB) newTagQ(tag interface{}) *bun.SelectQuery {
	logrus.Debug("newTagQ")
	return t.conn.
		NewSelect().
		Model(tag).
		Relation("Statuses")
}

func (t *tagDB) GetTagByID(ctx context.Context, id string) (*gtsmodel.Tag, db.Error) {
	return t.getTag(
		ctx,
		func() (*gtsmodel.Tag, bool) {
			return t.cache.GetByID(id)
		},
		func(tag *gtsmodel.Tag) error {
			return t.newTagQ(tag).Where("tag.id = ?", id).Scan(ctx)
		},
	)
}

func (t *tagDB) GetTagByName(ctx context.Context, name string) (*gtsmodel.Tag, db.Error) {
	tag := new(gtsmodel.Tag)
	if err := t.newTagQ(tag).
		Where("tag.name = ?", name).
		Scan(ctx); err != nil {
		return nil, t.conn.ProcessError(err)
	}
	return tag, nil
}

func (t *tagDB) getTag(ctx context.Context, cacheGet func() (*gtsmodel.Tag, bool), dbQuery func(*gtsmodel.Tag) error) (*gtsmodel.Tag, db.Error) {
	tag, cached := cacheGet()

	if !cached {
		tag = &gtsmodel.Tag{}

		// Not cached! Perform database query
		err := dbQuery(tag)
		if err != nil {
			return nil, t.conn.ProcessError(err)
		}

		// If there is boosted, fetch from DB also
		// if tag. {

		// }

		// t.cache.Put(tag)
	}

	return tag, nil
}
