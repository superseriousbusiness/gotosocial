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

const Promise = require("bluebird");

const { unwrapRes } = require("../lib");

module.exports = (build) => ({
	listEmoji: build.query({
		query: (params = {}) => ({
			url: "/api/v1/admin/custom_emojis",
			params: {
				limit: 0,
				...params
			}
		}),
		providesTags: (res) =>
			res
				? [...res.map((emoji) => ({ type: "Emoji", id: emoji.id })), { type: "Emoji", id: "LIST" }]
				: [{ type: "Emoji", id: "LIST" }]
	}),

	getEmoji: build.query({
		query: (id) => ({
			url: `/api/v1/admin/custom_emojis/${id}`
		}),
		providesTags: (res, error, id) => [{ type: "Emoji", id }]
	}),

	addEmoji: build.mutation({
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

	editEmoji: build.mutation({
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

	deleteEmoji: build.mutation({
		query: (id) => ({
			method: "DELETE",
			url: `/api/v1/admin/custom_emojis/${id}`
		}),
		invalidatesTags: (res, error, id) => [{ type: "Emoji", id }]
	}),

	searchStatusForEmoji: build.mutation({
		queryFn: (url, api, _extraOpts, baseQuery) => {
			return Promise.try(() => {
				return baseQuery({
					url: `/api/v2/search?q=${encodeURIComponent(url)}&resolve=true&limit=1`
				}).then(unwrapRes);
			}).then((searchRes) => {
				return emojiFromSearchResult(searchRes);
			}).then(({ type, domain, list }) => {
				const state = api.getState();
				if (domain == new URL(state.oauth.instance).host) {
					throw "LOCAL_INSTANCE";
				}

				// search for every mentioned emoji with the admin api to get their ID
				return Promise.map(list, (emoji) => {
					return baseQuery({
						url: `/api/v1/admin/custom_emojis`,
						params: {
							filter: `domain:${domain},shortcode:${emoji.shortcode}`,
							limit: 1
						}
					}).then((unwrapRes)).then((list) => list[0]);
				}, { concurrency: 5 }).then((listWithIDs) => {
					return {
						data: {
							type,
							domain,
							list: listWithIDs
						}
					};
				});
			}).catch((e) => {
				return { error: e };
			});
		}
	}),

	patchRemoteEmojis: build.mutation({
		queryFn: ({ action, ...formData }, _api, _extraOpts, baseQuery) => {
			const data = [];
			const errors = [];

			return Promise.each(formData.selectedEmoji, (emoji) => {
				return Promise.try(() => {
					let body = {
						type: action
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
					data.push([emoji.shortcode, res]);
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
});

function emojiFromSearchResult(searchRes) {
	/* Parses the search response, prioritizing a toot result,
			and returns referenced custom emoji
	*/
	let type;

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
		list: data.emojis
	};
}