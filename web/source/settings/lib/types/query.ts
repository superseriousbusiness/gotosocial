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

/**
 * Pass into a query when you don't
 * want to provide an argument to it.
 */
export const NoArg = undefined;

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

export type Action = (
	_draft: Draft<any>,
	_updated: any,
	_params: ActionParams,
) => void;

export interface ActionParams {
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
export type CacheMutation = (
	_queryName: string | ((_arg: any) => string),
	_params?: ActionParams,
) => { onQueryStarted: OnMutationStarted }
