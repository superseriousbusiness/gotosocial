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

const { createApi, fetchBaseQuery } = require("@reduxjs/toolkit/query/react");
const { isPlainObject } = require("is-plain-object");

function convertToForm(obj) {
	const formData = new FormData();
	Object.entries(obj).forEach(([key, val]) => {
		if (isPlainObject(val)) {
			Object.entries(val).forEach(([key2, val2]) => {
				if (val2 != undefined) {
					formData.set(`${key}[${key2}]`, val2);
				}
			});
		} else if (val != undefined) {
			formData.set(key, val);
		}
	});
	return formData;
}

function instanceBasedQuery(args, api, extraOptions) {
	const state = api.getState();
	const { instance, token } = state.oauth;

	if (args.baseUrl == undefined) {
		args.baseUrl = instance;
	}

	if (args.discardEmpty) {
		if (args.body == undefined || Object.keys(args.body).length == 0) {
			return { data: null };
		}
		delete args.discardEmpty;
	}

	if (args.asForm) {
		delete args.asForm;
		args.body = convertToForm(args.body);
	}

	return fetchBaseQuery({
		baseUrl: args.baseUrl,
		prepareHeaders: (headers) => {
			if (token != undefined) {
				headers.set('Authorization', token);
			}
			headers.set("Accept", "application/json");
			return headers;
		},
	})(args, api, extraOptions);
}

module.exports = createApi({
	reducerPath: "api",
	baseQuery: instanceBasedQuery,
	tagTypes: ["Auth", "Emoji"],
	endpoints: (build) => ({
		instance: build.query({
			query: () => ({
				url: `/api/v1/instance`
			})
		})
	})
});