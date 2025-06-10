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

package interaction_test

import (
	"strconv"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

const (
	rMediaPath    = "../../../testrig/media"
	rTemplatePath = "../../../web/template"
)

type InteractionTestSuite struct {
	suite.Suite

	testStatuses map[string]*gtsmodel.Status
	testAccounts map[string]*gtsmodel.Account
}

func (suite *InteractionTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.testStatuses = testrig.NewTestStatuses()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *InteractionTestSuite) TestInteractable() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	// Take zork's introduction post
	// as the base post for these tests.
	modelStatus := suite.testStatuses["local_account_1_status_1"]

	ctx := suite.T().Context()
	for i, test := range []struct {
		policy    *gtsmodel.InteractionPolicy
		account   *gtsmodel.Account
		likeable  gtsmodel.PolicyPermission
		replyable gtsmodel.PolicyPermission
		boostable gtsmodel.PolicyPermission
	}{
		{
			// Nil policy. Should all be fine as
			// it will fall back to the default then.
			policy:    nil,
			account:   suite.testAccounts["admin_account"],
			likeable:  gtsmodel.PolicyPermissionAutomaticApproval,
			replyable: gtsmodel.PolicyPermissionAutomaticApproval,
			boostable: gtsmodel.PolicyPermissionAutomaticApproval,
		},
		{
			// Nil canLike, everything else
			// restricted to author only.
			// Only the nil sub-policy should be OK.
			policy: &gtsmodel.InteractionPolicy{
				CanLike: nil,
				CanReply: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
				},
				CanAnnounce: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
				},
			},
			account:   suite.testAccounts["admin_account"],
			likeable:  gtsmodel.PolicyPermissionAutomaticApproval,
			replyable: gtsmodel.PolicyPermissionForbidden,
			boostable: gtsmodel.PolicyPermissionForbidden,
		},
		{
			// All restricted it's the author's own
			// account checking, so all should be fine.
			policy: &gtsmodel.InteractionPolicy{
				CanLike: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
				},
				CanReply: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
				},
				CanAnnounce: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
				},
			},
			account:   suite.testAccounts["local_account_1"],
			likeable:  gtsmodel.PolicyPermissionAutomaticApproval,
			replyable: gtsmodel.PolicyPermissionAutomaticApproval,
			boostable: gtsmodel.PolicyPermissionAutomaticApproval,
		},
		{
			// Followers can like automatically,
			// everything else requires manual approval.
			policy: &gtsmodel.InteractionPolicy{
				CanLike: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
						gtsmodel.PolicyValueFollowers,
					},
					ManualApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValuePublic,
					},
				},
				CanReply: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
					ManualApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValuePublic,
					},
				},
				CanAnnounce: &gtsmodel.PolicyRules{
					AutomaticApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValueAuthor,
					},
					ManualApproval: gtsmodel.PolicyValues{
						gtsmodel.PolicyValuePublic,
					},
				},
			},
			account:   suite.testAccounts["admin_account"],
			likeable:  gtsmodel.PolicyPermissionAutomaticApproval,
			replyable: gtsmodel.PolicyPermissionManualApproval,
			boostable: gtsmodel.PolicyPermissionManualApproval,
		},
	} {
		// Copy model status.
		status := new(gtsmodel.Status)
		*status = *modelStatus

		// Set test policy on it.
		status.InteractionPolicy = test.policy

		// Check likeableRes.
		likeableRes, err := testStructs.InteractionFilter.StatusLikeable(ctx, test.account, status)
		if err != nil {
			suite.FailNow(err.Error())
		}
		if likeableRes.Permission != test.likeable {
			suite.Fail(
				"failure in case "+strconv.FormatInt(int64(i), 10),
				"expected likeable result \"%s\", got \"%s\"",
				likeableRes.Permission, test.likeable,
			)
		}

		// Check replable.
		replyableRes, err := testStructs.InteractionFilter.StatusReplyable(ctx, test.account, status)
		if err != nil {
			suite.FailNow(err.Error())
		}
		if replyableRes.Permission != test.replyable {
			suite.Fail(
				"failure in case "+strconv.FormatInt(int64(i), 10),
				"expected replyable result \"%s\", got \"%s\"",
				replyableRes.Permission, test.replyable,
			)
		}

		// Check boostable.
		boostableRes, err := testStructs.InteractionFilter.StatusBoostable(ctx, test.account, status)
		if err != nil {
			suite.FailNow(err.Error())
		}
		if boostableRes.Permission != test.boostable {
			suite.Fail(
				"failure in case "+strconv.FormatInt(int64(i), 10),
				"expected boostable result \"%s\", got \"%s\"",
				boostableRes.Permission, test.boostable,
			)
		}
	}
}

func TestInteractionTestSuite(t *testing.T) {
	suite.Run(t, new(InteractionTestSuite))
}
