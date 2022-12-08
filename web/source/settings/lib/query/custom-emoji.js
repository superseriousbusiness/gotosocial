/*
	 GoToSocial
	 Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

const base = require("./base");

function unwrap(res) {
	if (res.error != undefined) {
		throw res.error;
	} else {
		return res.data;
	}
}

const endpoints = (build) => ({
	getAllEmoji: build.query({
		query: (params = {}) => ({
			url: "/api/v1/admin/custom_emojis",
			params: {
				limit: 0,
				...params
			}
		}),
		providesTags: (res) => 
			res
				? [...res.map((emoji) => ({type: "Emojis", id: emoji.id})), {type: "Emojis", id: "LIST"}]
				: [{type: "Emojis", id: "LIST"}]
	}),
	getEmoji: build.query({
		query: (id) => ({
			url: `/api/v1/admin/custom_emojis/${id}`
		}),
		providesTags: (res, error, id) => [{type: "Emojis", id}]
	}),
	addEmoji: build.mutation({
		query: (form) => {
			return {
				method: "POST",
				url: `/api/v1/admin/custom_emojis`,
				asForm: true,
				body: form
			};
		},
		invalidatesTags: (res) => 
			res
				? [{type: "Emojis", id: "LIST"}, {type: "Emojis", id: res.id}]
				: [{type: "Emojis", id: "LIST"}]
	}),
	editEmoji: build.mutation({
		query: ({id, ...patch}) => {
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
				? [{type: "Emojis", id: "LIST"}, {type: "Emojis", id: res.id}]
				: [{type: "Emojis", id: "LIST"}]
	}),
	deleteEmoji: build.mutation({
		query: (id) => ({
			method: "DELETE",
			url: `/api/v1/admin/custom_emojis/${id}`
		}),
		invalidatesTags: (res, error, id) => [{type: "Emojis", id}]
	}),
	searchStatusForEmoji: build.mutation({
		query: (url) => ({
			method: "GET",
			url: `/api/v2/search?q=${encodeURIComponent(url)}&resolve=true&limit=1`
		}),
		transformResponse: (res) => {
			/* Parses search response, prioritizing a toot result,
			   and returns referenced custom emoji
			*/
			let type;

			if (res.statuses.length > 0) {
				type = "statuses";
			} else if (res.accounts.length > 0) {
				type = "accounts";
			} else {
				return {
					type: "none"
				};
			}

			let data = res[type][0];

			return {
				type,
				domain: (new URL(data.url)).host, // to get WEB_DOMAIN, see https://github.com/superseriousbusiness/gotosocial/issues/1225
				list: data.emojis
			};
		}
	}),
	patchRemoteEmojis: build.mutation({
		queryFn: ({action, domain, list, category}, api, _extraOpts, baseQuery) => {
			const data = [];
			const errors = [];

			return Promise.each(list, (emoji) => {
				return Promise.try(() => {
					return baseQuery({
						method: "GET",
						url: `/api/v1/admin/custom_emojis`,
						params: {
							filter: `domain:${domain},shortcode:${emoji.shortcode}`,
							limit: 1
						}
					}).then(unwrap);
				}).then(([lookup]) => {
					if (lookup == undefined) { throw "not found"; }

					let body = {
						type: action
					};

					if (action == "copy") {
						body.shortcode = emoji.localShortcode ?? emoji.shortcode;
						if (category.trim().length != 0) {
							body.category = category;
						}
					}

					return baseQuery({
						method: "PATCH",
						url: `/api/v1/admin/custom_emojis/${lookup.id}`,
						asForm: true,
						body: body
					}).then(unwrap);
				}).then((res) => {
					data.push([emoji.shortcode, res]);
				}).catch((e) => {
					console.error("emoji lookup for", emoji.shortcode, "failed:", e);
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
		invalidatesTags: () => [{type: "Emojis", id: "LIST"}]
	})
});

module.exports = base.injectEndpoints({endpoints});