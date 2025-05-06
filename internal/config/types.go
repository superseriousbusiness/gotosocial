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

package config

import (
	"errors"
	"net/netip"
)

// IPPrefixes is a type-alias for []netip.Prefix
// to allow parsing by CLI "flag"-like utilities.
type IPPrefixes []netip.Prefix

func (p *IPPrefixes) Set(in string) error {
	if p == nil {
		return errors.New("nil receiver")
	}
	prefix, err := netip.ParsePrefix(in)
	if err != nil {
		return err
	}
	(*p) = append((*p), prefix)
	return nil
}

func (p *IPPrefixes) Strings() []string {
	if p == nil || len(*p) == 0 {
		return nil
	}
	strs := make([]string, len(*p))
	for i, prefix := range *p {
		strs[i] = prefix.String()
	}
	return strs
}
