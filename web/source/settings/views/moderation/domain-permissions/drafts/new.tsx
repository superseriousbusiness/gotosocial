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
import useFormSubmit from "../../../../lib/form/submit";
import { useCreateDomainPermissionDraftMutation } from "../../../../lib/query/admin/domain-permissions/drafts";
import { useBoolInput, useRadioInput, useTextInput } from "../../../../lib/form";
import { formDomainValidator } from "../../../../lib/util/formvalidators";
import MutationButton from "../../../../components/form/mutation-button";
import { Checkbox, RadioGroup, TextArea, TextInput } from "../../../../components/form/inputs";
import { useLocation } from "wouter";
import { DomainPermissionDraftDocsLink, DomainPermissionDraftHelpText } from "./common";

export default function DomainPermissionDraftNew() {
	const [ _location, setLocation ] = useLocation();
	
	const form = {
		domain: useTextInput("domain", {
			validator: formDomainValidator,
		}),
		permission_type: useRadioInput("permission_type", { 
			options: {
				block: "Block domain",
				allow: "Allow domain",
			}
		}),
		obfuscate: useBoolInput("obfuscate"),
		public_comment: useTextInput("public_comment"),
		private_comment: useTextInput("private_comment"),
	};
		
	const [formSubmit, result] = useFormSubmit(
		form,
		useCreateDomainPermissionDraftMutation(),
		{
			changedOnly: false,
			onFinish: (res) => {
				if (res.data) {
					// Creation successful,
					// redirect to drafts overview.
					setLocation(`/drafts/search`);
				}
			},
		});

	return (
		<form
			onSubmit={formSubmit}
			// Prevent password managers
			// trying to fill in fields.
			autoComplete="off"
		>
			<div className="form-section-docs">
				<h2>New Domain Permission Draft</h2>
				<p><DomainPermissionDraftHelpText /></p>
				<DomainPermissionDraftDocsLink />
			</div>

			<RadioGroup
				field={form.permission_type}
			/>

			<TextInput
				field={form.domain}
				label={`Domain (without "https://" prefix)`}
				placeholder="example.org"
				autoCapitalize="none"
				spellCheck="false"
			/>

			<TextArea
				field={form.private_comment}
				label={"Private comment (will be shown to admins only)"}
				placeholder="This domain is like unto a clown car full of clowns, I suggest we block it forthwith."
				autoCapitalize="sentences"
				rows={3}
			/>

			<TextArea
				field={form.public_comment}
				label={"Public comment (will be shown to members of this instance via the instance info page, and on the web if enabled)"}
				placeholder="Bad posters"
				autoCapitalize="sentences"
				rows={3}
			/>

			<Checkbox
				field={form.obfuscate}
				label="Obfuscate domain in public lists"
			/>

			<MutationButton
				label="Save"
				result={result}
				disabled={
					!form.domain.value ||
					!form.domain.valid ||
					!form.permission_type.value
				}
			/>
		</form>
	);
}
