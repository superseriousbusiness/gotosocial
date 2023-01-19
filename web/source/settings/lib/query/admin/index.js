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

const {
	replaceCacheOnMutation,
	removeFromCacheOnMutation,
	domainListToObject
} = require("../lib");
const base = require("../base");

const endpoints = (build) => ({
	updateInstance: build.mutation({
		query: (formData) => ({
			method: "PATCH",
			url: `/api/v1/instance`,
			asForm: true,
			body: formData,
			discardEmpty: true
		}),
		...replaceCacheOnMutation("instance")
	}),
	mediaCleanup: build.mutation({
		query: (days) => ({
			method: "POST",
			url: `/api/v1/admin/media_cleanup`,
			params: {
				remote_cache_days: days
			}
		})
	}),
	instanceBlocks: build.query({
		query: () => ({
			url: `/api/v1/admin/domain_blocks`
		}),
		transformResponse: domainListToObject
	}),
	addInstanceBlock: build.mutation({
		query: (formData) => ({
			method: "POST",
			url: `/api/v1/admin/domain_blocks`,
			asForm: true,
			body: formData,
			discardEmpty: true
		}),
		transformResponse: (data) => {
			return {
				[data.domain]: data
			};
		},
		...replaceCacheOnMutation("instanceBlocks")
	}),
	removeInstanceBlock: build.mutation({
		query: (id) => ({
			method: "DELETE",
			url: `/api/v1/admin/domain_blocks/${id}`,
		}),
		...removeFromCacheOnMutation("instanceBlocks", {
			findKey: (_draft, newData) => {
				return newData.domain;
			}
		})
	}),
	...require("./import-export")(build),
	...require("./custom-emoji")(build)
});

module.exports = base.injectEndpoints({ endpoints });