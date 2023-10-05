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

import { try as bbTry, map, each } from "bluebird";

import { unwrapRes } from "../lib";
import { gtsApi } from "../gts-api";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { RootState } from "../../../redux/store";

export interface CustomEmoji {
	id?: string;
	shortcode: string;
	category?: string;
}

function emojiFromSearchResult(searchRes) {
	/* Parses the search response, prioritizing a toot result,
			and returns referenced custom emoji
	*/
	let type: "statuses" | "accounts";

	if (searchRes.statuses.length > 0) {
		type = "statuses";
	} else if (searchRes.accounts.length > 0) {
		type = "accounts";
	} else {
		throw "NONE_FOUND";
	}

	let data = searchRes[type][0];

	return {
		type,
		domain: (new URL(data.url)).host, // to get WEB_DOMAIN, see https://github.com/superseriousbusiness/gotosocial/issues/1225
		emojis: data.emojis as CustomEmoji[]
	};
}

const extended = gtsApi.injectEndpoints({
	endpoints: (builder) => ({
		/**
		 * List all custom emojis uploaded on our local instance.
		 */
		listEmoji: builder.query<CustomEmoji[], Object>({
			query: (params = {}) => ({
				url: "/api/v1/admin/custom_emojis",
				params: {
					limit: 0,
					...params
				}
			}),
			providesTags: (res, _error, _arg) =>
				res
					? [
						...res.map((emoji) => ({ type: "Emoji" as const, id: emoji.id })),
						{ type: "Emoji", id: "LIST" }
					]
					: [{ type: "Emoji", id: "LIST" }]
		}),

		/**
		 * Get a single custom emoji uploaded
		 * on our local instance, by its ID.
		 */
		getEmoji: builder.query<CustomEmoji, string>({
			query: (id) => ({
				url: `/api/v1/admin/custom_emojis/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "Emoji", id }]
		}),

		/**
		 * Add a new custom emoji by uploading
		 * it to our local instance.
		 */
		addEmoji: builder.mutation<CustomEmoji, Object>({
			query: (form) => {
				return {
					method: "POST",
					url: `/api/v1/admin/custom_emojis`,
					asForm: true,
					body: form,
					discardEmpty: true
				};
			},
			invalidatesTags: (res) =>
			res
					? [{ type: "Emoji", id: "LIST" }, { type: "Emoji", id: res.id }]
					: [{ type: "Emoji", id: "LIST" }]
		}),

		/**
		 * Edit an existing custom emoji that's
		 * already been uploaded to our local instance.
		 */
		editEmoji: builder.mutation<CustomEmoji, any>({
			query: ({ id, ...patch }) => {
				return {
					method: "PATCH",
					url: `/api/v1/admin/custom_emojis/${id}`,
					asForm: true,
					body: {
						type: "modify",
						...patch
					}
				};
			},
			invalidatesTags: (res) =>
				res
					? [{ type: "Emoji", id: "LIST" }, { type: "Emoji", id: res.id }]
					: [{ type: "Emoji", id: "LIST" }]
		}),

		/**
		 * Delete a single custom emoji from
		 * our local instance using its id.
		 */
		deleteEmoji: builder.mutation<any, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/custom_emojis/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [{ type: "Emoji", id }]
		}),

		searchStatusForEmoji: builder.mutation({
			async queryFn(url: string, api, _extraOpts, fetchWithBQ) {
				// First search for given url.
				const searchRes = await fetchWithBQ({
					url: `/api/v2/search?q=${encodeURIComponent(url)}&resolve=true&limit=1`
				})
				if (searchRes.error) {
					return { error: searchRes.error as FetchBaseQueryError };
				}
				
				const { type, domain, emojis } = emojiFromSearchResult(searchRes.data)
				
				// Ensure emoji domain is not OUR domain. If it
				// is, we already have the emojis by definition.
				const state = api.getState() as RootState;
				const oauthState = state.oauth;

				if (oauthState.instanceUrl === undefined) {
					throw "AAAAAAAAAAAAAAA";
				}

				if (domain == new URL(oauthState.instanceUrl).host) {
					throw "LOCAL_INSTANCE";
				}	

				// Search for each listed emoji w/the
				// admin api to map them to versions
				// that include their ID.
				const errors: FetchBaseQueryError[] = [];
				const emojisWithIDs: CustomEmoji[] = [];

				emojis.forEach(async(emoji) => {
					// Request admin view of this emoji.
					const filter = `domain:${domain},shortcode:${emoji.shortcode}`
					const emojiRes = await fetchWithBQ({
						url: `/api/v1/admin/custom_emojis`,
						params: {
							filter: filter,
							limit: 1
						}
					})
					if (emojiRes.error) {
						errors.push(emojiRes.error)
					}
					
					// Got it!
					emojis.push(emojiRes.data as CustomEmoji)
				});

				if (errors.length !== 0) {
					return {
						error: {
							status: 400,
							statusText: 'Bad Request',
							data: {"error":`One or more errors fetching custom emojis: ${errors}`},
						},
					}	
				}
				
				return {
					data: {
						type,
						domain,
						list: emojisWithIDs,
					}
				};
			}
		}),

		patchRemoteEmojis: builder.mutation({
			queryFn: ({ action, ...formData }, _api, _extraOpts, baseQuery) => {
				const data = [];
				const errors = [];

				return each(formData.selectedEmoji, (emoji: CustomEmoji) => {
					return bbTry(() => {
						let body = {
							type: action,
							shortcode: "",
							category: undefined,
						};

						if (action == "copy") {
							body.shortcode = emoji.shortcode;
							if (formData.category.trim().length != 0) {
								body.category = formData.category;
							}
						}

						return baseQuery({
							method: "PATCH",
							url: `/api/v1/admin/custom_emojis/${emoji.id}`,
							asForm: true,
							body: body
						}).then(unwrapRes);
					}).then((res) => {
						data.push([emoji.id, res]);
					}).catch((e) => {
						let msg = e.message ?? e;
						if (e.data.error) {
							msg = e.data.error;
						}
						errors.push([emoji.shortcode, msg]);
					});
				}).then(() => {
					if (errors.length == 0) {
						return { data };
					} else {
						return {
							error: errors
						};
					}
				});
			},
			invalidatesTags: () => [{ type: "Emoji", id: "LIST" }]
		})
	})
});

