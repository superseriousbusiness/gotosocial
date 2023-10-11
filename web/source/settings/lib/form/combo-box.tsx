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

import { useComboboxState } from "ariakit/combobox";
import {
	ComboboxFormInputHook,
	CreateHookNames,
	HookOpts,
} from "./types";

const _default = "";
export default function useComboBoxInput(
	{ name, Name }: CreateHookNames,
	{ initialValue = _default }: HookOpts<string>
): ComboboxFormInputHook {
	const [isNew, setIsNew] = useState(false);

	const state = useComboboxState({
		defaultValue: initialValue,
		gutter: 0,
		sameWidth: true
	});

	function reset() {
		state.setValue(initialValue);
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
		reset,
		name,
		Name: "", // Will be set by inputHook function.
		state,
		value: state.value,
		setter: (val: string) => state.setValue(val),
		hasChanged: () => state.value != initialValue,
		isNew,
		setIsNew,
		_default
	});
}
