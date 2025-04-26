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

package paging_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/stretchr/testify/suite"
)

type PagingSuite struct {
	suite.Suite
}

func (suite *PagingSuite) TestPagingStandard() {
	config.SetHost("example.org")

	params := paging.ResponseParams{
		Items: make([]interface{}, 10, 10),
		Path:  "/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses",
		Next:  nextPage("01H11KA1DM2VH3747YDE7FV5HN", 10),
		Prev:  prevPage("01H11KBBVRRDYYC5KEPME1NP5R", 10),
	}

	resp := paging.PackageResponse(params)

	suite.Equal(make([]interface{}, 10, 10), resp.Items)
	suite.Equal(`<https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&max_id=01H11KA1DM2VH3747YDE7FV5HN>; rel="next", <https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&min_id=01H11KBBVRRDYYC5KEPME1NP5R>; rel="prev"`, resp.LinkHeader)
	suite.Equal(`https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&max_id=01H11KA1DM2VH3747YDE7FV5HN`, resp.NextLink)
	suite.Equal(`https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&min_id=01H11KBBVRRDYYC5KEPME1NP5R`, resp.PrevLink)
}

func (suite *PagingSuite) TestPagingNoLimit() {
	config.SetHost("example.org")

	params := paging.ResponseParams{
		Items: make([]interface{}, 10, 10),
		Path:  "/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses",
		Next:  nextPage("01H11KA1DM2VH3747YDE7FV5HN", 0),
		Prev:  prevPage("01H11KBBVRRDYYC5KEPME1NP5R", 0),
	}

	resp := paging.PackageResponse(params)

	suite.Equal(make([]interface{}, 10, 10), resp.Items)
	suite.Equal(`<https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?max_id=01H11KA1DM2VH3747YDE7FV5HN>; rel="next", <https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?min_id=01H11KBBVRRDYYC5KEPME1NP5R>; rel="prev"`, resp.LinkHeader)
	suite.Equal(`https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?max_id=01H11KA1DM2VH3747YDE7FV5HN`, resp.NextLink)
	suite.Equal(`https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?min_id=01H11KBBVRRDYYC5KEPME1NP5R`, resp.PrevLink)
}

func (suite *PagingSuite) TestPagingNoNextID() {
	config.SetHost("example.org")

	params := paging.ResponseParams{
		Items: make([]interface{}, 10, 10),
		Path:  "/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses",
		Prev:  prevPage("01H11KBBVRRDYYC5KEPME1NP5R", 10),
	}

	resp := paging.PackageResponse(params)

	suite.Equal(make([]interface{}, 10, 10), resp.Items)
	suite.Equal(`<https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&min_id=01H11KBBVRRDYYC5KEPME1NP5R>; rel="prev"`, resp.LinkHeader)
	suite.Equal(``, resp.NextLink)
	suite.Equal(`https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&min_id=01H11KBBVRRDYYC5KEPME1NP5R`, resp.PrevLink)
}

func (suite *PagingSuite) TestPagingNoPrevID() {
	config.SetHost("example.org")

	params := paging.ResponseParams{
		Items: make([]interface{}, 10, 10),
		Path:  "/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses",
		Next:  nextPage("01H11KA1DM2VH3747YDE7FV5HN", 10),
	}

	resp := paging.PackageResponse(params)

	suite.Equal(make([]interface{}, 10, 10), resp.Items)
	suite.Equal(`<https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&max_id=01H11KA1DM2VH3747YDE7FV5HN>; rel="next"`, resp.LinkHeader)
	suite.Equal(`https://example.org/api/v1/accounts/01H11KA68PM4NNYJEG0FJQ90R3/statuses?limit=10&max_id=01H11KA1DM2VH3747YDE7FV5HN`, resp.NextLink)
	suite.Equal(``, resp.PrevLink)
}

func TestPagingSuite(t *testing.T) {
	suite.Run(t, &PagingSuite{})
}

func nextPage(id string, limit int) *paging.Page {
	return &paging.Page{
		Max:   paging.MaxID(id),
		Limit: limit,
	}
}

func prevPage(id string, limit int) *paging.Page {
	return &paging.Page{
		Min:   paging.MinID(id),
		Limit: limit,
	}
}
