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

const React = require("react");
const syncpipe = require("syncpipe");
const { createSlice } = require("@reduxjs/toolkit");
const { enableMapSet } = require("immer");

enableMapSet(); // for use in reducers

const { reducer, actions } = createSlice({
	name: "checklist",
	initialState: {}, // not handled by slice itself
	reducers: {
		updateAll: (state, { payload: checked }) => {
			const selectedEntries = new Set();
			return {
				entries: syncpipe(state.entries, [
					(_) => Object.values(_),
					(_) => _.map((entry) => {
						if (checked) {
							selectedEntries.add(entry.key);
						}
						return [entry.key, {
							...entry,
							checked
						}];
					}),
					(_) => Object.fromEntries(_)
				]),
				selectedEntries
			};
		},
		update: (state, { payload: { key, value } }) => {
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
		updateMultiple: (state, { payload }) => {
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

function initialState({ entries, uniqueKey, initialValue }) {
	const selectedEntries = new Set();
	return {
		entries: syncpipe(entries, [
			(_) => _.map((entry) => {
				let key = entry[uniqueKey];
				let checked = entry.checked ?? initialValue;

				if (checked) {
					selectedEntries.add(key);
				} else {
					selectedEntries.delete(key);
				}

				return [
					key,
					{
						...entry,
						key,
						checked
					}
				];
			}),
			(_) => Object.fromEntries(_)
		]),
		selectedEntries
	};
}

module.exports = function useCheckListInput({ name }, { entries, uniqueKey = "key", initialValue = false }) {
	const [state, dispatch] = React.useReducer(reducer, null,
		() => initialState({ entries, uniqueKey, initialValue }) // initial state
	);

	const toggleAllRef = React.useRef(null);

	React.useEffect(() => {
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

	const reset = React.useCallback(
		() => dispatch(actions.updateAll(initialValue)),
		[initialValue]
	);

	const onChange = React.useCallback(
		(key, value) => dispatch(actions.update({ key, value })),
		[]
	);

	const updateMultiple = React.useCallback(
		(entries) => dispatch(actions.updateMultiple(entries)),
		[]
	);

	return React.useMemo(() => {
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
			name,
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