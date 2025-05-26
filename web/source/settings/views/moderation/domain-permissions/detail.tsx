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
import { useLocation, useParams, useSearch } from "wouter";

import { useTextInput, useBoolInput } from "../../../lib/form";

import useFormSubmit from "../../../lib/form/submit";

import { TextInput, Checkbox, TextArea } from "../../../components/form/inputs";

import Loading from "../../../components/loading";
import BackButton from "../../../components/back-button";
import MutationButton from "../../../components/form/mutation-button";

import { 
	useDomainAllowsQuery,
	useDomainBlocksQuery,
} from "../../../lib/query/admin/domain-permissions/get";
import {
	useAddDomainAllowMutation,
	useAddDomainBlockMutation,
	useRemoveDomainAllowMutation,
	useRemoveDomainBlockMutation,
	useUpdateDomainAllowMutation,
	useUpdateDomainBlockMutation,
} from "../../../lib/query/admin/domain-permissions/update";
import { DomainPerm } from "../../../lib/types/domain-permission";
import { NoArg } from "../../../lib/types/query";
import { Error } from "../../../components/error";
import { useBaseUrl } from "../../../lib/navigation/util";
import { PermType } from "../../../lib/types/perm";
import { useCapitalize } from "../../../lib/util";
import { formDomainValidator } from "../../../lib/util/formvalidators";
import UsernameLozenge from "../../../components/username-lozenge";
import { FormSubmitEvent } from "../../../lib/form/types";

export default function DomainPermView() {
	const baseUrl = useBaseUrl();
	const search = useSearch();

	// Parse perm type from routing params, converting
	// "blocks" => "block" and "allows" => "allow".
	const params = useParams();
	const permTypeRaw = params.permType;
	if (permTypeRaw !== "blocks" && permTypeRaw !== "allows") {
		throw "unrecognized perm type " + params.permType;
	}
	const permType = useMemo(() => {
		return permTypeRaw.slice(0, -1) as PermType;
	}, [permTypeRaw]);
	
	// Conditionally fetch either domain blocks or domain
	// allows depending on which perm type we're looking at.
	const {
		data: blocks = {},
		isLoading: loadingBlocks,
		isFetching: fetchingBlocks,
	} = useDomainBlocksQuery(NoArg, { skip: permType !== "block" });
	const {
		data: allows = {},
		isLoading: loadingAllows,
		isFetching: fetchingAllows,
	} = useDomainAllowsQuery(NoArg, { skip: permType !== "allow" });

	// Wait until we're done loading.
	const loading = permType === "block"
		? loadingBlocks || fetchingBlocks
		: loadingAllows || fetchingAllows;
	if (loading) {
		return <Loading />;
	}

	// Parse domain from routing params.
	let domain = params.domain ?? "unknown";
	if (domain === "view") {
		// Retrieve domain from form field submission.
		const searchParams = new URLSearchParams(search);
		const searchDomain = searchParams.get("domain");
		if (!searchDomain) {
			throw "empty view domain";
		}
		
		domain = searchDomain;
	}

	// Normalize / decode domain
	// (it may be URL-encoded).
	domain = decodeURIComponent(domain);

	// Check if we already have a perm
	// of the desired type for this domain.
	const existingPerm = permType === "block"
		? blocks[domain]
		: allows[domain];
	
	const title = <span>Domain {permType} for {domain}</span>;

	return (
		<div className="domain-permission-details">
			<h1><BackButton to={`~${baseUrl}/${permTypeRaw}`} /> {title}</h1>
			{ existingPerm
				? <DomainPermDetails perm={existingPerm} permType={permType} />
				: <span>No stored {permType} yet, you can add one below:</span>
			}
			<CreateOrUpdateDomainPerm
				defaultDomain={domain}
				perm={existingPerm}
				permType={permType}
			/>
		</div>
	);
}

interface DomainPermDetailsProps {
	perm: DomainPerm,
	permType: PermType,
}

function DomainPermDetails({
	perm,
	permType
}: DomainPermDetailsProps) {
	const baseUrl = useBaseUrl();
	const [ location ] = useLocation();
	
	const created = useMemo(() => {
		if (perm.created_at) {
			return new Date(perm.created_at).toDateString();
		}
		return "unknown";
	}, [perm.created_at]);

	return (
		<dl className="info-list">
			<div className="info-list-entry">
				<dt>Created</dt>
				<dd><time dateTime={perm.created_at}>{created}</time></dd>
			</div>
			<div className="info-list-entry">
				<dt>Created By</dt>
				<dd>
					<UsernameLozenge
						account={perm.created_by}
						linkTo={`~/settings/moderation/accounts/${perm.created_by}`}
						backLocation={`~${baseUrl}${location}`}
					/>
				</dd>
			</div>
			<div className="info-list-entry">
				<dt>Domain</dt>
				<dd>{perm.domain}</dd>
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
				<dt>Subscription ID</dt>
				<dd>{perm.subscription_id ?? "[none]"}</dd>
			</div>
		</dl>
	);
}

interface CreateOrUpdateDomainPermProps {
	defaultDomain: string;
	perm?: DomainPerm;
	permType: PermType;
}

