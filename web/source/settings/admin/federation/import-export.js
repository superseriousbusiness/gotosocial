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
const isValidDomain = require("is-valid-domain");
const FormWithData = require("../../lib/form/form-with-data");
const { Error } = require("../../components/error");

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
				<h2>Import / Export suspended domains</h2>

				<div>
					<form onSubmit={submitParse}>
						<TextArea
							field={form.domains}
							label="Domains, one per line (plaintext) or JSON"
							placeholder={`google.com\nfacebook.com`}
							rows={8}
						/>

						<div className="row">
							<MutationButton label="Import" result={parseResult} showError={false} />
							<button type="button" className="with-padding">
								<label>
									Import file
									<input className="hidden" type="file" onChange={fileChanged} accept="application/json,text/plain" />
								</label>
							</button>
						</div>
					</form>
					<form onSubmit={submitExport}>
						<div className="row">
							<MutationButton name="export" label="Export" result={exportResult} showError={false} />
							<MutationButton name="export-file" label="Export file" result={exportResult} showError={false} />
							<Select
								field={form.exportType}
								options={<>
									<option value="plain">Text</option>
									<option value="json">JSON</option>
								</>}
							/>
						</div>
					</form>
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
	let commentName = "";
	if (showComment.value == "public_comment") { commentName = "Public comment"; }
	if (showComment.value == "private_comment") { commentName = "Private comment"; }

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

				<CheckList
					field={form.domains}
					Component={DomainEntry}
					header={
						<>
							<b>Domain</b>
							<b></b>
							<b>{commentName}</b>
						</>
					}
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

function DomainEntry({ entry, onChange, blockedInstances, commentType }) {
	const domainField = useTextInput("domain", {
		defaultValue: entry.domain,
		validator: (value) => {
			return (entry.checked && !isValidDomain(value, { wildcard: true, allowUnicode: true }))
				? "Invalid domain"
				: "";
		}
	});

	React.useEffect(() => {
		onChange({ valid: domainField.valid });
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [domainField.valid]);

	let icon = null;

	if (blockedInstances[domainField.value] != undefined) {
		icon = (
			<>
				<i className="fa fa-history already-blocked" aria-hidden="true" title="Domain block already exists"></i>
				<span className="sr-only">Domain block already exists.</span>
			</>
		);
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
			<span id="icon">{icon}</span>
			<p>{entry[commentType]}</p>
		</>
	);
}