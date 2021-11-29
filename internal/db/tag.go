package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type Tag interface {
	GetTagByID(context.Context, string) (*gtsmodel.Tag, Error)
	GetTagByName(context.Context, string) (*gtsmodel.Tag, Error)
}
