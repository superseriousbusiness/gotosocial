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
const { Link, Switch, Route } = require("wouter");

const query = require("../../lib/query");

const FormWithData = require("../../lib/form/form-with-data");

const ReportDetail = require("./detail");
const Username = require("./username");

const baseUrl = "/settings/admin/reports";

module.exports = function Reports() {
	return (
		<div className="reports">
			<Switch>
				<Route path={`${baseUrl}/:reportId`}>
					<ReportDetail baseUrl={baseUrl} />
				</Route>
				<ReportOverview baseUrl={baseUrl} />
			</Switch>
		</div>
	);
};

function ReportOverview({ _baseUrl }) {
	return (
		<>
			<h1>Reports</h1>
			<div>
				<div className="info">
					<i className="fa fa-fw fa-exclamation-triangle" aria-hidden="true"></i>
					<p>
						<b>This interface is currently very limited</b>, only providing a basic overview. <br />
						Work is in progress on a more full-fledged moderation experience.
					</p>
				</div>
				<p>
					Here you can view and resolve reports made to your instance, originating from local and remote users.
				</p>
			</div>
			<FormWithData
				dataQuery={query.useListReportsQuery}
				DataForm={ReportsList}
			/>
		</>
	);
}

function ReportsList({ data: reports }) {
	return (
		<div className="list">
			{reports.map((report) => (
				<ReportEntry key={report.id} report={report} />
			))}
		</div>
	);
}

function ReportEntry({ report }) {
	const from = report.account;
	const target = report.target_account;

	let comment = report.comment.length > 200
		? report.comment.slice(0, 200) + "..."
		: report.comment;

	return (
		<Link to={`${baseUrl}/${report.id}`}>
			<a className={`report entry${report.action_taken ? " resolved" : ""}`}>
				<div className="byline">
					<div className="users">
						<Username user={from} link={false} /> reported <Username user={target} link={false} />
					</div>
					<h3 className="status">
						{report.action_taken ? "Resolved" : "Open"}
					</h3>
				</div>
				<div className="details">
					<b>Created: </b>
					<span>{new Date(report.created_at).toLocaleString()}</span>

					<b>Reason: </b>
					{comment.length > 0
						? <p>{comment}</p>
						: <i className="no-comment">none provided</i>
					}
				</div>
			</a>
		</Link>
	);
}