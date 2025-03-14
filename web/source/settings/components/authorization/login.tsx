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

import { useAuthorizeFlowMutation } from "../../lib/query/login";
import { useTextInput, useValue } from "../../lib/form";
import useFormSubmit from "../../lib/form/submit";
import MutationButton from "../form/mutation-button";
import Loading from "../loading";
import { TextInput } from "../form/inputs";

export default function Login({ }) {
	const form = {
		instance: useTextInput("instance", {
			defaultValue: window.location.origin
		}),
		scopes: useValue("scopes", "read write admin"),
	};

	const [formSubmit, result] = useFormSubmit(form, useAuthorizeFlowMutation(), { 
		changedOnly: false,
	});

	if (result.isLoading) {
		return (
			<div>
				<Loading /> Checking instance.
			</div>
		);
	} else if (result.isSuccess) {
		return (
			<div>
				<Loading /> Redirecting to instance authorization page.
			</div>
		);
	}

	return (
		<form onSubmit={formSubmit}>
			<TextInput
				field={form.instance}
				label="Instance"
				name="instance"
			/>
			<MutationButton
				label="Login"
				result={result}
				disabled={false}
			/>
		</form>
	);
}