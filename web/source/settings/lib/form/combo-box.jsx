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

const { useComboboxState } = require("ariakit/combobox");

module.exports = function useComboBoxInput({ name, Name }, { defaultValue } = {}) {
	const [isNew, setIsNew] = React.useState(false);

	const state = useComboboxState({
		defaultValue,
		gutter: 0,
		sameWidth: true
	});

	function reset() {
		state.setValue("");
	}

	return Object.assign([
		state,
		reset,
		{
			[name]: state.value,
			name,
			[`${name}IsNew`]: isNew,
			[`set${Name}IsNew`]: setIsNew
		}
	], {
		name,
		state,
		value: state.value,
		hasChanged: () => state.value != defaultValue,
		isNew,
		setIsNew,
		reset
	});
};