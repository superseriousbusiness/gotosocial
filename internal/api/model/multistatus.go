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

package model

// MultiStatus models a multistatus HTTP response body.
// This model should be transmitted along with http code
// 207 MULTI-STATUS to indicate a mixture of responses.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/207
//
// swagger:model multiStatus
type MultiStatus struct {
	Data     []MultiStatusEntry  `json:"data"`
	Metadata MultiStatusMetadata `json:"metadata"`
}

// MultiStatusEntry models one entry in multistatus data.
// It can model either a success or a failure. The type
// and value of `Resource` is left to the discretion of
// the caller, but at minimum it should be expected to be
// JSON-serializable.
//
// swagger:model multiStatusEntry
type MultiStatusEntry struct {
	// The resource/result for this entry.
	// Value may be any type, check the docs
	// per endpoint to see which to expect.
	Resource any `json:"resource"`
	// Message/error message for this entry.
	Message string `json:"message"`
	// HTTP status code of this entry.
	Status int `json:"status"`
}

// MultiStatusMetadata models an at-a-glance summary of
// the data contained in the MultiStatus.
//
// swagger:model multiStatusMetadata
type MultiStatusMetadata struct {
	// Success count + failure count.
	Total int `json:"total"`
	// Count of successful results (2xx).
	Success int `json:"success"`
	// Count of unsuccessful results (!2xx).
	Failure int `json:"failure"`
}

// NewMultiStatus returns a new MultiStatus API model with
// the provided entries, which will be iterated through to
// look for 2xx and non 2xx status codes, in order to count
// successes and failures.
func NewMultiStatus(entries []MultiStatusEntry) *MultiStatus {
	var (
		successCount int
		failureCount int
		total        = len(entries)
	)

	for _, e := range entries {
		// Outside 2xx range = failure.
		if e.Status > 299 || e.Status < 200 {
			failureCount++
		} else {
			successCount++
		}
	}

	return &MultiStatus{
		Data: entries,
		Metadata: MultiStatusMetadata{
			Total:   total,
			Success: successCount,
			Failure: failureCount,
		},
	}
}
