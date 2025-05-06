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

import React, { forwardRef, useCallback, useMemo, useRef } from "react";
import {
	useDefaultInteractionPoliciesQuery,
	useResetDefaultInteractionPoliciesMutation,
	useUpdateDefaultInteractionPoliciesMutation,
} from "../../../../lib/query/user";
import Loading from "../../../../components/loading";
import { Error } from "../../../../components/error";
import MutationButton from "../../../../components/form/mutation-button";
import {
	DefaultInteractionPolicies,
	InteractionPolicy,
	InteractionPolicyEntry,
	InteractionPolicyValue,
	PolicyValueAuthor,
	PolicyValueFollowers,
	PolicyValueFollowing,
	PolicyValueMentioned,
	PolicyValuePublic,
} from "../../../../lib/types/interaction";
import { useTextInput } from "../../../../lib/form";
import { Select } from "../../../../components/form/inputs";
import { TextFormInputHook } from "../../../../lib/form/types";
import { useBasicFor } from "./basic";
import { PolicyFormSomethingElse, useSomethingElseFor } from "./something-else";
import { Action, PolicyFormSub, SomethingElseValue, Visibility } from "./types";

export default function InteractionPolicySettings() {
	const {
		data: defaultPolicies,
		isLoading,
		isFetching,
		isError,
		error,
	} = useDefaultInteractionPoliciesQuery();

	if (isLoading || isFetching) {
		return <Loading />;
	}

	if (isError) {
		return <Error error={error} />;
	}

	if (!defaultPolicies) {
		throw "default policies undefined";
	}

	return (
		<InteractionPoliciesForm defaultPolicies={defaultPolicies} />
	);
}

interface InteractionPoliciesFormProps {
	defaultPolicies: DefaultInteractionPolicies;
}

function InteractionPoliciesForm({ defaultPolicies }: InteractionPoliciesFormProps) {
	// Sub-form for visibility "public".
	const formPublic = useFormForVis(defaultPolicies.public, "public");
	const assemblePublic = useCallback(() => {
		return {
			can_favourite: assemblePolicyEntry("public", "favourite", formPublic),
			can_reply: assemblePolicyEntry("public", "reply", formPublic),
			can_reblog: assemblePolicyEntry("public", "reblog", formPublic),
		};
	}, [formPublic]);
	
	// Sub-form for visibility "unlisted".
	const formUnlisted = useFormForVis(defaultPolicies.unlisted, "unlisted");
	const assembleUnlisted = useCallback(() => {
		return {
			can_favourite: assemblePolicyEntry("unlisted", "favourite", formUnlisted),
			can_reply: assemblePolicyEntry("unlisted", "reply", formUnlisted),
			can_reblog: assemblePolicyEntry("unlisted", "reblog", formUnlisted),
		};
	}, [formUnlisted]);
	
	// Sub-form for visibility "private".
	const formPrivate = useFormForVis(defaultPolicies.private, "private");
	const assemblePrivate = useCallback(() => {
		return {
			can_favourite: assemblePolicyEntry("private", "favourite", formPrivate),
			can_reply: assemblePolicyEntry("private", "reply", formPrivate),
			can_reblog: assemblePolicyEntry("private", "reblog", formPrivate),
		};
	}, [formPrivate]);

	const selectedVis = useTextInput("selectedVis", { defaultValue: "public" });
	
	const [updatePolicies, updateResult] = useUpdateDefaultInteractionPoliciesMutation();
	const [resetPolicies, resetResult] = useResetDefaultInteractionPoliciesMutation();

	const onSubmit = (e) => {
		e.preventDefault();
		updatePolicies({
			public: assemblePublic(),
			unlisted: assembleUnlisted(),
			private: assemblePrivate(),
			// Always use the
			// default for direct.
			direct: null,
		});
	};

	return (
		<form className="interaction-default-settings" onSubmit={onSubmit}>
			<div className="form-section-docs">
				<h3>Default Interaction Policies</h3>
				<p>
					You can use this section to customize the default interaction
					policy for posts created by you, per visibility setting.
					<br/>
					These settings apply only for new posts created by you <em>after</em> applying
					these settings; they do not apply retroactively.
					<br/>
					The word "anyone" in the below options means <em>anyone with
					permission to see the post</em>, taking account of blocks.
					<br/>
					Bear in mind that no matter what you set below, you will always
					be able to like, reply-to, and boost your own posts.
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings#default-interaction-policies"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<div className="tabbable-sections">
				<PolicyPanelsTablist selectedVis={selectedVis} />
				<PolicyPanel
					policyForm={formPublic}
					forVis={"public"}
					isActive={selectedVis.value === "public"}
				/>
				<PolicyPanel
					policyForm={formUnlisted}
					forVis={"unlisted"}
					isActive={selectedVis.value === "unlisted"}
				/>
				<PolicyPanel
					policyForm={formPrivate}
					forVis={"private"}
					isActive={selectedVis.value === "private"}
				/>
			</div>

			<div className="action-buttons row">
				<MutationButton
					disabled={false}
					label="Save policies"
					result={updateResult}
				/>

				<MutationButton
					disabled={false}
					type="button"
					onClick={() => resetPolicies()}
					label="Reset to defaults"
					result={resetResult}
					className="button danger"
					showError={false}
				/>
			</div>

		</form>
	);
}

