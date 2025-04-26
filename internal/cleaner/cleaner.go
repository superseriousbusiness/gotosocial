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

package cleaner

import (
	"context"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
)

const (
	selectLimit = 50
)

type Cleaner struct {
	state *state.State
	emoji Emoji
	media Media
}

func New(state *state.State) *Cleaner {
	c := new(Cleaner)
	c.state = state
	c.emoji.Cleaner = c
	c.media.Cleaner = c
	return c
}

// Emoji returns the emoji set of cleaner utilities.
func (c *Cleaner) Emoji() *Emoji {
	return &c.emoji
}

// Media returns the media set of cleaner utilities.
func (c *Cleaner) Media() *Media {
	return &c.media
}

// haveFiles returns whether all of the provided files exist within current storage.
func (c *Cleaner) haveFiles(ctx context.Context, files ...string) (bool, error) {
	for _, path := range files {
		if path == "" {
			// File not stored.
			return false, nil
		}

		// Check whether each file exists in storage.
		have, err := c.state.Storage.Has(ctx, path)
		if err != nil {
			return false, gtserror.Newf("error checking storage for %s: %w", path, err)
		}

		if !have {
			// Missing file(s).
			return false, nil
		}
	}
	return true, nil
}

// removeFiles removes the provided files, returning the number of them returned.
func (c *Cleaner) removeFiles(ctx context.Context, files ...string) (int, error) {
	if gtscontext.DryRun(ctx) {
		// Dry run, do nothing.
		return len(files), nil
	}

	var (
		errs  gtserror.MultiError
		count int
	)

	for _, path := range files {
		if path == "" {
			// not stored.
			continue
		}

		// Remove each provided storage path.
		log.Debugf(ctx, "removing file: %s", path)
		err := c.state.Storage.Delete(ctx, path)
		if err != nil && !storage.IsNotFound(err) {
			errs.Appendf("error removing %s: %w", path, err)
			continue
		}

		// Incr.
		count++
	}

	// Wrap the combined error slice.
	if err := errs.Combine(); err != nil {
		return count, gtserror.Newf("error(s) removing files: %w", err)
	}

	return count, nil
}

// ScheduleJobs schedules cleaning
// jobs using configured parameters.
//
// Returns an error if `MediaCleanupFrom`
// is not a valid format (hh:mm:ss).
func (c *Cleaner) ScheduleJobs() error {
	const hourMinute = "15:04"

	var (
		now            = time.Now()
		cleanupEvery   = config.GetMediaCleanupEvery()
		cleanupFromStr = config.GetMediaCleanupFrom()
	)

	// Parse cleanupFromStr as hh:mm.
	// Resulting time will be on 1 Jan year zero.
	cleanupFrom, err := time.Parse(hourMinute, cleanupFromStr)
	if err != nil {
		return gtserror.Newf(
			"error parsing '%s' in time format 'hh:mm': %w",
			cleanupFromStr, err,
		)
	}

	// Time travel from
	// year zero, groovy.
	firstCleanupAt := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		cleanupFrom.Hour(),
		cleanupFrom.Minute(),
		0,
		0,
		now.Location(),
	)

	// Ensure first cleanup is in the future.
	for firstCleanupAt.Before(now) {
		firstCleanupAt = firstCleanupAt.Add(cleanupEvery)
	}

	fn := func(ctx context.Context, start time.Time) {
		log.Info(ctx, "starting media clean")
		c.Media().All(ctx, config.GetMediaRemoteCacheDays())
		c.Emoji().All(ctx, config.GetMediaRemoteCacheDays())
		log.Infof(ctx, "finished media clean after %s", time.Since(start))
	}

	log.Infof(nil,
		"scheduling media clean to run every %s, starting from %s; next clean will run at %s",
		cleanupEvery, cleanupFromStr, firstCleanupAt,
	)

	// Schedule the cleaning to execute according to schedule.
	if !c.state.Workers.Scheduler.AddRecurring(
		"@mediacleanup",
		firstCleanupAt,
		cleanupEvery,
		fn,
	) {
		panic("failed to schedule @mediacleanup")
	}

	return nil
}
