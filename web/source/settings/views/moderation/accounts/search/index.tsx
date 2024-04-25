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

import { useLazySearchAccountsQuery } from "../../../../lib/query/admin";
import { useTextInput } from "../../../../lib/form";
import { AccountList } from "../../../../components/account-list";
import { SearchAccountParams } from "../../../../lib/types/account";
import { Select, TextInput } from "../../../../components/form/inputs";
import MutationButton from "../../../../components/form/mutation-button";

export function AccountSearchForm() {
	const form = {
		origin: useTextInput("origin"),
		status: useTextInput("status"),
		permissions: useTextInput("permissions"),
		username: useTextInput("username"),
		display_name: useTextInput("display_name"),
		by_domain: useTextInput("by_domain"),
		email: useTextInput("email"),
		ip: useTextInput("ip"),
	};

	function submitSearch(e) {
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
		const params: SearchAccountParams = Object.fromEntries(entries);
		searchAcct(params);
	}

	const [ searchAcct, searchRes ] = useLazySearchAccountsQuery();

	return (
		<>
			<form
				onSubmit={submitSearch}
				// Prevent password managers trying
				// to fill in username/email fields.
				autoComplete="off"
			>
				<TextInput
					field={form.username}
					label={"(Optional) username (without leading '@' symbol)"}
					placeholder="someone"
				/>
				<TextInput
					field={form.by_domain}
					label={"(Optional) domain"}
					placeholder="example.org"
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
					label={"(Optional) email address (local accounts only)"}
					placeholder={"someone@example.org"}
					// Get email validation for free.
					{...{type: "email"}}
				/>
				<TextInput
					field={form.ip}
					label={"(Optional) IP address (local accounts only)"}
					placeholder={"198.51.100.0"}
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
			<AccountList
				isLoading={searchRes.isLoading}
				isSuccess={searchRes.isSuccess}
				data={searchRes.data}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage="No accounts found that match your query"
			/>
		</>
	);
}
