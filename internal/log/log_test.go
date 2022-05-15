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

package log_test

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func TestOutputSplitFunc(t *testing.T) {
	var outbuf, errbuf bytes.Buffer

	out := log.SplitErrOutputs(&outbuf, &errbuf)

	log := logrus.New()
	log.SetOutput(out)
	log.SetLevel(logrus.TraceLevel)

	for _, lvl := range logrus.AllLevels {
		func() {
			defer func() { recover() }()
			log.Log(lvl, "hello world")
		}()

		t.Logf("outbuf=%q errbuf=%q", outbuf.String(), errbuf.String())

		switch lvl {
		case logrus.PanicLevel:
			if outbuf.Len() > 0 || errbuf.Len() == 0 {
				t.Error("expected panic to log to OutputSplitter.Err")
			}
		case logrus.FatalLevel:
			if outbuf.Len() > 0 || errbuf.Len() == 0 {
				t.Error("expected fatal to log to OutputSplitter.Err")
			}
		case logrus.ErrorLevel:
			if outbuf.Len() > 0 || errbuf.Len() == 0 {
				t.Error("expected error to log to OutputSplitter.Err")
			}
		default:
			if outbuf.Len() == 0 || errbuf.Len() > 0 {
				t.Errorf("expected %s to log to OutputSplitter.Out", lvl)
			}
		}

		// Reset buffers
		outbuf.Reset()
		errbuf.Reset()
	}
}
