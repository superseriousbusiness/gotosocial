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
	useBoolInput
} = require("../lib/form");

const useFormSubmit = require("../lib/form/submit");

const {
	Select,
	TextInput,
	Checkbox
} = require("../components/form/inputs");

const FormWithData = require("../lib/form/form-with-data");
const Languages = require("../components/languages");
const MutationButton = require("../components/form/mutation-button");

module.exports = function UserSettings() {
	return (
		<FormWithData
			dataQuery={query.useVerifyCredentialsQuery}
			DataForm={UserSettingsForm}
		/>
	);
};

function UserSettingsForm({ data }) {
	const { source } = data;
	/* form keys
		- string source[privacy]
		- bool source[sensitive]
		- string source[language]
		- string source[status_format]
	 */

	const form = {
		defaultPrivacy: useTextInput("source[privacy]", { defaultValue: source.privacy ?? "unlisted" }),
		isSensitive: useBoolInput("source[sensitive]", { defaultValue: source.sensitive }),
		language: useTextInput("source[language]", { defaultValue: source.language?.toUpperCase() ?? "EN" }),
		format: useTextInput("source[status_format]", { defaultValue: source.status_format ?? "plain" }),
	};

	const [submitForm, result] = useFormSubmit(form, query.useUpdateCredentialsMutation());

	return (
		<>
			<form className="user-settings" onSubmit={submitForm}>
				<h1>Post settings</h1>
				<Select field={form.language} label="Default post language" options={
					<Languages />
				}>
				</Select>
				<Select field={form.defaultPrivacy} label="Default post privacy" options={
					<>
						<option value="private">Private / followers-only</option>
						<option value="unlisted">Unlisted</option>
						<option value="public">Public</option>
					</>
				}>
					<a href="https://docs.gotosocial.org/en/latest/user_guide/posts/#privacy-settings" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about post privacy settings (opens in a new tab)</a>
				</Select>
				<Select field={form.format} label="Default post (and bio) format" options={
					<>
						<option value="plain">Plain (default)</option>
						<option value="markdown">Markdown</option>
					</>
				}>
					<a href="https://docs.gotosocial.org/en/latest/user_guide/posts/#input-types" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about post format settings (opens in a new tab)</a>
				</Select>
				<Checkbox
					field={form.isSensitive}
					label="Mark my posts as sensitive by default"
				/>

				<MutationButton label="Save settings" result={result} />
			</form>
			<div>
				<PasswordChange />
			</div>
		</>
	);
}

function PasswordChange() {
	const form = {
		oldPassword: useTextInput("old_password"),
		newPassword: useTextInput("new_password", {
			validator(val) {
				if (val != "" && val == form.oldPassword.value) {
					return "New password same as old password";
				}
				return "";
			}
		})
	};

	const verifyNewPassword = useTextInput("verifyNewPassword", {
		validator(val) {
			if (val != "" && val != form.newPassword.value) {
				return "Passwords do not match";
			}
			return "";
		}
	});

	const [submitForm, result] = useFormSubmit(form, query.usePasswordChangeMutation());

	return (
		<form className="change-password" onSubmit={submitForm}>
			<h1>Change password</h1>
			<TextInput type="password" field={form.oldPassword} label="Current password" />
			<TextInput type="password" field={form.newPassword} label="New password" />
			<TextInput type="password" field={verifyNewPassword} label="Confirm new password" />
			<MutationButton label="Change password" result={result} />
		</form>
	);
}