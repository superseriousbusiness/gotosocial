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

package testrig

import (
	"git.iim.gay/grufwub/go-store/kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

// NewTestProcessor returns a Processor suitable for testing purposes
func NewTestProcessor(db db.DB, storage *kv.KVStore, federator federation.Federator) processing.Processor {
	return processing.NewProcessor(NewTestConfig(), NewTestTypeConverter(db), federator, NewTestOauthServer(db), NewTestMediaHandler(db, storage), storage, NewTestTimelineManager(db), db, NewTestLog())
}
