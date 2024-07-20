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
import { InteractionPolicyValue, PolicyValueFollowers, PolicyValueFollowing, PolicyValuePublic } from "../../../../lib/types/interaction";
import { useTextInput } from "../../../../lib/form";
import { Action, Audience, PolicyFormSub, SomethingElseValue, Visibility } from "./types";

export interface PolicyFormSomethingElse {
	followers: PolicyFormSub,
	following: PolicyFormSub,
	mentioned: PolicyFormSub,
	everyoneElse: PolicyFormSub,
}

function useSomethingElseOptions(
	forVis: Visibility,
	forAction: Action,
	forAudience: Audience,
) {
	return (
		<>
			{ forAudience !== "everyone_else" &&
				<option value="always">Always</option>
			}
			<option value="with_approval">With my approval</option>
			<option value="no">No</option>
		</>
	);
}

export function useSomethingElseFor(
	forVis: Visibility,
	forAction: Action,
	currentAlways: InteractionPolicyValue[],
	currentWithApproval: InteractionPolicyValue[],
): PolicyFormSomethingElse {	
	const followersDefaultValue: SomethingElseValue = useMemo(() => {
		if (currentAlways.includes(PolicyValueFollowers)) {
			return "always";
		}

		if (currentWithApproval.includes(PolicyValueFollowers)) {
			return "with_approval";
		}
		
		return "no";
	}, [currentAlways, currentWithApproval]);
	
	const followingDefaultValue: SomethingElseValue = useMemo(() => {
		if (currentAlways.includes(PolicyValueFollowing)) {
			return "always";
		}

		if (currentWithApproval.includes(PolicyValueFollowing)) {
			return "with_approval";
		}
		
		return "no";
	}, [currentAlways, currentWithApproval]);
	
	const mentionedDefaultValue: SomethingElseValue = useMemo(() => {
		if (currentAlways.includes(PolicyValueFollowing)) {
			return "always";
		}

		if (currentWithApproval.includes(PolicyValueFollowing)) {
			return "with_approval";
		}
		
		return "no";
	}, [currentAlways, currentWithApproval]);
	
	const everyoneElseDefaultValue: SomethingElseValue = useMemo(() => {
		if (currentAlways.includes(PolicyValuePublic)) {
			return "always";
		}

		if (currentWithApproval.includes(PolicyValuePublic)) {
			return "with_approval";
		}
		
		return "no";
	}, [currentAlways, currentWithApproval]);

	return {
		followers: {
			field: useTextInput("followers", { defaultValue: followersDefaultValue }),
			label: "My followers",
			options: useSomethingElseOptions(forVis, forAction, "followers"),
		},
		following: {
			field: useTextInput("following", { defaultValue: followingDefaultValue }),
			label: "Accounts I follow",
			options: useSomethingElseOptions(forVis, forAction, "following"),
		},
		mentioned: {
			field: useTextInput("mentioned_accounts", { defaultValue: mentionedDefaultValue }),
			label: "Accounts mentioned in the post",
			options: useSomethingElseOptions(forVis, forAction, "mentioned_accounts"),
		},
		everyoneElse: {
			field: useTextInput("everyone_else", { defaultValue: everyoneElseDefaultValue }),
			label: "Everyone else",
			options: useSomethingElseOptions(forVis, forAction, "everyone_else"),
		},
	};
}