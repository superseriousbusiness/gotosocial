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
	"github.com/sirupsen/logrus"
)

// trimHook is a wrapper round a logrus hook that trims the *entry.Message
// to no more than 1700 characters before sending it through to the wrapped hook,
// to avoid spamming syslog with messages that are too long for it.
type trimHook struct {
	wrappedHook logrus.Hook
}

func (t *trimHook) Fire(e *logrus.Entry) error {
	// only copy/truncate if we need to
	if len(e.Message) < 1700 {
		return t.wrappedHook.Fire(e)
	}

	// it's too long, truncate + fire a copy of the entry so we don't meddle with the original
	return t.wrappedHook.Fire(&logrus.Entry{
		Logger:  e.Logger,
		Data:    e.Data,
		Time:    e.Time,
		Level:   e.Level,
		Caller:  e.Caller,
		Message: e.Message[:1696] + "...", // truncate
		Buffer:  e.Buffer,
		Context: e.Context,
	})
}

func (t *trimHook) Levels() []logrus.Level {
	return t.wrappedHook.Levels()
}
