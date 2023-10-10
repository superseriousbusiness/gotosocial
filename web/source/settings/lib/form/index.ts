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

import { useMemo } from "react";
import getByDot from "get-by-dot";

import text from "./text";
import file from "./file";
import bool from "./bool";
import radio from "./radio";
import combobox from "./combo-box";
import checklist from "./check-list";
import fieldarray from "./field-array";

import type {
	HookFunction,
	UseInputHook,
	UseFormInputHookOpts,
	FormInputHook,
} from "./types";

function capitalizeFirst(str: string) {
	return str.slice(0, 1).toUpperCase + str.slice(1);
}

function selectorByKey(key: string) {
	if (key.includes("[")) {
		// get-by-dot does not support 'nested[deeper][key]' notation, convert to 'nested.deeper.key'
		key = key
			.replace(/\[/g, ".") // nested.deeper].key]
			.replace(/\]/g, ""); // nested.deeper.key
	}

	return function selector(obj) {
		if (obj == undefined) {
			return undefined;
		} else {
			return getByDot(obj, key);
		}
	};
}

function makeHook(hookFunction: HookFunction): UseInputHook {
	return function(name: string, opts: UseFormInputHookOpts): FormInputHook {
		// for dynamically generating attributes like 'setName'
		const Name = useMemo(() => capitalizeFirst(name), [name]);
		const selector = useMemo(() => selectorByKey(name), [name]);
		const valueSelector = opts.valueSelector ?? selector;

		opts.initialValue = useMemo(() => {
			if (opts.source == undefined) {
				return opts.defaultValue;
			} else {
				return valueSelector(opts.source) ?? opts.defaultValue;
			}
		}, [opts.source, opts.defaultValue, valueSelector]);

		const hook = hookFunction({ name, Name }, opts);

		return Object.assign(hook, { name, Name });
	};
}

export const useTextInput = makeHook(text);
export const useFileInput = makeHook(file);
export const useBoolInput = makeHook(bool);
export const useRadioInput = makeHook(radio);
export const useComboBoxInput = makeHook(combobox);
export const useCheckListInput = makeHook(checklist);
export const useFieldArrayInput = makeHook(fieldarray);

export function useValue(name, value) {
	return {
		name,
		value,
		hasChanged: () => true // always included
	};
}
