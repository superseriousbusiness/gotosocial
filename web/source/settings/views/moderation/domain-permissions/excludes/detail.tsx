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
import { useLocation, useParams } from "wouter";
import Loading from "../../../../components/loading";
import { useBaseUrl } from "../../../../lib/navigation/util";
import BackButton from "../../../../components/back-button";
import { Error as ErrorC } from "../../../../components/error";
import UsernameLozenge from "../../../../components/username-lozenge";
import { useDeleteDomainPermissionExcludeMutation, useGetDomainPermissionExcludeQuery } from "../../../../lib/query/admin/domain-permissions/excludes";
import MutationButton from "../../../../components/form/mutation-button";

export default function DomainPermissionExcludeDetail() {
	const baseUrl = useBaseUrl();
	const backLocation: string = history.state?.backLocation ?? `~${baseUrl}`;

	const params = useParams();
	let id = params.excludeId as string | undefined;
	if (!id) {
		throw "no perm ID";
	}

	const {
		data: permExclude,
		isLoading,
		isFetching,
		isError,
		error,
	} = useGetDomainPermissionExcludeQuery(id);

	if (isLoading || isFetching) {
		return <Loading />;
	} else if (isError) {
		return <ErrorC error={error} />;
	} else if (permExclude === undefined) {
		return <ErrorC error={new Error("permission exclude was undefined")} />;
	}

	const created = permExclude.created_at ? new Date(permExclude.created_at).toDateString(): "unknown";
	const domain = permExclude.domain;
	const privateComment = permExclude.private_comment ?? "[none]";

	return (
		<div className="domain-permission-exclude-details">
			<h1><BackButton to={backLocation} /> Domain Permission Exclude Detail</h1>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Created</dt>
					<dd><time dateTime={permExclude.created_at}>{created}</time></dd>
				</div>
				<div className="info-list-entry">
					<dt>Created By</dt>
					<dd>
						<UsernameLozenge
							account={permExclude.created_by}
							linkTo={`~/settings/moderation/accounts/${permExclude.created_by}`}
							backLocation={`~${location}`}
						/>
					</dd>
				</div>
				<div className="info-list-entry">
					<dt>Domain</dt>
					<dd>{domain}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Private comment</dt>
					<dd>{privateComment}</dd>
				</div>
			</dl>
			<HandleExclude
				id={id}
				backLocation={backLocation}
			/>
		</div>
	);
}

function HandleExclude({ id, backLocation}: {id: string, backLocation: string}) {
	const [_location, setLocation] = useLocation();
	const [ deleteExclude, deleteResult ] = useDeleteDomainPermissionExcludeMutation();
	
	return (
		<MutationButton
			label={`Delete exclude`}
			title={`Delete exclude`}
			type="button"
			className="button danger"
			onClick={(e) => {
				e.preventDefault();
				e.stopPropagation();
				deleteExclude(id).then(res => {
					if ("data" in res) {
						setLocation(backLocation);
					}
				});
			}}
			disabled={false}
			showError={true}
			result={deleteResult}
		/>
	);
}
