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

import React from "react";

import {
	useReducer,
	useRef,
	useEffect,
	useCallback,
	useMemo,
} from "react";

import { PayloadAction, createSlice } from "@reduxjs/toolkit";

import type {
	Checkable,
	ChecklistInputHook,
	CreateHookNames,
	HookOpts,
} from "./types";

// https://immerjs.github.io/immer/installation#pick-your-immer-version
import { enableMapSet } from "immer";
enableMapSet();

interface ChecklistState {
	entries: { [k: string]: Checkable },
	selectedEntries: Set<string>,
}

const initialState: ChecklistState = {
	entries: {},
	selectedEntries: new Set(),
}

const { reducer, actions } = createSlice({
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
			)
			
			return { entries, selectedEntries };
		},
		update: (state, { payload: { key, value } }: PayloadAction<{key: string, value: Checkable}>) => {
			if (value.checked !== undefined) {
				if (value.checked === true) {
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
		updateMultiple: (state, { payload }: PayloadAction<Array<[key: string, value: Checkable]>>) => {
			payload.forEach(([key, value]) => {
				if (value.checked !== undefined) {
					if (value.checked === true) {
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
	)

	return {
		entries: mappedEntries,
		selectedEntries
	};
}

const _default: { [k: string]: Checkable } = {}

export default function useCheckListInput(
	{ name, Name }: CreateHookNames,
	{
		entries = [],
		uniqueKey = "key",
		initialValue = false,
	}: HookOpts<boolean>
): ChecklistInputHook {
	const [state, dispatch] = useReducer(
		reducer,
		initialState,
		(_) => initialHookState({ entries, uniqueKey, initialValue }) // initial state
	);

	const toggleAllRef = useRef<any>(null);

	useEffect(() => {
		if (toggleAllRef.current != null) {
			let some = state.selectedEntries.size > 0;
			let all = false;
			if (some) {
				all = state.selectedEntries.size == Object.values(state.entries).length;
			}
			toggleAllRef.current.checked = all;
			toggleAllRef.current.indeterminate = some && !all;
		}
		// only needs to update when state.selectedEntries changes, not state.entries
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [state.selectedEntries]);

	const reset = useCallback(
		() => dispatch(actions.updateAll(initialValue)),
		[initialValue]
	);

	const onChange = useCallback(
		(key, value) => dispatch(actions.update({ key, value })),
		[]
	);

	const updateMultiple = useCallback(
		(entries) => dispatch(actions.updateMultiple(entries)),
		[]
	);

	return useMemo(() => {
		function toggleAll(e) {
			let checked = e.target.checked;
			if (e.target.indeterminate) {
				checked = false;
			}
			dispatch(actions.updateAll(checked));
		}

		function selectedValues() {
			return Array.from((state.selectedEntries)).map((key) => ({
				...state.entries[key] // returned as new object, because reducer state is immutable
			}));
		}

		return Object.assign([
			state,
			reset,
			{ name }
		], {
			_default,
			hasChanged: () => true,
			name,
			Name: "",
			value: state.entries,
			onChange,
			selectedValues,
			reset,
			someSelected: state.selectedEntries.size > 0,
			updateMultiple,
			toggleAll: {
				ref: toggleAllRef,
				onChange: toggleAll
			}
		});
	}, [state, reset, name, onChange, updateMultiple]);
};