// A tablist of tab buttons, one for each visibility.
function PolicyPanelsTablist({ selectedVis }: { selectedVis: TextFormInputHook}) {
	const publicRef = useRef<HTMLButtonElement>(null);
	const unlistedRef = useRef<HTMLButtonElement>(null);
	const privateRef = useRef<HTMLButtonElement>(null);
	
	return (
		<div className="tab-buttons" role="tablist">
			<Tab
				label="Public"
				selectedVis={selectedVis}
				prevVis="private"
				thisVis="public"
				nextVis="unlisted"
				prevRef={privateRef}
				thisRef={publicRef}
				nextRef={unlistedRef}
			/>
			<Tab
				label="Unlisted"
				selectedVis={selectedVis}
				prevVis="public"
				thisVis="unlisted"
				nextVis="private"
				prevRef={publicRef}
				thisRef={unlistedRef}
				nextRef={privateRef}
			/>
			<Tab
				label="Followers-only"
				selectedVis={selectedVis}
				prevVis="unlisted"
				thisVis="private"
				nextVis="public"
				prevRef={unlistedRef}
				thisRef={privateRef}
				nextRef={publicRef}
			/>
		</div>
	);
}

interface TabProps {
	label: string;
	selectedVis: TextFormInputHook;
	prevVis: string;
	thisVis: string;
	nextVis: string;
	prevRef: React.RefObject<HTMLButtonElement>;
	thisRef: React.RefObject<HTMLButtonElement>;
	nextRef: React.RefObject<HTMLButtonElement>;
}

// One tab in a tablist, corresponding to the given thisVisibility.
const Tab = forwardRef(
	function Tab({
		label,
		selectedVis,
		prevVis,
		thisVis,
		nextVis,
		prevRef,
		thisRef,
		nextRef,
	}: TabProps) {
		const selected = useMemo(() => {
			return selectedVis.value === thisVis;
		}, [selectedVis, thisVis]);

		return (
			<button
				id={`tab-${thisVis}`}
				title={label}
				role="tab"
				ref={thisRef}
				className={`tab-button ${selected && "active"}`}
				onClick={(e) => {
					// Allow tab to be clicked.
					e.preventDefault();
					selectedVis.setter(thisVis);
				}}
				onKeyDown={(e) => {
					// Allow cycling through
					// tabs with arrow keys.
					if (e.key === "ArrowLeft") {
						// Select and set
						// focus on previous tab.
						selectedVis.setter(prevVis);
						prevRef.current?.focus();
					} else if (e.key === "ArrowRight") {
						// Select and set
						// focus on next tab.
						selectedVis.setter(nextVis);
						nextRef.current?.focus();
					}
				}}
				aria-selected={selected}
				aria-controls={`panel-${thisVis}`}
				tabIndex={selected ? 0 : -1}
			>
				{label}
			</button>
		);
	}
);

