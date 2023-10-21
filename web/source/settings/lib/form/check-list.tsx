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

import {
	useRef,
	useEffect,
	useCallback,
	useMemo,
} from "react";

import type {
	Checkable,
	ChecklistInputHook,
	CreateHookNames,
	HookOpts,
} from "./types";

import {
	useChecklistReducer,
	actions,
} from "../../redux/checklist";

const _default: { [k: string]: Checkable } = {};
export default function useCheckListInput(
	/* eslint-disable no-unused-vars */
	{ name, Name }: CreateHookNames,
	{
		entries = [],
		uniqueKey = "key",
		initialValue = false,
	}: HookOpts<boolean>
): ChecklistInputHook {
	const [state, dispatch] = useChecklistReducer(entries, uniqueKey, initialValue);
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
		(key: string, value: Checkable) => dispatch(actions.update({ key, value })),
		[]
	);

	const updateMultiple = useCallback(
		(entries: [key: string, value: Partial<Checkable>][]) => dispatch(actions.updateMultiple(entries)),
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
}
