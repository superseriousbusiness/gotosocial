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

import { useActionAccountMutation, useHandleSignupMutation } from "../../../../lib/query/admin";
import MutationButton from "../../../../components/form/mutation-button";
import useFormSubmit from "../../../../lib/form/submit";
import {
	useValue,
	useTextInput,
	useBoolInput,
} from "../../../../lib/form";
import { Checkbox, Select, TextInput } from "../../../../components/form/inputs";
import { AdminAccount } from "../../../../lib/types/account";
import { useLocation } from "wouter";

export interface AccountActionsProps {
	account: AdminAccount,
	backLocation: string,
}

export function AccountActions({ account, backLocation }: AccountActionsProps) {
	const local = !account.domain;
	
	// Available actions differ depending
	// on the account's current status.
	switch (true) {
		case account.suspended:
			// Can't do anything with
			// suspended accounts currently.
			return null;
		case local && !account.approved:
			// Unapproved local account sign-up,
			// only show HandleSignup form.
			return (
				<HandleSignup
					account={account}
					backLocation={backLocation}
				/>
			);
		default:
			// Normal local or remote account, show
			// full range of moderation options.
			return <ModerateAccount account={account} />;
	}
}

function ModerateAccount({ account }: { account: AdminAccount }) {
	const form = {
		id: useValue("id", account.id),
		reason: useTextInput("text")
	};
	
	const reallySuspend = useBoolInput("reallySuspend");
	const [accountAction, result] = useFormSubmit(form, useActionAccountMutation());
	
	return (
		<form
			onSubmit={accountAction}
			aria-labelledby="account-moderation-actions"
		>
			<h3 id="account-moderation-actions">Account Moderation Actions</h3>
			<div>
				Currently only the "suspend" action is implemented.
				<br/>
				Suspending an account will delete it from your server,
				and remove all of its media, posts, relationships, etc.
				<br/>
				If the suspended account is local, suspending will also
				send out a "delete" message to other servers, requesting
				them to remove its data from their instance as well.
				<br/>
				<b>Account suspension cannot be reversed.</b>
			</div>
			<TextInput
				field={form.reason}
				placeholder="Reason for this action"
				autoCapitalize="sentences"
			/>
			<div className="action-buttons">
				<MutationButton
					disabled={account.suspended || reallySuspend.value === undefined || reallySuspend.value === false}
					label="Suspend"
					name="suspend"
					result={result}
				/>
				<Checkbox
					label="Really suspend"
					field={reallySuspend}
				></Checkbox>
			</div>
		</form>
	);
}

function HandleSignup({ account, backLocation }: { account: AdminAccount, backLocation: string }) {
	const form = {
		id: useValue("id", account.id),
		approveOrReject: useTextInput("approve_or_reject", { defaultValue: "approve" }),
		privateComment: useTextInput("private_comment"),
		message: useTextInput("message"),
		sendEmail: useBoolInput("send_email"),
	};

	const [_location, setLocation] = useLocation();

	const [handleSignup, result] = useFormSubmit(form, useHandleSignupMutation(), {
		changedOnly: false,
		// After submitting the form, redirect back to
		// /settings/admin/accounts if rejecting, since
		// account will no longer be available at
		// /settings/admin/accounts/:accountID endpoint.
		onFinish: (res) => {			
			if (form.approveOrReject.value === "approve") {
				// An approve request:
				// stay on this page and
				// serve updated details.
				return;
			}

			if (res.data) {
				// "reject" successful,
				// redirect to accounts page.
				setLocation(backLocation);
			}
		}
	});

	return (
		<form
			onSubmit={handleSignup}
			aria-labelledby="account-handle-signup"
		>
			<h3 id="account-handle-signup">Handle Account Sign-Up</h3>
			<Select
				field={form.approveOrReject}
				label="Approve or Reject"
				options={
					<>
						<option value="approve">Approve</option>
						<option value="reject">Reject</option>
					</>
				}
			>
			</Select>
			{ form.approveOrReject.value === "reject" &&
			// Only show form fields relevant
			// to "reject" if rejecting.
			// On "approve" these fields will
			// be ignored anyway.
			<>
				<TextInput
					field={form.privateComment}
					label="(Optional) private comment on why sign-up was rejected (shown to other admins only)"
				/>
				<Checkbox
					field={form.sendEmail}
					label="Send email to applicant"
				/>
				<TextInput
					field={form.message}
					label={"(Optional) message to include in email to applicant, if send email is checked"}
				/>
			</> }
			<MutationButton
				disabled={false}
				label={form.approveOrReject.value === "approve" ? "Approve" : "Reject"}
				result={result}
			/>
		</form>
	);
}
