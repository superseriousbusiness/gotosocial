/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

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

import React from "react";
import { Link } from "wouter";
import FormWithData from "../../../lib/form/form-with-data";
import Username from "../../../components/username";
import { useListReportsQuery } from "../../../lib/query/admin/reports";

export function ReportOverview({ }) {
	return (
		<FormWithData
			dataQuery={useListReportsQuery}
			DataForm={ReportsList}
		/>
	);
}

function ReportsList({ data: reports }) {
	return (
		<div className="reports">
			<div className="form-section-docs">
				<h1>Reports</h1>
				<p>
					Here you can view and resolve reports made to your
					instance, originating from local and remote users.
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/admin/settings/#reports"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about this (opens in a new tab)
				</a>
			</div>
			<div className="list">
				{reports.map((report) => (
					<ReportEntry key={report.id} report={report} />
				))}
			</div>
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
		<Link
			to={`/${report.id}`}
			className="nounderline"
		>
			<div className={`report entry${report.action_taken ? " resolved" : ""}`}>
				<div className="byline">
					<div className="usernames">
						<Username account={from} /> reported <Username account={target} />
					</div>
					<h3 className="report-status">
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
			</div>
		</Link>
	);
}
