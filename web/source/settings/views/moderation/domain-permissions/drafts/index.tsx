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

import { useTextInput } from "../../../../lib/form";
import { PageableList } from "../../../../components/pageable-list";
import MutationButton from "../../../../components/form/mutation-button";
import { useLocation, useSearch } from "wouter";
import { useAcceptDomainPermissionDraftMutation, useLazySearchDomainPermissionDraftsQuery, useRemoveDomainPermissionDraftMutation } from "../../../../lib/query/admin/domain-permissions/drafts";
import { DomainPerm } from "../../../../lib/types/domain-permission";
import { Error as ErrorC } from "../../../../components/error";
import { Select, TextInput } from "../../../../components/form/inputs";
import { formDomainValidator } from "../../../../lib/util/formvalidators";
import { useCapitalize } from "../../../../lib/util";
import { DomainPermissionDraftDocsLink, DomainPermissionDraftHelpText } from "./common";

export default function DomainPermissionDraftsSearch() {
	return (
		<div className="domain-permission-drafts-view">
			<div className="form-section-docs">
				<h1>Domain Permission Drafts</h1>
				<p>
					You can use the form below to search through domain permission drafts.
					<br/>
					<DomainPermissionDraftHelpText />
				</p>
				<DomainPermissionDraftDocsLink />
			</div>
			<DomainPermissionDraftsSearchForm />
		</div>
	);
}

function DomainPermissionDraftsSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const hasParams = urlQueryParams.size != 0;
	const [ searchDrafts, searchRes ] = useLazySearchDomainPermissionDraftsQuery();

	const form = {
		subscription_id: useTextInput("subscription_id", { defaultValue: urlQueryParams.get("subscription_id") ?? "" }),
		domain: useTextInput("domain", {
			defaultValue: urlQueryParams.get("domain") ?? "",
			validator: formDomainValidator,
		}),
		permission_type: useTextInput("permission_type", { defaultValue: urlQueryParams.get("permission_type") ?? "" }),
		limit: useTextInput("limit", { defaultValue: urlQueryParams.get("limit") ?? "20" })
	};

	// On mount, if urlQueryParams were provided,
	// trigger the search. For example, if page
	// was accessed at /search?origin=local&limit=20,
	// then run a search with origin=local and
	// limit=20 and immediately render the results.
	//
	// If no urlQueryParams set, trigger default
	// search (first page, no filtering).
	useEffect(() => {
		if (hasParams) {
			searchDrafts(Object.fromEntries(urlQueryParams));
		} else {
			setLocation(location + "?limit=20");
		}
	}, [
		urlQueryParams,
		hasParams,
		searchDrafts,
		location,
		setLocation,
	]);

	// Rather than triggering the search directly,
	// the "submit" button changes the location
	// based on form field params, and lets the
	// useEffect hook above actually do the search.
	function submitQuery(e) {
		e.preventDefault();
		
		// Parse query parameters.
		const entries = Object.entries(form).map(([k, v]) => {
			// Take only defined form fields.
			if (v.value === undefined || v.value.length === 0 || v.value === "any") {
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
	const backLocation = location + (hasParams ? `?${urlQueryParams}` : "");
	
	// Function to map an item to a list entry.
	function itemToEntry(draft: DomainPerm): ReactNode {
		return (
			<DraftListEntry
				key={draft.id}	
				permDraft={draft}
				linkTo={`/drafts/${draft.id}`}
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
					field={form.permission_type}
					label="Permission type"
					options={
						<>
							<option value="">Any</option>
							<option value="block">Block</option>
							<option value="allow">Allow</option>
						</>
					}
				></Select>
				<TextInput
					field={form.domain}
					label={`Domain (without "https://" prefix)`}
					placeholder="example.org"
					autoCapitalize="none"
					spellCheck="false"
				/>
				<Select
					field={form.limit}
					label="Items per page"
					options={
						<>
							<option value="20">20</option>
							<option value="50">50</option>
							<option value="100">100</option>
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
				items={searchRes.data?.drafts}
				itemToEntry={itemToEntry}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage={<b>No drafts found that match your query.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}

interface DraftEntryProps {
	permDraft: DomainPerm;
	linkTo: string;
	backLocation: string;
}

function DraftListEntry({ permDraft, linkTo, backLocation }: DraftEntryProps) {
	const [ _location, setLocation ] = useLocation();
	const [ accept, acceptResult ] = useAcceptDomainPermissionDraftMutation();
	const [ remove, removeResult ] = useRemoveDomainPermissionDraftMutation();

	const domain = permDraft.domain;
	const permType = permDraft.permission_type;
	const permTypeUpper = useCapitalize(permType);
	if (!permType) {
		return <ErrorC error={new Error("permission_type was undefined")} />;
	}

	const publicComment = permDraft.public_comment ?? "[none]";
	const privateComment = permDraft.private_comment ?? "[none]";
	const subscriptionID = permDraft.subscription_id ?? "[none]";
	const id = permDraft.id;
	if (!id) {
		return <ErrorC error={new Error("id was undefined")} />;
	}

	const title = `${permTypeUpper} ${domain}`;

	return (
		<span
			className={`pseudolink domain-permission-draft entry ${permType}`}
			aria-label={title}
			title={title}
			onClick={() => {
				// When clicking on a draft, direct
				// to the detail view for that draft.
				setLocation(linkTo, {
					// Store the back location in history so
					// the detail view can use it to return to
					// this page (including query parameters).
					state: { backLocation: backLocation }
				});
			}}
			role="link"
			tabIndex={0}
		>
			<h3>{title}</h3>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Domain:</dt>
					<dd className="text-cutoff">{domain}</dd>
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
					<dt>Private comment:</dt>
					<dd className="text-cutoff">{privateComment}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Public comment:</dt>
					<dd>{publicComment}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Subscription:</dt>
					<dd className="text-cutoff">{subscriptionID}</dd>
				</div>
			</dl>
			<div className="action-buttons">
				<MutationButton
					label={`Accept ${permType}`}
					title={`Accept ${permType}`}
					type="button"
					className="button"
					onClick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						accept({ id, permType });
					}}
					disabled={false}
					showError={true}
					result={acceptResult}
				/>
				<MutationButton
					label={`Remove draft`}
					title={`Remove draft`}
					type="button"
					className="button danger"
					onClick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						remove({ id });
					}}
					disabled={false}
					showError={true}
					result={removeResult}
				/>
			</div>
		</span>
	);
}
