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
import { HeaderPermission } from "../../../types/http-header-permissions";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		
		/* HTTP HEADER ALLOWS */
		
		getHeaderAllows: build.query<HeaderPermission[], void>({
			query: () => ({
				url: `/api/v1/admin/header_allows`
			}),
			providesTags: (res) =>
				res
					? [
						...res.map(({ id }) => ({ type: "HTTPHeaderAllows" as const, id })),
						{ type: "HTTPHeaderAllows", id: "LIST" },
					]
					: [{ type: "HTTPHeaderAllows", id: "LIST" }],
		}),

		getHeaderAllow: build.query<HeaderPermission, string>({
			query: (id) => ({
				url: `/api/v1/admin/header_allows/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "HTTPHeaderAllows", id }],
		}),

		postHeaderAllow: build.mutation<HeaderPermission, { header: string, regex: string }>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/header_allows`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			invalidatesTags: [{ type: "HTTPHeaderAllows", id: "LIST" }],
		}),

		deleteHeaderAllow: build.mutation<HeaderPermission, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/header_allows/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [{ type: "HTTPHeaderAllows", id }],
		}),
		
		/* HTTP HEADER BLOCKS */

		getHeaderBlocks: build.query<HeaderPermission[], void>({
			query: () => ({
				url: `/api/v1/admin/header_blocks`
			}),
			providesTags: (res) =>
				res
					? [
						...res.map(({ id }) => ({ type: "HTTPHeaderBlocks" as const, id })),
						{ type: "HTTPHeaderBlocks", id: "LIST" },
					]
					: [{ type: "HTTPHeaderBlocks", id: "LIST" }],
		}),

		postHeaderBlock: build.mutation<HeaderPermission, { header: string, regex: string }>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/header_blocks`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			invalidatesTags: [{ type: "HTTPHeaderBlocks", id: "LIST" }],
		}),

		getHeaderBlock: build.query<HeaderPermission, string>({
			query: (id) => ({
				url: `/api/v1/admin/header_blocks/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "HTTPHeaderBlocks", id }],
		}),

		deleteHeaderBlock: build.mutation<HeaderPermission, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/header_blocks/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [{ type: "HTTPHeaderBlocks", id }],
		}),
	}),
});

/**
 * Get admin view of all HTTP header allow regexes.
 */
const useGetHeaderAllowsQuery = extended.useGetHeaderAllowsQuery;

/**
 * Get admin view of one HTTP header allow regex.
 */
const useGetHeaderAllowQuery = extended.useGetHeaderAllowQuery;

/**
 * Create a new HTTP header allow regex.
 */
const usePostHeaderAllowMutation = extended.usePostHeaderAllowMutation;

/**
 * Delete one HTTP header allow regex.
 */
const useDeleteHeaderAllowMutation = extended.useDeleteHeaderAllowMutation;

/**
 * Get admin view of all HTTP header block regexes.
 */
const useGetHeaderBlocksQuery = extended.useGetHeaderBlocksQuery;

/**
 * Get admin view of one HTTP header block regex.
 */
const useGetHeaderBlockQuery = extended.useGetHeaderBlockQuery;

/**
 * Create a new HTTP header block regex.
 */
const usePostHeaderBlockMutation = extended.usePostHeaderBlockMutation;

/**
 * Delete one HTTP header block regex.
 */
const useDeleteHeaderBlockMutation = extended.useDeleteHeaderBlockMutation;

export {
	useGetHeaderAllowsQuery,
	useGetHeaderAllowQuery,
	usePostHeaderAllowMutation,
	useDeleteHeaderAllowMutation,
	useGetHeaderBlocksQuery,
	useGetHeaderBlockQuery,
	usePostHeaderBlockMutation,
	useDeleteHeaderBlockMutation,
};
