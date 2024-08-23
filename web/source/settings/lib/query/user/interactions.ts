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

import {
	InteractionRequest,
	SearchInteractionRequestsParams,
	SearchInteractionRequestsResp,
} from "../../types/interaction";
import { gtsApi } from "../gts-api";
import parse from "parse-link-header";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		getInteractionRequest: build.query<InteractionRequest, string>({
			query: (id) => ({
				method: "GET",
				url: `/api/v1/interaction_requests/${id}`,
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'InteractionRequest', id }
			],
		}),
		
		searchInteractionRequests: build.query<SearchInteractionRequestsResp, SearchInteractionRequestsParams>({
			query: (form) => {
				const params = new(URLSearchParams);
				Object.entries(form).forEach(([k, v]) => {
					if (v !== undefined) {
						params.append(k, v);
					}
				});

				let query = "";
				if (params.size !== 0) {
					query = `?${params.toString()}`;
				}

				return {
					url: `/api/v1/interaction_requests${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: InteractionRequest[], meta) => {
				const requests = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { requests, links };
			},
			providesTags: [{ type: "InteractionRequest", id: "TRANSFORMED" }]
		}),

		approveInteractionRequest: build.mutation<InteractionRequest, string>({
			query: (id) => ({
				method: "POST",
				url: `/api/v1/interaction_requests/${id}/authorize`,
			}),
			invalidatesTags: (res) =>
				res
					? [{ type: "InteractionRequest", id: "TRANSFORMED" }, { type: "InteractionRequest", id: res.id }]
					: [{ type: "InteractionRequest", id: "TRANSFORMED" }]
		}),

		rejectInteractionRequest: build.mutation<any, string>({
			query: (id) => ({
				method: "POST",
				url: `/api/v1/interaction_requests/${id}/reject`,
			}),
			invalidatesTags: (res) =>
				res
					? [{ type: "InteractionRequest", id: "TRANSFORMED" }, { type: "InteractionRequest", id: res.id }]
					: [{ type: "InteractionRequest", id: "TRANSFORMED" }]
		}),
	})
});

export const {
	useGetInteractionRequestQuery,
	useLazySearchInteractionRequestsQuery,
	useApproveInteractionRequestMutation,
	useRejectInteractionRequestMutation,
} = extended;
