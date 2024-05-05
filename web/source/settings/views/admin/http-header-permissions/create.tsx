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
import { usePostHeaderAllowMutation, usePostHeaderBlockMutation } from "../../../lib/query/admin/http-header-permissions";
import { useTextInput } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import { TextInput } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import { PermType } from "../../../lib/types/perm";

export default function HeaderPermCreateForm({ permType }: { permType: PermType }) {
	const form = {
		header: useTextInput("header", {
			validator: (val: string) => {
				// Technically invalid but avoid
				// showing red outline when user
				// hasn't entered anything yet.
				if (val.length === 0) {
					return "";
				}

				// Only requirement is that header
				// must be less than 1024 chars.
				if (val.length > 1024) {
					return "header must be less than 1024 characters";
				}

				return "";
			}
		}),
		regex: useTextInput("regex", {
			validator: (val: string) => {
				// Technically invalid but avoid
				// showing red outline when user
				// hasn't entered anything yet.
				if (val.length === 0) {
					return "";
				}

				// Ensure regex compiles.
				try {
					new RegExp(val);
				} catch (e) {
					return e;
				}

				return "";
			}
		}),
	};

	// Use appropriate mutation for given permType.
	const [ postAllowTrigger, postAllowResult ] = usePostHeaderAllowMutation();
	const [ postBlockTrigger, postBlockResult ] = usePostHeaderBlockMutation();

	let mutationTrigger;
	let mutationResult;

	if (permType === "block") {
		mutationTrigger = postBlockTrigger;
		mutationResult = postBlockResult;
	} else {
		mutationTrigger = postAllowTrigger;
		mutationResult = postAllowResult;
	}

	const [formSubmit, result] = useFormSubmit(
		form,
		[mutationTrigger, mutationResult],
		{
			changedOnly: false,
			onFinish: ({ _data }) => {
				form.header.reset();
				form.regex.reset();
			},
		});

	return (
		<form onSubmit={formSubmit}>
			<h2>Create new HTTP header {permType}</h2>
			<TextInput
				field={form.header}
				label={
					<>
						HTTP Header Name&nbsp;
						 <a
							href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers"
							target="_blank"
							className="docslink"
							rel="noreferrer"
						>
							Learn more about HTTP request headers (opens in a new tab)
						</a>
					</>
				}
				placeholder={"User-Agent"}
			/>
			<TextInput
				field={form.regex}
				label={
					<>
						HTTP Header Value Regex&nbsp;
						<a
							href="https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_expressions"
							target="_blank"
							className="docslink"
							rel="noreferrer"
						>
							Learn more about regular expressions (opens in a new tab)
						</a>
					</>
				}
				placeholder={"^.*Some-User-Agent.*$"}
				{...{className: "monospace"}}
			/>
			<MutationButton
				label="Save"
				result={result}
				disabled={
					(!form.header.value || !form.regex.value) ||
					(!form.header.valid || !form.regex.valid)
				}
			/>
		</form>
	);
}
