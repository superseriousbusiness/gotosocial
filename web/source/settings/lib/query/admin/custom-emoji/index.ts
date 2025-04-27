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
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { RootState } from "../../../../redux/store";

import type { CustomEmoji, EmojisFromItem, ListEmojiParams } from "../../../types/custom-emoji";

/**
 * Parses the search response, prioritizing a status
 * result, and returns any referenced custom emoji.
 * 
 * Due to current API constraints, the returned emojis
 * will not have their ID property set, so further
 * processing is required to retrieve the IDs.
 * 
 * @param searchRes 
 * @returns 
 */
function emojisFromSearchResult(searchRes): EmojisFromItem {
	// We don't know in advance whether a searched URL
	// is the URL for a status, or the URL for an account,
	// but we can derive this by looking at which search
	// result field actually has entries in it (if any).
	let type: "statuses" | "accounts";
	if (searchRes.statuses.length > 0) {
		// We had status results,
		// so this was a status URL.
		type = "statuses";
	} else if (searchRes.accounts.length > 0) {
		// We had account results,
		// so this was an account URL.
		type = "accounts";
	} else {
		// Nada, zilch, we can't do
		// anything with this.
		throw "NONE_FOUND";
	}

	// Narrow type to discard all the other
	// data on the result that we don't need.
	const data: {
		url: string;
		emojis: CustomEmoji[];
	} = searchRes[type][0];

	return {
		type,
		// Workaround to get host rather than account domain.
		// See https://codeberg.org/superseriousbusiness/gotosocial/issues/1225.
		domain: (new URL(data.url)).host,
		list: data.emojis,
	};
}

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		listEmoji: build.query<CustomEmoji[], ListEmojiParams | void>({
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

		getEmoji: build.query<CustomEmoji, string>({
			query: (id) => ({
				url: `/api/v1/admin/custom_emojis/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "Emoji", id }]
		}),

		addEmoji: build.mutation<CustomEmoji, Object>({
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

		editEmoji: build.mutation<CustomEmoji, any>({
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

		deleteEmoji: build.mutation<any, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/custom_emojis/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [{ type: "Emoji", id }]
		}),

		searchItemForEmoji: build.mutation<EmojisFromItem, string>({
			async queryFn(url, api, _extraOpts, fetchWithBQ) {
				const state = api.getState() as RootState;
				const loginState = state.login;
				
				// First search for given url.
				const searchRes = await fetchWithBQ({
					url: `/api/v2/search?q=${encodeURIComponent(url)}&resolve=true&limit=1`
				});
				if (searchRes.error) {
					return { error: searchRes.error as FetchBaseQueryError };
				}
				
				// Parse initial results of search.
				// These emojis will not have IDs set.
				const {
					type,
					domain,
					list: withoutIDs,
				} = emojisFromSearchResult(searchRes.data);
				
				// Ensure emojis domain is not OUR domain. If it
				// is, we already have the emojis by definition.
				if (loginState.instanceUrl !== undefined) {
					if (domain == new URL(loginState.instanceUrl).host) {
						throw "LOCAL_INSTANCE";
					}
				}

				// Search for each listed emoji with the admin
				// api to get the version that includes an ID.
				const errors: FetchBaseQueryError[] = [];
				const withIDs: CustomEmoji[] = (
					await Promise.all(
						withoutIDs.map(async(emoji) => {
							// Request admin view of this emoji.
							const emojiRes = await fetchWithBQ({
								url: `/api/v1/admin/custom_emojis`,
								params: {
									filter: `domain:${domain},shortcode:${emoji.shortcode}`,
									limit: 1
								}
							});
						
							if (emojiRes.error) {
								// Put error in separate array so
								// the null can be filtered nicely.
								errors.push(emojiRes.error);
								return null;
							}
							
							// Got it!
							return emojiRes.data as CustomEmoji;
						})
					)
				).flatMap((emoji) => {
					// Remove any nulls.
					return emoji || [];
				});

				if (errors.length !== 0) {
					const errData = errors.map(e => JSON.stringify(e.data)).join(",");
					return {
						error: {
							status: 400,
							statusText: 'Bad Request',
							data: {
								error: `One or more errors fetching custom emojis: [${errData}]`
							},
						},
					};	
				}
				
				// Return our ID'd
				// emojis list.
				return {
					data: {
						type,
						domain,
						list: withIDs,
					}
				};
			}
		}),

		patchRemoteEmojis: build.mutation({
			async queryFn({ action, ...formData }, _api, _extraOpts, fetchWithBQ) {
				const errors: FetchBaseQueryError[] = [];
				const selectedEmoji: CustomEmoji[] = formData.selectedEmoji;
				
				// Map function to get a promise
				// of an emoji (or null).
				const copyEmoji = async(emoji: CustomEmoji) => {
					let body: {
						type: string;
						shortcode?: string;
						category?: string;
					} = {
						type: action,
					};

					if (action == "copy") {
						body.shortcode = emoji.shortcode;
						if (formData.category.trim().length != 0) {
							body.category = formData.category;
						}
					}

					const emojiRes = await fetchWithBQ({
						method: "PATCH",
						url: `/api/v1/admin/custom_emojis/${emoji.id}`,
						asForm: true,
						body: body,
					});

					if (emojiRes.error) {
						errors.push(emojiRes.error);
						return null;
					}
					
					// Instead of mapping to the emoji we just got in emojiRes.data,
					// we map here to the existing emoji. The reason for this is that
					// if we return the new emoji, it has a new ID, and the checklist
					// component calling this function gets its state mixed up.
					//
					// For example, say you copy an emoji with ID "some_emoji"; the
					// result would return an emoji with ID "some_new_emoji_id". The
					// checklist state would then contain one emoji with ID "some_emoji",
					// and the new copy of the emoji with ID "some_new_emoji_id", leading
					// to weird-looking bugs where it suddenly appears as if the searched
					// status has another blank emoji attached to it.
					return emoji;
				};

				// Wait for all the promises to
				// resolve and remove any nulls.
				const data = (
					await Promise.all(selectedEmoji.map(copyEmoji))
				).flatMap((emoji) => emoji || []);

				if (errors.length !== 0) {
					const errData = errors.map(e => JSON.stringify(e.data)).join(",");
					return {
						error: {
							status: 400,
							statusText: 'Bad Request',
							data: {
								error: `One or more errors patching custom emojis: [${errData}]`
							},
						},
					};	
				}
				
				return { data };
			},
			invalidatesTags: () => [{ type: "Emoji", id: "LIST" }]
		})
	})
});

/**
 * List all custom emojis uploaded on our local instance.
 */
const useListEmojiQuery = extended.useListEmojiQuery;

/**
 * Get a single custom emoji uploaded on our local instance, by its ID.
 */
const useGetEmojiQuery = extended.useGetEmojiQuery;

/**
 * Add a new custom emoji by uploading it to our local instance.
 */
const useAddEmojiMutation = extended.useAddEmojiMutation;

/**
 * Edit an existing custom emoji that's already been uploaded to our local instance.
 */
const useEditEmojiMutation = extended.useEditEmojiMutation;

/**
 * Delete a single custom emoji from our local instance using its id.
 */
const useDeleteEmojiMutation = extended.useDeleteEmojiMutation;

/**
 * "Steal this look" function for selecting remote emoji from a status or account.
 */
const useSearchItemForEmojiMutation = extended.useSearchItemForEmojiMutation;

/**
 * Update/patch a bunch of remote emojis.
 */
const usePatchRemoteEmojisMutation = extended.usePatchRemoteEmojisMutation;

export {
	useListEmojiQuery,
	useGetEmojiQuery,
	useAddEmojiMutation,
	useEditEmojiMutation,
	useDeleteEmojiMutation,
	useSearchItemForEmojiMutation,
	usePatchRemoteEmojisMutation,
};
