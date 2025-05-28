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

import React, { useMemo } from "react";
import { useInstanceV2Query } from "../../../lib/query/gts-api";
import Loading from "../../../components/loading";
import { InstanceV2 } from "../../../lib/types/instance";
import { useHumanReadableBytes, useHumanReadableDuration, yesOrNo } from "../../../lib/util";
import { HighlightedCode } from "../../../components/highlightedcode";
import { useInstanceDomainAllowsQuery, useInstanceDomainBlocksQuery } from "../../../lib/query/user/domainperms";

export default function InstanceInfo() {
	// Load instance v2 data.
	const {
		data,
		isFetching,
		isLoading,
	} = useInstanceV2Query();
	
	if (isFetching || isLoading) {
		return <Loading />;
	}

	if (data === undefined) {
		throw "could not fetch instance v2";
	}

	return (
		<div className="instance-info-view">
			<div className="form-section-docs">
				<h1>Instance Info</h1>
				<p>
					On this page you can see information about this instance, and view domain blocks
					and domain allows that have been created by the admin(s) of the instance.
				</p>
			</div>
			<Instance instance={data} />
			<Allowlist />
			<Blocklist />
		</div>
	);
}

function Instance({ instance }: { instance: InstanceV2 }) {
	const emojiSizeLimit = useHumanReadableBytes(instance.configuration.emojis.emoji_size_limit);
	const accountsCustomCSS = yesOrNo(instance.configuration.accounts.allow_custom_css);
	const imageSizeLimit = useHumanReadableBytes(instance.configuration.media_attachments.image_size_limit);
	const videoSizeLimit = useHumanReadableBytes(instance.configuration.media_attachments.video_size_limit);
	const pollMinExpiry = useHumanReadableDuration(instance.configuration.polls.min_expiration);
	const pollMaxExpiry = useHumanReadableDuration(instance.configuration.polls.max_expiration);

	return (
		<>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Software version:</dt>
					<dd>
						<a
							href={instance.source_url}
							target="_blank"
							rel="noreferrer"
						>
							{instance.version}
						</a>
					</dd>
				</div>

				<div className="info-list-entry">
					<dt>Streaming URL:</dt>
					<dd className="monospace">{instance.configuration.urls.streaming}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Emoji size limit:</dt>
					<dd>{emojiSizeLimit}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Accounts custom CSS:</dt>
					<dd>{accountsCustomCSS}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Accounts max featured tags:</dt>
					<dd>{instance.configuration.accounts.max_featured_tags}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Accounts max profile fields:</dt>
					<dd>{instance.configuration.accounts.max_profile_fields}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Posts max characters:</dt>
					<dd>{instance.configuration.statuses.max_characters}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Posts max attachments:</dt>
					<dd>{instance.configuration.statuses.max_media_attachments}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Posts supported types:</dt>
					<dd className="monospace">
						{ useJoinWithNewlines(instance.configuration.statuses.supported_mime_types) }
					</dd>
				</div>

				<div className="info-list-entry">
					<dt>Polls max options:</dt>
					<dd>{instance.configuration.polls.max_options}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Polls max characters per option:</dt>
					<dd>{instance.configuration.polls.max_characters_per_option}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Polls min expiration:</dt>
					<dd>{pollMinExpiry}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Polls max expiration:</dt>
					<dd>{pollMaxExpiry}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Media max description characters:</dt>
					<dd>{instance.configuration.media_attachments.description_limit}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Media max image size:</dt>
					<dd>{imageSizeLimit}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Media max video size:</dt>
					<dd>{videoSizeLimit}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Media supported types:</dt>
					<dd className="monospace">
						{ useJoinWithNewlines(instance.configuration.media_attachments.supported_mime_types) }
					</dd>
				</div>
			</dl>

			{ instance.custom_css &&
				<>
					<div className="form-section-docs">
						<h3>Custom CSS</h3>
						<p>The following custom CSS has been set by the admin(s) of this instance, and will be loaded on each web page:</p>
					</div>
					<HighlightedCode code={instance.custom_css} lang="css" />
				</>
			}
		</>
	);
}

function Allowlist() {
	// Load allows.
	const {
		data,
		isFetching,
		isLoading,
	} = useInstanceDomainAllowsQuery();
	
	if (isFetching || isLoading) {
		return <Loading />;
	}

	if (data === undefined) {
		throw "could not fetch domain allows";
	}
	
	return (
		<>
			<div className="form-section-docs">
				<h3>Domain Allows</h3>
				<p>
					The following list of domains has been explicitly allowed by the administrator(s) of this instance.
					<br/>This extends to subdomains, so an allowlist entry for domain 'example.com' includes domain 'social.example.com' etc as well. 
				</p>
			</div>
			{ data.length !== 0
				? <div className="list domain-perm-list">
					<div className="header entry">
						<div className="domain">Domain</div>
						<div className="public_comment">Public comment</div>
					</div>
					{ data.map(e => {
						return (
							<div className="entry" id={e.domain} key={e.domain}>
								<div className="domain text-cutoff">{e.domain}</div>
								<div className="public_comment">{e.comment}</div>
							</div>
						);
					}) }
				</div>
				: <b>No explicit allows.</b>
			}
		</>
	);
}

function Blocklist() {
	// Load blocks.
	const {
		data,
		isFetching,
		isLoading,
	} = useInstanceDomainBlocksQuery();
	
	if (isFetching || isLoading) {
		return <Loading />;
	}

	if (data === undefined) {
		throw "could not fetch domain blocks";
	}
	
	return (
		<>
			<div className="form-section-docs">
				<h3>Domain Blocks</h3>
				<p>
					The following list of domains has been blocked by the administrator(s) of this instance.
					<br/>All past, present, and future accounts at blocked domains are forbidden from interacting with this instance or accounts on this instance.
					<br/>No data will be sent to the server at the remote domain, and no data will be received from it.
					<br/>This extends to subdomains, so a blocklist entry for domain 'example.com' includes domain 'social.example.com' etc as well.
				</p>
			</div>
			{ data.length !== 0
				? <div className="list domain-perm-list">
					<div className="header entry">
						<div className="domain">Domain</div>
						<div className="public_comment">Public comment</div>
					</div>
					{ data.map(e => {
						return (
							<div className="entry" id={e.domain} key={e.domain}>
								<div className="domain text-cutoff">{e.domain}</div>
								<div className="public_comment">{e.comment}</div>
							</div>
						);
					}) }
				</div>
				: <b>No domain blocks.</b>
			}
		</>
	);
}

function useJoinWithNewlines(a: string[]) {
	return useMemo(() => {
		const l = a.length;
		return a.map((v, i) => {
			const e = <span key={v}>{v}</span>;
			if (i+1 !== l) {
				return [e, <br key={v + "br"} />];
			}
			return [e];
		}).flat();
	}, [a]);
}
