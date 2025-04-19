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

import React, { useMemo, useState } from "react";
import { useVerifyCredentialsQuery } from "../lib/query/login";
import { MediaAttachment, Status as StatusType } from "../lib/types/status";
import sanitize from "sanitize-html";
import BlurhashCanvas from "./blurhash";

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

	const [ detailsOpen, setDetailsOpen ] = useState(false);

	return (
		<div className="status-body">
			<details
				className="text-spoiler"
				open={detailsOpen}
			>
				<summary tabIndex={-1}>
					<div
						className="spoiler-content"
						lang={status.language}
					>
						{ status.spoiler_text
							? status.spoiler_text + " "
							: "[no content warning set] "
						}
					</div>
					<span
						className="button"
						role="button"
						tabIndex={0}
						aria-label={detailsOpen ? "Hide content" : "Show content"}
						onClick={(e) => {
							e.preventDefault();
							setDetailsOpen(!detailsOpen);
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								setDetailsOpen(!detailsOpen);
							}
						}}
					>
						{detailsOpen ? "Hide content" : "Show content"}
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
	const [ detailsOpen, setDetailsOpen ] = useState(false);
	return (
		<div className="media-wrapper">
			<details
				className="image-spoiler media-spoiler"
				open={detailsOpen}
			>
				<summary tabIndex={-1}>
					<div
						className="show sensitive button"
						role="button"
						tabIndex={-1}
						aria-hidden="true"
						onClick={(e) => {
							e.preventDefault();
							setDetailsOpen(!detailsOpen);
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								setDetailsOpen(!detailsOpen);
							}
						}}
					>
						Show media
					</div>
					<span
						className="eye button"
						role="button"
						tabIndex={0}
						aria-label={detailsOpen ? "Hide media" : "Show media"}
						onClick={(e) => {
							e.preventDefault();
							setDetailsOpen(!detailsOpen);
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								setDetailsOpen(!detailsOpen);
							}
						}}
					>
						<i className="hide fa fa-fw fa-eye-slash" aria-hidden="true"></i>
						<i className="show fa fa-fw fa-eye" aria-hidden="true"></i>
					</span>
					<div className="blurhash-container">
						<BlurhashCanvas media={media} />
					</div>
				</summary>
				<a
					className="photoswipe-slide"
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

function useVisibilityIcon(visibility: string): string {
	return useMemo(() => {
		switch (true) {
			case visibility === "direct":
				return "fa-envelope";
			case visibility === "followers_only":
				return "fa-lock";
			case visibility === "unlisted":
				return "fa-unlock";
			case visibility === "public":
				return "fa-globe";
			default:
				return "fa-question";
		}
	}, [visibility]);
}

function StatusFooter({ status }: { status: StatusType }) {
	const visibilityIcon = useVisibilityIcon(status.visibility);	
	return (
		<aside className="status-info">
			<div className="status-stats">
				<dl className="stats-grouping text-cutoff">
					<div className="stats-item published-at text-cutoff">
						<dt className="sr-only">Published</dt>
						<dd className="text-cutoff">
							<a
								href={status.url}
								className="u-url text-cutoff"
							>
								<time
									className="dt-published text-cutoff"
									dateTime={status.created_at}
								>
									{new Date(status.created_at).toLocaleString(undefined, {
										year: 'numeric',
										month: 'short',
										day: '2-digit',
										hour: '2-digit',
										minute: '2-digit',
										hour12: false
									})}
								</time>{ status.edited_at && "*" }
							</a>
						</dd>
					</div>
					<div className="stats-grouping">
						<div className="stats-item visibility-level" title={status.visibility}>
							<dt className="sr-only">Visibility</dt>
							<dd>
								<i className={`fa ${visibilityIcon}`} aria-hidden="true"></i>
								<span className="sr-only">{status.visibility}</span>
							</dd>
						</div>
					</div>
				</dl>
				<details className="stats-more-info">
					<summary title="More info">
						<i className="fa fa-fw fa-info" aria-hidden="true"></i>
						<span className="sr-only">More info</span>
						<i className="fa fa-fw fa-chevron-right show" aria-hidden="true"></i>
						<i className="fa fa-fw fa-chevron-down hide" aria-hidden="true"></i>
					</summary>
					<dl className="stats-more-info-content">
						<div className="stats-grouping">
							{ status.language &&
								<div className="stats-item" title="Language">
									<dt>
										<span className="sr-only">Language</span>
										<i className="fa fa-language" aria-hidden="true"></i>
									</dt>
									<dd>{status.language}</dd>
								</div>
							}
							<div className="stats-item" title="Replies">
								<dt>
									<span className="sr-only">Replies</span>
									<i className="fa fa-reply-all" aria-hidden="true"></i>
								</dt>
								<dd>{status.replies_count}</dd>
							</div>
							<div className="stats-item" title="Faves">
								<dt>
									<span className="sr-only">Favourites</span>
									<i className="fa fa-star" aria-hidden="true"></i>
								</dt>
								<dd>{status.favourites_count}</dd>
							</div>
							<div className="stats-item" title="Boosts">
								<dt>
									<span className="sr-only">Reblogs</span>
									<i className="fa fa-retweet" aria-hidden="true"></i>
								</dt>
								<dd>{status.reblogs_count}</dd>
							</div>
						</div>
					</dl>
				</details>
			</div>
		</aside>
	);
}
