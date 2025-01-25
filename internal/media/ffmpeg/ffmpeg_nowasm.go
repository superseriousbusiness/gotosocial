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

//go:build nowasm

package ffmpeg

import (
	"context"
	"os/exec"
)

// ffmpegRunner limits the number of
// ffmpeg WebAssembly instances that
// may be concurrently running, in
// order to reduce memory usage.
var ffmpegRunner runner

// InitFfmpeg looks for a local copy of ffmpeg in path, and prepares
// the runner to only allow max given concurrent running instances.
func InitFfmpeg(ctx context.Context, max int) error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}
	ffmpegRunner.Init(max)
	return nil
}

// Ffmpeg runs the given arguments with an instance of ffmpeg.
func Ffmpeg(ctx context.Context, args Args) (uint32, error) {
	return ffmpegRunner.Run(ctx, func() (uint32, error) {
		return runCmd(ctx, "ffmpeg", args)
	})
}
