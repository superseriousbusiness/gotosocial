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
import { memo, useMemo, useCallback, useEffect } from "react";
import { isValidDomainPermission, hasBetterScope } from "../../../lib/util/domain-permission";

import {
	useTextInput,
	useBoolInput,
	useCheckListInput,
} from "../../../lib/form";

import {
	Select,
	TextArea,
	Checkbox,
	TextInput,
} from "../../../components/form/inputs";

import useFormSubmit from "../../../lib/form/submit";

import CheckList from "../../../components/check-list";
import MutationButton from "../../../components/form/mutation-button";
import FormWithData from "../../../lib/form/form-with-data";

import { useImportDomainPermsMutation } from "../../../lib/query/admin/domain-permissions/import";
import {
	useDomainAllowsQuery,
	useDomainBlocksQuery
} from "../../../lib/query/admin/domain-permissions/get";

import type { DomainPerm, MappedDomainPerms } from "../../../lib/types/domain-permission";
import type { ChecklistInputHook, RadioFormInputHook } from "../../../lib/form/types";

export interface ProcessImportProps {
	list: DomainPerm[],
	permType: RadioFormInputHook,
}

export const ProcessImport = memo(
	function ProcessImport({ list, permType }: ProcessImportProps) {
		return (
			<FormWithData
				dataQuery={permType.value == "allow"
					? useDomainAllowsQuery
					: useDomainBlocksQuery
				}
				DataForm={ImportList}
				{...{ list, permType }}
			/>
		);
	}
);

export interface ImportListProps {
	list: Array<DomainPerm>,
	data: MappedDomainPerms,
	permType: RadioFormInputHook,
}

function ImportList({ list, data: domainPerms, permType }: ImportListProps) {
	const hasComment = useMemo(() => {
		let hasPublic = false;
		let hasPrivate = false;

		list.some((entry) => {
			if (entry.public_comment) {
				hasPublic = true;
			}

			if (entry.private_comment) {
				hasPrivate = true;
			}

			return hasPublic && hasPrivate;
		});

		if (hasPublic && hasPrivate) {
			return { both: true };
		} else if (hasPublic) {
			return { type: "public_comment" };
		} else if (hasPrivate) {
			return { type: "private_comment" };
		} else {
			return {};
		}
	}, [list]);

	const showComment = useTextInput("showComment", { defaultValue: hasComment.type ?? "public_comment" });

	const form = {
		domains: useCheckListInput("domains", { entries: list }), // DomainPerm is actually also a Checkable.
		obfuscate: useBoolInput("obfuscate"),
		privateComment: useTextInput("private_comment", {
			defaultValue: `Imported on ${new Date().toLocaleString()}`
		}),
		replacePrivateComment: useBoolInput("replace_private_comment", { defaultValue: false }),
		publicComment: useTextInput("public_comment"),
		replacePublicComment: useBoolInput("replace_public_comment", { defaultValue: false }),
		permType: permType,
	};

	const [importDomains, importResult] = useFormSubmit(
		form,
		useImportDomainPermsMutation(),
		{ changedOnly: false },
	);

	return (
		<form
			onSubmit={importDomains}
			className="domain-perm-import-list"
		>
			<span>{list.length} domain{list.length != 1 ? "s" : ""} in this list</span>

			{hasComment.both &&
					<Select field={showComment} options={
						<>
							<option value="public_comment">Show public comments</option>
							<option value="private_comment">Show private comments</option>
						</>
					} />
			}

			<div className="checkbox-list-wrapper">
				<DomainCheckList
					field={form.domains}
					domainPerms={domainPerms}
					commentType={showComment.value as "public_comment" | "private_comment"}
					permType={form.permType}
				/>
			</div>

			<Checkbox
				field={form.obfuscate}
				label="Obfuscate domains in public lists"
			/>

			<div className="set-comment-checkbox">
				<Checkbox
					field={form.replacePrivateComment}
					label="Set/replace private comment(s) to:"
				/>
				<TextArea
					field={form.privateComment}
					rows={3}
					disabled={!form.replacePrivateComment.value}
					placeholder="Private comment"
				/>
			</div>

			<div className="set-comment-checkbox">
				<Checkbox
					field={form.replacePublicComment}
					label="Set/replace public comment(s) to:"
				/>
				<TextArea
					field={form.publicComment}
					rows={3}
					disabled={!form.replacePublicComment.value}
					placeholder="Public comment"
				/>
			</div>


			<MutationButton
				label="Import"
				disabled={false}
				result={importResult}
			/>
		</form>
	);
}

interface DomainCheckListProps {
	field: ChecklistInputHook,
	domainPerms: MappedDomainPerms,
	commentType: "public_comment" | "private_comment",
	permType: RadioFormInputHook,
}

