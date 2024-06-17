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

import React, { useState } from "react";
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
	
	const from = report.account;
	const target = report.target_account;
	const comment = report.comment;
	const status = report.action_taken ? "Resolved" : "Unresolved";
	const created = new Date(report.created_at).toLocaleString();

	return (
		<dl className="info-list">
			<div className="info-list-entry">
				<dt>Target</dt>
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
		</dl>
	);

	// return (
	// 	<>
	// 		<div className="usernames">
	// 			<Username
	// 				account={from}
	// 				linkTo={`~/settings/moderation/accounts/${from.id}`}
	// 				backLocation={`~/settings/moderation/reports/${report.id}`}
	// 			/>
	// 			<> reported </>
	// 			<Username
	// 				account={target}
	// 				linkTo={`~/settings/moderation/accounts/${target.id}`}
	// 				backLocation={`~/settings/moderation/reports/${report.id}`}
	// 			/>
	// 		</div>

	// 		{report.action_taken &&
	// 			<div className="info">
	// 				<h3>Resolved by @{report.action_taken_by_account.account.acct}</h3>
	// 				<span className="timestamp">at {new Date(report.action_taken_at).toLocaleString()}</span>
	// 				<br />
	// 				<b>Comment: </b><span>{report.action_taken_comment}</span>
	// 			</div>
	// 		}

	// 		<div className="info-block">
	// 			<h3>Report info:</h3>
	// 			<div className="details">
	// 				<b>Created: </b>
	// 				<span>{new Date(report.created_at).toLocaleString()}</span>

	// 				<b>Forwarded: </b> <span>{report.forwarded ? "Yes" : "No"}</span>
	// 				<b>Category: </b> <span>{report.category}</span>

	// 				<b>Reason: </b>
	// 				{report.comment.length > 0
	// 					? <p>{report.comment}</p>
	// 					: <i className="no-comment">none provided</i>
	// 				}

	// 			</div>
	// 		</div>

	// 		{!report.action_taken && <ReportActionForm report={report} />}

	// 		{
	// 			report.statuses.length > 0 &&
	// 			<div className="info-block">
	// 				<h3>Reported toots ({report.statuses.length}):</h3>
	// 				<div className="reported-toots">
	// 					{report.statuses.map((status) => (
	// 						<ReportedToot key={status.id} toot={status} />
	// 					))}
	// 				</div>
	// 			</div>
	// 		}
	// 	</>
	// );
}

function ReportActionForm({ report }) {
	const form = {
		id: useValue("id", report.id),
		comment: useTextInput("action_taken_comment")
	};

	const [submit, result] = useFormSubmit(form, useResolveReportMutation(), { changedOnly: false });

	return (
		<form onSubmit={submit} className="info-block">
			<h3>Resolving this report</h3>
			<p>
				An optional comment can be included while resolving this report.
				Useful for providing an explanation about what action was taken (if any) before the report was marked as resolved.<br />
				<b>This will be visible to the user that created the report!</b>
			</p>
			<TextArea
				field={form.comment}
				label="Comment"
			/>
			<MutationButton
				disabled={false}
				label="Resolve"
				result={result}
			/>
		</form>
	);
}

function ReportedToot({ toot }) {
	const account = toot.account;

	return (
		<article className="status expanded">
			<header className="status-header">
				<address>
					<a style={{margin: 0}}>
						<img className="avatar" src={account.avatar} alt="" />
						<dl className="author-strap">
							<dt className="sr-only">Display name</dt>
							<dd className="displayname text-cutoff">
								{account.display_name.trim().length > 0 ? account.display_name : account.username}
							</dd>
							<dt className="sr-only">Username</dt>
							<dd className="username text-cutoff">@{account.username}</dd>
						</dl>
					</a>
				</address>
			</header>
			<section className="status-body">
				<div className="text">
					<div className="content">
						{toot.spoiler_text?.length > 0
							? <TootCW content={toot.content} note={toot.spoiler_text} />
							: toot.content
						}
					</div>
				</div>
				{toot.media_attachments?.length > 0 &&
					<TootMedia media={toot.media_attachments} sensitive={toot.sensitive} />
				}
			</section>
			<aside className="status-info">
				<dl className="status-stats">
					<div className="stats-grouping">
						<div className="stats-item published-at text-cutoff">
							<dt className="sr-only">Published</dt>
							<dd>
								<time dateTime={toot.created_at}>{new Date(toot.created_at).toLocaleString()}</time>
							</dd>
						</div>
					</div>
				</dl>
			</aside>
		</article>
	);
}

function TootCW({ note, content }) {
	const [visible, setVisible] = useState(false);

	function toggleVisible() {
		setVisible(!visible);
	}

	return (
		<>
			<div className="spoiler">
				<span>{note}</span>
				<label className="button spoiler-label" onClick={toggleVisible}>Show {visible ? "less" : "more"}</label>
			</div>
			{visible && content}
		</>
	);
}

function TootMedia({ media, sensitive }) {
	let classes = (media.length % 2 == 0) ? "even" : "odd";
	if (media.length == 1) {
		classes += " single";
	}

	return (
		<div className={`media photoswipe-gallery ${classes}`}>
			{media.map((m) => (
				<div key={m.id} className="media-wrapper">
					{sensitive && <>
						<input id={`sensitiveMedia-${m.id}`} type="checkbox" className="sensitive-checkbox hidden" />
						<div className="sensitive">
							<div className="open">
								<label htmlFor={`sensitiveMedia-${m.id}`} className="button" role="button" tabIndex={0}>
									<i className="fa fa-eye-slash" title="Hide sensitive media"></i>
								</label>
							</div>
							<div className="closed" title={m.description}>
								<label htmlFor={`sensitiveMedia-${m.id}`} className="button" role="button" tabIndex={0}>
									Show sensitive media
								</label>
							</div>
						</div>
					</>}
					<a
						href={m.url}
						title={m.description}
						target="_blank"
						rel="noreferrer"
						data-cropped="true"
						data-pswp-width={`${m.meta?.original.width}px`}
						data-pswp-height={`${m.meta?.original.height}px`}
					>
						<img
							alt={m.description}
							src={m.url}
							// thumb={m.preview_url}
							sizes={m.meta?.original}
						/>
					</a>
				</div>
			))}
		</div>
	);
}
