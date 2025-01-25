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

import React, { useMemo } from "react";

import {
	useTextInput,
	useFileInput,
	useBoolInput,
	useFieldArrayInput,
} from "../../lib/form";

import useFormSubmit from "../../lib/form/submit";
import { useWithFormContext, FormContext } from "../../lib/form/context";

import {
	TextInput,
	TextArea,
	FileInput,
	Checkbox,
	Select
} from "../../components/form/inputs";

import FormWithData from "../../lib/form/form-with-data";
import FakeProfile from "../../components/profile";
import MutationButton from "../../components/form/mutation-button";

import { useAccountThemesQuery } from "../../lib/query/user";
import { useUpdateCredentialsMutation } from "../../lib/query/user";
import { useVerifyCredentialsQuery } from "../../lib/query/oauth";
import { useInstanceV1Query } from "../../lib/query/gts-api";
import { Account } from "../../lib/types/account";

export default function UserProfile() {
	return (
		<FormWithData
			dataQuery={useVerifyCredentialsQuery}
			DataForm={UserProfileForm}
		/>
	);
}

interface UserProfileFormProps {
	data: Account;
}

function UserProfileForm({ data: profile }: UserProfileFormProps) {
	/*
		User profile update form keys
		- bool bot
		- bool locked
		- string display_name
		- string note
		- file avatar
		- file header
		- bool enable_rss
		- bool hide_collections
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
	const themeOptions = useMemo(() => {
		let themeOptions = [
			<option key="" value="">
				Default
			</option>
		];

		themes?.forEach((theme) => {
			const value = theme.file_name;
			let text = theme.title;
			if (theme.description) {
				text += " - " + theme.description;
			}
			themeOptions.push(
				<option key={value} value={value}>
					{text}
				</option>
			);
		});

		return themeOptions;
	}, [themes]);

	const form = {
		avatar: useFileInput("avatar", { withPreview: true }),
		avatarDescription: useTextInput("avatar_description", { source: profile }),
		header: useFileInput("header", { withPreview: true }),
		headerDescription: useTextInput("header_description", { source: profile }),
		displayName: useTextInput("display_name", { source: profile }),
		note: useTextInput("note", { source: profile, valueSelector: (p) => p.source?.note }),
		bot: useBoolInput("bot", { source: profile }),
		locked: useBoolInput("locked", { source: profile }),
		discoverable: useBoolInput("discoverable", { source: profile}),
		enableRSS: useBoolInput("enable_rss", { source: profile }),
		hideCollections: useBoolInput("hide_collections", { source: profile }),
		webVisibility: useTextInput("web_visibility", { source: profile, valueSelector: (p) => p.source?.web_visibility }),
		fields: useFieldArrayInput("fields_attributes", {
			defaultValue: profile?.source?.fields,
			length: instanceConfig.maxPinnedFields
		}),
		customCSS: useTextInput("custom_css", { source: profile, nosubmit: !instanceConfig.allowCustomCSS }),
		theme: useTextInput("theme", { source: profile }),
	};

	const [submitForm, result] = useFormSubmit(form, useUpdateCredentialsMutation(), {
		changedOnly: true,
		onFinish: () => {
			form.avatar.reset();
			form.header.reset();
		}
	});

	const noAvatarSet = !profile.avatar_media_id;
	const noHeaderSet = !profile.header_media_id;

	return (
		<form className="user-profile" onSubmit={submitForm}>
			<h1>Profile</h1>
			<div className="overview">
				<FakeProfile
					avatar={form.avatar.previewValue ?? profile.avatar}
					header={form.header.previewValue ?? profile.header}
					display_name={form.displayName.value ?? profile.username}
					bot={profile.bot}
					username={profile.username}
					role={profile.role}
				/>

				<fieldset className="file-input-with-image-description">
					<legend>Header</legend>
					<FileInput
						label="Upload file"
						field={form.header}
						accept="image/png, image/jpeg, image/webp, image/gif"
					/>
					<TextInput
						field={form.headerDescription}
						label="Image description; only settable if not using default header"
						placeholder="A green field with pink flowers."
						autoCapitalize="sentences"
						disabled={noHeaderSet && !form.header.value}
					/>
				</fieldset>
				
				<fieldset className="file-input-with-image-description">
					<legend>Avatar</legend>
					<FileInput
						label="Upload file (1:1 images look best)"
						field={form.avatar}
						accept="image/png, image/jpeg, image/webp, image/gif"
					/>
					<TextInput
						field={form.avatarDescription}
						label="Image description; only settable if not using default avatar"
						placeholder="A cute drawing of a smiling sloth."
						autoCapitalize="sentences"
						disabled={noAvatarSet && !form.avatar.value}
					/>
				</fieldset>

				<div className="theme">
					<div>
						<b id="theme-label">Theme</b>
						<br/>
						<span>After choosing theme and saving, <a href={profile.url} target="_blank">open your profile</a> and refresh to see changes.</span>
					</div>
					<Select
						aria-labelledby="theme-label"
						field={form.theme}
						options={<>{themeOptions}</>}
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
			<Checkbox
				field={form.bot}
				label="Mark as bot account; this indicates to other users that this is an automated account"
			/>
			<TextInput
				field={form.displayName}
				label="Display name"
				placeholder="A GoToSocial User"
				autoCapitalize="words"
				spellCheck="false"
			/>
			<TextArea
				field={form.note}
				label="Bio"
				placeholder="Just trying out GoToSocial, my pronouns are they/them and I like sloths."
				autoCapitalize="sentences"
				rows={8}
			/>
			<fieldset>
				<legend>Profile fields</legend>
				<ProfileFields
					field={form.fields}
				/>
			</fieldset>

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
			<Select
				field={form.webVisibility}
				label="Visibility level of posts to show on your profile, and in your RSS feed (if enabled)."
				options={
					<>
						<option value="public">Show Public posts only (the GoToSocial default)</option>
						<option value="unlisted">Show Public and Unlisted posts (the Mastodon default)</option>
						<option value="none">Show no posts</option>
					</>
				}
			/>
			<Checkbox
				field={form.locked}
				label="Manually approve follow requests."
			/>
			<Checkbox
				field={form.discoverable}
				label="Mark account as discoverable by search engines and directories."
			/>
			<Checkbox
				field={form.enableRSS}
				label="Enable RSS feed of posts."
			/>
			<Checkbox
				field={form.hideCollections}
				label="Hide who you follow / are followed by."
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
				autoCapitalize="none"
				spellCheck="false"
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
				autoCapitalize="none"
				spellCheck="false"
			/>
			<TextInput
				field={form.value}
				placeholder="Value"
				autoCapitalize="none"
				spellCheck="false"
			/>
		</div>
	);
}
