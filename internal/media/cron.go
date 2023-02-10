/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package media

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type cronLogger struct{}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Info("media manager cron logger: ", msg, keysAndValues)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Error("media manager cron logger: ", err, msg, keysAndValues)
}

func scheduleCleanup(m *manager) error {
	pruneCtx, pruneCancel := context.WithCancel(context.Background())

	c := cron.New(cron.WithLogger(new(cronLogger)))
	defer c.Start()

	if _, err := c.AddFunc("@midnight", func() {
		if err := m.PruneAll(pruneCtx, config.GetMediaRemoteCacheDays(), true); err != nil {
			log.Error(err)
			return
		}
	}); err != nil {
		pruneCancel()
		return fmt.Errorf("error starting media manager cleanup job: %s", err)
	}

	m.stopCronJobs = func() error {
		// Try to stop jobs gracefully by waiting til they're finished.
		stopCtx := c.Stop()

		select {
		case <-stopCtx.Done():
			log.Infof("media manager: cron finished jobs and stopped gracefully")
		case <-time.After(1 * time.Minute):
			log.Warnf("media manager: cron didn't stop after 60 seconds, force closing jobs")
			pruneCancel()
		}

		return nil
	}

	return nil
}
