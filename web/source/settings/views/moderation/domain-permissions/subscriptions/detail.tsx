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

import React, { ReactNode, useState } from "react";
import { useLocation, useParams } from "wouter";
import { useBaseUrl } from "../../../../lib/navigation/util";
import BackButton from "../../../../components/back-button";
import { useGetDomainPermissionSubscriptionQuery, useRemoveDomainPermissionSubscriptionMutation, useTestDomainPermissionSubscriptionMutation, useUpdateDomainPermissionSubscriptionMutation } from "../../../../lib/query/admin/domain-permissions/subscriptions";
import { useBoolInput, useNumberInput, useTextInput } from "../../../../lib/form";
import FormWithData from "../../../../lib/form/form-with-data";
import { DomainPerm, DomainPermSub } from "../../../../lib/types/domain-permission";
import MutationButton from "../../../../components/form/mutation-button";
import { Checkbox, NumberInput, Select, TextInput } from "../../../../components/form/inputs";
import useFormSubmit from "../../../../lib/form/submit";
import UsernameLozenge from "../../../../components/username-lozenge";
import { urlValidator } from "../../../../lib/util/formvalidators";
import { PageableList } from "../../../../components/pageable-list";

export default function DomainPermissionSubscriptionDetail() {
	const params = useParams();
	let id = params.permSubId as string | undefined;
	if (!id) {
		throw "no permSub ID";
	}
	
	return (
		<FormWithData
			dataQuery={useGetDomainPermissionSubscriptionQuery}
			queryArg={id}
			DataForm={DomainPermSubForm}
		/>
	);
}

function DomainPermSubForm({ data: permSub }: { data: DomainPermSub }) {
	const baseUrl = useBaseUrl();
	const backLocation: string = history.state?.backLocation ?? `~${baseUrl}/subscriptions/search`;
	
	return (
		<div className="domain-permission-subscription-details">
			<h1><BackButton to={backLocation} /> Domain Permission Subscription Detail</h1>
			<DomainPermSubDetails permSub={permSub} />
			<UpdateDomainPermSub permSub={permSub} />
			<TestDomainPermSub permSub={permSub} />
			<DeleteDomainPermSub permSub={permSub} backLocation={backLocation} />
		</div>
	);
}

function DomainPermSubDetails({ permSub }: { permSub: DomainPermSub }) {
	const [ location ] = useLocation();
	const baseUrl = useBaseUrl();
	
	const permType = permSub.permission_type;
	if (!permType) {
		throw "permission_type was undefined";
	}

	const created = new Date(permSub.created_at).toDateString();
	let fetchedAtStr = "never";
	if (permSub.fetched_at) {
		fetchedAtStr = new Date(permSub.fetched_at).toDateString();
	}

	let successfullyFetchedAtStr = "never";
	if (permSub.successfully_fetched_at) {
		successfullyFetchedAtStr = new Date(permSub.successfully_fetched_at).toDateString();
	}
	
	return (
		<dl className="info-list">
			<div className="info-list-entry">
				<dt>Permission type:</dt>
				<dd className={`permission-type ${permType}`}>
					<i
						aria-hidden={true}
						className={`fa fa-${permType === "allow" ? "check" : "close"}`}
					></i>
					{permType}
				</dd>
			</div>
			<div className="info-list-entry">
				<dt>ID</dt>
				<dd className="monospace">{permSub.id}</dd>
			</div>
			<div className="info-list-entry">
				<dt>Created</dt>
				<dd><time dateTime={permSub.created_at}>{created}</time></dd>
			</div>
			<div className="info-list-entry">
				<dt>Created By</dt>
				<dd>
					<UsernameLozenge
						account={permSub.created_by}
						linkTo={`~/settings/moderation/accounts/${permSub.created_by}`}
						backLocation={`~${baseUrl}${location}`}
					/>
				</dd>
			</div>
			<div className="info-list-entry">
				<dt>Last fetch attempt:</dt>
				<dd>{fetchedAtStr}</dd>
			</div>
			<div className="info-list-entry">
				<dt>Last successful fetch:</dt>
				<dd>{successfullyFetchedAtStr}</dd>
			</div>
			<div className="info-list-entry">
				<dt>Discovered {permSub.permission_type}s:</dt>
				<dd>{permSub.count}</dd>
			</div>
		</dl>
	);
}

