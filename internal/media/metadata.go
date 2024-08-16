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

package media

import (
	"context"
	"os"
	"strings"

	terminator "codeberg.org/superseriousbusiness/exif-terminator"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// clearMetadata performs our best-attempt at cleaning metadata from
// input file. Depending on file type this may perform a full EXIF clean,
// or just a clean of global container level metadata records.
func clearMetadata(ctx context.Context, filepath string) error {
	var ext, outpath string

	// Generate cleaned output path MAINTAINING extension.
	if i := strings.IndexByte(filepath, '.'); i != -1 {
		outpath = filepath[:i] + "_cleaned" + filepath[i:]
		ext = filepath[i+1:]
	} else {
		return gtserror.New("input file missing extension")
	}

	switch ext {
	case "jpeg", "png", "webp":
		// For these few file types, we actually support
		// cleaning exif data using a native Go library.
		log.Debug(ctx, "cleaning with exif-terminator")
		err := terminateExif(outpath, filepath, ext)
		if err == nil {
			// No problem.
			break
		}

		log.Warnf(ctx, "error cleaning with exif-terminator, falling back to ffmpeg: %v", err)
		fallthrough

	default:
		// For all other types, best-effort clean with ffmpeg.
		log.Debug(ctx, "cleaning with ffmpeg -map_metadata -1")
		err := ffmpegClearMetadata(ctx, outpath, filepath)
		if err != nil {
			return err
		}
	}

	// Move the new output file path to original location.
	if err := os.Rename(outpath, filepath); err != nil {
		return gtserror.Newf("error renaming %s -> %s: %w", outpath, filepath, err)
	}

	return nil
}

// terminateExif cleans exif data from file at input path, into file
// at output path, using given file extension to determine cleaning type.
func terminateExif(outpath, inpath string, ext string) error {
	// Open input file at given path.
	inFile, err := os.Open(inpath)
	if err != nil {
		return gtserror.Newf("error opening input file %s: %w", inpath, err)
	}
	defer inFile.Close()

	// Open output file at given path.
	outFile, err := os.Create(outpath)
	if err != nil {
		return gtserror.Newf("error opening output file %s: %w", outpath, err)
	}
	defer outFile.Close()

	// Terminate EXIF data from 'inFile' -> 'outFile'.
	err = terminator.TerminateInto(outFile, inFile, ext)
	if err != nil {
		return gtserror.Newf("error terminating exif data: %w", err)
	}

	return nil
}