function DomainCheckList({ field, domainPerms, commentType, permType }: DomainCheckListProps) {
	const getExtraProps = useCallback((entry: DomainPerm) => {
		return {
			comment: entry[commentType],
			alreadyExists: entry.domain in domainPerms,
			permType: permType,
		};
	}, [domainPerms, commentType, permType]);

	const entriesWithSuggestions = useMemo(() => {
		const fieldValue = (field.value ?? {}) as { [k: string]: DomainPerm; };
		return Object.values(fieldValue).filter((entry) => entry.suggest);
	}, [field.value]);

	return (
		<>
			<CheckList
				field={field as ChecklistInputHook}
				header={<>
					<b>Domain</b>
					<b>
						{commentType == "public_comment" && "Public comment"}
						{commentType == "private_comment" && "Private comment"}
					</b>
				</>}
				EntryComponent={DomainEntry}
				getExtraProps={getExtraProps}
			/>
			<UpdateHint
				entries={entriesWithSuggestions}
				updateEntry={field.onChange}
				updateMultiple={field.updateMultiple}
			/>
		</>
	);
}

interface UpdateHintProps {
	entries,
	updateEntry,
	updateMultiple,
}

const UpdateHint = memo(
	function UpdateHint({ entries, updateEntry, updateMultiple }: UpdateHintProps) {
		if (entries.length == 0) {
			return null;
		}

		function changeAll() {
			updateMultiple(
				entries.map((entry) => [entry.key, { domain: entry.suggest, suggest: null }])
			);
		}

		return (
			<div className="update-hints">
				<p>
					{entries.length} {entries.length == 1 ? "entry uses" : "entries use"} a specific subdomain,
					which you might want to change to the main domain, as that includes all it's (future) subdomains.
				</p>
				<div className="hints">
					{entries.map((entry) => (
						<UpdateableEntry key={entry.key} entry={entry} updateEntry={updateEntry} />
					))}
				</div>
				{entries.length > 0 && <a onClick={changeAll}>change all</a>}
			</div>
		);
	}
);

interface UpdateableEntryProps {
	entry,
	updateEntry,
}

const UpdateableEntry = memo(
	function UpdateableEntry({ entry, updateEntry }: UpdateableEntryProps) {
		return (
			<>
				<span className="text-cutoff">{entry.domain}</span>
				<i className="fa fa-long-arrow-right" aria-hidden="true"></i>
				<span>{entry.suggest}</span>
				<a role="button" onClick={() =>
					updateEntry(entry.key, { domain: entry.suggest, suggest: null })
				}>change</a>
			</>
		);
	}
);

function domainValidationError(isValid) {
	return isValid ? "" : "Invalid domain";
}

interface DomainEntryProps {
	entry;
	onChange;
	extraProps: {
		alreadyExists: boolean;
		comment: string;
		permType: RadioFormInputHook;
	};
}

function DomainEntry({ entry, onChange, extraProps: { alreadyExists, comment, permType } }: DomainEntryProps) {
	const domainField = useTextInput("domain", {
		defaultValue: entry.domain,
		showValidation: entry.checked,
		initValidation: domainValidationError(entry.valid),
		validator: (value) => domainValidationError(isValidDomainPermission(value))
	});

	useEffect(() => {
		if (entry.valid != domainField.valid) {
			onChange({ valid: domainField.valid });
		}
	}, [onChange, entry.valid, domainField.valid]);

	useEffect(() => {
		if (entry.domain != domainField.value) {
			domainField.setter(entry.domain);
		}
		// domainField.setter is enough, eslint wants domainField
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [entry.domain, domainField.setter]);

	useEffect(() => {
		onChange({ suggest: hasBetterScope(domainField.value ?? "") });
		// only need this update if it's the entry.checked that updated, not onChange
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [domainField.value]);

	function clickIcon(e) {
		if (entry.suggest) {
			e.stopPropagation();
			e.preventDefault();
			domainField.setter(entry.suggest);
			onChange({ domain: entry.suggest, checked: true });
		}
	}

	return (
		<>
			<div className="domain-input">
				<TextInput
					field={domainField}
					onChange={(e) => {
						domainField.onChange(e);
						onChange({ domain: e.target.value, checked: true });
					}}
				/>
				<span id="icon" onClick={clickIcon}>
					<DomainEntryIcon
						alreadyExists={alreadyExists}
						suggestion={entry.suggest}
						permTypeString={permType.value?? ""}
					/>
				</span>
			</div>
			<p>{comment}</p>
		</>
	);
}

interface DomainEntryIconProps {
	alreadyExists: boolean;
	suggestion: string;
	permTypeString: string; 
}

function DomainEntryIcon({ alreadyExists, suggestion, permTypeString }: DomainEntryIconProps) {
	let icon;
	let text;

	if (suggestion) {
		icon = "fa-info-circle suggest-changes";
		text = `Entry targets a specific subdomain, consider changing it to '${suggestion}'.`;
	} else if (alreadyExists) {
		icon = "fa-history permission-already-exists";
		text = `Domain ${permTypeString} already exists.`;
	}

	if (!icon) {
		return null;
	}

	return (
		<>
			<i className={`fa fa-fw ${icon}`} aria-hidden="true" title={text}></i>
			<span className="sr-only">{text}</span>
		</>
	);
}
