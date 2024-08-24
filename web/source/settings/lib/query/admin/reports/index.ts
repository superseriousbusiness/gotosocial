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

import type {
	AdminReport,
	AdminSearchReportParams,
	AdminReportResolveParams,
	AdminSearchReportResp,
} from "../../../types/report";
import parse from "parse-link-header";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		searchReports: build.query<AdminSearchReportResp, AdminSearchReportParams>({
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
					url: `/api/v1/admin/reports${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: AdminReport[], meta) => {
				const accounts = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { accounts, links };
			},
			// Only provide LIST tag id since this model is not the
			// same as getReport model (due to transformResponse).
			providesTags: [{ type: "Report", id: "TRANSFORMED" }]
		}),

		getReport: build.query<AdminReport, string>({
			query: (id) => ({
				url: `/api/v1/admin/reports/${id}`
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'Report', id }
			],
		}),
	
		resolveReport: build.mutation<AdminReport, AdminReportResolveParams>({
			query: (formData) => ({
				url: `/api/v1/admin/reports/${formData.id}/resolve`,
				method: "POST",
				asForm: true,
				body: formData
			}),
			invalidatesTags: (res) =>
				res
					? [{ type: "Report", id: "TRANSFORMED" }, { type: "Report", id: res.id }]
					: [{ type: "Report", id: "TRANSFORMED" }]
		})
	})
});

/**
 * List reports received on this instance, filtered using given parameters.
 */
const useLazySearchReportsQuery = extended.useLazySearchReportsQuery;

/**
 * Get a single report by its ID.
 */
const useGetReportQuery = extended.useGetReportQuery;

/**
 * Mark an open report as resolved.
 */
const useResolveReportMutation = extended.useResolveReportMutation;

export {
	useLazySearchReportsQuery,
	useGetReportQuery,
	useResolveReportMutation,
};
