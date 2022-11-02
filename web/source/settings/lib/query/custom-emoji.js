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
const { getChanges } = require("../api");

const endpoints = (build) => ({
	getAllEmoji: build.query({
		query: (params = {}) => ({
			url: "/api/v1/admin/custom_emojis",
			params: {
				limit: 0,
				...params
			},
			providesTags: (res) => 
				res
					? [...res.map((id) => ({type: "Emojis", id})), {type: "Emojis", id: "LIST"}]
					: [{type: "Emojis", id: "LIST"}]
		})
	}),
	getEmoji: build.query({
		query: (id) => ({
			url: `/api/v1/admin/custom_emojis/${id}`
		}),
		providesTags: (res, error, id) => [{type: "Emojis", id}]
	}),
	addEmoji: build.query({
		query: (form) => {
			const body = getChanges(form, {
				formKeys: ["shortcode"],
				fileKeys: ["image"]
			});
			return {
				method: "POST",
				url: `/api/v1/admin/custom_emojis`,
				asForm: true,
				body
			};
		},
		providesTags: (res, error, id) => [{type: "Emojis", id}]
	}),
	deleteEmoji: build.mutation({
		query: (id) => ({
			method: "DELETE",
			url: `/api/v1/admin/custom_emojis/${id}`
		}),
		invalidatesTags: (res, error, id) => [{type: "Emojis", id}]
	})
});

module.exports = base.injectEndpoints({endpoints});