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
import { useTextInput } from "../../lib/form";
import useFormSubmit from "../../lib/form/submit";
import { TextInput } from "../../components/form/inputs";
import MutationButton from "../../components/form/mutation-button";
import { useEmailChangeMutation, usePasswordChangeMutation, useUserQuery } from "../../lib/query/user";
import Loading from "../../components/loading";
import { User } from "../../lib/types/user";
import { useInstanceV1Query } from "../../lib/query/gts-api";

export default function EmailPassword() {
	return (
		<>
			<h1>Email & Password Settings</h1>
			<EmailChange />
			<PasswordChange />
		</>
	);
}

function PasswordChange() {
	// Load instance data.
	const {
		data: instance,
		isFetching: isFetchingInstance,
		isLoading: isLoadingInstance
	} = useInstanceV1Query();
	if (isFetchingInstance || isLoadingInstance) {
		return <Loading />;
	}

	if (instance === undefined) {
		throw "could not fetch instance";
	}

	return <PasswordChangeForm oidcEnabled={instance.configuration.oidc_enabled} />;
}

function PasswordChangeForm({ oidcEnabled }: { oidcEnabled?: boolean }) {
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

function EmailChange() {
	// Load instance data.
	const {
		data: instance,
		isFetching: isFetchingInstance,
		isLoading: isLoadingInstance
	} = useInstanceV1Query();
	
	// Load user data.
	const {
		data: user,
		isFetching: isFetchingUser,
		isLoading: isLoadingUser
	} = useUserQuery();

	if (
		(isFetchingInstance || isLoadingInstance) ||
		(isFetchingUser || isLoadingUser)
	) {
		return <Loading />;
	}

	if (user === undefined) {
		throw "could not fetch user";
	}

	if (instance === undefined) {
		throw "could not fetch instance";
	}

	return <EmailChangeForm user={user} oidcEnabled={instance.configuration.oidc_enabled} />;
}

function EmailChangeForm({user, oidcEnabled}: { user: User, oidcEnabled?: boolean }) {
	const form = {
		currentEmail: useTextInput("current_email", {
			defaultValue: user.email,
			nosubmit: true
		}),
		newEmail: useTextInput("new_email", {
			validator: (value: string | undefined) => {
				if (!value) {
					return "";
				}

				if (value.toLowerCase() === user.email?.toLowerCase()) {
					return "cannot change to your existing address";
				}

				if (value.toLowerCase() === user.unconfirmed_email?.toLowerCase()) {
					return "you already have a pending email address change to this address";
				}

				return "";
			},
		}),
		password: useTextInput("password"),
	};
	const [submitForm, result] = useFormSubmit(form, useEmailChangeMutation());

	return (
		<form className="change-email" onSubmit={submitForm}>
			<div className="form-section-docs">
				<h3>Change Email</h3>
				{ oidcEnabled && <p>
					This instance is running with OIDC as its authorization + identity provider.
					<br/>
					You can still change your email address using this settings panel,
					but it will only affect which address GoToSocial uses to contact you,
					not the email address you use to log in.
					<br/>
					To change the email address you use to log in, contact your OIDC provider.
				</p> }
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#email-change"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about this (opens in a new tab)
				</a>
			</div>

			{ (user.unconfirmed_email && user.unconfirmed_email !== user.email) && <>
				<div className="info">
					<i className="fa fa-fw fa-info-circle" aria-hidden="true"></i>
					<b>
						You currently have a pending email address
						change to the address: {user.unconfirmed_email}
						<br />
						To confirm {user.unconfirmed_email} as your new
						address for this account, please check your email inbox.
					</b>
				</div>
			</> }

			<TextInput
				type="email"
				name="current-email"
				field={form.currentEmail}
				label="Current email address"
				autoComplete="none"
				disabled={true}
			/>

			<TextInput
				type="password"
				name="password"
				field={form.password}
				label="Current password"
				autoComplete="current-password"
			/>

			<TextInput
				type="email"
				name="new-email"
				field={form.newEmail}
				label="New email address"
				autoComplete="none"
			/>
			
			<MutationButton
				disabled={!form.password || !form.newEmail || !form.newEmail.valid}
				label="Change email address"
				result={result}
			/>
		</form>
	);
}