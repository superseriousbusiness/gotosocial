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

package status

import (
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/text"
)

// getFilterableFields returns text fields from
// a status that we might want to filter on:
//
//   - content warning
//   - content (converted to plaintext from HTML)
//   - media descriptions
//   - poll options
//
// Each field should be filtered separately. This avoids
// scenarios where false-positive multiple-word matches
// can be made by matching the last word of one field
// combined with the first word of the next field together.
func getFilterableFields(status *gtsmodel.Status) []string {

	// Estimate expected no of status fields.
	fieldCount := 2 + len(status.Attachments)
	if status.Poll != nil {
		fieldCount += len(status.Poll.Options)
	}
	fields := make([]string, 0, fieldCount)

	// Append content warning / title.
	if status.ContentWarning != "" {
		fields = append(fields, status.ContentWarning)
	}

	// Status content. Though we have raw text
	// available for statuses created on our
	// instance, use the plaintext version to
	// remove markdown-formatting characters
	// and ensure more consistent filtering.
	if status.Content != "" {
		text := text.ParseHTMLToPlain(status.Content)
		if text != "" {
			fields = append(fields, text)
		}
	}

	// Media descriptions, only where they are set.
	for _, attachment := range status.Attachments {
		if attachment.Description != "" {
			fields = append(fields, attachment.Description)
		}
	}

	// Non-empty poll options.
	if status.Poll != nil {
		for _, opt := range status.Poll.Options {
			if opt != "" {
				fields = append(fields, opt)
			}
		}
	}

	return fields
}
