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
import useFormSubmit from "../../../lib/form/submit";
import { useTextInput } from "../../../lib/form";
import MutationButton from "../../../components/form/mutation-button";
import { TextArea, TextInput } from "../../../components/form/inputs";
import { useLocation } from "wouter";
import { useCreateAppMutation } from "../../../lib/query/user/applications";
import { urlValidator, useScopesValidator } from "../../../lib/util/formvalidators";
import { useCallbackURL } from "./common";
import { HighlightedCode } from "../../../components/highlightedcode";

export default function NewApp() {
	const [ _location, setLocation ] = useLocation();
	const callbackURL = useCallbackURL();
	const scopesValidator = useScopesValidator();

	const form = {
		name: useTextInput("client_name"),
		redirect_uris: useTextInput("redirect_uris", {
			validator: (redirectURIs: string) => {
				if (redirectURIs === "") {
					return "";
				}

				const invalids = redirectURIs.
					split("\n").
					map(redirectURI => redirectURI === "urn:ietf:wg:oauth:2.0:oob" ? "" : urlValidator(redirectURI)).
					flatMap((invalid) => invalid || []);

				return invalids.join(", ");
			}
		}),
		scopes: useTextInput("scopes", {
			validator: (scopesStr: string) => {
				if (scopesStr === "") {
					return "";
				}
				return scopesValidator(scopesStr.split(" "));
			}
		}),
		website: useTextInput("website", {
			validator: urlValidator,
		}),
	};

	const [formSubmit, result] = useFormSubmit(
		form,
		useCreateAppMutation(),
		{
			changedOnly: false,
			onFinish: (res) => {
				if (res.data) {
					// Creation successful,
					// redirect to apps overview.
					setLocation(`/search`);
				}
			},
		});

	return (
		<form
			className="application-new"
			onSubmit={formSubmit}
			// Prevent password managers
			// trying to fill in fields.
			autoComplete="off"
		>
			<div className="form-section-docs">
				<h2>New Application</h2>
				<p>
					On this page you can create a new managed OAuth client application, with the specified redirect URIs and scopes.
					<br/>If not specified, redirect URIs defaults to <span className="monospace">urn:ietf:wg:oauth:2.0:oob</span>, and scopes defaults to <span className="monospace">read</span>.
					<br/>If you want to obtain an access token for your application here in the settings panel, include this settings panel callback URL in your redirect URIs:
					<HighlightedCode code={callbackURL} lang="url" />
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#applications"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about application redirect URIs and scopes (opens in a new tab)
				</a>
			</div>

			<TextInput
				field={form.name}
				label="Application name (required)"
				placeholder="My Cool Application"
				autoCapitalize="words"
				spellCheck="false"
				maxLength={1024}
			/>

			<TextInput
				field={form.website}
				label="Application website (optional)"
				placeholder="https://example.org/my_cool_application"
				autoCapitalize="none"
				spellCheck="false"
				type="url"
				maxLength={1024}
			/>

			<TextArea
				className="monospace"
				field={form.redirect_uris}
				label="Redirect URIs (optional, newline-separated entries)"
				placeholder={`https://example.org/my_cool_application`}
				autoCapitalize="none"
				spellCheck="false"
				rows={5}
				maxLength={2056}
			/>

			<TextInput
				className="monospace"
				field={form.scopes}
				label="Scopes (optional, space-separated entries)"
				placeholder={`read write push`}
				autoCapitalize="none"
				spellCheck="false"
				maxLength={1024}
			/>

			<MutationButton
				label="Create"
				result={result}
				disabled={!form.name.value}
			/>
		</form>
	);
}
