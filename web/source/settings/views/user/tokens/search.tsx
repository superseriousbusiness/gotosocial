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

import { useTextInput } from "../../../lib/form";
import { PageableList } from "../../../components/pageable-list";
import MutationButton from "../../../components/form/mutation-button";
import { useLocation, useSearch } from "wouter";
import { Select } from "../../../components/form/inputs";
import { useInvalidateTokenMutation, useLazySearchTokenInfoQuery } from "../../../lib/query/user/tokens";
import { TokenInfo } from "../../../lib/types/tokeninfo";

export default function TokensSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const [ searchTokenInfo, searchRes ] = useLazySearchTokenInfoQuery();

	// Populate search form using values from
	// urlQueryParams, to allow paging.
	const form = {
		limit: useTextInput("limit", { defaultValue: urlQueryParams.get("limit") ?? "25" })
	};

	// On mount, trigger search.
	useEffect(() => {
		searchTokenInfo(Object.fromEntries(urlQueryParams), true);
	}, [urlQueryParams, searchTokenInfo]);

	// Rather than triggering the search directly,
	// the "submit" button changes the location
	// based on form field params, and lets the
	// useEffect hook above actually do the search.
	function submitQuery(e) {
		e.preventDefault();

		// Parse query parameters.
		const entries = Object.entries(form).map(([k, v]) => {
			// Take only defined form fields.
			if (v.value === undefined) {
				return null;
			} else if (typeof v.value === "string" && v.value.length === 0) {
				return null;
			}

			return [[k, v.value.toString()]];
		}).flatMap(kv => {
			// Remove any nulls.
			return kv !== null ? kv : [];
		});

		const searchParams = new URLSearchParams(entries);
		setLocation(location + "?" + searchParams.toString());
	}
	
	// Function to map an item to a list entry.
	function itemToEntry(tokenInfo: TokenInfo): ReactNode {
		return (
			<TokenInfoListEntry
				key={tokenInfo.id}
				tokenInfo={tokenInfo}
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
					field={form.limit}
					label="Items per page"
					options={
						<>
							<option value="25">25</option>
							<option value="50">50</option>
							<option value="75">75</option>
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
				items={searchRes.data?.tokens}
				itemToEntry={itemToEntry}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage={<b>No tokens found.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}

interface TokenInfoListEntryProps {
	tokenInfo: TokenInfo;
}

function TokenInfoListEntry({ tokenInfo }: TokenInfoListEntryProps) {
	const appWebsite = useMemo(() => {
		if (!tokenInfo.application.website) {
			return "";
		}

		try {
			// Try to parse nicely and return link.
			const websiteURL = new URL(tokenInfo.application.website);
			const websiteURLStr = websiteURL.toString();
			return (
				<a
					href={websiteURLStr}
					target="_blank"
					rel="nofollow noreferrer noopener"
				>{websiteURLStr}</a>
			);
		} catch {
			// Fall back to returning string.
			return tokenInfo.application.website;
		}
	}, [tokenInfo.application.website]);
	
	const created = useMemo(() => {
		const createdAt = new Date(tokenInfo.created_at);
		return <time dateTime={tokenInfo.created_at}>{createdAt.toDateString()}</time>;
	}, [tokenInfo.created_at]);

	const lastUsed = useMemo(() => {
		if (!tokenInfo.last_used) {
			return "unknown/never";
		}

		const lastUsed = new Date(tokenInfo.last_used);
		return <time dateTime={tokenInfo.last_used}>{lastUsed.toDateString()}</time>;
	}, [tokenInfo.last_used]);

	const [ invalidate, invalidateResult ] = useInvalidateTokenMutation();

	return (
		<span
			className={`token-info entry`}
			aria-label={`${tokenInfo.application.name}, scope: ${tokenInfo.scope}`}
			title={`${tokenInfo.application.name}, scope: ${tokenInfo.scope}`}
		>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>App name:</dt>
					<dd className="text-cutoff">{tokenInfo.application.name}</dd>
				</div>
				{ appWebsite && 
					<div className="info-list-entry">
						<dt>App website:</dt>
						<dd className="text-cutoff">{appWebsite}</dd>
					</div>
				}
				<div className="info-list-entry">
					<dt>Scope:</dt>
					<dd className="text-cutoff monospace">{tokenInfo.scope}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Created:</dt>
					<dd className="text-cutoff">{created}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Last used:</dt>
					<dd className="text-cutoff">{lastUsed}</dd>
				</div>
			</dl>
			<div className="action-buttons">
				<MutationButton
					label={`Invalidate token`}
					title={`Invalidate token`}
					type="button"
					className="button danger"
					onClick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						invalidate(tokenInfo.id);
					}}
					disabled={false}
					showError={true}
					result={invalidateResult}
				/>
			</div>
		</span>
	);
}