interface PolicyPanelProps {
	policyForm: PolicyForm;
	forVis: Visibility;
	isActive: boolean;
}

// Tab panel for one policy form of the given visibility.
function PolicyPanel({ policyForm, forVis, isActive }: PolicyPanelProps) {
	return (
		<div
			className={`interaction-policy-section ${isActive && "active"}`}
			role="tabpanel"
			hidden={!isActive}
		>
			<PolicyComponent
				form={policyForm.favourite}
				forAction="favourite"
			/>
			<PolicyComponent
				form={policyForm.reply}
				forAction="reply"
			/>
			{ forVis !== "private" &&
				<PolicyComponent
					form={policyForm.reblog}
					forAction="reblog"
				/>
			}
		</div>
	);
}

interface PolicyComponentProps {
	form: {
		basic: PolicyFormSub;
		somethingElse: PolicyFormSomethingElse;
	};
	forAction: Action;
}

// A component of one policy of the given
// visibility, corresponding to the given action.
function PolicyComponent({ form, forAction }: PolicyComponentProps) {	
	const legend = useLegend(forAction);
	return (
		<fieldset>
			<legend>{legend}</legend>
			{ forAction === "reply" &&
				<div className="info">
					<i className="fa fa-fw fa-info-circle" aria-hidden="true"></i>
					<b>Mentioned accounts can always reply.</b>
				</div>	
			}
			<Select
				field={form.basic.field}
				label={form.basic.label}
				options={form.basic.options}
			/>
			{/* Include advanced "something else" options if appropriate */}
			{ (form.basic.field.value === "something_else") &&
				<>
					<hr />
					<div className="something-else">
						<Select
							field={form.somethingElse.followers.field}
							label={form.somethingElse.followers.label}
							options={form.somethingElse.followers.options}
						/>
						<Select
							field={form.somethingElse.following.field}
							label={form.somethingElse.following.label}
							options={form.somethingElse.following.options}
						/>
						{/*
							Skip mentioned accounts field for reply action,
							since mentioned accounts can always reply.
						*/}
						{ forAction !== "reply" &&
							<Select
								field={form.somethingElse.mentioned.field}
								label={form.somethingElse.mentioned.label}
								options={form.somethingElse.mentioned.options}
							/>
						}
						<Select
							field={form.somethingElse.everyoneElse.field}
							label={form.somethingElse.everyoneElse.label}
							options={form.somethingElse.everyoneElse.options}
						/>
					</div>
				</>
			}
		</fieldset>
	);
}

/*
	UTILITY FUNCTIONS
*/

// useLegend returns an appropriate
// fieldset legend for the given action.
function useLegend(action: Action) {
	return useMemo(() => {
		switch (action) {
			case "favourite":
				return (
					<>
						<i className="fa fa-fw fa-star" aria-hidden="true"></i>
						<span>Like</span>
					</>
				);
			case "reply":
				return (
					<>
						<i className="fa fa-fw fa-reply-all" aria-hidden="true"></i>
						<span>Reply</span>
					</>
				);
			case "reblog":
				return (
					<>
						<i className="fa fa-fw fa-retweet" aria-hidden="true"></i>
						<span>Boost</span>
					</>
				);
		}
	}, [action]);
}

// Form encapsulating the different
// actions for one visibility.
interface PolicyForm {
	favourite: {
		basic: PolicyFormSub,
		somethingElse: PolicyFormSomethingElse,
	}
	reply: {
		basic: PolicyFormSub,
		somethingElse: PolicyFormSomethingElse,
	}
	reblog: {
		basic: PolicyFormSub,
		somethingElse: PolicyFormSomethingElse,
	}
}

