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

import { useMemo } from "react";
import { Link, useLocation, useParams } from "wouter";
import { matchSorter } from "match-sorter";
import { useTextInput } from "../../../lib/form";
import { TextInput } from "../../../components/form/inputs";
import Loading from "../../../components/loading";
import { useDomainAllowsQuery, useDomainBlocksQuery } from "../../../lib/query/admin/domain-permissions/get";
import type { MappedDomainPerms } from "../../../lib/types/domain-permission";
import { NoArg } from "../../../lib/types/query";
import { PermType } from "../../../lib/types/perm";
import { useBaseUrl } from "../../../lib/navigation/util";

export default function DomainPermissionsOverview() {	
	const baseUrl = useBaseUrl();
	
	// Parse perm type from routing params.
	let params = useParams();
	if (params.permType !== "blocks" && params.permType !== "allows") {
		throw "unrecognized perm type " + params.permType;
	}
	const permType = params.permType.slice(0, -1) as PermType;

	// Uppercase first letter of given permType.
	const permTypeUpper = useMemo(() => {
		return permType.charAt(0).toUpperCase() + permType.slice(1); 
	}, [permType]);

	// Fetch / wait for desired perms to load.
	const { data: blocks, isLoading: isLoadingBlocks } = useDomainBlocksQuery(NoArg, { skip: permType !== "block" });
	const { data: allows, isLoading: isLoadingAllows } = useDomainAllowsQuery(NoArg, { skip: permType !== "allow" });
	
	let data: MappedDomainPerms | undefined;
	let isLoading: boolean;

	if (permType == "block") {
		data = blocks;
		isLoading = isLoadingBlocks;
	} else {
		data = allows;
		isLoading = isLoadingAllows;
	}

	if (isLoading || data === undefined) {
		return <Loading />;
	}
	
	return (
		<div className={`domain-${permType}`}>
			<div className="form-section-docs">
				<h1>Domain {permTypeUpper}s</h1>
				{ permType == "block" ? <BlockHelperText/> : <AllowHelperText/> }
			</div>
			<DomainPermsList
				data={data}
				permType={permType}
				permTypeUpper={permTypeUpper}
			/>
			<Link to={`~${baseUrl}/import-export`}>
				Or use the bulk import/export interface
			</Link>
		</div>
	);
}

interface DomainPermsListProps {
	data: MappedDomainPerms;
	permType: PermType;
	permTypeUpper: string;
}

function DomainPermsList({ data, permType, permTypeUpper }: DomainPermsListProps) {
	// Format perms into a list.
	const perms = useMemo(() => {
		return Object.values(data);
	}, [data]);

	const [_location, setLocation] = useLocation();
	const filterField = useTextInput("filter");
	
	function filterFormSubmit(e) {
		e.preventDefault();
		setLocation(`/${permType}s/${filter}`);
	}
	
	const filter = filterField.value ?? "";
	const filteredPerms = useMemo(() => {
		return matchSorter(perms, filter, { keys: ["domain"] });
	}, [perms, filter]);
	const filtered = perms.length - filteredPerms.length;
	
	const filterInfo = (
		<span>
			{perms.length} {permType}ed domain{perms.length != 1 ? "s" : ""} {filtered > 0 && `(${filtered} filtered by search)`}
		</span>
	);

	const entries = filteredPerms.map((entry) => {
		return (
			<Link
				className="entry nounderline"
				key={entry.domain}
				to={`/${permType}s/${entry.domain}`}
			>
				<span id="domain">{entry.domain}</span>
				<span id="date">{new Date(entry.created_at ?? "").toLocaleString()}</span>
			</Link>
		);
	});

	return (
		<div className="domain-permissions-list">
			<form className="filter" role="search" onSubmit={filterFormSubmit}>
				<TextInput
					field={filterField}
					placeholder="example.org"
					label={`Search or add domain ${permType}`}
				/>
				<button
					type="submit"
					disabled={
						filterField.value === undefined ||
						filterField.value.length == 0
					}
				>
					{permTypeUpper}&nbsp;{filter}
				</button>
			</form>
			<div>
				{filterInfo}
				<div className="list">
					<div className="entries scrolling">
						{entries}
					</div>
				</div>
			</div>
		</div>
	);
}

function BlockHelperText() {
	return (
		<p>
			Blocking a domain blocks interaction between your instance, and all current and future accounts on
			instance(s) running on the blocked domain. Stored content will be removed, and no more data is sent to
			the remote server. This extends to all subdomains as well, so blocking 'example.com' also blocks 'social.example.com'.
			<br/>
			<a
				href="https://docs.gotosocial.org/en/latest/admin/domain_blocks/"
				target="_blank"
				className="docslink"
				rel="noreferrer"
			>
				Learn more about domain blocks (opens in a new tab)
			</a>
			<br/>
		</p>
	);
}

function AllowHelperText() {
	return (
		<p>
			Allowing a domain explicitly allows instance(s) running on that domain to interact with your instance.
			If you're running in allowlist mode, this is how you "allow" instances through.
			If you're running in blocklist mode (the default federation mode), you can use explicit domain allows
			to override domain blocks. In blocklist mode, explicitly allowed instances will be able to interact with
			your instance regardless of any domain blocks in place.  This extends to all subdomains as well, so allowing
			'example.com' also allows 'social.example.com'. This is useful when you're importing a block list but
			there are some domains on the list you don't want to block: just create an explicit allow for those domains
			before importing the list.
			<br/>
			<a
				href="https://docs.gotosocial.org/en/latest/admin/federation_modes/"
				target="_blank"
				className="docslink"
				rel="noreferrer"
			>
				Learn more about federation modes (opens in a new tab)
			</a>
		</p>
	);
}
