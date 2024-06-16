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

import React, { ReactNode, useEffect, useMemo } from "react";

import { useLazySearchReportsQuery } from "../../../lib/query/admin/reports";
import { useTextInput } from "../../../lib/form";
import { PageableList } from "../../../components/pageable-list";
import { Select } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import { Link, useLocation, useSearch } from "wouter";
import Username from "../../../components/username";
import { AdminReport } from "../../../lib/types/report";

export default function ReportsSearch() {
	return (
		<div className="reports-view">
			<h1>Reports Search</h1>
			<span>
				You can use the form below to search through reports
				created by, or directed towards, accounts on this instance.
			</span>
			<ReportSearchForm />
		</div>
	);
}

function ReportSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const [ searchReports, searchRes ] = useLazySearchReportsQuery();

	// Populate search form using values from
	// urlQueryParams, to allow paging.
	const form = {
		resolved: useTextInput("resolved", { defaultValue: urlQueryParams.get("resolved") ?? "" }),
		account_id: useTextInput("account_id", { defaultValue: urlQueryParams.get("account_id") ?? "" }),
		target_account_id: useTextInput("target_account_id", { defaultValue: urlQueryParams.get("target_account_id") ?? "" }),
		limit: useTextInput("limit", { defaultValue: urlQueryParams.get("limit") ?? "20" })
	};

	// On mount, if urlQueryParams were provided,
	// trigger the search. For example, if page
	// was accessed at /search?origin=local&limit=20,
	// then run a search with origin=local and
	// limit=20 and immediately render the results.
	useEffect(() => {
		if (urlQueryParams.size > 0) {
			searchReports(Object.fromEntries(urlQueryParams), true);
		}
	}, [urlQueryParams, searchReports]);

	// Rather than triggering the search directly,
	// the "submit" button changes the location
	// based on form field params, and lets the
	// useEffect hook above actually do the search.
	function submitQuery(e) {
		e.preventDefault();
		
		// Parse query parameters.
		const entries = Object.entries(form).map(([k, v]) => {
			// Take only defined form fields.
			if (v.value === undefined || v.value.length === 0) {
				return null;
			}
			return [[k, v.value]];
		}).flatMap(kv => {
			// Remove any nulls.
			return kv || [];
		});

		const searchParams = new URLSearchParams(entries);
		setLocation(location + "?" + searchParams.toString());
	}

	// Location to return to when user clicks "back" on the detail view.
	const backLocation = location + (urlQueryParams ? `?${urlQueryParams}` : "");
	
	// Function to map an item to a list entry.
	function itemToEntry(report: AdminReport): ReactNode {
		return (
			<ReportEntry
				report={report}
				backLocation={backLocation}
			/>
		);
	}

	return (
		<>
			<form
				onSubmit={submitQuery}
				// Prevent password managers
				// trying to fill in fields.
				autoComplete="off"
			>
				<Select
					field={form.resolved}
					label="Report status"
					options={
						<>
							<option value="">Any</option>
							<option value="false">Unresolved only</option>
							<option value="true">Resolved only</option>
						</>
					}
				></Select>
				<MutationButton
					disabled={false}
					label={"Search"}
					result={searchRes}
				/>
			</form>
			<PageableList
				isLoading={searchRes.isLoading}
				isFetching={searchRes.isFetching}
				isSuccess={searchRes.isSuccess}
				items={searchRes.data?.accounts}
				itemToEntry={itemToEntry}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage={<b>No reports found that match your query.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}


// function ReportsList({ data: reports }) {
// 	return (
// 		<div className="reports">
// 			<div className="form-section-docs">
// 				<h1>Reports</h1>
// 				<p>
// 					Here you can view and resolve reports made to your
// 					instance, originating from local and remote users.
// 				</p>
// 				<a
// 					href="https://docs.gotosocial.org/en/latest/admin/settings/#reports"
// 					target="_blank"
// 					className="docslink"
// 					rel="noreferrer"
// 				>
// 					Learn more about this (opens in a new tab)
// 				</a>
// 			</div>
// 			<div className="list">
// 				{reports.map((report) => (
// 					<ReportEntry key={report.id} report={report} />
// 				))}
// 			</div>
// 		</div>
// 	);
// }

interface ReportEntryProps {
	report: AdminReport;
	linkTo?: string;
	backLocation?: string;
}

function ReportEntry({ report }: ReportEntryProps) {
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
