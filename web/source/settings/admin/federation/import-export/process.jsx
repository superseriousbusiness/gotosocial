/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

"use strict";

const React = require("react");

const query = require("../../../lib/query");
const { isValidDomainBlock, hasBetterScope } = require("../../../lib/domain-block");

const {
	useTextInput,
	useBoolInput,
	useRadioInput,
	useCheckListInput
} = require("../../../lib/form");

const useFormSubmit = require("../../../lib/form/submit");

const {
	TextInput,
	TextArea,
	Checkbox,
	Select,
	RadioGroup
} = require("../../../components/form/inputs");

const CheckList = require("../../../components/check-list");
const MutationButton = require("../../../components/form/mutation-button");
const FormWithData = require("../../../lib/form/form-with-data");

module.exports = React.memo(
	function ProcessImport({ list }) {
		return (
			<div className="without-border">
				<FormWithData
					dataQuery={query.useInstanceBlocksQuery}
					DataForm={ImportList}
					list={list}
				/>
			</div>
		);
	}
);

function ImportList({ list, data: blockedInstances }) {
	const hasComment = React.useMemo(() => {
		let hasPublic = false;
		let hasPrivate = false;

		list.some((entry) => {
			if (entry.public_comment?.length > 0) {
				hasPublic = true;
			}

			if (entry.private_comment?.length > 0) {
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
		domains: useCheckListInput("domains", { entries: list }),
		obfuscate: useBoolInput("obfuscate"),
		privateComment: useTextInput("private_comment", {
			defaultValue: `Imported on ${new Date().toLocaleString()}`
		}),
		privateCommentBehavior: useRadioInput("private_comment_behavior", {
			defaultValue: "append",
			options: {
				append: "Append to",
				replace: "Replace"
			}
		}),
		publicComment: useTextInput("public_comment"),
		publicCommentBehavior: useRadioInput("public_comment_behavior", {
			defaultValue: "append",
			options: {
				append: "Append to",
				replace: "Replace"
			}
		}),
	};

	const [importDomains, importResult] = useFormSubmit(form, query.useImportDomainListMutation(), { changedOnly: false });

	return (
		<>
			<form onSubmit={importDomains} className="suspend-import-list">
				<span>{list.length} domain{list.length != 1 ? "s" : ""} in this list</span>

				{hasComment.both &&
					<Select field={showComment} options={
						<>
							<option value="public_comment">Show public comments</option>
							<option value="private_comment">Show private comments</option>
						</>
					} />
				}

				<DomainCheckList
					field={form.domains}
					blockedInstances={blockedInstances}
					commentType={showComment.value}
				/>

				<TextArea
					field={form.privateComment}
					label="Private comment"
					rows={3}
				/>
				<RadioGroup
					field={form.privateCommentBehavior}
					label="imported private comment"
				/>

				<TextArea
					field={form.publicComment}
					label="Public comment"
					rows={3}
				/>
				<RadioGroup
					field={form.publicCommentBehavior}
					label="imported public comment"
				/>

				<Checkbox
					field={form.obfuscate}
					label="Obfuscate domains in public lists"
				/>

				<MutationButton label="Import" result={importResult} />
			</form>
		</>
	);
}

function DomainCheckList({ field, blockedInstances, commentType }) {
	const getExtraProps = React.useCallback((entry) => {
		return {
			comment: entry[commentType],
			alreadyExists: blockedInstances[entry.domain] != undefined
		};
	}, [blockedInstances, commentType]);

	const entriesWithSuggestions = React.useMemo(() => (
		Object.values(field.value).filter((entry) => entry.suggest)
	), [field.value]);

	return (
		<>
			<CheckList
				field={field}
				header={<>
					<b>Domain</b>
					<b></b>
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

const UpdateHint = React.memo(
	function UpdateHint({ entries, updateEntry, updateMultiple }) {
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

const UpdateableEntry = React.memo(
	function UpdateableEntry({ entry, updateEntry }) {
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

function DomainEntry({ entry, onChange, extraProps: { alreadyExists, comment } }) {
	const domainField = useTextInput("domain", {
		defaultValue: entry.domain,
		showValidation: entry.checked,
		initValidation: domainValidationError(entry.valid),
		validator: (value) => domainValidationError(isValidDomainBlock(value))
	});

	React.useEffect(() => {
		if (entry.valid != domainField.valid) {
			onChange({ valid: domainField.valid });
		}
	}, [onChange, entry.valid, domainField.valid]);

	React.useEffect(() => {
		if (entry.domain != domainField.value) {
			domainField.setter(entry.domain);
		}
		// domainField.setter is enough, eslint wants domainField
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [entry.domain, domainField.setter]);

	React.useEffect(() => {
		onChange({ suggest: hasBetterScope(domainField.value) });
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
			<TextInput
				field={domainField}
				onChange={(e) => {
					domainField.onChange(e);
					onChange({ domain: e.target.value, checked: true });
				}}
			/>
			<span id="icon" onClick={clickIcon}>
				<DomainEntryIcon alreadyExists={alreadyExists} suggestion={entry.suggest} onChange={onChange} />
			</span>
			<p>{comment}</p>
		</>
	);
}

function DomainEntryIcon({ alreadyExists, suggestion }) {
	let icon;
	let text;

	if (suggestion) {
		icon = "fa-info-circle suggest-changes";
		text = `Entry targets a specific subdomain, consider changing it to '${suggestion}'.`;
	} else if (alreadyExists) {
		icon = "fa-history already-blocked";
		text = "Domain block already exists.";
	}

	if (!icon) {
		return null;
	}

	return (
		<>
			<i className={`fa ${icon}`} aria-hidden="true" title={text}></i>
			<span className="sr-only">{text}</span>
		</>
	);
}