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

const slice = createSlice({
	name: "checklist",
	initialState: {},
	reducers: {
		create: (state, { payload }) => {
			return createState(payload, state);
		},
		updateAll: (state, { payload: value }) => {
			return syncpipe(state, [
				(_) => Object.values(_),
				(_) => _.map((entry) => [entry.key, {
					...entry,
					checked: value
				}]),
				(_) => Object.fromEntries(_)
			]);
		},
		update: (state, { payload: { key, value } }) => {
			state[key] = {
				...state[key],
				...value
			};
		}
	}
});

const { reducer: reducer2, actions } = slice;

function reducer() {
	console.log("REDUCING", ...arguments);
	return reducer2(...arguments);
}

function createState({ entries, uniqueKey, defaultValue }, oldState) {
	console.log("creating state", oldState);
	return syncpipe(entries, [
		(_) => _.map((entry) => {
			let key = entry[uniqueKey];
			return [
				key,
				{
					...entry,
					key,
					checked: oldState?.[key]?.checked ?? entry.checked ?? defaultValue
				}
			];
		}),
		(_) => Object.fromEntries(_)
	]);
}

module.exports = function useCheckListInput({ name }, { entries, uniqueKey = "key", defaultValue = false }) {
	const [state, dispatch] = React.useReducer(reducer, null, () => createState({ entries, uniqueKey, defaultValue }));

	const [someSelected, setSomeSelected] = React.useState(false);
	const [toggleAllState, setToggleAllState] = React.useState(0);
	const toggleAllRef = React.useRef(null);

	React.useEffect(() => {
		/* Updates (un)check all checkbox, based on shortcode checkboxes
			 Can be 0 (not checked), 1 (checked) or 2 (indeterminate)
		 */
		if (toggleAllRef.current == null) {
			return;
		}

		let values = Object.values(state);
		/* one or more boxes are checked */
		let some = values.some((v) => v.checked);

		let all = false;
		if (some) {
			/* there's not at least one unchecked box */
			all = !values.some((v) => v.checked == false);
		}

		setSomeSelected(some);

		if (some && !all) {
			setToggleAllState(2);
			toggleAllRef.current.indeterminate = true;
		} else {
			setToggleAllState(all ? 1 : 0);
			toggleAllRef.current.indeterminate = false;
		}
	}, [state, toggleAllRef]);

	const reset = React.useCallback(
		() => dispatch(actions.updateAll(defaultValue)),
		[defaultValue]
	);

	const onChange = React.useCallback(
		(key, value) => dispatch(actions.update({ key, value })),
		[]
	);

	return React.useMemo(() => {
		function toggleAll(e) {
			let selectAll = e.target.checked;

			if (toggleAllState == 2) { // indeterminate
				selectAll = false;
			}

			dispatch(actions.updateAll(selectAll));
			setToggleAllState(selectAll);
		}

		function selectedValues() {
			return syncpipe(state, [
				(_) => Object.values(_),
				(_) => _.filter((entry) => entry.checked),
				(_) => _.map((entry) => ({ ...entry }))
			]);
		}

		return Object.assign([
			state,
			reset,
			{ name }
		], {
			name,
			value: state,
			onChange,
			selectedValues,
			reset,
			someSelected,
			toggleAll: {
				ref: toggleAllRef,
				value: toggleAllState,
				onChange: toggleAll
			}
		});
	}, [name, onChange, reset, someSelected, state, toggleAllState]);
};