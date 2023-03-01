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
	useFileInput,
	useBoolInput
} = require("../lib/form");

const useFormSubmit = require("../lib/form/submit");

const {
	TextInput,
	TextArea,
	FileInput,
	Checkbox
} = require("../components/form/inputs");

const FormWithData = require("../lib/form/form-with-data");
const FakeProfile = require("../components/fake-profile");
const MutationButton = require("../components/form/mutation-button");

module.exports = function UserProfile() {
	return (
		<FormWithData
			dataQuery={query.useVerifyCredentialsQuery}
			DataForm={UserProfileForm}
		/>
	);
};

function UserProfileForm({ data: profile }) {
	/*
		User profile update form keys
		- bool bot
		- bool locked
		- string display_name
		- string note
		- file avatar
		- file header
		- bool enable_rss
		- string custom_css (if enabled)
	*/

	const { data: instance } = query.useInstanceQuery();
	const allowCustomCSS = React.useMemo(() => {
		return instance?.configuration?.accounts?.allow_custom_css === true;
	}, [instance]);

	const form = {
		avatar: useFileInput("avatar", { withPreview: true }),
		header: useFileInput("header", { withPreview: true }),
		displayName: useTextInput("display_name", { source: profile }),
		note: useTextInput("note", { source: profile, valueSelector: (p) => p.source?.note }),
		customCSS: useTextInput("custom_css", { source: profile }),
		bot: useBoolInput("bot", { source: profile }),
		locked: useBoolInput("locked", { source: profile }),
		enableRSS: useBoolInput("enable_rss", { source: profile }),
	};

	const [submitForm, result] = useFormSubmit(form, query.useUpdateCredentialsMutation());

	return (
		<form className="user-profile" onSubmit={submitForm}>
			<h1>Profile</h1>
			<div className="overview">
				<FakeProfile
					avatar={form.avatar.previewValue ?? profile.avatar}
					header={form.header.previewValue ?? profile.header}
					display_name={form.displayName.value ?? profile.username}
					username={profile.username}
					role={profile.role.name}
				/>
				<div className="files">
					<div>
						<h3>Header</h3>
						<FileInput
							field={form.header}
							accept="image/*"
						/>
					</div>
					<div>
						<h3>Avatar</h3>
						<FileInput
							field={form.avatar}
							accept="image/*"
						/>
					</div>
				</div>
			</div>
			<TextInput
				field={form.displayName}
				label="Name"
				placeholder="A GoToSocial user"
			/>
			<TextArea
				field={form.note}
				label="Bio"
				placeholder="Just trying out GoToSocial, my pronouns are they/them and I like sloths."
				rows={8}
			/>
			<Checkbox
				field={form.locked}
				label="Manually approve follow requests"
			/>
			<Checkbox
				field={form.enableRSS}
				label="Enable RSS feed of Public posts"
			/>
			{!allowCustomCSS ? null :
				<TextArea
					field={form.customCSS}
					label="Custom CSS"
					className="monospace"
					rows={8}
				>
					<a href="https://docs.gotosocial.org/en/latest/user_guide/custom_css" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about custom profile CSS (opens in a new tab)</a>
				</TextArea>
			}
			<MutationButton label="Save profile info" result={result} />
		</form>
	);
}