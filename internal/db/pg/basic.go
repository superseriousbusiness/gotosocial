package pg

import (
	"errors"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

func (ps *postgresService) Put(i interface{}) error {
	_, err := ps.conn.Model(i).Insert(i)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return db.ErrAlreadyExists{}
	}
	return err
}

func (ps *postgresService) GetByID(id string, i interface{}) error {
	if err := ps.conn.Model(i).Where("id = ?", id).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err

	}
	return nil
}

func (ps *postgresService) GetWhere(where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := ps.conn.Model(i)
	for _, w := range where {

		if w.Value == nil {
			q = q.Where("? IS NULL", pg.Ident(w.Key))
		} else {
			if w.CaseInsensitive {
				q = q.Where("LOWER(?) = LOWER(?)", pg.Safe(w.Key), w.Value)
			} else {
				q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
			}
		}
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAll(i interface{}) error {
	if err := ps.conn.Model(i).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) DeleteByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).Where("id = ?", id).Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}

func (ps *postgresService) DeleteWhere(where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := ps.conn.Model(i)
	for _, w := range where {
		q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
	}

	if _, err := q.Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}
