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
import {
	useAcceptDomainPermissionDraftMutation,
	useGetDomainPermissionDraftQuery,
	useRemoveDomainPermissionDraftMutation
} from "../../../../lib/query/admin/domain-permissions/drafts";
import { Error as ErrorC } from "../../../../components/error";
import UsernameLozenge from "../../../../components/username-lozenge";
import MutationButton from "../../../../components/form/mutation-button";
import { useBoolInput, useTextInput } from "../../../../lib/form";
import { Checkbox, Select } from "../../../../components/form/inputs";
import { PermType } from "../../../../lib/types/perm";

export default function DomainPermissionDraftDetail() {
	const baseUrl = useBaseUrl();
	const backLocation: string = history.state?.backLocation ?? `~${baseUrl}`;
	const params = useParams();

	let id = params.permDraftId as string | undefined;
	if (!id) {
		throw "no perm ID";
	}

	const {
		data: permDraft,
		isLoading,
		isFetching,
		isError,
		error,
	} = useGetDomainPermissionDraftQuery(id);

	if (isLoading || isFetching) {
		return <Loading />;
	} else if (isError) {
		return <ErrorC error={error} />;
	} else if (permDraft === undefined) {
		return <ErrorC error={new Error("permission draft was undefined")} />;
	}

	const created = permDraft.created_at ? new Date(permDraft.created_at).toDateString(): "unknown";
	const domain = permDraft.domain;
	const permType = permDraft.permission_type;
	if (!permType) {
		return <ErrorC error={new Error("permission_type was undefined")} />;
	}
	const publicComment = permDraft.public_comment ?? "[none]";
	const privateComment = permDraft.private_comment ?? "[none]";
	const subscriptionID = permDraft.subscription_id ?? "[none]";

	return (
		<div className="domain-permission-draft-details">
			<h1><BackButton to={backLocation} /> Domain Permission Draft Detail</h1>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Created</dt>
					<dd><time dateTime={permDraft.created_at}>{created}</time></dd>
				</div>
				<div className="info-list-entry">
					<dt>Created By</dt>
					<dd>
						<UsernameLozenge
							account={permDraft.created_by}
							linkTo={`~/settings/moderation/accounts/${permDraft.created_by}`}
							backLocation={`~${location}`}
						/>
					</dd>
				</div>
				<div className="info-list-entry">
					<dt>Domain</dt>
					<dd>{domain}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Permission type</dt>
					<dd className={`permission-type ${permType}`}>
						<i
							aria-hidden={true}
							className={`fa fa-${permType === "allow" ? "check" : "close"}`}
						></i>
						{permType}
					</dd>
				</div>
				<div className="info-list-entry">
					<dt>Private comment</dt>
					<dd>{privateComment}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Public comment</dt>
					<dd>{publicComment}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Subscription ID</dt>
					<dd>{subscriptionID}</dd>
				</div>
			</dl>
			<HandleDraft
				id={id}
				permType={permType}
				backLocation={backLocation}
			/> 
		</div>
	);
}

function HandleDraft({ id, permType, backLocation }: { id: string, permType: PermType, backLocation: string }) {
	const [ accept, acceptResult ] = useAcceptDomainPermissionDraftMutation();
	const [ remove, removeResult ] = useRemoveDomainPermissionDraftMutation();
	const [_location, setLocation] = useLocation();
	const form = {
		acceptOrRemove: useTextInput("accept_or_remove", { defaultValue: "accept" }),
		overwrite: useBoolInput("overwrite"),
		exclude_target: useBoolInput("exclude_target"),
	};

	const onClick = (e) => {
		e.preventDefault();
		if (form.acceptOrRemove.value === "accept") {
			const overwrite = form.overwrite.value;
			accept({id, overwrite, permType}).then(res => {
				if ("data" in res) {
					setLocation(backLocation);
				}
			});
		} else {
			const exclude_target = form.exclude_target.value;
			remove({id, exclude_target}).then(res => {
				if ("data" in res) {
					setLocation(backLocation);
				}
			});	
		}
	};

	return (
		<form>
			<Select
				field={form.acceptOrRemove}
				label="Accept or remove draft"
				options={
					<>
						<option value="accept">Accept</option>
						<option value="remove">Remove</option>
					</>
				}
			></Select>
			
			{ form.acceptOrRemove.value === "accept" &&
				<>
					<Checkbox
						field={form.overwrite}
						label={`Overwrite any existing ${permType} for this domain`}
					/>
				</>
			}

			{ form.acceptOrRemove.value === "remove" &&
				<>
					<Checkbox
						field={form.exclude_target}
						label={`Add a domain permission exclude for this domain`}
					/>
				</>
			}

			<MutationButton
				label={
					form.acceptOrRemove.value === "accept"
						? `Accept ${permType}`
						: "Remove draft"
				}
				type="button"
				className={
					form.acceptOrRemove.value === "accept"
						? "button"
						: "button danger"
				}
				onClick={onClick}
				disabled={false}
				showError={true}
				result={
					form.acceptOrRemove.value === "accept"
						? acceptResult
						: removeResult
				}
			/>
		</form>
	);
}
