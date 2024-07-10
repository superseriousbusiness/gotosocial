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

import React, { useEffect, useMemo } from "react";
import { useLocation, useParams } from "wouter";
import { PermType } from "../../../lib/types/perm";
import { useDeleteHeaderAllowMutation, useDeleteHeaderBlockMutation, useGetHeaderAllowQuery, useGetHeaderBlockQuery } from "../../../lib/query/admin/http-header-permissions";
import { HeaderPermission } from "../../../lib/types/http-header-permissions";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import Loading from "../../../components/loading";
import { Error } from "../../../components/error";
import { useLazyGetAccountQuery } from "../../../lib/query/admin";
import Username from "../../../components/username";
import { useBaseUrl } from "../../../lib/navigation/util";
import BackButton from "../../../components/back-button";
import MutationButton from "../../../components/form/mutation-button";

const testString = `/* To test this properly, set "flavor" to "Golang", as that's the language GoToSocial uses for regular expressions */

/* Amazon crawler User-Agent example */
Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/600.2.5 (KHTML\\, like Gecko) Version/8.0.2 Safari/600.2.5 (Amazonbot/0.1; +https://developer.amazon.com/support/amazonbot)

/* Some other test strings */
Some Test Value
Another Test Value`;

export default function HeaderPermDetail() {
	let params = useParams();
	if (params.permType !== "blocks" && params.permType !== "allows") {
		throw "unrecognized perm type " + params.permType;
	}
	const permType = useMemo(() => {
		return params.permType?.slice(0, -1) as PermType;
	}, [params]);

	let permID = params.permId as string | undefined;
	if (!permID) {
		throw "no perm ID";
	}

	if (permType === "block") {
		return <BlockDetail id={permID} />;
	} else {
		return <AllowDetail id={permID} />;
	}
}

function BlockDetail({ id }: { id: string }) {
	return (
		<PermDeets
			permType={"Block"}
			{...useGetHeaderBlockQuery(id)}
		/>
	);
}

function AllowDetail({ id }: { id: string }) {
	return (
		<PermDeets
			permType={"Allow"}
			{...useGetHeaderAllowQuery(id)}
		/>
	);
}

interface PermDeetsProps {
	permType: string;
	data?: HeaderPermission;
	isLoading: boolean;
	isFetching: boolean;
	isError: boolean;
	error?: FetchBaseQueryError | SerializedError;
}

function PermDeets({
	permType,
	data: perm,
	isLoading: isLoadingPerm,
	isFetching: isFetchingPerm,
	isError: isErrorPerm,
	error: errorPerm,
}: PermDeetsProps) {
	const [ location ] = useLocation();
	const baseUrl = useBaseUrl();
	
	// Once we've loaded the perm, trigger
	// getting the account that created it.
	const [ getAccount, getAccountRes ] = useLazyGetAccountQuery();
	useEffect(() => {
		if (!perm) {
			return;
		}
		getAccount(perm.created_by, true);
	}, [getAccount, perm]);

	// Load the createdByAccount if possible,
	// returning a username lozenge with
	// a link to the account.
	const createdByAccount = useMemo(() => {
		const {
			data: account,
			isLoading: isLoadingAccount,
			isFetching: isFetchingAccount,
			isError: isErrorAccount,
		} = getAccountRes;
		
		// Wait for query to finish, returning
		// loading spinner in the meantime.
		if (isLoadingAccount || isFetchingAccount || !perm) {
			return <Loading />;
		} else if (isErrorAccount || account === undefined) {
			// Fall back to account ID.
			return perm?.created_by;
		}

		return (
			<Username
				account={account}
				linkTo={`~/settings/moderation/accounts/${account.id}`}
				backLocation={`~${baseUrl}${location}`}
			/>
		);
	}, [getAccountRes, perm, baseUrl, location]);

	// Now wait til the perm itself is loaded.
	if (isLoadingPerm || isFetchingPerm) {
		return <Loading />;
	} else if (isErrorPerm) {
		return <Error error={errorPerm} />;
	} else if (perm === undefined) {
		throw "perm undefined";
	}

	const created = new Date(perm.created_at).toDateString();	
	
	// Create parameters to link to regex101
	// with this regular expression prepopulated.
	const testParams = new URLSearchParams();
	testParams.set("regex", perm.regex);
	testParams.set("flags", "gm");
	testParams.set("testString", testString);
	const regexLink = `https://regex101.com/?${testParams.toString()}`;	

	return (
		<div className="http-header-permission-details">
			<h1><BackButton to={`~${baseUrl}/${permType.toLowerCase()}s`} /> HTTP Header {permType} Detail</h1>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>ID</dt>
					<dd className="monospace">{perm.id}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Created</dt>
					<dd><time dateTime={perm.created_at}>{created}</time></dd>
				</div>
				<div className="info-list-entry">
					<dt>Created By</dt>
					<dd>{createdByAccount}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Header Name</dt>
					<dd className="monospace">{perm.header}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Value Regex</dt>
					<dd className="monospace">{perm.regex}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Test This Regex</dt>
					<dd>
						<a
							href={regexLink}
							target="_blank"
							rel="noreferrer"
						>
							<i className="fa fa-fw fa-external-link" aria-hidden="true"></i> Link to Regex101 (opens in a new tab)
						</a>
					</dd>
				</div>
			</dl>
			{ permType === "Block"
				? <DeleteBlock id={perm.id} />
				: <DeleteAllow id={perm.id} />
			}
		</div>
	);
}

function DeleteBlock({ id }: { id: string }) {
	const [ _location, setLocation ] = useLocation();
	const baseUrl = useBaseUrl();
	const [ removeTrigger, removeResult ] = useDeleteHeaderBlockMutation();
	
	return (
		<MutationButton
			type="button"
			onClick={() => {
				removeTrigger(id);
				setLocation(`~${baseUrl}/blocks`);
			}}
			label="Remove this block"
			result={removeResult}
			className="button danger"
			showError={false}
			disabled={false}
		/>
	);
}

function DeleteAllow({ id }: { id: string }) {
	const [ _location, setLocation ] = useLocation();
	const baseUrl = useBaseUrl();
	const [ removeTrigger, removeResult ] = useDeleteHeaderAllowMutation();
	
	return (
		<MutationButton
			type="button"
			onClick={() => {
				removeTrigger(id);
				setLocation(`~${baseUrl}/allows`);
			}}
			label="Remove this allow"
			result={removeResult}
			className="button danger"
			showError={false}
			disabled={false}
		/>
	);
}
