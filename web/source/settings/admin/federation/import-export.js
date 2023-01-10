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
	FileInput
} = require("../../components/form/inputs");
const FormWithData = require("../../lib/form/form-with-data");

module.exports = function ImportExport() {
	return (
		<div className="import-export">
			<h2>Import / Export</h2>
			<FormWithData
				dataQuery={query.useInstanceBlocksQuery}
				DataForm={ImportExportForm}
			/>
		</div>
	);
};

function ImportExportForm({ data: blockedInstances }) {
	const form = {
		list: useTextInput("list"),
		obfuscate: useBoolInput("obfuscate"),
		commentPrivate: useTextInput("private_comment"),
		commentPublic: useTextInput("public_comment"),
		json: useFileInput("json")
	};

	return (
		<form>
			<TextArea
				field={form.list}
				label="Domains, one per line"
				placeholder={`google.com\nfacebook.com`}
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
		</form>
	);
}