function UpdateDomainPermSub({ permSub }: { permSub: DomainPermSub }) {
	const [ showPassword, setShowPassword ] = useState(false);
	const form = {
		priority: useNumberInput("priority", { source: permSub }),
		uri: useTextInput("uri", {
			source: permSub,
			validator: urlValidator,
		}),
		content_type: useTextInput("content_type", { source: permSub }),
		title: useTextInput("title", { source: permSub }),
		remove_retracted: useBoolInput("remove_retracted", { source: permSub }),
		as_draft: useBoolInput("as_draft", { source: permSub }),
		adopt_orphans: useBoolInput("adopt_orphans", { source: permSub }),
		useBasicAuth: useBoolInput("useBasicAuth", {
			defaultValue: 
				(permSub.fetch_password !== undefined && permSub.fetch_password !== "") ||
				(permSub.fetch_username !== undefined && permSub.fetch_username !== ""),
			nosubmit: true
		}),
		fetch_username: useTextInput("fetch_username", {
			source: permSub
		}),
		fetch_password: useTextInput("fetch_password", {
			source: permSub
		}),
	};

	const [submitUpdate, updateResult] = useFormSubmit(
		form,
		useUpdateDomainPermissionSubscriptionMutation(),
		{
			changedOnly: true,
			customizeMutationArgs: (mutationData) => {
				// Clear username + password if they were set,
				// but user has selected to not use basic auth.
				if (!form.useBasicAuth.value) {
					if (permSub.fetch_username !== undefined && permSub.fetch_username !== "") {
						mutationData["fetch_username"] = "";
					}
					if (permSub.fetch_password !== undefined && permSub.fetch_password !== "") {
						mutationData["fetch_password"] = "";
					}
				}

				// Remove useBasicAuth if included.
				delete mutationData["useBasicAuth"];

				// Modify mutation argument to
				// include ID and permission type.
				return {
					id: permSub.id,
					permType: permSub.permission_type,
					formData: mutationData,
				};
			},
			onFinish: res => {
				// On a successful response that returns data,
				// clear the fetch_username and fetch_password
				// fields if they weren't set on the returned sub.
				if (res.data) {
					if (res.data.fetch_username === undefined || res.data.fetch_username === "") {
						form.fetch_username.setter("");
					}
					if (res.data.fetch_password === undefined || res.data.fetch_password === "") {
						form.fetch_password.setter("");
					}
				}
			}
		}
	);

	const submitDisabled = () => {
		// If no basic auth, we don't care what
		// fetch_password and fetch_username are.
		if (!form.useBasicAuth.value) {
			return false;
		}

		// Either of fetch_password or fetch_username must be set.
		return !(form.fetch_password.value || form.fetch_username.value);
	};

	return (
		<form
			className="domain-permission-subscription-update"
			onSubmit={submitUpdate}
			// Prevent password managers
			// trying to fill in fields.
			autoComplete="off"
		>
			<h2>Edit Subscription</h2>
			<TextInput
				field={form.title}
				label={`Subscription title`}
				placeholder={`Some List of ${permSub.permission_type === "block" ? "Baddies" : "Goodies"}`}
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
				field={form.useBasicAuth}
			/>

			{ form.useBasicAuth.value &&
				<>
					<TextInput
						field={form.fetch_username}
						label={`Basic auth username`}
						autoCapitalize="none"
						spellCheck="false"
						autoComplete="off"
						required={form.useBasicAuth.value && !form.fetch_password.value}
					/>
					<div className="password-show-hide">
						<TextInput
							field={form.fetch_password}
							label={`Basic auth password`}
							autoCapitalize="none"
							spellCheck="false"
							type={showPassword ? "" : "password"}
							autoComplete="off"
							required={form.useBasicAuth.value && !form.fetch_username.value}
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
				label="Remove retracted permissions"
				field={form.remove_retracted}
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
				result={updateResult}
				disabled={submitDisabled()}
			/>

		</form>
	);
}

function DeleteDomainPermSub({ permSub, backLocation }: { permSub: DomainPermSub, backLocation: string }) {
	const permType = permSub.permission_type;
	if (!permType) {
		throw "permission_type was undefined";
	}
	
	const [_location, setLocation] = useLocation();
	const [ removeSub, result ] = useRemoveDomainPermissionSubscriptionMutation();
	const removeChildren = useBoolInput("remove_children", { defaultValue: false });

	return (
		<form className="domain-permission-subscription-remove">
			<h2>Remove Subscription</h2>
			
			<Checkbox
				label={`Also remove any ${permType}s created by this subscription`}
				field={removeChildren}
			/>

			<MutationButton
				label={`Remove`}
				title={`Remove`}
				type="button"
				className="button danger"
				onClick={(e) => {
					e.preventDefault();
					const id = permSub.id;
					const remove_children = removeChildren.value as boolean;
					removeSub({ id, remove_children }).then(res => {
						if ("data" in res) {
							setLocation(backLocation);
						}
					});
				}}
				disabled={false}
				showError={true}
				result={result}
			/>
		</form>
	);
}

function TestDomainPermSub({ permSub }: { permSub: DomainPermSub }) {
	const permType = permSub.permission_type;
	if (!permType) {
		throw "permission_type was undefined";
	}
	
	const [ testSub, testRes ] = useTestDomainPermissionSubscriptionMutation();
	const onSubmit = (e) => {
		e.preventDefault();
		testSub(permSub.id);
	};

	// Function to map an item to a list entry.
	function itemToEntry(perm: DomainPerm): ReactNode {
		return (
			<span className="text-cutoff entry perm-preview">
				<strong>{ perm.domain }</strong>
				{ perm.public_comment && <>({ perm.public_comment })</> }
			</span>
		);
	}

	return (
		<>
			<form
				className="domain-permission-subscription-test"
				onSubmit={onSubmit}
			>
				<h2>Test Subscription</h2>
				Click the "test" button to instruct your instance to do a test
				fetch and parse of the {permType} list at the subscription URI.
				<br/>
				If the fetch is successful, you will see a list of {permType}s
				(or {permType} drafts) that *would* be created by this subscription,
				along with the public comment for each {permType} (if applicable).
				<br/>
				The test does not actually create those {permType}s in your database.
				<MutationButton
					disabled={false}
					label={"Test"}
					result={testRes}
				/>
			</form>
			{ testRes.data && "error" in testRes.data
				? <div className="info perm-issue">
					<i className="fa fa-fw fa-exclamation-circle" aria-hidden="true"></i>
					<b>
						The following issue was encountered when doing a fetch + parse:
						<br/><code>{ testRes.data.error }</code>
						<br/>This may be due to a temporary outage at the remote URL,
						or you may wish to check your subscription settings and test again.
					</b>
				</div>
				: <>
					{ testRes.data && `${testRes.data?.length} ${permType}s would be created by this subscription:`}
					<PageableList
						isLoading={testRes.isLoading}
						isSuccess={testRes.isSuccess}
						items={testRes.data}
						itemToEntry={itemToEntry}
						isError={testRes.isError}
						error={testRes.error}
						emptyMessage={<b>No entries!</b>}
					/>
				</>
			}
		</>
	);
}
