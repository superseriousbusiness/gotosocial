/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package log

import (
	"io"
	"sync"
)

// safewriter wraps a writer to provide mutex safety on write.
type safewriter struct {
	w io.Writer
	m sync.Mutex
}

func (w *safewriter) Write(b []byte) (int, error) {
	w.m.Lock()
	n, err := w.w.Write(b)
	w.m.Unlock()
	return n, err
}
