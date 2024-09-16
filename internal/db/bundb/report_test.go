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

package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ReportTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *ReportTestSuite) TestGetReportByID() {
	report, err := suite.db.GetReportByID(context.Background(), suite.testReports["local_account_2_report_remote_account_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(report)
	suite.NotNil(report.Account)
	suite.NotNil(report.TargetAccount)
	suite.Zero(report.ActionTakenAt)
	suite.Nil(report.ActionTakenByAccount)
	suite.Empty(report.ActionTakenByAccountID)
	suite.NotEmpty(report.URI)
}

func (suite *ReportTestSuite) TestGetReportByURI() {
	report, err := suite.db.GetReportByID(context.Background(), suite.testReports["remote_account_1_report_local_account_2"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(report)
	suite.NotNil(report.Account)
	suite.NotNil(report.TargetAccount)
	suite.NotZero(report.ActionTakenAt)
	suite.NotNil(report.ActionTakenByAccount)
	suite.NotEmpty(report.ActionTakenByAccountID)
	suite.NotEmpty(report.URI)
}

func (suite *ReportTestSuite) TestGetAllReports() {
	reports, err := suite.db.GetReports(
		context.Background(),
		nil,
		"",
		"",
		&paging.Page{},
	)
	suite.NoError(err)
	suite.NotEmpty(reports)
}

func (suite *ReportTestSuite) TestReportPagingDown() {
	// Get one from the top.
	reports1, err := suite.db.GetReports(
		context.Background(),
		nil,
		"",
		"",
		&paging.Page{
			Limit: 1,
		},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if l := len(reports1); l != 1 {
		suite.FailNowf("", "expected reports len 1, got %d", l)
	}
	id1 := reports1[0].ID

	// Use this one to page down.
	reports2, err := suite.db.GetReports(
		context.Background(),
		nil,
		"",
		"",
		&paging.Page{
			Limit: 1,
			Max:   paging.MaxID(id1),
		},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if l := len(reports2); l != 1 {
		suite.FailNowf("", "expected reports len 1, got %d", l)
	}
	id2 := reports2[0].ID

	suite.Greater(id1, id2)
}

func (suite *ReportTestSuite) TestReportPagingUp() {
	// Get one from the bottom.
	reports1, err := suite.db.GetReports(
		context.Background(),
		nil,
		"",
		"",
		&paging.Page{
			Limit: 1,
			Min:   paging.MinID(id.Lowest),
		},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if l := len(reports1); l != 1 {
		suite.FailNowf("", "expected reports len 1, got %d", l)
	}
	id1 := reports1[0].ID

	// Use this one to page up.
	reports2, err := suite.db.GetReports(
		context.Background(),
		nil,
		"",
		"",
		&paging.Page{
			Limit: 1,
			Min:   paging.MinID(id1),
		},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if l := len(reports2); l != 1 {
		suite.FailNowf("", "expected reports len 1, got %d", l)
	}
	id2 := reports2[0].ID

	suite.Less(id1, id2)
}

func (suite *ReportTestSuite) TestGetAllReportsByAccountID() {
	accountID := suite.testAccounts["local_account_2"].ID
	reports, err := suite.db.GetReports(
		context.Background(),
		nil,
		accountID,
		"",
		&paging.Page{},
	)
	suite.NoError(err)
	suite.NotEmpty(reports)
	for _, r := range reports {
		suite.Equal(accountID, r.AccountID)
	}
}

func (suite *ReportTestSuite) TestPutReport() {
	ctx := context.Background()

	reportID := "01GP3ECY8QJD8DBJSS8B1CR0AX"
	report := &gtsmodel.Report{
		ID:              reportID,
		CreatedAt:       testrig.TimeMustParse("2022-05-14T12:20:03+02:00"),
		UpdatedAt:       testrig.TimeMustParse("2022-05-14T12:20:03+02:00"),
		URI:             "http://localhost:8080/01GP3ECY8QJD8DBJSS8B1CR0AX",
		AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
		TargetAccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
		Comment:         "another report",
		StatusIDs:       []string{"01FVW7JHQFSFK166WWKR8CBA6M"},
		Forwarded:       util.Ptr(true),
	}

	err := suite.db.PutReport(ctx, report)
	suite.NoError(err)
}

func (suite *ReportTestSuite) TestUpdateReport() {
	ctx := context.Background()

	report := &gtsmodel.Report{}
	*report = *suite.testReports["local_account_2_report_remote_account_1"]
	report.ActionTaken = "nothing"
	report.ActionTakenByAccountID = suite.testAccounts["admin_account"].ID
	report.ActionTakenAt = testrig.TimeMustParse("2022-05-14T12:20:03+02:00")

	if err := suite.db.UpdateReport(ctx, report, "action_taken", "action_taken_by_account_id", "action_taken_at"); err != nil {
		suite.FailNow(err.Error())
	}

	dbReport, err := suite.db.GetReportByID(ctx, report.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(dbReport)
	suite.NotNil(dbReport.Account)
	suite.NotNil(dbReport.TargetAccount)
	suite.NotZero(dbReport.ActionTakenAt)
	suite.NotNil(dbReport.ActionTakenByAccount)
	suite.NotEmpty(dbReport.ActionTakenByAccountID)
	suite.NotEmpty(dbReport.URI)
}

func (suite *ReportTestSuite) TestUpdateReportAllColumns() {
	ctx := context.Background()

	report := &gtsmodel.Report{}
	*report = *suite.testReports["local_account_2_report_remote_account_1"]
	report.ActionTaken = "nothing"
	report.ActionTakenByAccountID = suite.testAccounts["admin_account"].ID
	report.ActionTakenAt = testrig.TimeMustParse("2022-05-14T12:20:03+02:00")

	if err := suite.db.UpdateReport(ctx, report); err != nil {
		suite.FailNow(err.Error())
	}

	dbReport, err := suite.db.GetReportByID(ctx, report.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(dbReport)
	suite.NotNil(dbReport.Account)
	suite.NotNil(dbReport.TargetAccount)
	suite.NotZero(dbReport.ActionTakenAt)
	suite.NotNil(dbReport.ActionTakenByAccount)
	suite.NotEmpty(dbReport.ActionTakenByAccountID)
	suite.NotEmpty(dbReport.URI)
}

func (suite *ReportTestSuite) TestDeleteReport() {
	if err := suite.db.DeleteReportByID(context.Background(), suite.testReports["remote_account_1_report_local_account_2"].ID); err != nil {
		suite.FailNow(err.Error())
	}

	report, err := suite.db.GetReportByID(context.Background(), suite.testReports["remote_account_1_report_local_account_2"].ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(report)
}

func TestReportTestSuite(t *testing.T) {
	suite.Run(t, new(ReportTestSuite))
}
