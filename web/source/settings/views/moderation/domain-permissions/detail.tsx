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

import { useDomainAllowsQuery, useDomainBlocksQuery } from "../../../lib/query/admin/domain-permissions/get";
import { useAddDomainAllowMutation, useAddDomainBlockMutation, useRemoveDomainAllowMutation, useRemoveDomainBlockMutation } from "../../../lib/query/admin/domain-permissions/update";
import { DomainPerm } from "../../../lib/types/domain-permission";
import { NoArg } from "../../../lib/types/query";
import { Error } from "../../../components/error";
import { useBaseUrl } from "../../../lib/navigation/util";
import { PermType } from "../../../lib/types/perm";
import { useCapitalize } from "../../../lib/util";
import { formDomainValidator } from "../../../lib/util/formvalidators";

export default function DomainPermDetail() {
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
	
	// Render different into content depending on
	// if we have a perm already for this domain.
	let infoContent: React.JSX.Element;
	if (existingPerm === undefined) {
		infoContent = (
			<span>
				No stored {permType} yet, you can add one below:
			</span>
		);
	} else {
		infoContent = (
			<div className="info">
				<i className="fa fa-fw fa-exclamation-triangle" aria-hidden="true"></i>
				<b>Editing existing domain {permTypeRaw} isn't implemented yet, <a href="https://github.com/superseriousbusiness/gotosocial/issues/1198" target="_blank" rel="noopener noreferrer">check here for progress</a></b>
			</div>
		);
	}

	return (
		<div>
			<h1 className="text-cutoff">
				<BackButton to={`~${baseUrl}/${permTypeRaw}`} />
				{" "}
				Domain {permType} for {domain}
			</h1>
			{infoContent}
			<DomainPermForm
				defaultDomain={domain}
				perm={existingPerm}
				permType={permType}
			/>
		</div>
	);
}

interface DomainPermFormProps {
	defaultDomain: string;
	perm?: DomainPerm;
	permType: PermType;
}

function DomainPermForm({ defaultDomain, perm, permType }: DomainPermFormProps) {
	const isExistingPerm = perm !== undefined;
	const disabledForm = isExistingPerm
		? {
			disabled: true,
			title: "Domain permissions currently cannot be edited."
		}
		: {
			disabled: false,
			title: "",
		};

	const form = {
		domain: useTextInput("domain", {
			source: perm,
			defaultValue: defaultDomain,
			validator: formDomainValidator,
		}),
		obfuscate: useBoolInput("obfuscate", { source: perm }),
		commentPrivate: useTextInput("private_comment", { source: perm }),
		commentPublic: useTextInput("public_comment", { source: perm })
	};

	// Check which perm type we're meant to be handling
	// here, and use appropriate mutations and results.
	// We can't call these hooks conditionally because
	// react is like "weh" (mood), but we can decide
	// which ones to use conditionally.
	const [ addBlock, addBlockResult ] = useAddDomainBlockMutation();
	const [ removeBlock, removeBlockResult] = useRemoveDomainBlockMutation({ fixedCacheKey: perm?.id });
	const [ addAllow, addAllowResult ] = useAddDomainAllowMutation();
	const [ removeAllow, removeAllowResult ] = useRemoveDomainAllowMutation({ fixedCacheKey: perm?.id });
	
	const [
		addTrigger,
		addResult,
		removeTrigger,
		removeResult,
	] = useMemo(() => {
		return permType == "block"
			? [
				addBlock,
				addBlockResult,
				removeBlock,
				removeBlockResult,
			]
			: [
				addAllow,
				addAllowResult,
				removeAllow,
				removeAllowResult,
			];
	}, [permType,
		addBlock, addBlockResult, removeBlock, removeBlockResult,
		addAllow, addAllowResult, removeAllow, removeAllowResult,
	]);

	// Use appropriate submission params for this permType.
	const [submitForm, submitFormResult] = useFormSubmit(form, [addTrigger, addResult], { changedOnly: false });

	// Uppercase first letter of given permType.
	const permTypeUpper = useCapitalize(permType);

	const [location, setLocation] = useLocation();

	function verifyUrlThenSubmit(e) {
		// Adding a new domain permissions happens on a url like
		// "/settings/admin/domain-permissions/:permType/domain.com",
		// but if domain input changes, that doesn't match anymore
		// and causes issues later on so, before submitting the form,
		// silently change url, and THEN submit.
		let correctUrl = `/${permType}s/${form.domain.value}`;
		if (location != correctUrl) {
			setLocation(correctUrl);
		}
		return submitForm(e);
	}

	return (
		<form onSubmit={verifyUrlThenSubmit}>
			<TextInput
				field={form.domain}
				label="Domain"
				placeholder="example.com"
				autoCapitalize="none"
				spellCheck="false"
				{...disabledForm}
			/>

			<Checkbox
				field={form.obfuscate}
				label="Obfuscate domain in public lists"
				{...disabledForm}
			/>

			<TextArea
				field={form.commentPrivate}
				label="Private comment"
				autoCapitalize="sentences"
				rows={3}
				{...disabledForm}
			/>

			<TextArea
				field={form.commentPublic}
				label="Public comment"
				autoCapitalize="sentences"
				rows={3}
				{...disabledForm}
			/>

			<div className="action-buttons row">
				<MutationButton
					label={permTypeUpper}
					result={submitFormResult}
					showError={false}
					{...disabledForm}
				/>

				{
					isExistingPerm &&
					<MutationButton
						type="button"
						onClick={() => removeTrigger(perm.id?? "")}
						label="Remove"
						result={removeResult}
						className="button danger"
						showError={false}
						disabled={!isExistingPerm}
					/>
				}
			</div>

			<>
				{addResult.error && <Error error={addResult.error} />}
				{removeResult.error && <Error error={removeResult.error} />}
			</>

		</form>
	);
}
