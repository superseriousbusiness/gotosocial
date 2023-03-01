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

const query = require("../lib/query");

const {
	useTextInput,
	useFileInput
} = require("../lib/form");

const useFormSubmit = require("../lib/form/submit");

const {
	TextInput,
	TextArea,
	FileInput
} = require("../components/form/inputs");

const FormWithData = require("../lib/form/form-with-data");
const MutationButton = require("../components/form/mutation-button");

module.exports = function AdminSettings() {
	return (
		<FormWithData
			dataQuery={query.useInstanceQuery}
			DataForm={AdminSettingsForm}
		/>
	);
};

function AdminSettingsForm({ data: instance }) {
	const form = {
		title: useTextInput("title", {
			source: instance,
			validator: (val) => val.length <= 40 ? "" : "Instance title must be 40 characters or less"
		}),
		thumbnail: useFileInput("thumbnail", { withPreview: true }),
		thumbnailDesc: useTextInput("thumbnail_description", { source: instance }),
		shortDesc: useTextInput("short_description", { source: instance }),
		description: useTextInput("description", { source: instance }),
		contactUser: useTextInput("contact_username", { source: instance, valueSelector: (s) => s.contact_account?.username }),
		contactEmail: useTextInput("contact_email", { source: instance, valueSelector: (s) => s.email }),
		terms: useTextInput("terms", { source: instance })
	};

	const [submitForm, result] = useFormSubmit(form, query.useUpdateInstanceMutation());

	return (
		<form onSubmit={submitForm}>
			<h1>Instance Settings</h1>
			<TextInput
				field={form.title}
				label="Title"
				placeholder="My GoToSocial instance"
			/>

			<div className="file-upload">
				<h3>Instance thumbnail</h3>
				<div>
					<img className="preview avatar" src={form.thumbnail.previewValue ?? instance.thumbnail} alt={form.thumbnailDesc.value ?? (instance.thumbnail ? `Thumbnail image for the instance` : "No instance thumbnail image set")} />
					<FileInput
						field={form.thumbnail}
						accept="image/*"
					/>
				</div>
			</div>

			<TextInput
				field={form.thumbnailDesc}
				label="Instance thumbnail description"
				placeholder="A cute drawing of a smiling sloth."
			/>

			<TextArea
				field={form.shortDesc}
				label="Short description"
				placeholder="A small testing instance for the GoToSocial alpha software."
			/>

			<TextArea
				field={form.description}
				label="Full description"
				placeholder="A small testing instance for the GoToSocial alpha software. Just trying it out, my main instance is https://example.com"
			/>

			<TextInput
				field={form.contactUser}
				label="Contact user (local account username)"
				placeholder="admin"
			/>

			<TextInput
				field={form.contactEmail}
				label="Contact email"
				placeholder="admin@example.com"
			/>

			<TextArea
				field={form.terms}
				label="Terms & Conditions"
				placeholder=""
			/>

			<MutationButton label="Save" result={result} />
		</form>
	);
}