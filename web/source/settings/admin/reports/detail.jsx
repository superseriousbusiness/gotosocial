/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

"use strict";

const React = require("react");
const { useRoute, Redirect } = require("wouter");

const query = require("../../lib/query");

const FormWithData = require("../../lib/form/form-with-data");
const BackButton = require("../../components/back-button");

const Username = require("./username");

module.exports = function ReportDetail({ baseUrl }) {
	let [_match, params] = useRoute(`${baseUrl}/:reportId`);
	if (params?.reportId == undefined) {
		return <Redirect to={baseUrl} />;
	} else {
		return (
			<div className="report-detail">
				<h1>
					<BackButton to={baseUrl} /> Report
				</h1>
				<FormWithData
					dataQuery={query.useGetReportQuery}
					queryArg={params.reportId}
					DataForm={ReportDetailForm}
				/>
			</div>
		);
	}
};

function ReportDetailForm({ data: report }) {
	const from = report.account;
	const target = report.target_account;

	return (
		<div className="report detail">
			<Username user={from} /> reported <Username user={target} />
		</div>
	);
}