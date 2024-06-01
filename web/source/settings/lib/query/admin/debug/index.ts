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

import { ApURLResponse } from "../../../types/debug";
import { gtsApi } from "../../gts-api";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		ApURL: build.query<ApURLResponse, string>({
			query: (url) => {
				// Get the url in a SearchParam
				// so we can escape it.
				const urlParam = new URLSearchParams();
				urlParam.set("url", url);

				return {
					url: `/api/v1/admin/debug/apurl?${urlParam.toString()}`,
				};
			}
		}),
		ClearCaches: build.mutation<{}, void>({
			query: () => ({
				method: "POST",
				url: `/api/v1/admin/debug/caches/clear`
			})
		}),
	}),
});

/**
 * Lazy GET to /api/v1/admin/debug/apurl.
 */
const useLazyApURLQuery = extended.useLazyApURLQuery;

/**
 * POST to /api/v1/admin/debug/caches/clear to empty in-memory caches.
 */
const useClearCachesMutation = extended.useClearCachesMutation;

export {
	useLazyApURLQuery,
	useClearCachesMutation,
};
