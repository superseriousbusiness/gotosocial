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

import { gtsApi } from "../../gts-api";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		mediaCleanup: build.mutation({
			query: (days) => ({
				method: "POST",
				url: `/api/v1/admin/media_cleanup`,
				params: {
					remote_cache_days: days
				}
			})
		}),

		instanceKeysExpire: build.mutation({
			query: (domain) => ({
				method: "POST",
				url: `/api/v1/admin/domain_keys_expire`,
				params: {
					domain: domain
				}
			})
		}),

		sendTestEmail: build.mutation<any, { email: string, message?: string }>({
			query: (params) => ({
				method: "POST",
				url: `/api/v1/admin/email/test`,
				params: params,
			})
		}),
	}),
});

/**
 * POST to /api/v1/admin/media_cleanup to trigger manual cleanup.
 */
const useMediaCleanupMutation = extended.useMediaCleanupMutation;

/**
 * POST to /api/v1/admin/domain_keys_expire to expire domain keys for the given domain.
 */
const useInstanceKeysExpireMutation = extended.useInstanceKeysExpireMutation;

const useSendTestEmailMutation = extended.useSendTestEmailMutation;

export {
	useMediaCleanupMutation,
	useInstanceKeysExpireMutation,
	useSendTestEmailMutation,
};
