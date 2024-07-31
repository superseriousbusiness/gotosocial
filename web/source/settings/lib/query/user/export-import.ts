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

import fileDownload from "js-file-download";

import { gtsApi } from "../gts-api";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { AccountExportStats } from "../../types/account";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		exportStats: build.query<AccountExportStats, void>({
			query: () => ({
				url: `/api/v1/exports/stats`
			})
		}),
		
		exportFollowing: build.mutation<string | null, void>({
			async queryFn(_arg, _api, _extraOpts, fetchWithBQ) {
				const csvRes = await fetchWithBQ({
					url: `/api/v1/exports/following.csv`,
					acceptContentType: "text/csv",
				});
				if (csvRes.error) {
					return { error: csvRes.error as FetchBaseQueryError };
				}

				if (csvRes.meta?.response?.status !== 200) {
					return { error: csvRes.data };
				}

				fileDownload(csvRes.data, "following.csv", "text/csv");
				return { data: null };
			}
		}),

		exportFollowers: build.mutation<string | null, void>({
			async queryFn(_arg, _api, _extraOpts, fetchWithBQ) {
				const csvRes = await fetchWithBQ({
					url: `/api/v1/exports/followers.csv`,
					acceptContentType: "text/csv",
				});
				if (csvRes.error) {
					return { error: csvRes.error as FetchBaseQueryError };
				}

				if (csvRes.meta?.response?.status !== 200) {
					return { error: csvRes.data };
				}

				fileDownload(csvRes.data, "followers.csv", "text/csv");
				return { data: null };
			}
		}),

		exportLists: build.mutation<string | null, void>({
			async queryFn(_arg, _api, _extraOpts, fetchWithBQ) {
				const csvRes = await fetchWithBQ({
					url: `/api/v1/exports/lists.csv`,
					acceptContentType: "text/csv",
				});
				if (csvRes.error) {
					return { error: csvRes.error as FetchBaseQueryError };
				}

				if (csvRes.meta?.response?.status !== 200) {
					return { error: csvRes.data };
				}

				fileDownload(csvRes.data, "lists.csv", "text/csv");
				return { data: null };
			}
		}),

		exportBlocks: build.mutation<string | null, void>({
			async queryFn(_arg, _api, _extraOpts, fetchWithBQ) {
				const csvRes = await fetchWithBQ({
					url: `/api/v1/exports/blocks.csv`,
					acceptContentType: "text/csv",
				});
				if (csvRes.error) {
					return { error: csvRes.error as FetchBaseQueryError };
				}

				if (csvRes.meta?.response?.status !== 200) {
					return { error: csvRes.data };
				}

				fileDownload(csvRes.data, "blocks.csv", "text/csv");
				return { data: null };
			}
		}),

		exportMutes: build.mutation<string | null, void>({
			async queryFn(_arg, _api, _extraOpts, fetchWithBQ) {
				const csvRes = await fetchWithBQ({
					url: `/api/v1/exports/mutes.csv`,
					acceptContentType: "text/csv",
				});
				if (csvRes.error) {
					return { error: csvRes.error as FetchBaseQueryError };
				}

				if (csvRes.meta?.response?.status !== 200) {
					return { error: csvRes.data };
				}

				fileDownload(csvRes.data, "mutes.csv", "text/csv");
				return { data: null };
			}
		}),

		importData: build.mutation({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/import`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
		}),
	})
});

export const {
	useExportStatsQuery,
	useExportFollowingMutation,
	useExportFollowersMutation,
	useExportListsMutation,
	useExportBlocksMutation,
	useExportMutesMutation,
	useImportDataMutation,
} = extended;
