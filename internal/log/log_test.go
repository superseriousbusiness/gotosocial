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

package log_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-kv/v2"
)

func TestInlineability(t *testing.T) {
	t.Skip()

	// to check the output of this run:
	// go test -gcflags=all='-m=2' ./internal/log/ -run=TestInlineability 2>&1 | grep 'internal/log/log_test.go.*cannot inline'
	//
	// the output should not include any
	// of the below log package func calls.

	ctx := t.Context()
	const s = "hello world"

	log.Debug(ctx, s)
	log.Debugf(ctx, s)
	log.DebugKV(ctx, "key", s)
	log.DebugKVs(ctx, kv.Fields{{K: "key", V: s}}...)

	log.Info(ctx, s)
	log.Infof(ctx, s)
	log.InfoKV(ctx, "key", s)
	log.InfoKVs(ctx, kv.Fields{{K: "key", V: s}}...)

	log.Warn(ctx, s)
	log.Warnf(ctx, s)
	log.WarnKV(ctx, "key", s)
	log.WarnKVs(ctx, kv.Fields{{K: "key", V: s}}...)

	log.Error(ctx, s)
	log.Errorf(ctx, s)
	log.ErrorKV(ctx, "key", s)
	log.ErrorKVs(ctx, kv.Fields{{K: "key", V: s}}...)

	log.Panic(ctx, s)
	log.Panicf(ctx, s)
	log.PanicKV(ctx, "key", s)
	log.PanicKVs(ctx, kv.Fields{{K: "key", V: s}}...)

	log.Print(s)
	log.Printf(s)

	e := log.New()
	e = e.WithContext(ctx)
	e = e.WithField("key", s)
	e = e.WithFields(kv.Fields{{K: "key", V: s}}...)

	e.Debug(s)
	e.Debugf(s)

	e.Info(s)
	e.Infof(s)

	e.Warn(s)
	e.Warnf(s)

	e.Error(s)
	e.Errorf(s)

	e.Panic(s)
	e.Panicf(s)

	e.Print(s)
	e.Printf(s)
}
