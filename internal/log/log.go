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

package log

import (
	"bytes"
	"os"

	"github.com/sirupsen/logrus"
)

// New returns a new logrus logger with the specified level,
// or an error if that level can't be parsed. It also sets
// the output to log.outputSplitter, so you get error logs
// on stderr and normal logs on stdout.
func New(level string) (*logrus.Logger, error) {
	log := logrus.New()

	log.SetOutput(&outputSplitter{})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	log.SetLevel(logLevel)

	if logLevel == logrus.TraceLevel {
		log.SetReportCaller(true)
	}

	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		ForceQuote:    true,
		FullTimestamp: true,
	})

	return log, nil
}

// outputSplitter implements the io.Writer interface for use with Logrus, and simply
// splits logs between stdout and stderr depending on their severity.
// See: https://github.com/sirupsen/logrus/issues/403#issuecomment-346437512
type outputSplitter struct{}

func (splitter *outputSplitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("level=error")) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}
