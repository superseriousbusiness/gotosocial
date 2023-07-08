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

package oidc

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// Claims represents claims as found in an id_token returned from an OIDC flow.
//
// All known claims are stored in their respective field. All unknown claims, as
// long as they were either a string or a list of string, are stored keyed by the
// claim in the Attrs map.
type Claims struct {
	Sub               string              `json:"sub"`
	Email             string              `json:"email"`
	EmailVerified     bool                `json:"email_verified"`
	Name              string              `json:"name"`
	PreferredUsername string              `json:"preferred_username"`
	Attrs             map[string][]string `json:"-"`
}

// UnmarshalJSON implements the necessary logic to store any unknown
// keys in the Attrs map
func (c *Claims) UnmarshalJSON(data []byte) error {
	// Create a new type so we don't recursively call UnmarshalJSON, but
	// can reuse the type and decode based on the existing struct tags
	type claims2 Claims
	if err := json.Unmarshal(data, (*claims2)(c)); err != nil {
		return fmt.Errorf("failed to decode JSON as a claims object: %s", err)
	}

	// Decode again, but now as json.RawMessage. Though we decode twice
	// json.RawMessage is just []byte, so no actual decoding occurs. Only
	// the keys in the top level object are collected
	tmp := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		// Given that we already successfully unmarshalled into Claims, this error
		// can realistically never occur as we know it's an object and keys in JSON
		// must be strings
		return fmt.Errorf("failed to decode as JSON object: %s", err)
	}

	// Remove any known fields, so all that is left are fields that we
	// didn't have a direct struct mapping for. Delete is a no-op for
	// keys that don't exist or are nil so can be safely called for all
	// known fields even if some weren't present
	delete(tmp, "sub")
	delete(tmp, "email")
	delete(tmp, "email_verified")
	delete(tmp, "name")
	delete(tmp, "preferred_username")

	// initialise the Attrs map so we can set values on it
	c.Attrs = map[string][]string{}

	for key, value := range tmp {
		var ls []string
		if err := json.Unmarshal(value, &ls); err != nil {
			var s string
			if nerr := json.Unmarshal(value, &s); nerr != nil {
				// The value couldn't be unmarshalled as either a string or a slice of
				// string. Whatever it is, it's in a format we don't know or understand
				// so we won't be able to use it. Move on to the next key.
				continue
			}
			c.Attrs[key] = []string{s}
			// The value decoded successfully as a string, move on to the next key
			continue
		}
		c.Attrs[key] = ls
	}
	return nil
}

func init() {
	gob.Register(&Claims{})
}
