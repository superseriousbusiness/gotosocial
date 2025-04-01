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
import { useTextInput } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import { TextInput } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import { usePasswordChangeMutation } from "../../../lib/query/user";

export default function PasswordChange({ oidcEnabled }: { oidcEnabled?: boolean }) {
	const form = {
		oldPassword: useTextInput("old_password"),
		newPassword: useTextInput("new_password", {
			validator(val) {
				if (val != "" && val == form.oldPassword.value) {
					return "New password same as old password";
				}
				return "";
			}
		})
	};

	const verifyNewPassword = useTextInput("verifyNewPassword", {
		validator(val) {
			if (val != "" && val != form.newPassword.value) {
				return "Passwords do not match";
			}
			return "";
		}
	});

	const [submitForm, result] = useFormSubmit(form, usePasswordChangeMutation());

	return (
		<form className="change-password" onSubmit={submitForm}>
			<div className="form-section-docs">
				<h3>Change Password</h3>
				{ oidcEnabled && <p>
					This instance is running with OIDC as its authorization + identity provider.
					<br/>
					This means <strong>you cannot change your password using this settings panel</strong>.
					<br/>
					To change your password, you should instead contact your OIDC provider.
				</p> }
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#password-change"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about this (opens in a new tab)
				</a>
			</div>
			
			<TextInput
				type="password"
				name="password"
				field={form.oldPassword}
				label="Current password"
				autoComplete="current-password"
				disabled={oidcEnabled}
			/>
			<TextInput
				type="password"
				name="newPassword"
				field={form.newPassword}
				label="New password"
				autoComplete="new-password"
				disabled={oidcEnabled}
			/>
			<TextInput
				type="password"
				name="confirmNewPassword"
				field={verifyNewPassword}
				label="Confirm new password"
				autoComplete="new-password"
				disabled={oidcEnabled}
			/>
			<MutationButton
				label="Change password"
				result={result}
				disabled={oidcEnabled ?? false}
			/>
		</form>
	);
}
