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

package bundb

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type workerTaskDB struct{ db *bun.DB }

func (w *workerTaskDB) GetWorkerTasks(ctx context.Context) ([]*gtsmodel.WorkerTask, error) {
	var tasks []*gtsmodel.WorkerTask
	if err := w.db.NewSelect().
		Model(&tasks).
		OrderExpr("? ASC", bun.Ident("created_at")).
		Scan(ctx); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (w *workerTaskDB) PutWorkerTasks(ctx context.Context, tasks []*gtsmodel.WorkerTask) error {
	var errs []error
	for _, task := range tasks {
		_, err := w.db.NewInsert().Model(task).Exec(ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (w *workerTaskDB) DeleteWorkerTaskByID(ctx context.Context, id uint) error {
	_, err := w.db.NewDelete().
		Table("worker_tasks").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return err
}
