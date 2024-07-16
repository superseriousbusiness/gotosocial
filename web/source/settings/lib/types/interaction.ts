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
