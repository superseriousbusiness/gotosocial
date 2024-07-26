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

package util_test

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type PunyTestSuite struct {
	suite.Suite
}

func (suite *PunyTestSuite) TestMatches() {
	for i, testCase := range []struct {
		expect *url.URL
		actual []*url.URL
		match  bool
	}{
		{
			expect: testrig.URLMustParse("https://%D5%A9%D5%B8%D6%82%D5%A9.%D5%B0%D5%A1%D5%B5/@ankap"),
			actual: []*url.URL{
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/users/ankap"),
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/@ankap"),
			},
			match: true,
		},
		{
			expect: testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/@ankap"),
			actual: []*url.URL{
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/users/ankap"),
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/@ankap"),
			},
			match: true,
		},
		{
			expect: testrig.URLMustParse("https://թութ.հայ/@ankap"),
			actual: []*url.URL{
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/users/ankap"),
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/@ankap"),
			},
			match: true,
		},
		{
			expect: testrig.URLMustParse("https://թութ.հայ/@ankap"),
			actual: []*url.URL{
				testrig.URLMustParse("https://example.org/users/ankap"),
				testrig.URLMustParse("https://%D5%A9%D5%B8%D6%82%D5%A9.%D5%B0%D5%A1%D5%B5/@ankap"),
			},
			match: true,
		},
		{
			expect: testrig.URLMustParse("https://example.org/@ankap"),
			actual: []*url.URL{
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/users/ankap"),
				testrig.URLMustParse("https://xn--69aa8bzb.xn--y9a3aq/@ankap"),
			},
			match: false,
		},
	} {
		matches, err := util.URIMatches(
			testCase.expect,
			testCase.actual...,
		)
		if err != nil {
			suite.FailNow(err.Error())
		}

		if matches != testCase.match {
			suite.Failf(
				"case "+strconv.Itoa(i)+" matches not equal expected",
				"wanted %t, got %t",
				testCase.match, matches,
			)
		}
	}
}

func TestPunyTestSuite(t *testing.T) {
	suite.Run(t, new(PunyTestSuite))
}
