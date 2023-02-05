/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

"use strict";

module.exports = (build) => ({
	listReports: build.query({
		query: (params = {}) => ({
			url: "/api/v1/admin/reports",
			params: {
				limit: 100,
				...params
			}
		}),
		providesTags: ["Reports"]
	}),

	getReport: build.query({
		query: (id) => ({
			url: `/api/v1/admin/reports/${id}`
		}),
		providesTags: (res, error, id) => [{ type: "Reports", id }]
	}),

	resolveReport: build.mutation({
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
});