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
	"bytes"
	"io"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Initialize initializes the global Logrus logger, reading the desired
// log level from the viper store, or using a default if the level
// has not been set in viper.
//
// It also sets the output to log.SplitErrOutputs(...)
// so you get error logs on stderr and normal logs on stdout.
//
// If syslog settings are also in viper, then Syslog will be initialized as well.
func Initialize() error {
	out := SplitErrOutputs(os.Stdout, os.Stderr)
	logrus.SetOutput(out)

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		DisableQuote:  true,
		FullTimestamp: true,
	})

	// check if a desired log level has been set
	if lvl := config.GetLogLevel(); lvl != "" {
		level, err := logrus.ParseLevel(lvl)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)

		if level == logrus.TraceLevel {
			logrus.SetReportCaller(true)
		}
	}

	// check if syslog has been enabled, and configure it if so
	if config.GetSyslogEnabled() {
		protocol := config.GetSyslogProtocol()
		address := config.GetSyslogAddress()

		hook, err := lSyslog.NewSyslogHook(protocol, address, syslog.LOG_INFO, "")
		if err != nil {
			return err
		}

		logrus.AddHook(&trimHook{hook})
	}

	return nil
}

// SplitErrOutputs returns an OutputSplitFunc that splits output to either one of
// two given outputs depending on whether the level is "error","fatal","panic".
func SplitErrOutputs(out, err io.Writer) OutputSplitFunc {
	return func(lvl []byte) io.Writer {
		switch string(lvl) /* convert to str for compare is no-alloc */ {
		case "error", "fatal", "panic":
			return err
		default:
			return out
		}
	}
}

// OutputSplitFunc implements the io.Writer interface for use with Logrus, and simply
// splits logs between stdout and stderr depending on their severity.
type OutputSplitFunc func(lvl []byte) io.Writer

var levelBytes = []byte("level=")

func (fn OutputSplitFunc) Write(b []byte) (int, error) {
	var lvl []byte
	if i := bytes.Index(b, levelBytes); i >= 0 {
		blvl := b[i+len(levelBytes):]
		if i := bytes.IndexByte(blvl, ' '); i >= 0 {
			lvl = blvl[:i]
		}
	}
	return fn(lvl).Write(b)
}
