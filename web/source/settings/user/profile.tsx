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

import {
	useTextInput,
	useFileInput,
	useBoolInput,
	useFieldArrayInput,
	useRadioInput
} from "../lib/form";

import useFormSubmit from "../lib/form/submit";
import { useWithFormContext, FormContext } from "../lib/form/context";

import {
	TextInput,
	TextArea,
	FileInput,
	Checkbox,
	RadioGroup
} from "../components/form/inputs";

import FormWithData from "../lib/form/form-with-data";
import FakeProfile from "../components/fake-profile";
import MutationButton from "../components/form/mutation-button";

import { useAccountThemesQuery, useInstanceV1Query } from "../lib/query";
import { useUpdateCredentialsMutation } from "../lib/query/user";
import { useVerifyCredentialsQuery } from "../lib/query/oauth";

export default function UserProfile() {
	return (
		<FormWithData
			dataQuery={useVerifyCredentialsQuery}
			DataForm={UserProfileForm}
		/>
	);
}

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
		- string theme
	*/

	const { data: instance } = useInstanceV1Query();
	const instanceConfig = React.useMemo(() => {
		return {
			allowCustomCSS: instance?.configuration?.accounts?.allow_custom_css === true,
			maxPinnedFields: instance?.configuration?.accounts?.max_profile_fields ?? 6
		};
	}, [instance]);
	
	// Parse out available theme options into nice format.
	const { data: themes } = useAccountThemesQuery();
	let themeOptions = { "": "Default" };
	themes?.forEach((theme) => {
		let key = theme.file_name;
		let value = theme.title;
		if (theme.description) {
			value += " - " + theme.description;
		}
		themeOptions[key] = value;
	});

	const form = {
		avatar: useFileInput("avatar", { withPreview: true }),
		header: useFileInput("header", { withPreview: true }),
		displayName: useTextInput("display_name", { source: profile }),
		note: useTextInput("note", { source: profile, valueSelector: (p) => p.source?.note }),
		bot: useBoolInput("bot", { source: profile }),
		locked: useBoolInput("locked", { source: profile }),
		discoverable: useBoolInput("discoverable", { source: profile}),
		enableRSS: useBoolInput("enable_rss", { source: profile }),
		fields: useFieldArrayInput("fields_attributes", {
			defaultValue: profile?.source?.fields,
			length: instanceConfig.maxPinnedFields
		}),
		customCSS: useTextInput("custom_css", { source: profile, nosubmit: !instanceConfig.allowCustomCSS }),
		theme: useRadioInput("theme", {
			source: profile,
			options: themeOptions,
		}),
	};

	const [submitForm, result] = useFormSubmit(form, useUpdateCredentialsMutation(), {
		changedOnly: true,
		onFinish: () => {
			form.avatar.reset();
			form.header.reset();
		}
	});

	return (
		<form className="user-profile" onSubmit={submitForm}>
			<h1>Profile</h1>
			<div className="overview">
				<FakeProfile
					avatar={form.avatar.previewValue ?? profile.avatar}
					header={form.header.previewValue ?? profile.header}
					display_name={form.displayName.value ?? profile.username}
					username={profile.username}
					role={profile.role}
				/>
				<div className="files">
					<div>
						<FileInput
							label="Header"
							field={form.header}
							accept="image/*"
						/>
					</div>
					<div>
						<FileInput
							label="Avatar"
							field={form.avatar}
							accept="image/*"
						/>
					</div>
				</div>

				<div className="theme">
					<div>
						<b id="theme-label">Theme</b>
						<br/>
						<span>After choosing theme and saving, <a href={profile.url} target="_blank">open your profile</a> and refresh to see changes.</span>
					</div>
					<RadioGroup
						aria-labelledby="theme-label"
						field={form.theme}
					/>
				</div>
			</div>

			<div className="form-section-docs">
				<h3>Basic Information</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#basic-information"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<TextInput
				field={form.displayName}
				label="Display name"
				placeholder="A GoToSocial user"
			/>
			<TextArea
				field={form.note}
				label="Bio"
				placeholder="Just trying out GoToSocial, my pronouns are they/them and I like sloths."
				rows={8}
			/>
			<b>Profile fields</b>
			<ProfileFields
				field={form.fields}
			/>

			<div className="form-section-docs">
				<h3>Visibility and privacy</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#visibility-and-privacy"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<Checkbox
				field={form.locked}
				label="Manually approve follow requests"
			/>
			<Checkbox
				field={form.discoverable}
				label="Mark account as discoverable by search engines and directories"
			/>
			<Checkbox
				field={form.enableRSS}
				label="Enable RSS feed of Public posts"
			/>

			<div className="form-section-docs">
				<h3>Advanced</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#advanced"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<TextArea
				field={form.customCSS}
				label={`Custom CSS` + (!instanceConfig.allowCustomCSS ? ` (not enabled on this instance)` : ``)}
				className="monospace"
				rows={8}
				disabled={!instanceConfig.allowCustomCSS}
			/>
			<MutationButton
				disabled={false}
				label="Save profile info"
				result={result}
			/>
		</form>
	);
}

function ProfileFields({ field: formField }) {
	return (
		<div className="fields">
			<FormContext.Provider value={formField.ctx}>
				{formField.value.map((data, i) => (
					<Field
						key={i}
						index={i}
						data={data}
					/>
				))}
			</FormContext.Provider>
		</div>
	);
}

function Field({ index, data }) {
	const form = useWithFormContext(index, {
		name: useTextInput("name", { defaultValue: data.name }),
		value: useTextInput("value", { defaultValue: data.value })
	});

	return (
		<div className="entry">
			<TextInput
				field={form.name}
				placeholder="Name"
			/>
			<TextInput
				field={form.value}
				placeholder="Value"
			/>
		</div>
	);
}