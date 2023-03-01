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

const { useValue, useTextInput } = require("../../lib/form");
const useFormSubmit = require("../../lib/form/submit");

const { TextArea } = require("../../components/form/inputs");

const MutationButton = require("../../components/form/mutation-button");
const Username = require("./username");

module.exports = function ReportDetail({ baseUrl }) {
	let [_match, params] = useRoute(`${baseUrl}/:reportId`);
	if (params?.reportId == undefined) {
		return <Redirect to={baseUrl} />;
	} else {
		return (
			<div className="report-detail">
				<h1>
					<BackButton to={baseUrl} /> Report Details
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
			<div className="usernames">
				<Username user={from} /> reported <Username user={target} />
			</div>

			{report.action_taken &&
				<div className="info">
					<h3>Resolved by @{report.action_taken_by_account.account.acct}</h3>
					<span className="timestamp">at {new Date(report.action_taken_at).toLocaleString()}</span>
					<br />
					<b>Comment: </b><span>{report.action_taken_comment}</span>
				</div>
			}

			<div className="info-block">
				<h3>Report info:</h3>
				<div className="details">
					<b>Created: </b>
					<span>{new Date(report.created_at).toLocaleString()}</span>

					<b>Forwarded: </b> <span>{report.forwarded ? "Yes" : "No"}</span>
					<b>Category: </b> <span>{report.category}</span>

					<b>Reason: </b>
					{report.comment.length > 0
						? <p>{report.comment}</p>
						: <i className="no-comment">none provided</i>
					}

				</div>
			</div>

			{!report.action_taken && <ReportActionForm report={report} />}

			{
				report.statuses.length > 0 &&
				<div className="info-block">
					<h3>Reported toots ({report.statuses.length}):</h3>
					<div className="reported-toots">
						{report.statuses.map((status) => (
							<ReportedToot key={status.id} toot={status} />
						))}
					</div>
				</div>
			}
		</div>
	);
}

function ReportActionForm({ report }) {
	const form = {
		id: useValue("id", report.id),
		comment: useTextInput("action_taken_comment")
	};

	const [submit, result] = useFormSubmit(form, query.useResolveReportMutation(), { changedOnly: false });

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
			<MutationButton label="Resolve" result={result} />
		</form>
	);
}

function ReportedToot({ toot }) {
	const account = toot.account;

	return (
		<div className="toot expanded">
			<div className="contentgrid">
				<span className="avatar">
					<img src={account.avatar} alt="" />
				</span>
				<span className="displayname">{account.display_name.trim().length > 0 ? account.display_name : account.username}</span>
				<span className="username">@{account.username}</span>
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
			</div>
			<div className="toot-info">
				<a
					href={toot.url}
					target="_blank"
					rel="noreferrer"
				>{new Date(toot.created_at).toLocaleString()}</a>
			</div>
		</div>
	);
}

function TootCW({ note, content }) {
	const [visible, setVisible] = React.useState(false);

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
								<label htmlFor={`sensitiveMedia-${m.id}`} className="button" role="button" tabIndex="0">
									<i className="fa fa-eye-slash" title="Hide sensitive media"></i>
								</label>
							</div>
							<div className="closed" title={m.description}>
								<label htmlFor={`sensitiveMedia-${m.id}`} className="button" role="button" tabIndex="0">
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
							size={m.meta?.original}
							type={m.type}
						/>
					</a>
				</div>
			))}
		</div>
	);
}