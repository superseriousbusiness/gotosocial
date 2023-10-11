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

import { useState } from "react";
import { CreateHookNames, HookOpts, RadioFormInputHook } from "./types";

const _default = "";
export default function useRadioInput(
	{ name, Name }: CreateHookNames,
	{
		initialValue = _default,
		options = {},
	}: HookOpts<string>
): RadioFormInputHook {
	const [value, setValue] = useState(initialValue);

	function onChange(e) {
		setValue(e.target.value);
	}

	function reset() {
		setValue(initialValue);
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
		onChange,
		reset,
		name,
		Name: "",
		value,
		setter: setValue,
		options,
		hasChanged: () => value != initialValue,
		_default
	});
}