function CreateOrUpdateDomainPerm({
	defaultDomain,
	perm,
	permType
}: CreateOrUpdateDomainPermProps) {
	const isExistingPerm = perm !== undefined;

	const form = {
		domain: useTextInput("domain", {
			source: perm,
			defaultValue: defaultDomain,
			validator: formDomainValidator,
		}),
		obfuscate: useBoolInput("obfuscate", { source: perm }),
		privateComment: useTextInput("private_comment", { source: perm }),
		publicComment: useTextInput("public_comment", { source: perm })
	};

	// Check which perm type we're meant to be handling
	// here, and use appropriate mutations and results.
	// We can't call these hooks conditionally because
	// react is like "weh" (mood), but we can decide
	// which ones to use conditionally.
	const [ addBlock, addBlockResult ] = useAddDomainBlockMutation();
	const [ updateBlock, updateBlockResult ] = useUpdateDomainBlockMutation({ fixedCacheKey: perm?.id });
	const [ removeBlock, removeBlockResult] = useRemoveDomainBlockMutation({ fixedCacheKey: perm?.id });
	const [ addAllow, addAllowResult ] = useAddDomainAllowMutation();
	const [ updateAllow, updateAllowResult ] = useUpdateDomainAllowMutation({ fixedCacheKey: perm?.id });
	const [ removeAllow, removeAllowResult ] = useRemoveDomainAllowMutation({ fixedCacheKey: perm?.id });
	
	const [
		createOrUpdateTrigger,
		createOrUpdateResult,
		removeTrigger,
		removeResult,
	] = useMemo(() => {
		switch (true) {
			case (permType === "block" && !isExistingPerm):
				return [ addBlock, addBlockResult, removeBlock, removeBlockResult ];
			case (permType === "block"):
				return [ updateBlock, updateBlockResult, removeBlock, removeBlockResult ];
			case !isExistingPerm:
				return [ addAllow, addAllowResult, removeAllow, removeAllowResult ];
			default:
				return [ updateAllow, updateAllowResult, removeAllow, removeAllowResult ];
		}
	}, [permType, isExistingPerm,
		addBlock, addBlockResult, updateBlock, updateBlockResult, removeBlock, removeBlockResult,
		addAllow, addAllowResult, updateAllow, updateAllowResult, removeAllow, removeAllowResult,
	]);

	// Use appropriate submission params for this
	// permType, and whether we're creating or updating.
	const [submit, submitResult] = useFormSubmit(
		form,
		[ createOrUpdateTrigger, createOrUpdateResult ],
		{
			changedOnly: isExistingPerm,
			// If we're updating an existing perm,
			// insert the perm ID into the mutation
			// data before submitting. Otherwise just
			// return the mutationData unmodified.
			customizeMutationArgs: (mutationData) => {
				if (isExistingPerm) {
					return {
						id: perm?.id,
						...mutationData,
					};
				} else {
					return mutationData;
				}
			},
		},
	);

	// Uppercase first letter of given permType.
	const permTypeUpper = useCapitalize(permType);

	const [location, setLocation] = useLocation();
	function onSubmit(e: FormSubmitEvent) {
		// Adding a new domain permissions happens on a url like
		// "/settings/admin/domain-permissions/:permType/domain.com",
		// but if domain input changes, that doesn't match anymore
		// and causes issues later on so, before submitting the form,
		// silently change url, and THEN submit.
		if (!isExistingPerm) {
			let correctUrl = `/${permType}s/${form.domain.value}`;
			if (location != correctUrl) {
				setLocation(correctUrl);
			}
		}
		return submit(e);
	}

	return (
		<form onSubmit={onSubmit}>
			{ !isExistingPerm && 
				<TextInput
					field={form.domain}
					label="Domain"
					placeholder="example.com"
					autoCapitalize="none"
					spellCheck="false"
				/>
			}

			<Checkbox
				field={form.obfuscate}
				label="Obfuscate domain in public lists"
			/>

			<TextArea
				field={form.privateComment}
				label="Private comment (shown to admins only)"
				autoCapitalize="sentences"
				rows={3}
			/>

			<TextArea
				field={form.publicComment}
				label="Public comment (shown to members of this instance via the instance info page, and on the web if enabled)"
				autoCapitalize="sentences"
				rows={3}
			/>

			<div className="action-buttons row">
				<MutationButton
					label={isExistingPerm ? "Update " + permType.toString() : permTypeUpper}
					result={submitResult}
					disabled={
						isExistingPerm &&
						!form.obfuscate.hasChanged() &&
						!form.privateComment.hasChanged() &&
						!form.publicComment.hasChanged()
					}
				/>

				{ isExistingPerm &&
					<button
						type="button"
						onClick={() => removeTrigger(perm.id?? "")}
						className="button danger"
					>
						Remove {permType.toString()}
					</button>
				}
			</div>

			<>
				{createOrUpdateResult.error && <Error error={createOrUpdateResult.error} />}
				{removeResult.error && <Error error={removeResult.error} />}
			</>

		</form>
	);
}
