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
import { useVerifyCredentialsQuery } from "../lib/query/oauth";
import { MediaAttachment, Status as StatusType } from "../lib/types/status";
import sanitize from "sanitize-html";

export function FakeStatus({ children }) {
	const { data: account = {
		avatar: "/assets/default_avatars/GoToSocial_icon1.webp",
		display_name: "",
		username: ""
	} } = useVerifyCredentialsQuery();

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
						{children}
					</div>
				</div>
			</section>
		</article>
	);
}

export function Status({ status }: { status: StatusType }) {
	return (
		<article
			className="status expanded"
			id={status.id}
			role="region"
		>
			<StatusHeader status={status} />
			<StatusBody status={status} />
			<StatusFooter status={status} />
			<a
				href={status.url}
				target="_blank"
				className="status-link"
				data-nosnippet
				title="Open this status (opens in new tab)"
			>
				Open this status (opens in new tab)
			</a>
		</article>
	);
}

function StatusHeader({ status }: { status: StatusType }) {
	const author = status.account;
	
	return (
		<header className="status-header">
			<address>
				<a
					href={author.url}
					rel="author"
					title="Open profile"
					target="_blank"
				>
					<img
						className="avatar"
						aria-hidden="true"
						src={author.avatar}
						alt={`Avatar for ${author.username}`}
						title={`Avatar for ${author.username}`}
					/>
					<div className="author-strap">
						<span className="displayname text-cutoff">{author.display_name}</span>
						<span className="sr-only">,</span>
						<span className="username text-cutoff">@{author.acct}</span>
					</div>
					<span className="sr-only">(open profile)</span>
				</a>
			</address>
		</header>
	);
}

function StatusBody({ status }: { status: StatusType }) {
	let content: string;
	if (status.content.length === 0) {
		content = "[no content set]";
	} else {
		// HTML has already been through
		// the instance sanitizer by now,
		// but do it again just in case.
		content = sanitize(status.content);
	}

	return (
		<div className="status-body">
			<details className="text-spoiler">
				<summary>
					<span
						className="spoiler-text"
						lang={status.language}
					>
						{ status.spoiler_text
							? status.spoiler_text + " "
							: "[no content warning set] "
						}
					</span>
					<span
						className="button"
						role="button"
						tabIndex={0}
						aria-label="Toggle content visibility"
					>
						Toggle content visibility
					</span>
				</summary>
				<div
					className="text"
					dangerouslySetInnerHTML={{__html: content}}
				/>
			</details>
			<StatusMedia status={status} />
		</div>
	);
}

function StatusMedia({ status }: { status: StatusType }) {
	if (status.media_attachments.length === 0) {
		return null;
	}

	const count = status.media_attachments.length;
	const aria_label = count === 1 ? "1 attachment" : `${count} attachments`;
	const oddOrEven = count % 2 === 0 ? "even" : "odd";
	const single = count === 1 ? " single" : "";

	return (
		<div
			className={`media ${oddOrEven}${single}`}
			role="group"
			aria-label={aria_label}
		>
			{ status.media_attachments.map((media) => {
				return (
					<StatusMediaEntry
						key={media.id}
						media={media}
					/>
				);
			})}
		</div>
	);
}

function StatusMediaEntry({ media }: { media: MediaAttachment }) {
	return (
		<div className="media-wrapper">
			<details className="image-spoiler media-spoiler">
				<summary>
					<div className="show sensitive button" aria-hidden="true">Show media</div>
					<span className="eye button" role="button" tabIndex={0} aria-label="Toggle show media">
						<i className="hide fa fa-fw fa-eye-slash" aria-hidden="true"></i>
						<i className="show fa fa-fw fa-eye" aria-hidden="true"></i>
					</span>
					<img
						src={media.preview_url}
						loading="lazy"
						alt={media.description}
						title={media.description}
						width={media.meta.small.width}
						height={media.meta.small.height}
					/>
				</summary>
				<a
					href={media.url}
					target="_blank"
				>
					<img
						src={media.url}
						loading="lazy"
						alt={media.description}
						width={media.meta.original.width}
						height={media.meta.original.height}
					/>
				</a>
			</details>
		</div>
	);
}

function StatusFooter({ status }: { status: StatusType }) {
	return (
		<aside className="status-info">
			<dl className="status-stats">
				<div className="stats-grouping">
					<div className="stats-item published-at text-cutoff">
						<dt className="sr-only">Published</dt>
						<dd>
							<time dateTime={status.created_at}>
								{ new Date(status.created_at).toLocaleString() }
							</time>
						</dd>
					</div>
				</div>
				<div className="stats-item language">
					<dt className="sr-only">Language</dt>
					<dd>{status.language}</dd>
				</div>
			</dl>
		</aside>
	);
}
