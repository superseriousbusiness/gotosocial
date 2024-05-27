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
import { TextInput } from "../../../../components/form/inputs";
import MutationButton from "../../../../components/form/mutation-button";
import { useTextInput } from "../../../../lib/form";
import { useSendTestEmailMutation } from "../../../../lib/query/admin/actions";
import { useInstanceV1Query } from "../../../../lib/query/gts-api";
import useFormSubmit from "../../../../lib/form/submit";

export default function Test({}) {
	const { data: instance } = useInstanceV1Query();

	const form = {
		email: useTextInput("email", { defaultValue: instance?.email }),
		message: useTextInput("message")
	};

	const [submit, result] = useFormSubmit(form, useSendTestEmailMutation(), { changedOnly: false });

	return (
		<form onSubmit={submit}>
			<div className="form-section-docs">
				<h2>Send test email</h2>
				<p>
					To check whether your instance email configuration is correct, you can
					try sending a test email to the given address, with an optional message.
					<br/>
					If you do not have SMTP configured for your instance, this will do nothing.
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/configuration/smtp/"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about SMTP configuration (opens in a new tab)
				</a>
			</div>
			<TextInput
				field={form.email}
				label="Email"
				placeholder="someone@example.org"
				// Get email validation for free.
				type="email"
				required={true}
			/>
			<TextInput
				field={form.message}
				label="Message (optional)"
				placeholder="Please disregard this test email, thanks!"
			/>
			<MutationButton
				disabled={!form.email.value}
				label="Send"
				result={result}
			/>
		</form>
	);
}
