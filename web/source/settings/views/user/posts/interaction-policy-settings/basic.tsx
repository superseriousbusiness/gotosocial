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

import React, { useMemo } from "react";
import {
	InteractionPolicyValue,
	PolicyValueAuthor,
	PolicyValueFollowers,
	PolicyValueMentioned,
	PolicyValuePublic,
} from "../../../../lib/types/interaction";
import { useTextInput } from "../../../../lib/form";
import { Action, BasicValue, PolicyFormSub, Visibility } from "./types";

// Based on the given visibility, action, and states,
// derives what the initial basic Select value should be.
function useBasicValue(
	forVis: Visibility,
	forAction: Action,
	always: InteractionPolicyValue[],
	withApproval: InteractionPolicyValue[],
): BasicValue {
	// Check if "always" value is just the author
	// (and possibly mentioned accounts when dealing
	// with replies -- still counts as "just_me").
	const alwaysJustAuthor = useMemo(() => {
		if (
			always.length === 1 &&
			always[0] === PolicyValueAuthor
		) {
			return true;
		}

		if (
			forAction === "reply" &&
			always.length === 2 &&
			always.includes(PolicyValueAuthor) &&
			always.includes(PolicyValueMentioned)
		) {
			return true;
		}

		return false;
	}, [forAction, always]);

	// Check if "always" includes the widest
	// possible audience for this visibility.
	const alwaysWidestAudience = useMemo(() => {
		return (
			(forVis === "private" && always.includes(PolicyValueFollowers)) ||
			always.includes(PolicyValuePublic)
		);
	}, [forVis, always]);

	// Check if "withApproval" includes the widest
	// possible audience for this visibility.
	const withApprovalWidestAudience = useMemo(() => {
		return (
			(forVis === "private" && withApproval.includes(PolicyValueFollowers)) ||
			withApproval.includes(PolicyValuePublic)
		);
	}, [forVis, withApproval]);

	return useMemo(() => {
		// Simplest case: if "always" includes the
		// widest possible audience for this visibility,
		// then we don't need to check anything else.
		if (alwaysWidestAudience) {
			return "anyone";
		}

		// Next simplest case: there's no "with approval"
		// URIs set, so check if it's always just author.
		if (withApproval.length === 0 && alwaysJustAuthor) {
			return "just_me";
		}

		// Third simplest case: always is just us, and with
		// approval is addressed to the widest possible audience.
		if (alwaysJustAuthor && withApprovalWidestAudience) {
			return "anyone_with_approval";
		}

		// We've exhausted the
		// simple possibilities.
		return "something_else";
	}, [
		withApproval.length,
		alwaysJustAuthor,
		alwaysWidestAudience,
		withApprovalWidestAudience,
	]);
}

// Derive wording for the basic label for 
// whatever visibility and action we're handling.
function useBasicLabel(visibility: Visibility, action: Action) {
	return useMemo(() => {
		let visPost = "";
		switch (visibility) {
			case "public":
				visPost = "a public post";
				break;
			case "unlisted":
				visPost = "an unlisted post";
				break;
			case "private":
				visPost = "a followers-only post";
				break;
		}
		
		switch (action) {
			case "favourite":
				return "Who can like " + visPost + "?";
			case "reply":
				return "Who else can reply to " + visPost + "?";
			case "reblog":
				return "Who can boost " + visPost + "?";
		}
	}, [visibility, action]);
}

// Return whatever the "basic" options should
// be in the basic Select for this visibility.
function useBasicOptions(visibility: Visibility) {
	return useMemo(() => {
		const audience = visibility === "private"
			? "My followers"
			: "Anyone";
		
		return (
			<>
				<option value="anyone">{audience}</option>
				<option value="anyone_with_approval">{audience} (approval required)</option>
				<option value="just_me">Just me</option>
				{ visibility !== "private" &&
					<option value="something_else">Something else...</option>
				}
			</>
		);
	}, [visibility]);
}

export function useBasicFor(
	forVis: Visibility,
	forAction: Action,
	currentAlways: InteractionPolicyValue[],
	currentWithApproval: InteractionPolicyValue[],
): PolicyFormSub {
	// Determine who's currently *basically* allowed
	// to do this action for this visibility.
	const defaultValue = useBasicValue(
		forVis,
		forAction,
		currentAlways,
		currentWithApproval,
	);

	return {
		field: useTextInput("basic", { defaultValue: defaultValue }),
		label: useBasicLabel(forVis, forAction),
		options: useBasicOptions(forVis),
	};
}
