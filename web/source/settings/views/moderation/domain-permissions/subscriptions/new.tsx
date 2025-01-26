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

import React, { useState } from "react";
import useFormSubmit from "../../../../lib/form/submit";
import { useCreateDomainPermissionSubscriptionMutation } from "../../../../lib/query/admin/domain-permissions/subscriptions";
import { useBoolInput, useNumberInput, useTextInput } from "../../../../lib/form";
import { urlValidator } from "../../../../lib/util/formvalidators";
import MutationButton from "../../../../components/form/mutation-button";
import { Checkbox, NumberInput, Select, TextInput } from "../../../../components/form/inputs";
import { useLocation } from "wouter";
import { DomainPermissionSubscriptionDocsLink, DomainPermissionSubscriptionHelpText } from "./common";

export default function DomainPermissionSubscriptionNew() {
	const [ _location, setLocation ] = useLocation();
	
	const useBasicAuth = useBoolInput("useBasicAuth", { defaultValue: false });
	const form = {
		priority: useNumberInput("priority", { defaultValue: 0 }),
		uri: useTextInput("uri", {
			validator: urlValidator,
		}),
		content_type: useTextInput("content_type", { defaultValue: "text/csv" }),
		permission_type: useTextInput("permission_type", { defaultValue: "block" }),
		title: useTextInput("title"),
		as_draft: useBoolInput("as_draft", { defaultValue: true }),
		adopt_orphans: useBoolInput("adopt_orphans", { defaultValue: false }),
		fetch_username: useTextInput("fetch_username", {
			nosubmit: !useBasicAuth.value
		}),
		fetch_password: useTextInput("fetch_password", {
			nosubmit: !useBasicAuth.value
		}),
	};

	const [ showPassword, setShowPassword ] = useState(false);

	const [formSubmit, result] = useFormSubmit(
		form,
		useCreateDomainPermissionSubscriptionMutation(),
		{
			changedOnly: false,
			onFinish: (res) => {
				if (res.data) {
					// Creation successful,
					// redirect to subscription detail.
					setLocation(`/subscriptions/${res.data.id}`);
				}
			},
		});

	const submitDisabled = () => {
		// URI required.
		if (!form.uri.value || !form.uri.valid) {
			return true;
		}
		
		// If no basic auth, we don't care what
		// fetch_password and fetch_username are.
		if (!useBasicAuth.value) {
			return false;
		}

		// Either of fetch_password or fetch_username must be set.
		return !(form.fetch_password.value || form.fetch_username.value);
	};

	return (
		<form
			className="domain-permission-subscription-create"
			onSubmit={formSubmit}
			// Prevent password managers
			// trying to fill in fields.
			autoComplete="off"
		>
			<div className="form-section-docs">
				<h2>New Domain Permission Subscription</h2>
				<p><DomainPermissionSubscriptionHelpText /></p>
				<DomainPermissionSubscriptionDocsLink />
			</div>

			<TextInput
				field={form.title}
				label={`Subscription title`}
				placeholder={`Some List of ${form.permission_type.value === "block" ? "Baddies" : "Goodies"}`}
				autoCapitalize="words"
				spellCheck="false"
			/>

			<NumberInput
				field={form.priority}
				label={`Subscription priority (0-255)`}
				type="number"
				min="0"
				max="255"
			/>

			<Select
				field={form.permission_type}
				label="Permission type"
				options={ 
					<>
						<option value="block">Block</option>
						<option value="allow">Allow</option>
					</>
				}
			/>

			<TextInput
				field={form.uri}
				label={`Permission list URL (http or https)`}
				placeholder="https://example.org/files/some_list_somewhere"
				autoCapitalize="none"
				spellCheck="false"
				type="url"
			/>

			<Select
				field={form.content_type}
				label="Content type"
				options={ 
					<>
						<option value="text/csv">CSV</option>
						<option value="application/json">JSON</option>
						<option value="text/plain">Plain</option>
					</>
				}
			/>

			<Checkbox
				label={
					<>
						<>Use </> 
						<a
							href="https://en.wikipedia.org/wiki/Basic_access_authentication"
							target="_blank"
							rel="noreferrer"
						>basic auth</a>
						<> when fetching</>
					</>
				}
				field={useBasicAuth}
			/>

			{ useBasicAuth.value &&
				<>
					<TextInput
						field={form.fetch_username}
						label={`Basic auth username`}
						autoCapitalize="none"
						spellCheck="false"
						autoComplete="off"
						required={useBasicAuth.value && !form.fetch_password.value}
					/>
					<div className="password-show-hide">
						<TextInput
							field={form.fetch_password}
							label={`Basic auth password`}
							autoCapitalize="none"
							spellCheck="false"
							type={showPassword ? "" : "password"}
							autoComplete="off"
							required={useBasicAuth.value && !form.fetch_username.value}
						/>
						<button
							className="password-show-hide-toggle"
							type="button"
							title={!showPassword ? "Show password" : "Hide password"}
							onClick={e => {
								e.preventDefault();
								setShowPassword(!showPassword);
							}}
						>
							{ !showPassword ? "Show" : "Hide" }
						</button>
					</div>
				</>
			}

			<Checkbox
				label="Adopt orphan permissions"
				field={form.adopt_orphans}
			/>

			<Checkbox
				label="Create permissions as drafts"
				field={form.as_draft}
			/>

			{ !form.as_draft.value && 
				<div className="info">
					<i className="fa fa-fw fa-exclamation-circle" aria-hidden="true"></i>
					<b>
						Unchecking "create permissions as drafts" means that permissions found on the
						subscribed list will be enforced immediately the next time the list is fetched.
						<br/>
						If you're subscribing to a block list, this means that blocks will be created
						automatically from the given list, potentially severing any existing follow
						relationships with accounts on the blocked domain.
						<br/>
						Before saving, make sure this is what you really want to do, and consider
						creating domain excludes for domains that you want to manage manually.
					</b>
				</div>
			}

			<MutationButton
				label="Save"
				result={result}
				disabled={submitDisabled()}
			/>
		</form>
	);
}
