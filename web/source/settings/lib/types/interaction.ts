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

import { Links } from "parse-link-header";
import { Account } from "./account";
import { Status } from "./status";

export interface DefaultInteractionPolicies {
	direct: InteractionPolicy;
	private: InteractionPolicy;
	unlisted: InteractionPolicy;
	public: InteractionPolicy;
}

export interface UpdateDefaultInteractionPolicies {
	direct: InteractionPolicy | null;
	private: InteractionPolicy | null;
	unlisted: InteractionPolicy | null;
	public: InteractionPolicy | null;
}

export interface InteractionPolicy {
	can_favourite: InteractionPolicyEntry;
	can_reply: InteractionPolicyEntry;
	can_reblog: InteractionPolicyEntry;
}

export interface InteractionPolicyEntry {
	always: InteractionPolicyValue[];
	with_approval: InteractionPolicyValue[];
}

export type InteractionPolicyValue = string;

const PolicyValuePublic: InteractionPolicyValue = "public";
const PolicyValueFollowers: InteractionPolicyValue = "followers";
const PolicyValueFollowing: InteractionPolicyValue = "following";
const PolicyValueMutuals: InteractionPolicyValue = "mutuals";
const PolicyValueMentioned: InteractionPolicyValue = "mentioned";
const PolicyValueAuthor: InteractionPolicyValue = "author";
const PolicyValueMe: InteractionPolicyValue = "me";

export {
	PolicyValuePublic,
	PolicyValueFollowers,
	PolicyValueFollowing,
	PolicyValueMutuals,
	PolicyValueMentioned,
	PolicyValueAuthor,
	PolicyValueMe,
};


/**
 * Interaction request targeting a status by an account.
 */
export interface InteractionRequest {
    /**
	 * ID of the request.
	 */
	id: string;
	/**
	 * Type of interaction being requested.
	 */
	type: "favourite" | "reply" | "reblog";
	/**
	 * Time when the request was created.
	 */
	created_at: string;
	/**
	 * Account that created the request.
	 */
	account: Account;
	/**
	 * Status being interacted with.
	 */
	status: Status;
	/**
	 * Replying status, if type = "reply".
	 */
	reply?: Status;
}

/**
 * Parameters for GET to /api/v1/interaction_requests.
 */
export interface SearchInteractionRequestsParams {
	/**
	 * If set, show only requests targeting the given status_id.
	 */
	status_id?: string;
	/**
	 * If true or not set, include favourites in the results.
	 */
	favourites?: boolean;
	/**
	 * If true or not set, include replies in the results.
	 */
	replies?: boolean;
	/**
	 * If true or not set, include reblogs in the results.
	 */
	reblogs?: boolean;
	/**
	 * If set, show only requests older (ie., lower) than the given ID.
	 * Request with the given ID will not be included in response.
	 */
	max_id?: string;
	/**
	 * If set, show only requests newer (ie., higher) than the given ID.
	 * Request with the given ID will not be included in response.
	 */
	since_id?: string;
	/**
	 * If set, show only requests *immediately newer* than the given ID.
	 * Request with the given ID will not be included in response.
	 */
	min_id?: string;
	/**
	 * If set, limit returned requests to this number.
	 * Else, fall back to GtS API defaults.
	 */
	limit?: number;
}

export interface SearchInteractionRequestsResp {
	requests: InteractionRequest[];
	links: Links | null;
}
