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

module.exports = function useBoolInput({ name, Name }, { defaultValue = false } = {}) {
	const [value, setValue] = React.useState(defaultValue);

	function onChange(e) {
		setValue(e.target.checked);
	}

	function reset() {
		setValue(defaultValue);
	}

	// Array / Object hybrid, for easier access in different contexts
	return Object.assign([
		onChange,
		reset,
		{
			[name]: value,
			[`set${Name}`]: setValue
		}
	], {
		name,
		onChange,
		reset,
		value,
		setter: setValue,
		hasChanged: () => value != defaultValue
	});
};