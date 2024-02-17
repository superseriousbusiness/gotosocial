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

import FormWithData from "../lib/form/form-with-data";

import { useVerifyCredentialsQuery } from "../lib/query/oauth";
import { useArrayInput, useTextInput } from "../lib/form";
import { TextInput } from "../components/form/inputs";
import useFormSubmit from "../lib/form/submit";
import MutationButton from "../components/form/mutation-button";
import { useAliasAccountMutation, useMoveAccountMutation } from "../lib/query/user";
import { FormContext, useWithFormContext } from "../lib/form/context";
import { store } from "../redux/store";

export default function UserMigration() {
	return (
		<FormWithData
			dataQuery={useVerifyCredentialsQuery}
			DataForm={UserMigrationForm}
		/>
	);
}

function UserMigrationForm({ data: profile }) {
	let urlStr = store.getState().oauth.instanceUrl ?? "";
	let url = new URL(urlStr);

	return (
		<>
			<h2>Account Migration Settings</h2>
			<div className="info">
				<i className="fa fa-fw fa-exclamation-triangle" aria-hidden="true"></i>
				<b>Moving your account to another instance isn't implemented yet, <a href="https://github.com/superseriousbusiness/gotosocial/issues/130" target="_blank" rel="noopener noreferrer">see here for progress</a></b>
			</div>
			<p>
				The following settings allow you to <strong>alias</strong> your account to another account
				elsewhere, and to <strong>move</strong> your followers and following lists to another account.
			</p>
			<p>
				Account <strong>aliasing</strong> is harmless and reversible; you can
				set and unset up to five account aliases as many times as you wish.
			</p>
			<p>
				The account <strong>move</strong> action, on the other hand, has serious and irreversible consequences.
			</p>
			<p>
				To move, you must set an alias from your account to the target account, using this settings panel.
			</p>
			<p>
				You must also set an alias from the target account back to your account, using
				the settings panel of the instance on which the target account resides.
			</p>
			<p>
				Provide the following details to the other instance: 
			</p>
			<dl className="migration-details">
				<div>
					<dt>Account handle/username:</dt>
					<dd>@{profile.acct}@{url.host}</dd>
				</div>
				<div>
					<dt>Account URI:</dt>
					<dd>{urlStr}/users/{profile.username}</dd>
				</div>
			</dl>
			<p>
				For more information on account migration, please see <a href="https://docs.gotosocial.org/en/latest/user_guide/settings/#alias-account" target="_blank" className="docslink" rel="noreferrer">the documentation</a>.
			</p>
			<AliasForm data={profile} />
			<MoveForm data={profile} />
		</>
	);
}

function AliasForm({ data: profile }) {
	const form = {
		alsoKnownAs: useArrayInput("also_known_as_uris", {
			source: profile,
			valueSelector: (p) => (
				p.source?.also_known_as_uris
					? p.source?.also_known_as_uris.map(entry => [entry])
					: []
			),
			length: 5,
		}),
	};

	const [submitForm, result] = useFormSubmit(form, useAliasAccountMutation());
	
	return (
		<form className="user-migration-alias" onSubmit={submitForm}>
			<div className="form-section-docs without-border">
				<h3>Alias Account</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#alias-account"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about account aliasing (opens in a new tab)
				</a>
			</div>
			<AlsoKnownAsURIs
				field={form.alsoKnownAs}
			/>
			<MutationButton
				disabled={false}
				label="Save account aliases"
				result={result}
			/>
		</form>
	);
}

function AlsoKnownAsURIs({ field: formField }) {	
	return (
		<div className="aliases">
			<FormContext.Provider value={formField.ctx}>
				{formField.value.map((data, i) => (
					<AlsoKnownAsURI
						key={i}
						index={i}
						data={data}
					/>
				))}
			</FormContext.Provider>
		</div>
	);
}

function AlsoKnownAsURI({ index, data }) {	
	const name = `${index}`;
	const form = useWithFormContext(index, {
		alsoKnownAsURI: useTextInput(
			name,
			// Only one field per entry.
			{ defaultValue: data[0] ?? "" },
		),
	}); 

	return (
		<TextInput
			label={`Alias #${index+1}`}
			field={form.alsoKnownAsURI}
			placeholder={`https://example.org/users/my_other_account_${index+1}`}
		/>
	);
}

function MoveForm({ data: profile }) {
	const form = {
		movedToURI: useTextInput("moved_to_uri", {
			source: profile,
			valueSelector: (p) => p.moved?.uri },
		),
		password: useTextInput("password"),
	};

	const [submitForm, result] = useFormSubmit(form, useMoveAccountMutation());
	
	return (
		<form className="user-migration-move" onSubmit={submitForm}>
			<div className="form-section-docs without-border">
				<h3>Move Account</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#move-account"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about moving your account (opens in a new tab)
				</a>
			</div>
			<TextInput
				disabled={true}
				field={form.movedToURI}
				label="Move target URI"
				placeholder="https://example.org/users/my_new_account"
			/>
			<TextInput
				disabled={true}
				type="password"
				name="password"
				field={form.password}
				label="Confirm account password"
			/>
			<MutationButton
				disabled={true}
				label="Confirm account move"
				result={result}
			/>
		</form>
	);
}
