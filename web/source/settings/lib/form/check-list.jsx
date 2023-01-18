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

function createState(entries, uniqueKey, oldState, defaultValue) {
	return syncpipe(entries, [
		(_) => _.map((entry) => {
			let key = entry[uniqueKey];
			return [
				key,
				{
					...entry,
					key,
					checked: oldState[key]?.checked ?? entry.checked ?? defaultValue
				}
			];
		}),
		(_) => Object.fromEntries(_)
	]);
}

function updateAllState(state, newValue) {
	return syncpipe(state, [
		(_) => Object.values(_),
		(_) => _.map((entry) => [entry.key, {
			...entry,
			checked: newValue
		}]),
		(_) => Object.fromEntries(_)
	]);
}

function updateState(state, key, newValue) {
	return {
		...state,
		[key]: {
			...state[key],
			...newValue
		}
	};
}

module.exports = function useCheckListInput({ name }, { entries, uniqueKey = "key", defaultValue = false }) {
	const [state, setState] = React.useState({});

	const [someSelected, setSomeSelected] = React.useState(false);
	const [toggleAllState, setToggleAllState] = React.useState(0);
	const toggleAllRef = React.useRef(null);

	React.useEffect(() => {
		/* 
			entries changed, update state,
			re-using old state if available for key
		*/
		setState(createState(entries, uniqueKey, state, defaultValue));

		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [entries]);

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

	function toggleAll(e) {
		let selectAll = e.target.checked;

		if (toggleAllState == 2) { // indeterminate
			selectAll = false;
		}

		setState(updateAllState(state, selectAll));
		setToggleAllState(selectAll);
	}

	function reset() {
		setState(updateAllState(state, defaultValue));
	}

	function selectedValues() {
		return syncpipe(state, [
			(_) => Object.values(_),
			(_) => _.filter((entry) => entry.checked)
		]);
	}

	return Object.assign([
		state,
		reset,
		{ name }
	], {
		name,
		value: state,
		onChange: (key, newValue) => setState(updateState(state, key, newValue)),
		selectedValues,
		reset,
		someSelected,
		toggleAll: {
			ref: toggleAllRef,
			value: toggleAllState,
			onChange: toggleAll
		}
	});
};