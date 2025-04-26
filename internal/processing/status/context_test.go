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

package status_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/processing/status"
	"github.com/stretchr/testify/suite"
)

type topoSortTestSuite struct {
	suite.Suite
}

func statusIDs(apiStatuses []*gtsmodel.Status) []string {
	ids := make([]string, 0, len(apiStatuses))
	for _, apiStatus := range apiStatuses {
		ids = append(ids, apiStatus.ID)
	}
	return ids
}

func (suite *topoSortTestSuite) TestBranched() {
	// https://commons.wikimedia.org/wiki/File:Sorted_binary_tree_ALL_RGB.svg
	f := &gtsmodel.Status{ID: "F"}
	b := &gtsmodel.Status{ID: "B", InReplyToID: f.ID}
	a := &gtsmodel.Status{ID: "A", InReplyToID: b.ID}
	d := &gtsmodel.Status{ID: "D", InReplyToID: b.ID}
	c := &gtsmodel.Status{ID: "C", InReplyToID: d.ID}
	e := &gtsmodel.Status{ID: "E", InReplyToID: d.ID}
	g := &gtsmodel.Status{ID: "G", InReplyToID: f.ID}
	i := &gtsmodel.Status{ID: "I", InReplyToID: g.ID}
	h := &gtsmodel.Status{ID: "H", InReplyToID: i.ID}

	expected := statusIDs([]*gtsmodel.Status{f, b, a, d, c, e, g, i, h})
	list := []*gtsmodel.Status{a, b, c, d, e, f, g, h, i}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestBranchedWithSelfReplyChain() {
	targetAccount := &gtsmodel.Account{ID: "1"}
	otherAccount := &gtsmodel.Account{ID: "2"}

	f := &gtsmodel.Status{
		ID:      "F",
		Account: targetAccount,
	}
	b := &gtsmodel.Status{
		ID:                 "B",
		Account:            targetAccount,
		AccountID:          targetAccount.ID,
		InReplyToID:        f.ID,
		InReplyToAccountID: f.Account.ID,
	}
	d := &gtsmodel.Status{
		ID:                 "D",
		Account:            targetAccount,
		AccountID:          targetAccount.ID,
		InReplyToID:        b.ID,
		InReplyToAccountID: b.Account.ID,
	}
	e := &gtsmodel.Status{
		ID:                 "E",
		Account:            targetAccount,
		AccountID:          targetAccount.ID,
		InReplyToID:        d.ID,
		InReplyToAccountID: d.Account.ID,
	}
	c := &gtsmodel.Status{
		ID:                 "C",
		Account:            otherAccount,
		AccountID:          otherAccount.ID,
		InReplyToID:        d.ID,
		InReplyToAccountID: d.Account.ID,
	}
	a := &gtsmodel.Status{
		ID:                 "A",
		Account:            otherAccount,
		AccountID:          otherAccount.ID,
		InReplyToID:        b.ID,
		InReplyToAccountID: b.Account.ID,
	}
	g := &gtsmodel.Status{
		ID:                 "G",
		Account:            otherAccount,
		AccountID:          otherAccount.ID,
		InReplyToID:        f.ID,
		InReplyToAccountID: f.Account.ID,
	}
	i := &gtsmodel.Status{
		ID:                 "I",
		Account:            targetAccount,
		AccountID:          targetAccount.ID,
		InReplyToID:        g.ID,
		InReplyToAccountID: g.Account.ID,
	}
	h := &gtsmodel.Status{
		ID:                 "H",
		Account:            otherAccount,
		AccountID:          otherAccount.ID,
		InReplyToID:        i.ID,
		InReplyToAccountID: i.Account.ID,
	}

	expected := statusIDs([]*gtsmodel.Status{f, b, d, e, c, a, g, i, h})
	list := []*gtsmodel.Status{a, b, c, d, e, f, g, h, i}
	status.TopoSort(list, targetAccount.ID)
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestDisconnected() {
	f := &gtsmodel.Status{ID: "F"}
	b := &gtsmodel.Status{ID: "B", InReplyToID: f.ID}
	dID := "D"
	e := &gtsmodel.Status{ID: "E", InReplyToID: dID}

	expected := statusIDs([]*gtsmodel.Status{e, f, b})
	list := []*gtsmodel.Status{b, e, f}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestTrivialCycle() {
	xID := "X"
	x := &gtsmodel.Status{ID: xID, InReplyToID: xID}

	expected := statusIDs([]*gtsmodel.Status{x})
	list := []*gtsmodel.Status{x}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.ElementsMatch(expected, actual)
}

func (suite *topoSortTestSuite) TestCycle() {
	yID := "Y"
	x := &gtsmodel.Status{ID: "X", InReplyToID: yID}
	y := &gtsmodel.Status{ID: yID, InReplyToID: x.ID}

	expected := statusIDs([]*gtsmodel.Status{x, y})
	list := []*gtsmodel.Status{x, y}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.ElementsMatch(expected, actual)
}

func (suite *topoSortTestSuite) TestMixedCycle() {
	yID := "Y"
	x := &gtsmodel.Status{ID: "X", InReplyToID: yID}
	y := &gtsmodel.Status{ID: yID, InReplyToID: x.ID}
	z := &gtsmodel.Status{ID: "Z"}

	expected := statusIDs([]*gtsmodel.Status{x, y, z})
	list := []*gtsmodel.Status{x, y, z}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.ElementsMatch(expected, actual)
}

func (suite *topoSortTestSuite) TestEmpty() {
	expected := statusIDs([]*gtsmodel.Status{})
	list := []*gtsmodel.Status{}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestNil() {
	expected := statusIDs(nil)
	var list []*gtsmodel.Status
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func TestTopoSortTestSuite(t *testing.T) {
	suite.Run(t, &topoSortTestSuite{})
}
