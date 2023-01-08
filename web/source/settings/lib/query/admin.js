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

const { updateCacheOnMutation } = require("./lib");
const base = require("./base");

const endpoints = (build) => ({
	updateInstance: build.mutation({
		query: (formData) => ({
			method: "PATCH",
			url: `/api/v1/instance`,
			asForm: true,
			body: formData
		}),
		...updateCacheOnMutation("instance")
	}),
	mediaCleanup: build.mutation({
		query: (days) => ({
			method: "POST",
			url: `/api/v1/admin/media_cleanup`,
			params: {
				remote_cache_days: days 
			}
		})
	})
});

module.exports = base.injectEndpoints({endpoints});