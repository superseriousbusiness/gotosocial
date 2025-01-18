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
import { useLazySearchDomainPermissionSubscriptionsQuery } from "../../../../lib/query/admin/domain-permissions/subscriptions";
import { DomainPermSub } from "../../../../lib/types/domain-permission";
import { Select } from "../../../../components/form/inputs";
import { DomainPermissionSubscriptionDocsLink, DomainPermissionSubscriptionHelpText, SubscriptionListEntry } from "./common";

export default function DomainPermissionSubscriptionsSearch() {
	return (
		<div className="domain-permission-subscriptions-view">
			<div className="form-section-docs">
				<h1>Domain Permission Subscriptions</h1>
				<p>
					You can use the form below to search through domain permission
					subscriptions, sorted by creation time (newer to older).
					<br/>
					<DomainPermissionSubscriptionHelpText />
				</p>
				<DomainPermissionSubscriptionDocsLink />
			</div>
			<DomainPermissionSubscriptionsSearchForm />
		</div>
	);
}

function DomainPermissionSubscriptionsSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const hasParams = urlQueryParams.size != 0;
	const [ searchSubscriptions, searchRes ] = useLazySearchDomainPermissionSubscriptionsQuery();

	const form = {
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
			searchSubscriptions(Object.fromEntries(urlQueryParams));
		} else {
			setLocation(location + "?limit=20");
		}
	}, [
		urlQueryParams,
		hasParams,
		searchSubscriptions,
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
	function itemToEntry(permSub: DomainPermSub): ReactNode {
		return (
			<SubscriptionListEntry
				key={permSub.id}
				permSub={permSub}
				linkTo={`/subscriptions/${permSub.id}`}
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
				items={searchRes.data?.subs}
				itemToEntry={itemToEntry}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage={<b>No subscriptions found that match your query.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}
