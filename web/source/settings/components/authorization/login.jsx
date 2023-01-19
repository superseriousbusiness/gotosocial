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

const query = require("../../lib/query");
const { useTextInput, useValue } = require("../../lib/form");
const useFormSubmit = require("../../lib/form/submit");
const { TextInput } = require("../form/inputs");
const MutationButton = require("../form/mutation-button");
const Loading = require("../loading");

module.exports = function Login({ }) {
	const form = {
		instance: useTextInput("instance", {
			defaultValue: window.location.origin
		}),
		scopes: useValue("scopes", "user admin")
	};

	const [formSubmit, result] = useFormSubmit(
		form,
		query.useAuthorizeFlowMutation(),
		{ changedOnly: false }
	);

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
			<MutationButton label="Login" result={result} />
		</form>
	);
};