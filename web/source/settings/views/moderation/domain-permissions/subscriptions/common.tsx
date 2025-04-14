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
import { useLocation } from "wouter";
import { DomainPermSub } from "../../../../lib/types/domain-permission";
import { yesOrNo } from "../../../../lib/util";

export function DomainPermissionSubscriptionHelpText() {
	return (
		<>
			Domain permission subscriptions allow your instance to "subscribe" to a list of block or allows at a given url.
			<br/>
			Every 24 hours, each subscribed list is fetched by your instance, and any discovered
			permissions in each list are loaded into your instance as blocks/allows/drafts.
		</>
	);
}

export function DomainPermissionSubscriptionDocsLink() {
	return (
		<a
			href="https://docs.gotosocial.org/en/latest/admin/settings/#domain-permission-subscriptions"
			target="_blank"
			className="docslink"
			rel="noreferrer"
		>
			Learn more about domain permission subscriptions (opens in a new tab)
		</a>
	);
}

export interface SubscriptionEntryProps {
	permSub: DomainPermSub;
	linkTo: string;
	backLocation: string;
}

export function SubscriptionListEntry({ permSub, linkTo, backLocation }: SubscriptionEntryProps) {
	const [ _location, setLocation ] = useLocation();

	const permType = permSub.permission_type;
	if (!permType) {
		throw "permission_type was undefined";
	}

	const {
		priority,
		title,
		uri,
		as_draft: asDraft,
		adopt_orphans: adoptOrphans,
		content_type: contentType,
		fetched_at: fetchedAt,
		successfully_fetched_at: successfullyFetchedAt,
		count,
	} = permSub;

	const ariaLabel = useMemo(() => {
		let ariaLabel = "";
		
		// Prepend title.
		if (title.length !== 0) {
			ariaLabel += `${title}, create `;
		} else {
			ariaLabel += "Create ";
		}

		// Add perm type.
		ariaLabel += permType;
		
		// Alter wording
		// if using drafts.
		if (asDraft) {
			ariaLabel += " drafts from ";
		} else {
			ariaLabel += "s from ";
		}

		// Add url.
		ariaLabel += uri;

		return ariaLabel;
	}, [title, permType, asDraft, uri]);

	let fetchedAtStr = "never";
	if (fetchedAt) {
		fetchedAtStr = new Date(fetchedAt).toDateString();
	}

	let successfullyFetchedAtStr = "never";
	if (successfullyFetchedAt) {
		successfullyFetchedAtStr = new Date(successfullyFetchedAt).toDateString();
	}

	const onClick = (e) => {
		e.preventDefault();
		// When clicking on a subscription, direct
		// to the detail view for that subscription.
		setLocation(linkTo, {
			// Store the back location in history so
			// the detail view can use it to return to
			// this page (including query parameters).
			state: { backLocation: backLocation }
		});
	};

	return (
		<span
			className={`pseudolink domain-permission-subscription entry`}
			aria-label={ariaLabel}
			title={ariaLabel}
			onClick={onClick}
			onKeyDown={(e) => {
				if (e.key === "Enter") {
					e.preventDefault();
					onClick(e);
				}
			}}
			role="link"
			tabIndex={0}
		>
			<dl className="info-list">
				{ permSub.title !== "" &&
					<span className="domain-permission-subscription-title">
						{title}
					</span>
				}
				<div className="info-list-entry">
					<dt>Priority:</dt>
					<dd>{priority}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Permission type:</dt>
					<dd className={`permission-type ${permType}`}>
						<i
							aria-hidden={true}
							className={`fa fa-${permType === "allow" ? "check" : "close"}`}
						></i>
						{permType}
					</dd>
				</div>
				<div className="info-list-entry">
					<dt>URL:</dt>
					<dd className="text-cutoff">{uri}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Content type:</dt>
					<dd>{contentType}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Create as draft:</dt>
					<dd>{yesOrNo(asDraft)}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Adopt orphans:</dt>
					<dd>{yesOrNo(adoptOrphans)}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Last fetch attempt:</dt>
					<dd className="text-cutoff">{fetchedAtStr}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Last successful fetch:</dt>
					<dd className="text-cutoff">{successfullyFetchedAtStr}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Discovered {permType}s:</dt>
					<dd>{count}</dd>
				</div>
			</dl>
		</span>
	);
}
