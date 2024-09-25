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

import { useTextInput, useFileInput } from "../../../lib/form";
import { TextInput, TextArea, FileInput } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import { useInstanceV1Query } from "../../../lib/query/gts-api";
import { useUpdateInstanceMutation } from "../../../lib/query/admin";
import { InstanceV1 } from "../../../lib/types/instance";
import FormWithData from "../../../lib/form/form-with-data";
import useFormSubmit from "../../../lib/form/submit";

export default function InstanceSettings() {
	return (
		<FormWithData
			dataQuery={useInstanceV1Query}
			DataForm={InstanceSettingsForm}
		/>
	);
}

interface InstanceSettingsFormProps{
	data: InstanceV1;
}

function InstanceSettingsForm({ data: instance }: InstanceSettingsFormProps) {
	const titleLimit = 40;
	const shortDescLimit = 500;
	const descLimit = 5000;
	const termsLimit = 5000;

	const form = {
		title: useTextInput("title", {
			source: instance,
			validator: (val: string) => val.length <= titleLimit ? "" : `Instance title is ${val.length} characters; must be ${titleLimit} characters or less`
		}),
		thumbnail: useFileInput("thumbnail", { withPreview: true }),
		thumbnailDesc: useTextInput("thumbnail_description", { source: instance }),
		shortDesc: useTextInput("short_description", {
			source: instance,
			// Select "raw" text version of parsed field for editing.
			valueSelector: (s: InstanceV1) => s.short_description_text,
			validator: (val: string) => val.length <= shortDescLimit ? "" : `Instance short description is ${val.length} characters; must be ${shortDescLimit} characters or less`
		}),
		description: useTextInput("description", {
			source: instance,
			// Select "raw" text version of parsed field for editing.
			valueSelector: (s: InstanceV1) => s.description_text,
			validator: (val: string) => val.length <= descLimit ? "" : `Instance description is ${val.length} characters; must be ${descLimit} characters or less`
		}),
		customCSS: useTextInput("custom_css", {
			source: instance,
			valueSelector: (s: InstanceV1) => s.custom_css
		}),
		terms: useTextInput("terms", {
			source: instance,
			// Select "raw" text version of parsed field for editing.
			valueSelector: (s: InstanceV1) => s.terms_text,
			validator: (val: string) => val.length <= termsLimit ? "" : `Instance terms and conditions is ${val.length} characters; must be ${termsLimit} characters or less`
		}),
		contactUser: useTextInput("contact_username", { source: instance, valueSelector: (s) => s.contact_account?.username }),
		contactEmail: useTextInput("contact_email", { source: instance, valueSelector: (s) => s.email })
	};

	const [submitForm, result] = useFormSubmit(form, useUpdateInstanceMutation());

	return (
		<form
			onSubmit={submitForm}
			autoComplete="none"
		>
			<h1>Instance Settings</h1>

			<div className="form-section-docs">
				<h3>Appearance</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/admin/settings/#instance-appearance"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>

			<TextInput
				field={form.title}
				label={`Instance title (max ${titleLimit} characters)`}
				autoCapitalize="words"
				placeholder="My GoToSocial Instance"
			/>

			<div className="file-upload" aria-labelledby="avatar">
				<strong id="avatar">Instance avatar (1:1 images look best)</strong>
				<div className="file-upload-with-preview">
					<img
						className="preview avatar"
						src={form.thumbnail.previewValue ?? instance?.thumbnail}
						alt={form.thumbnailDesc.value ?? (instance?.thumbnail ? `Thumbnail image for the instance` : "No instance thumbnail image set")}
					/>
					<div className="file-input-with-image-description">
						<FileInput
							field={form.thumbnail}
							accept="image/png, image/jpeg, image/webp, image/gif"
						/>
						<TextInput
							field={form.thumbnailDesc}
							label="Avatar image description"
							placeholder="A cute drawing of a smiling sloth."
							autoCapitalize="sentences"
						/>
					</div>
				</div>

			</div>

			<div className="form-section-docs">
				<h3>Descriptors</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/admin/settings/#instance-descriptors"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>

			<TextArea
				field={form.shortDesc}
				label={`Short description (markdown accepted, max ${shortDescLimit} characters)`}
				placeholder="A small testing instance for GoToSocial."
				autoCapitalize="sentences"
				rows={6}
			/>

			<TextArea
				field={form.description}
				label={`Full description (markdown accepted, max ${descLimit} characters)`}
				placeholder="A small testing instance for GoToSocial. Just trying it out, my main instance is https://example.com"
				autoCapitalize="sentences"
				rows={6}
			/>

			<TextArea
				field={form.terms}
				label={`Terms & Conditions (markdown accepted, max ${termsLimit} characters)`}
				placeholder="Terms and conditions of using this instance, data policy, imprint, GDPR stuff, yadda yadda."
				autoCapitalize="sentences"
				rows={6}
			/>

			<div className="form-section-docs">
				<h3>Contact info</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/admin/settings/#instance-contact-info"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>

			<TextInput
				field={form.contactUser}
				label="Contact user (local account username)"
				placeholder="admin"
				autoCapitalize="none"
				spellCheck="false"
			/>

			<TextInput
				field={form.contactEmail}
				label="Contact email"
				placeholder="admin@example.com"
				type="email"
			/>

			<TextArea
				field={form.customCSS}
				label={"Custom CSS"}
				className="monospace"
				rows={8}
				autoCapitalize="none"
				spellCheck="false"
			/>

			<MutationButton label="Save" result={result} disabled={false} />
		</form>
	);
}
