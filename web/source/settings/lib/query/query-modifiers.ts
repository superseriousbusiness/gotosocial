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

import { gtsApi } from "./gts-api";

import type { 
	Action,
	CacheMutation,
} from "../types/query";

import { NoArg } from "../types/query";

/**
 * Cache mutation creator for pessimistic updates.
 * 
 * Feed it a function that you want to perform on the
 * given draft and updated data, using the given parameters.
 * 
 * https://redux-toolkit.js.org/rtk-query/api/createApi#onquerystarted
 * https://redux-toolkit.js.org/rtk-query/usage/manual-cache-updates#pessimistic-updates
 */
function makeCacheMutation(action: Action): CacheMutation {
	return function cacheMutation(
		queryName: string | ((_arg: any) => string),
		{ key } = {},
	) {
		return {
			onQueryStarted: async(mutationData, { dispatch, queryFulfilled }) => {
				// queryName might be a function that returns
				// a query name; trigger it if so. The returned
				// queryName has to match one of the API endpoints
				// we've defined. So if we have endpoints called
				// (for example) `instanceV1` and `getPosts` then
				// the queryName provided here has to line up with
				// one of those in order to actually do anything.
				if (typeof queryName !== "string") {
					queryName = queryName(mutationData);
				}
				
				if (queryName == "") {
					throw (
						"provided queryName resolved to an empty string;" +
						"double check your mutation definition!"
					);
				}

				try {
					// Wait for the mutation to finish (this
					// is why it's a pessimistic update).
					const { data: newData } = await queryFulfilled;	
					
					// In order for `gtsApi.util.updateQueryData` to
					// actually do something within a dispatch, the
					// first two arguments passed into it have to line
					// up with arguments that were used earlier to
					// fetch the data whose cached version we're now
					// trying to modify.
					// 
					// So, if we earlier fetched all reports with
					// queryName `getReports`, and arg `undefined`,
					// then we now need match those parameters in
					// `updateQueryData` in order to modify the cache.
					//
					// If you pass something like `null` or `""` here
					// instead, then the cache will not get modified!
					// Redux will just quietly discard the thunk action.
					dispatch(
						gtsApi.util.updateQueryData(queryName as any, NoArg, (draft) => {
							if (key != undefined && typeof key !== "string") {
								key = key(draft, newData);
							}
							action(draft, newData, { key });
						})
					);
				} catch (e) {
					// eslint-disable-next-line no-console
					console.error(`rolling back pessimistic update of ${queryName}: ${e}`);
				}
			}
		};
	};
}

/**
 * 
 */
const replaceCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, _params) => {	
	Object.assign(draft, newData);
});

const appendCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, _params) => {
	draft.push(newData);
});

const spliceCacheOnMutation: CacheMutation = makeCacheMutation((draft, _newData, { key }) => {
	if (key === undefined) {
		throw ("key undefined");
	}
	
	draft.splice(key, 1);
});

const updateCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, { key }) => {
	if (key === undefined) {
		throw ("key undefined");
	}

	if (typeof key !== "string") {
		key = key(draft, newData);
	}
	
	draft[key] = newData;
});

const removeFromCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, { key }) => {
	if (key === undefined) {
		throw ("key undefined");
	}

	if (typeof key !== "string") {
		key = key(draft, newData);
	}
	
	delete draft[key];
});


export {
	replaceCacheOnMutation,
	appendCacheOnMutation,
	spliceCacheOnMutation,
	updateCacheOnMutation,
	removeFromCacheOnMutation,
};
