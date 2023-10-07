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

import syncpipe from "syncpipe";
import { gtsApi } from "./gts-api";

export function unwrapRes(res) {
	if (res.error != undefined) {
		throw res.error;
	} else {
		return res.data;
	}
}

export function idListToObject(data) {
	// Turn flat Array into Object keyed by entry id field
	return syncpipe(data, [
		(_) => _.map((entry) => [entry.id, entry]),
		(_) => Object.fromEntries(_)
	]);
}

export const replaceCacheOnMutation = makeCacheMutation((draft, newData) => {
	Object.assign(draft, newData);
});

export const appendCacheOnMutation = makeCacheMutation((draft, newData) => {
	draft.push(newData);
});

export const spliceCacheOnMutation = makeCacheMutation((draft, _newData, { key }) => {
	draft.splice(key, 1);
});

export const updateCacheOnMutation = makeCacheMutation((draft, newData, { key }) => {
	draft[key] = newData;
});

export const removeFromCacheOnMutation = makeCacheMutation((draft, _newData, { key }) => {
	delete draft[key];
});

export const editCacheOnMutation = makeCacheMutation((draft, newData, { update }) => {
	update(draft, newData);
});

// https://redux-toolkit.js.org/rtk-query/usage/manual-cache-updates#pessimistic-updates
function makeCacheMutation(action) {
	return function cacheMutation(queryName, { key, findKey, arg, ...opts } = {}) {
		return {
			onQueryStarted: (_, { dispatch, queryFulfilled }) => {
				queryFulfilled.then(({ data: newData }) => {
					dispatch(gtsApi.util.updateQueryData(queryName, arg, (draft) => {
						if (findKey != undefined) {
							key = findKey(draft, newData);
						}
						action(draft, newData, { key, ...opts });
					}));
				});
			}
		};
	};
}