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
	AdminReportListParams,
	AdminReportResolveParams,
} from "../../../types/report";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		listReports: build.query<AdminReport[], AdminReportListParams | void>({
			query: (params) => ({
				url: "/api/v1/admin/reports",
				params: {
					// Override provided limit.
					limit: 100,
					...params
				}
			}),
			providesTags: [{ type: "Reports", id: "LIST" }]
		}),

		getReport: build.query<AdminReport, string>({
			query: (id) => ({
				url: `/api/v1/admin/reports/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "Reports", id }]
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
					? [{ type: "Reports", id: "LIST" }, { type: "Reports", id: res.id }]
					: [{ type: "Reports", id: "LIST" }]
		})
	})
});

/**
 * List reports received on this instance, filtered using given parameters.
 */
const useListReportsQuery = extended.useListReportsQuery;

/**
 * Get a single report by its ID.
 */
const useGetReportQuery = extended.useGetReportQuery;

/**
 * Mark an open report as resolved.
 */
const useResolveReportMutation = extended.useResolveReportMutation;

export {
	useListReportsQuery,
	useGetReportQuery,
	useResolveReportMutation,
};