// Return a PolicyForm for the given visibility,
// set already to whatever the defaultPolicies value is.
function useFormForVis(
	currentPolicy: InteractionPolicy,
	forVis: Visibility,
): PolicyForm {	
	return {
		favourite: {
			basic: useBasicFor(
				forVis,
				"favourite",
				currentPolicy.can_favourite.always,
				currentPolicy.can_favourite.with_approval,
			),
			somethingElse: useSomethingElseFor(
				forVis,
				"favourite",
				currentPolicy.can_favourite.always,
				currentPolicy.can_favourite.with_approval,
			),
		},
		reply: {
			basic: useBasicFor(
				forVis,
				"reply",
				currentPolicy.can_reply.always,
				currentPolicy.can_reply.with_approval,
			),
			somethingElse: useSomethingElseFor(
				forVis,
				"reply",
				currentPolicy.can_reply.always,
				currentPolicy.can_reply.with_approval,
			),
		},
		reblog: {
			basic: useBasicFor(
				forVis,
				"reblog",
				currentPolicy.can_reblog.always,
				currentPolicy.can_reblog.with_approval,
			),
			somethingElse: useSomethingElseFor(
				forVis,
				"reblog",
				currentPolicy.can_reblog.always,
				currentPolicy.can_reblog.with_approval,
			),
		},
	};
}

function assemblePolicyEntry(
	forVis: Visibility,
	forAction: Action,
	policyForm: PolicyForm,
): InteractionPolicyEntry {
	const basic = policyForm[forAction].basic;
	
	// If this is followers visibility then
	// "anyone" only means followers, not public.
	const anyone: InteractionPolicyValue =
		(forVis === "private")
			? PolicyValueFollowers
			: PolicyValuePublic;
	
	// If this is a reply action then "just me"
	// must include mentioned accounts as well,
	// since they can always reply.
	const justMe: InteractionPolicyValue[] =
		(forAction === "reply")
			? [PolicyValueAuthor, PolicyValueMentioned]
			: [PolicyValueAuthor];

	switch (basic.field.value) {
		case "anyone":
			return {
				// Anyone can do this.
				always: [anyone],
				with_approval: [],
			};
		case "anyone_with_approval":
			return {
				// Author and maybe mentioned can do
				// this, everyone else needs approval.
				always: justMe,
				with_approval: [anyone],
			};
		case "just_me":
			return {
				// Only author and maybe
				// mentioned can do this.
				always: justMe,
				with_approval: [],
			};
	}

	// Something else!
	const somethingElse = policyForm[forAction].somethingElse;
	
	// Start with basic "always"
	// and "with_approval" values.
	let always: InteractionPolicyValue[] = justMe;
	let withApproval: InteractionPolicyValue[] = [];
	
	// Add PolicyValueFollowers depending on choices made.
	switch (somethingElse.followers.field.value as SomethingElseValue) {
		case "always":
			always.push(PolicyValueFollowers);
			break;
		case "with_approval":
			withApproval.push(PolicyValueFollowers);
			break;
	}

	// Add PolicyValueFollowing depending on choices made.
	switch (somethingElse.following.field.value as SomethingElseValue) {
		case "always":
			always.push(PolicyValueFollowing);
			break;
		case "with_approval":
			withApproval.push(PolicyValueFollowing);
			break;
	}

	// Add PolicyValueMentioned depending on choices made.
	// Note: mentioned can always reply, and that's already
	// included above, so only do this if action is not reply.
	if (forAction !== "reply") {
		switch (somethingElse.mentioned.field.value as SomethingElseValue) {
			case "always":
				always.push(PolicyValueMentioned);
				break;
			case "with_approval":
				withApproval.push(PolicyValueMentioned);
				break;
		}
	}

	// Add anyone depending on choices made.
	switch (somethingElse.everyoneElse.field.value as SomethingElseValue) {
		case "with_approval":
			withApproval.push(anyone);
			break;
	}

	// Simplify a bit after
	// all the parsing above.
	if (always.includes(anyone)) {
		always = [anyone];
	}

	if (withApproval.includes(anyone)) {
		withApproval = [anyone];
	}

	return {
		always: always,
		with_approval: withApproval,
	};
}
