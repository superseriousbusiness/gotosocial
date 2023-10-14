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
import { Link, useLocation } from "wouter";
import { matchSorter } from "match-sorter";

import { useTextInput } from "../../lib/form";

import { TextInput } from "../../components/form/inputs";

import Loading from "../../components/loading";
import { useDomainAllowsQuery, useDomainBlocksQuery } from "../../lib/query/admin/domain-permissions/get";
import { TextFormInputHook } from "../../lib/form/types";
import { DomainPerm, MappedDomainPerms, PermType } from "../../lib/types/domain-permission";
import { NoArg } from "../../lib/types/query";

export interface DomainPermissionsOverviewProps {
	// Params injected by
	// the wouter router.
	permType: PermType;
	baseUrl: string,
}

export default function DomainPermissionsOverview({ permType, baseUrl }: DomainPermissionsOverviewProps) {	
	if (permType !== "block" && permType !== "allow") {
		throw "unrecognized perm type " + permType;
	}

	// Uppercase first letter of given permType.
	const permTypeUpper = useMemo(() => {
		return permType.charAt(0).toUpperCase() + permType.slice(1); 
	}, [permType])

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
		<div>
			<h1>Domain {permTypeUpper}s</h1>
			{ permType == "block" ? <BlockHelperText/> : <AllowHelperText/> }
			<DomainPermsList
				data={data}
				baseUrl={baseUrl}
				permType={permType}
				permTypeUpper={permTypeUpper}
			/>
			<Link to={`${baseUrl}/import-export`}>
				<a>Or use the bulk import/export interface</a>
			</Link>
		</div>
	);
};

interface DomainPermsListProps {
	data: MappedDomainPerms;
	baseUrl: string;
	permType: PermType;
	permTypeUpper: string;
}

function DomainPermsList({ data, baseUrl, permType, permTypeUpper }: DomainPermsListProps) {
	// Format perms into a list.
	const perms = useMemo(() => {
		return Object.values(data);
	}, [data]);

	const [_location, setLocation] = useLocation();
	const filterField = useTextInput("filter");
	
	function filterFormSubmit(e) {
		e.preventDefault();
		setLocation(`${baseUrl}/${filter}`);
	}
	
	const filter = filterField.value ?? "";
	const filteredPerms = useMemo(() => {
		return matchSorter(perms, filter, { keys: ["domain"] });
	}, [perms, filter]);
	const filtered = perms.length - filteredPerms.length;
	
	const filterInfo = (
		<span>
			{perms.length} {permType}ed instance{perms.length != 1 ? "s" : ""} {filtered > 0 && `(${filtered} filtered by search)`}
		</span>
	)

	const entries = filteredPerms.map((entry) => {
		return (
			<Link key={entry.domain} to={`${baseUrl}/${entry.domain}`}>
				<a className="entry nounderline">
					<span id="domain">{entry.domain}</span>
					<span id="date">{new Date(entry.created_at ?? "").toLocaleString()}</span>
				</a>
			</Link>
		);
	})

	return (
		<div className="domain-permissions-list">
			<form className="filter" role="search" onSubmit={filterFormSubmit}>
				<TextInput
					field={filterField}
					placeholder="example.org"
					label={`Search or add domain ${permType}`}
				/>
				<Link to={`${baseUrl}/${filter}`}>
					<a className="button">{permTypeUpper}&nbsp;{filter}</a>
				</Link>
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
	)
}

function BlockHelperText({}) {
	return (
		<>
			<br/>Blocking a domain blocks all current and future accounts on instance(s) running on that domain.
			<br/>Stored content will be removed, and no more data is sent to the remote server.
			<br/>This extends to all subdomains as well, so blocking 'example.com' also includes 'social.example.com'.
			<a
				href="https://docs.gotosocial.org/en/latest/admin/domain_blocks/"
				target="_blank"
				className="docslink"
				rel="noreferrer"
			>
				Learn more about domain blocks (opens in a new tab)
			</a>
			<br/>
		</>
	)
}

function AllowHelperText({}) {
	return (
		<>
			<br/>Blah blah blah blah blah.
			<a
				href="https://docs.gotosocial.org/en/latest/admin/federation_modes/"
				target="_blank"
				className="docslink"
				rel="noreferrer"
			>
				Learn more about allowlist federation mode (opens in a new tab)
			</a>
			<br/>
		</>
	)
}
