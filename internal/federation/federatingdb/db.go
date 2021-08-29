/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package federatingdb

import (
	"context"
	"sync"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// DB wraps the pub.Database interface with a couple of custom functions for GoToSocial.
type DB interface {
	pub.Database
	Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error
	Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error
	Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error
}

// FederatingDB uses the underlying DB interface to implement the go-fed pub.Database interface.
// It doesn't care what the underlying implementation of the DB interface is, as long as it works.
type federatingDB struct {
	mutex         sync.Mutex
	locks         map[string]*mutex
	pool          sync.Pool
	db            db.DB
	config        *config.Config
	log           *logrus.Logger
	typeConverter typeutils.TypeConverter
}

// New returns a DB interface using the given database, config, and logger.
func New(db db.DB, config *config.Config, log *logrus.Logger) DB {
	fdb := federatingDB{
		mutex:         sync.Mutex{},
		locks:         make(map[string]*mutex, 100),
		pool:          sync.Pool{New: func() interface{} { return &mutex{} }},
		db:            db,
		config:        config,
		log:           log,
		typeConverter: typeutils.NewConverter(config, db, log),
	}
	go fdb.cleanupLocks()
	return &fdb
}

func (db *federatingDB) cleanupLocks() {
	for {
		// Sleep for a minute...
		time.Sleep(time.Minute)

		// Delete unused locks from map
		db.mutex.Lock()
		for id, mu := range db.locks {
			if !mu.inUse() {
				delete(db.locks, id)
				db.pool.Put(mu)
			}
		}
		db.mutex.Unlock()
	}
}
