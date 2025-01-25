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
import { useGetHeaderAllowsQuery, useGetHeaderBlocksQuery } from "../../../lib/query/admin/http-header-permissions";
import { NoArg } from "../../../lib/types/query";
import { PageableList } from "../../../components/pageable-list";
import { HeaderPermission } from "../../../lib/types/http-header-permissions";
import { useLocation, useParams } from "wouter";
import { PermType } from "../../../lib/types/perm";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import HeaderPermCreateForm from "./create";
import { useCapitalize } from "../../../lib/util";

export default function HeaderPermsOverview() {
	const [ location, setLocation ] = useLocation();
	
	// Parse perm type from routing params.
	let params = useParams();
	if (params.permType !== "blocks" && params.permType !== "allows") {
		throw "unrecognized perm type " + params.permType;
	}
	const permType = useMemo(() => {
		return params.permType?.slice(0, -1) as PermType;
	}, [params]);

	// Uppercase first letter of given permType.
	const permTypeUpper = useCapitalize(permType);
	
	// Fetch desired perms, skipping
	// the ones we don't want.
	const {
		data: blocks,
		isLoading: isLoadingBlocks,
		isFetching: isFetchingBlocks,
		isSuccess: isSuccessBlocks,
		isError: isErrorBlocks,
		error: errorBlocks
	} = useGetHeaderBlocksQuery(NoArg, { skip: permType !== "block" });

	const {
		data: allows,
		isLoading: isLoadingAllows,
		isFetching: isFetchingAllows,
		isSuccess: isSuccessAllows,
		isError: isErrorAllows,
		error: errorAllows
	} = useGetHeaderAllowsQuery(NoArg, { skip: permType !== "allow" });

	const itemToEntry = (perm: HeaderPermission) => {
		return (
			<dl
				key={perm.id}
				className="entry pseudolink"
				onClick={() => {
					// When clicking on a header perm,
					// go to the detail view for perm.
					setLocation(`/${permType}s/${perm.id}`, {
						// Store the back location in
						// history so the detail view
						// can use it to return here.
						state: { backLocation: location }
					});
				}}
				role="link"
				tabIndex={0}
			>
				<dt>{perm.header}</dt>
				<dd>{perm.regex}</dd>
			</dl>
		);
	};

	const emptyMessage = (
		<div className="info">
			<i className="fa fa-fw fa-info-circle" aria-hidden="true"></i>
			<b>
				No HTTP header {permType}s exist yet.
				You can create one using the form below.
			</b>
		</div>
	);

	let isLoading: boolean;
	let isFetching: boolean;
	let isSuccess: boolean; 
	let isError: boolean;
	let error: FetchBaseQueryError | SerializedError | undefined;
	let items: HeaderPermission[] | undefined;

	if (permType === "block") {
		isLoading = isLoadingBlocks;
		isFetching = isFetchingBlocks;
		isSuccess = isSuccessBlocks;
		isError = isErrorBlocks;
		error = errorBlocks;
		items = blocks;
	} else {
		isLoading = isLoadingAllows;
		isFetching = isFetchingAllows;
		isSuccess = isSuccessAllows;
		isError = isErrorAllows;
		error = errorAllows;
		items = allows;
	}

	return (
		<div className="http-header-permissions">
			<div className="form-section-docs">
				<h1>HTTP Header {permTypeUpper}s</h1>
				<p>
					On this page, you can view, create, and remove HTTP header {permType} entries,
					<br/>
					Blocks and allows have different effects depending on the value you've set
					for <code>advanced-header-filter-mode</code> in your instance configuration.
					<br/>
					{ permType === "block" && <>
						<strong>
							When running in <code>block</code> mode, be very careful when creating
							your value regexes, as a too-broad match can cause your instance to
							deny all requests, locking you out of this settings panel.
						</strong>
						<br/>
						If you do this by accident, you can fix it by stopping your instance,
						changing <code>advanced-header-filter-mode</code> to an empty string
						(disabled), starting your instance again, and removing the block.
					</> }
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/admin/request_filtering_modes/"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about HTTP request filtering (opens in a new tab)
				</a>
			</div>
			<PageableList
				isLoading={isLoading}
				isFetching={isFetching}
				isSuccess={isSuccess}
				isError={isError}
				error={error}
				items={items}
				itemToEntry={itemToEntry}
				emptyMessage={emptyMessage}
			/>
			<HeaderPermCreateForm permType={permType} />
		</div>
	);
}
