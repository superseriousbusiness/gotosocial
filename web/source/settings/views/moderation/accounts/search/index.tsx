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

import { useLazySearchAccountsQuery } from "../../../../lib/query/admin";
import { useTextInput } from "../../../../lib/form";
import { PageableList } from "../../../../components/pageable-list";
import { Select, TextInput } from "../../../../components/form/inputs";
import MutationButton from "../../../../components/form/mutation-button";
import { useLocation, useSearch } from "wouter";
import { AdminAccount } from "../../../../lib/types/account";
import Username from "../../../../components/username";
import isValidDomain from "is-valid-domain";

export function AccountSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const [ searchAcct, searchRes ] = useLazySearchAccountsQuery();

	// Populate search form using values from
	// urlQueryParams, to allow paging.
	const form = {
		origin: useTextInput("origin", { defaultValue: urlQueryParams.get("origin") ?? ""}),
		status: useTextInput("status", { defaultValue: urlQueryParams.get("status") ?? ""}),
		permissions: useTextInput("permissions", { defaultValue: urlQueryParams.get("permissions") ?? ""}),
		username: useTextInput("username", { defaultValue: urlQueryParams.get("username") ?? ""}),
		display_name: useTextInput("display_name", { defaultValue: urlQueryParams.get("display_name") ?? ""}),
		by_domain: useTextInput("by_domain", {
			defaultValue: urlQueryParams.get("by_domain") ?? "",
			validator: (v: string) => {
				if (v.length === 0) {
					return "";
				}

				if (v[v.length-1] === ".") {
					return "invalid domain";
				}

				const valid = isValidDomain(v, {
					subdomain: true,
					wildcard: false,
					allowUnicode: true,
					topLevel: false,
				});

				if (valid) {
					return "";
				}

				return "invalid domain";
			}
		}),
		email: useTextInput("email", { defaultValue: urlQueryParams.get("email") ?? ""}),
		ip: useTextInput("ip", { defaultValue: urlQueryParams.get("ip") ?? ""}),
		limit: useTextInput("limit", { defaultValue: urlQueryParams.get("limit") ?? "50"})
	};

	// On mount, if urlQueryParams were provided,
	// trigger the search. For example, if page
	// was accessed at /search?origin=local&limit=20,
	// then run a search with origin=local and
	// limit=20 and immediately render the results.
	useEffect(() => {
		if (urlQueryParams.size > 0) {
			searchAcct(Object.fromEntries(urlQueryParams), true);
		}
	}, [urlQueryParams, searchAcct]);

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

	// Location to return to when user clicks "back" on the account detail view.
	const backLocation = location + (urlQueryParams.size > 0 ? `?${urlQueryParams}` : "");
	
	// Function to map an item to a list entry.
	function itemToEntry(account: AdminAccount): ReactNode {
		const acc = account.account;
		return (
			<Username
				key={acc.acct}
				account={account}
				linkTo={`/${account.id}`}
				backLocation={backLocation}
				classNames={["entry"]}
			/>
		);
	}

	return (
		<>
			<form
				onSubmit={submitQuery}
				// Prevent password managers trying
				// to fill in username/email fields.
				autoComplete="off"
			>
				<TextInput
					field={form.username}
					label={`Username (without "@" prefix) - case sensitive`}
					placeholder="someone"
					autoCapitalize="none"
					spellCheck="false"
				/>
				<TextInput
					field={form.by_domain}
					label={`Domain (without "https://" prefix)`}
					placeholder="example.org"
					autoCapitalize="none"
					spellCheck="false"
				/>
				<Select
					field={form.origin}
					label="Account origin"
					options={
						<>
							<option value="">Local or remote</option>
							<option value="local">Local only</option>
							<option value="remote">Remote only</option>
						</>
					}
				></Select>
				<TextInput
					field={form.email}
					label={"Email address (local accounts only)"}
					placeholder={"someone@example.org"}
					// Get email validation for free.
					type="email"
				/>
				<TextInput
					field={form.ip}
					label={"IP address (local accounts only)"}
					placeholder={"198.51.100.0"}
					autoCapitalize="none"
					spellCheck="false"
					className="monospace"
				/>
				<Select
					field={form.status}
					label="Account status"
					options={
						<>
							<option value="">Any</option>
							<option value="pending">Pending only</option>
							<option value="disabled">Disabled only</option>
							<option value="suspended">Suspended only</option>
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
				emptyMessage={<b>No accounts found that match your query.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}
