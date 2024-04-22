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

import { useActionAccountMutation } from "../../../lib/query";

import MutationButton from "../../../components/form/mutation-button";

import useFormSubmit from "../../../lib/form/submit";
import {
	useValue,
	useTextInput,
	useBoolInput,
} from "../../../lib/form";

import { Checkbox, TextInput } from "../../../components/form/inputs";
import { AdminAccount } from "../../../lib/types/account";

export interface AccountActionsProps {
	account: AdminAccount,
}

export function AccountActions({ account }: AccountActionsProps) {
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
				Currently only the "suspend" action is implemented.<br/>
				Suspending an account will delete it from your server, and remove all of its media, posts, relationships, etc.<br/>
				If the suspended account is local, suspending will also send out a "delete" message to other servers, requesting them to remove its data from their instance as well.<br/>
				<b>Account suspension cannot be reversed.</b>
			</div>
			<TextInput
				field={form.reason}
				placeholder="Reason for this action"
			/>
			<div className="action-buttons">
				{/* <MutationButton
					label="Disable"
					name="disable"
					result={result}
				/>
				<MutationButton
					label="Silence"
					name="silence"
					result={result}
				/> */}
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
