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
import { useTextInput, useBoolInput } from "../../../../lib/form";
import useFormSubmit from "../../../../lib/form/submit";
import { Select, Checkbox } from "../../../../components/form/inputs";
import Languages from "../../../../components/languages";
import MutationButton from "../../../../components/form/mutation-button";
import { useUpdateCredentialsMutation } from "../../../../lib/query/user";
import { Account } from "../../../../lib/types/account";

export default function BasicSettings({ account }: { account: Account }) {
	/* form keys
		- string source[privacy]
		- bool source[sensitive]
		- string source[language]
		- string source[status_content_type]
	 */
	const form = {
		defaultPrivacy: useTextInput("source[privacy]", { source: account, defaultValue: "unlisted" }),
		isSensitive: useBoolInput("source[sensitive]", { source: account }),
		language: useTextInput("source[language]", { source: account, valueSelector: (s: Account) => s.source?.language?.toUpperCase() ?? "EN" }),
		statusContentType: useTextInput("source[status_content_type]", { source: account, defaultValue: "text/plain" }),
	};
	
	const [submitForm, result] = useFormSubmit(form, useUpdateCredentialsMutation());
	
	return (
		<form className="post-settings" onSubmit={submitForm}>
			<div className="form-section-docs">
				<h3>Basic</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings#post-settings"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
				Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<Select field={form.language} label="Default post language" options={
				<Languages />
			}>
			</Select>
			<Select field={form.defaultPrivacy} label="Default post privacy" options={
				<>
					<option value="public">Public</option>
					<option value="unlisted">Unlisted</option>
					<option value="private">Followers-only</option>
				</>
			}>
			</Select>
			<Select field={form.statusContentType} label="Default post (and bio) format" options={
				<>
					<option value="text/plain">Plain (default)</option>
					<option value="text/markdown">Markdown</option>
				</>
			}>
			</Select>
			<Checkbox
				field={form.isSensitive}
				label="Mark my posts as sensitive by default"
			/>
			<MutationButton
				disabled={false}
				label="Save settings"
				result={result}
			/>
		</form>
	);
}