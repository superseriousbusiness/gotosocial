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

package db

// ErrNoEntries is to be returned from the DB interface when no entries are found for a given query.
type ErrNoEntries struct{}

func (e ErrNoEntries) Error() string {
	return "no entries"
}

// ErrAlreadyExists is to be returned from the DB interface when an entry already exists for a given query or its constraints.
type ErrAlreadyExists struct{}

func (e ErrAlreadyExists) Error() string {
	return "already exists"
}
