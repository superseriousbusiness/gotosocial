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
const { Switch, Route, Redirect, useLocation } = require("wouter");

const query = require("../../lib/query");
const { isValidDomainBlock } = require("../../lib/domain-block");

const {
	useTextInput,
	useBoolInput,
	useRadioInput,
	useCheckListInput
} = require("../../lib/form");

const useFormSubmit = require("../../lib/form/submit");

const {
	TextInput,
	TextArea,
	Checkbox,
	Select,
	RadioGroup
} = require("../../components/form/inputs");

const CheckList = require("../../components/check-list");
const MutationButton = require("../../components/form/mutation-button");
const FormWithData = require("../../lib/form/form-with-data");
const { Error } = require("../../components/error");
const ExportFormatTable = require("./export-format-table");

const baseUrl = "/settings/admin/federation/import-export";

module.exports = function ImportExport() {
	const [updateFromFile, setUpdateFromFile] = React.useState(false);
	const form = {
		domains: useTextInput("domains"),
		exportType: useTextInput("exportType", { defaultValue: "plain", dontReset: true })
	};

	const [submitParse, parseResult] = useFormSubmit(form, query.useProcessDomainListMutation());
	const [submitExport, exportResult] = useFormSubmit(form, query.useExportDomainListMutation());

	function fileChanged(e) {
		const reader = new FileReader();
		reader.onload = function (read) {
			form.domains.setter(read.target.result);
			setUpdateFromFile(true);
		};
		reader.readAsText(e.target.files[0]);
	}

	React.useEffect(() => {
		if (exportResult.isSuccess) {
			form.domains.setter(exportResult.data);
		}
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [exportResult]);

	const [_location, setLocation] = useLocation();

	if (updateFromFile) {
		setUpdateFromFile(false);
		submitParse();
	}

	return (
		<Switch>
			<Route path={`${baseUrl}/list`}>
				{!parseResult.isSuccess && <Redirect to={baseUrl} />}

				<h1>
					<span className="button" onClick={() => {
						parseResult.reset();
						setLocation(baseUrl);
					}}>
						&lt; back
					</span> Confirm import:
				</h1>
				<FormWithData
					dataQuery={query.useInstanceBlocksQuery}
					DataForm={ImportList}
					list={parseResult.data}
				/>
			</Route>

			<Route>
				{parseResult.isSuccess && <Redirect to={`${baseUrl}/list`} />}
				<h1>Import / Export suspended domains</h1>
				<p>
					This page can be used to import and export lists of domains to suspend.
					Exports can be done in various formats, with varying functionality and support in other software.
					Imports will automatically detect what format is being processed.
				</p>
				<ExportFormatTable />
				<div className="import-export">
					<TextArea
						field={form.domains}
						label="Domains"
						placeholder={`google.com\nfacebook.com`}
						rows={8}
					/>

					<div className="button-grid">
						<MutationButton
							label="Import"
							type="button"
							onClick={() => submitParse()}
							result={parseResult}
							showError={false}
						/>
						<button type="button" className="with-padding">
							<label>
								Import file
								<input
									type="file"
									className="hidden"
									onChange={fileChanged}
									accept="application/json,text/plain,text/csv"
								/>
							</label>
						</button>
						<b /> {/* grid filler */}
						<MutationButton
							label="Export"
							type="button"
							onClick={() => submitExport("export")}
							result={exportResult} showError={false}
						/>
						<MutationButton label="Export to file" type="button" onClick={() => submitExport("export-file")} result={exportResult} showError={false} />
						<div className="export-file">
							<span>
								as
							</span>
							<Select
								field={form.exportType}
								options={<>
									<option value="plain">Text</option>
									<option value="json">JSON</option>
									<option value="csv">CSV</option>
								</>}
							/>
						</div>
					</div>

					{parseResult.error && <Error error={parseResult.error} />}
					{exportResult.error && <Error error={exportResult.error} />}
				</div>
			</Route>
		</Switch>
	);
};

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
		domains: useCheckListInput("domains", {
			entries: list,
			uniqueKey: "domain"
		}),
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

	return (
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
	);
}

function domainValidationError(isValid) {
	return isValid ? "" : "Invalid domain";
}

function DomainEntry({ entry, onChange, extraProps: { alreadyExists, comment } }) {
	const domainField = useTextInput("domain", {
		defaultValue: entry.domain,
		initValidation: domainValidationError(entry.valid),
		validator: (value) => domainValidationError(
			!entry.checked || isValidDomainBlock(value)
		)
	});

	// FIXME: actually update suggestion when needed
	const suggest = (entry.suggest != entry.domain) && entry.suggest;

	React.useEffect(() => {
		if (entry.valid != domainField.valid) {
			onChange({ valid: domainField.valid });
		}
	}, [onChange, entry.valid, domainField.valid]);

	React.useEffect(() => {
		domainField.validate();
		// only need this update if it's the entry.checked that updated, not domainField
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [entry.checked]);

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
				<DomainEntryIcon alreadyExists={alreadyExists} suggestion={suggest} onChange={onChange} />
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