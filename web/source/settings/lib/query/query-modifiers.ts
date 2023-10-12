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

import { Draft } from "@reduxjs/toolkit";
import { gtsApi } from "./gts-api";

/**
 * Shadow the redux onQueryStarted function for mutations.
 * https://redux-toolkit.js.org/rtk-query/api/createApi#onquerystarted
 */
type OnMutationStarted = (
	_arg: any,
	_params: MutationStartedParams
) => Promise<void>;

/**
 * Shadow the redux onQueryStarted function parameters for mutations.
 * https://redux-toolkit.js.org/rtk-query/api/createApi#onquerystarted
 */
interface MutationStartedParams {
	/**
	 * The dispatch method for the store.
	 */
	dispatch,
	/**
	 * A method to get the current state for the store.
	 */
    getState,
	/**
	 * extra as provided as thunk.extraArgument to the configureStore getDefaultMiddleware option.
	 */
    extra,
	/**
	 * A unique ID generated for the query/mutation.
	 */
    requestId,
	/**
	 *  A Promise that will resolve with a data property (the transformed query result), and a
	 * meta property (meta returned by the baseQuery). If the query fails, this Promise will
	 * reject with the error. This allows you to await for the query to finish.
	 */
    queryFulfilled,
	/**
	 * A function that gets the current value of the cache entry.
	 */
    getCacheEntry,
}

type MutationAction = (
	_draft: Draft<any>,
	_updated: any,
	_params: MutationActionParams,
) => void;

interface MutationActionParams {
	/**
	 * Either a normal old string, or a custom
	 * function to derive the key to change based
	 * on the draft and updated data.
	 * 
	 * @param _draft
	 * @param _updated 
	 * @returns 
	 */
	key?: string | ((_draft: Draft<any>, _updated: any) => string),
}

/**
 * Custom cache mutation.
 */
type CacheMutation = (
	_queryName: string | ((_arg: any) => string),
	_params?: MutationActionParams,
) => { onQueryStarted: OnMutationStarted }

/**
 * Cache mutation creator for pessimistic updates.
 * 
 * Feed it a function that you want to perform on the
 * given draft and updated data, using the given parameters.
 * 
 * https://redux-toolkit.js.org/rtk-query/api/createApi#onquerystarted
 * https://redux-toolkit.js.org/rtk-query/usage/manual-cache-updates#pessimistic-updates
 */
function makeCacheMutation(action: MutationAction): CacheMutation {
	const cacheMutation: CacheMutation = (
		queryName: string | ((_arg: any) => string),
		{ key }: MutationActionParams = {},
	) => {
		// The actual meat + potatoes of this function: returning
		// an onQueryStarted function that satisfies the redux
		// onQueryStarted function signature.
		const onQueryStarted = async(arg, { dispatch, queryFulfilled }) => {
			try {
				const { data: newData } = await queryFulfilled;
				if (typeof queryName !== "string") {
					queryName = queryName(arg);
				}

				const patchResult = dispatch(
					gtsApi.util.updateQueryData(queryName as any, arg, (draft) => {
						if (key != undefined && typeof key !== "string") {
							key = key(draft, newData);
						}
						console.log("about to do the thing")
						action(draft, newData, { key });
					})
				);

				console.log(patchResult);
			} catch (e) {
				// eslint-disable-next-line no-console
				console.error(`rolling back pessimistic update of ${queryName}: ${e}`);
			}
		};

		return { onQueryStarted };
	};
	
	return cacheMutation;
}

export const replaceCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, params) => {	
	console.log(`draft: ${draft}`)
	console.log(`newData: ${newData}`)
	console.log(`params: ${params}`)
	Object.assign(draft, newData);
});

export const appendCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, _params) => {
	draft.push(newData);
});

export const spliceCacheOnMutation: CacheMutation = makeCacheMutation((draft, _newData, { key }) => {
	if (key === undefined) {
		throw ("key undefined");
	}
	
	draft.splice(key, 1);
});

export const updateCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, { key }) => {
	if (key === undefined) {
		throw ("key undefined");
	}

	if (typeof key !== "string") {
		key = key(draft, newData);
	}
	
	draft[key] = newData;
});

export const removeFromCacheOnMutation: CacheMutation = makeCacheMutation((draft, newData, { key }) => {
	if (key === undefined) {
		throw ("key undefined");
	}

	if (typeof key !== "string") {
		key = key(draft, newData);
	}
	
	delete draft[key];
});
