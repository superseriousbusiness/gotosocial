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

	"log/syslog"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Initialize initializes the global Logrus logger, reading the desired
// log level from the viper store, or using a default if the level
// has not been set in viper.
//
// It also sets the output to log.outputSplitter,
// so you get error logs on stderr and normal logs on stdout.
//
// If syslog settings are also in viper, then Syslog will be initialized as well.
func Initialize() error {
	logrus.SetOutput(&outputSplitter{})

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		DisableQuote:  true,
		FullTimestamp: true,
	})

	keys := config.Keys

	// check if a desired log level has been set
	logLevel := viper.GetString(keys.LogLevel)
	if logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)

		if level == logrus.TraceLevel {
			logrus.SetReportCaller(true)
		}
	}

	// check if syslog has been enabled, and configure it if so
	if syslogEnabled := viper.GetBool(keys.SyslogEnabled); syslogEnabled {
		protocol := viper.GetString(keys.SyslogProtocol)
		address := viper.GetString(keys.SyslogAddress)

		hook, err := lSyslog.NewSyslogHook(protocol, address, syslog.LOG_INFO, "")
		if err != nil {
			return err
		}

		logrus.AddHook(hook)
	}

	return nil
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
