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

package account

import (
	"cmp"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-bytesize"
)

var (
	themeTitleRegex       = regexp.MustCompile(`(?m)^\ *theme-title:(.*)$`)
	themeDescriptionRegex = regexp.MustCompile(`(?m)^\ *theme-description:(.*)$`)
)

// GetThemes returns available account css themes.
func (p *Processor) ThemesGet() []apimodel.Theme {
	return p.converter.ThemesToAPIThemes(p.themes.SortedByTitle)
}

// Themes represents an in-memory
// storage structure for themes.
type Themes struct {
	// Themes sorted alphabetically
	// by title (case insensitive).
	SortedByTitle []*gtsmodel.Theme

	// ByFileName contains themes retrievable
	// by their filename eg., `light-blurple.css`.
	ByFileName map[string]*gtsmodel.Theme
}

// PopulateThemes parses available account CSS
// themes from the web assets themes directory.
func PopulateThemes() *Themes {
	webAssetsAbsFilePath, err := filepath.Abs(config.GetWebAssetBaseDir())
	if err != nil {
		log.Panicf(nil, "error getting abs path for web assets: %v", err)
	}

	themesAbsFilePath := filepath.Join(webAssetsAbsFilePath, "themes")
	themesFiles, err := os.ReadDir(themesAbsFilePath)
	if err != nil {
		log.Warnf(nil, "error reading themes at %s: %v", themesAbsFilePath, err)
		return nil
	}

	themes := &Themes{
		ByFileName: make(map[string]*gtsmodel.Theme),
	}

	for _, f := range themesFiles {
		// Ignore nested directories.
		if f.IsDir() {
			continue
		}

		// Ignore weird files.
		info, err := f.Info()
		if err != nil {
			continue
		}

		// Ignore really big files.
		if info.Size() > int64(bytesize.MiB) {
			continue
		}

		// Get just the name of the
		// file, eg `blurple-light.css`.
		fileName := f.Name()

		// Get just the `.css` part.
		extensionWithDot := filepath.Ext(fileName)

		// Remove any leading `.`
		extension := strings.TrimPrefix(extensionWithDot, ".")

		// Ignore non-css files.
		if extension != "css" {
			continue
		}

		// Load the file contents.
		path := filepath.Join(themesAbsFilePath, fileName)
		contents, err := os.ReadFile(path)
		if err != nil {
			log.Warnf(nil, "error reading css theme at %s: %v", path, err)
			continue
		}

		// Try to parse a title and description
		// for this theme from the file itself.
		var themeTitle string
		titleMatches := themeTitleRegex.FindSubmatch(contents)
		if len(titleMatches) == 2 {
			themeTitle = strings.TrimSpace(string(titleMatches[1]))
		} else {
			// Fall back to file name
			// without `.css` suffix.
			themeTitle = strings.TrimSuffix(fileName, ".css")
		}

		var themeDescription string
		descMatches := themeDescriptionRegex.FindSubmatch(contents)
		if len(descMatches) == 2 {
			themeDescription = strings.TrimSpace(string(descMatches[1]))
		}

		theme := &gtsmodel.Theme{
			Title:       themeTitle,
			Description: themeDescription,
			FileName:    fileName,
		}

		themes.SortedByTitle = append(themes.SortedByTitle, theme)
		themes.ByFileName[fileName] = theme
	}

	// Sort themes alphabetically
	// by title (case insensitive).
	slices.SortFunc(themes.SortedByTitle, func(a, b *gtsmodel.Theme) int {
		return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})

	return themes
}
