/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

const query = require("../../lib/query");
const processDomainList = require("../../lib/import-export");

const {
	useTextInput,
	useBoolInput,
	useFileInput
} = require("../../lib/form");

const useFormSubmit = require("../../lib/form/submit");

const {
	TextInput,
	TextArea,
	Checkbox,
	FileInput,
	useCheckListInput
} = require("../../components/form/inputs");

const FormWithData = require("../../lib/form/form-with-data");
const CheckList = require("../../components/check-list");
const MutationButton = require("../../components/form/mutation-button");

module.exports = function ImportExport() {
	const [parsedList, setParsedList] = React.useState();

	const form = {
		domains: useTextInput("domains"),
		obfuscate: useBoolInput("obfuscate"),
		commentPrivate: useTextInput("private_comment"),
		commentPublic: useTextInput("public_comment"),
		// json: useFileInput("json")
	};

	function submitImport(e) {
		e.preventDefault();

		Promise.try(() => {
			return processDomainList(form.domains.value);
		}).then((processed) => {
			setParsedList(processed);
		}).catch((e) => {
			console.error(e);
		});
	}

	return (
		<div className="import-export">
			<h2>Import / Export</h2>
			<div>
				{
					parsedList
						? <ImportExportList list={parsedList} />
						: <ImportExportForm form={form} submitImport={submitImport} />
				}
			</div>
		</div>
	);
};

function ImportExportList({ list }) {
	const entryCheckList = useCheckListInput("selectedDomains", {
		entries: list,
		uniqueKey: "domain"
	});

	return (
		<CheckList
		/>
	);
}

function ImportExportForm({ form, submitImport }) {
	return (
		<form onSubmit={submitImport}>
			<TextArea
				field={form.domains}
				label="Domains, one per line (plaintext) or JSON"
				placeholder={`google.com\nfacebook.com`}
				rows={8}
			/>

			<TextArea
				field={form.commentPrivate}
				label="Private comment"
				rows={3}
			/>

			<TextArea
				field={form.commentPublic}
				label="Public comment"
				rows={3}
			/>

			<Checkbox
				field={form.obfuscate}
				label="Obfuscate domain in public lists"
			/>

			<div>
				<MutationButton label="Import" result={importResult} /> {/* default form action */}
			</div>
		</form>
	);
}