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

import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import type { Checkable } from "../lib/form/types";
import { useReducer } from "react";

// https://immerjs.github.io/immer/installation#pick-your-immer-version
import { enableMapSet } from "immer";
enableMapSet();

export interface ChecklistState {
	entries: { [k: string]: Checkable },
	selectedEntries: Set<string>,
}

const initialState: ChecklistState = {
	entries: {},
	selectedEntries: new Set(),
};

function initialHookState({
	entries,
	uniqueKey,
	initialValue,
}: {
	entries: Checkable[],
	uniqueKey: string,
	initialValue: boolean,
}): ChecklistState {
	const selectedEntries = new Set<string>();
	const mappedEntries = Object.fromEntries(
		entries.map((entry) => {
			const key = entry[uniqueKey];
			const checked = entry.checked ?? initialValue;
			
			if (checked) {
				selectedEntries.add(key);
			} else {
				selectedEntries.delete(key);
			}
		
			return [ key, { ...entry, key, checked } ];
		})
	);

	return {
		entries: mappedEntries,
		selectedEntries
	};
}

const checklistSlice = createSlice({
	name: "checklist",
	initialState, // not handled by slice itself
	reducers: {
		updateAll: (state, { payload: checked }: PayloadAction<boolean>) => {
			const selectedEntries = new Set<string>();
			const entries = Object.fromEntries(
				Object.values(state.entries).map((entry) => {
					if (checked) {
						// Cheekily add this to selected
						// entries while we're here.
						selectedEntries.add(entry.key);
					}

					return [entry.key, { ...entry, checked } ];
				})
			);
			
			return { entries, selectedEntries };
		},
		update: (state, { payload: { key, value } }: PayloadAction<{key: string, value: Partial<Checkable>}>) => {
			if (value.checked !== undefined) {
				if (value.checked) {
					state.selectedEntries.add(key);
				} else {
					state.selectedEntries.delete(key);
				}
			}

			state.entries[key] = {
				...state.entries[key],
				...value
			};
		},
		updateMultiple: (state, { payload }: PayloadAction<Array<[key: string, value: Partial<Checkable>]>>) => {						
			payload.forEach(([ key, value ]) => {								
				if (value.checked !== undefined) {
					if (value.checked) {
						state.selectedEntries.add(key);
					} else {
						state.selectedEntries.delete(key);
					}
				}

				state.entries[key] = {
					...state.entries[key],
					...value
				};
			});
		}
	}
});

export const actions = checklistSlice.actions;

/**
 * useChecklistReducer wraps the react 'useReducer'
 * hook with logic specific to the checklist reducer.
 * 
 * Use it in components where you need to keep track
 * of checklist state.
 * 
 * To update it, use dispatch with the actions
 * exported from this module.
 * 
 * @example
 * 
 * ```javascript
 * // Start with one entry with "checked" set to "false".
 * const initialEntries = [{ key: "some_key", id: "some_id", value: "some_value", checked: false }];
 * const [state, dispatch] = useChecklistReducer(initialEntries, "id", false);
 * 
 * // Dispatch an action to set "checked" of all entries to "true".
 * let checked = true;
 * dispatch(actions.updateAll(checked));
 * 
 * // Will log `["some_id"]`
 * console.log(state.selectedEntries)
 * 
 * // Will log `{ key: "some_key", id: "some_id", value: "some_value", checked: true }`
 * console.log(state.entries["some_id"])
 * ```
 */
export const useChecklistReducer = (entries: Checkable[], uniqueKey: string, initialValue: boolean) => {
	return useReducer(
		checklistSlice.reducer,
		initialState,
		(_) => initialHookState({ entries, uniqueKey, initialValue })
	);
}
