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
import { useLocation, useParams } from "wouter";
import FormWithData from "../../../lib/form/form-with-data";
import BackButton from "../../../components/back-button";
import { useValue, useTextInput } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import { TextArea } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import Username from "../../../components/username";
import { useGetReportQuery, useResolveReportMutation } from "../../../lib/query/admin/reports";
import { useBaseUrl } from "../../../lib/navigation/util";
import { AdminReport } from "../../../lib/types/report";
import { yesOrNo } from "../../../lib/util";
import { Status } from "../../../components/status";

export default function ReportDetail({ }) {
	const params: { reportId: string } = useParams();
	const baseUrl = useBaseUrl();
	const backLocation: String = history.state?.backLocation ?? `~${baseUrl}`;

	return (
		<div className="report-detail">
			<h1><BackButton to={backLocation}/> Report Details</h1>
			<FormWithData
				dataQuery={useGetReportQuery}
				queryArg={params.reportId}
				DataForm={ReportDetailForm}
				{...{ backLocation: backLocation }}
			/>
		</div>
	);
}

function ReportDetailForm({ data: report }: { data: AdminReport }) {
	const [ location ] = useLocation();
	const baseUrl = useBaseUrl();
	
	return (
		<>
			<ReportBasicInfo
				report={report}
				baseUrl={baseUrl}
				location={location}
			/>
			
			{ report.action_taken
				&& <ReportHistory 
					report={report}
					baseUrl={baseUrl}
					location={location}
				/>
			}

			{ report.statuses &&
				<ReportStatuses report={report} />
			}

			{ !report.action_taken &&
				<ReportActionForm report={report} />
			}
		</>
	);
}

interface ReportSectionProps {
	report: AdminReport;
	baseUrl: string;
	location: string;
}

function ReportBasicInfo({ report, baseUrl, location }: ReportSectionProps) {
	const from = report.account;
	const target = report.target_account;
	const comment = report.comment;
	const status = report.action_taken ? "Resolved" : "Unresolved";
	const created = new Date(report.created_at).toLocaleString();

	return (
		<dl className="info-list overview">
			<div className="info-list-entry">
				<dt>Reported account</dt>
				<dd>
					<Username
						account={target}
						linkTo={`~/settings/moderation/accounts/${target.id}`}
						backLocation={`~${baseUrl}${location}`}
					/>
				</dd>
			</div>
		
			<div className="info-list-entry">
				<dt>Reported by</dt>
				<dd>
					<Username
						account={from}
						linkTo={`~/settings/moderation/accounts/${from.id}`}
						backLocation={`~${baseUrl}${location}`}
					/>
				</dd>
			</div>

			<div className="info-list-entry">
				<dt>Status</dt>
				<dd>
					{ report.action_taken
						? <>{status}</>
						: <b>{status}</b>
					}
				</dd>
			</div>

			<div className="info-list-entry">
				<dt>Reason</dt>
				<dd>
					{ comment.length > 0
						? <>{comment}</>
						: <i>none provided</i>
					}
				</dd>
			</div>

			<div className="info-list-entry">
				<dt>Created</dt>
				<dd>
					<time dateTime={report.created_at}>{created}</time>
				</dd>
			</div>

			<div className="info-list-entry">
				<dt>Category</dt>
				<dd>{ report.category }</dd>
			</div>

			<div className="info-list-entry">
				<dt>Forwarded</dt>
				<dd>{ yesOrNo(report.forwarded) }</dd>
			</div>
		</dl>
	);
}

function ReportHistory({ report, baseUrl, location }: ReportSectionProps) {
	const handled_by = report.action_taken_by_account;
	if (!handled_by) {
		throw "report handled by action_taken_by_account undefined";
	}
	
	const handled = report.action_taken_at ? new Date(report.action_taken_at).toLocaleString() : "never";
	
	return (
		<>
			<h3>Moderation History</h3>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Handled by</dt>
					<dd>
						<Username
							account={handled_by}
							linkTo={`~/settings/moderation/accounts/${handled_by.id}`}
							backLocation={`~${baseUrl}${location}`}
						/>
					</dd>
				</div>

				<div className="info-list-entry">
					<dt>Handled</dt>
					<dd>
						<time dateTime={report.action_taken_at}>{handled}</time>
					</dd>
				</div>

				<div className="info-list-entry">
					<dt>Comment</dt>
					<dd>{ report.action_taken_comment ?? "none"}</dd>
				</div>
			</dl>
		</>
	);
}

function ReportActionForm({ report }) {
	const form = {
		id: useValue("id", report.id),
		comment: useTextInput("action_taken_comment")
	};

	const [submit, result] = useFormSubmit(form, useResolveReportMutation(), { changedOnly: false });

	return (
		<form onSubmit={submit}>
			<h3>Resolve this report</h3>
			<>
				An optional comment can be included while resolving this report.
				This is useful for providing an explanation about what action was
				taken (if any) before the report was marked as resolved.
				<br />
				<div className="info">
					<i className="fa fa-fw fa-exclamation-triangle" aria-hidden="true"></i>
					<b>
						If the report was created by a local account, then any
						comment made here will be emailed to that account's user!
					</b>
				</div>
			</>
			<TextArea
				field={form.comment}
				label="Comment"
				autoCapitalize="sentences"
			/>
			<MutationButton
				disabled={false}
				label="Resolve"
				result={result}
			/>
		</form>
	);
}

function ReportStatuses({ report }: { report: AdminReport }) {
	if (report.statuses.length === 0) {
		return null;
	}
	
	return (
		<div className="report-statuses">
			<h3>Reported Statuses</h3>
			<ul className="thread">
				{ report.statuses.map((status) => {
					return (
						<Status
							key={status.id}
							status={status}
					 	/>
					);
				})}
			</ul>
		</div>
	);
}
