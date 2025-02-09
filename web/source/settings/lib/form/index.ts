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
import array from "./array";
import fieldarray from "./field-array";

import type {
	CreateHook,
	FormInputHook,
	HookOpts,
	TextFormInputHook,
	RadioFormInputHook,
	FileFormInputHook,
	BoolFormInputHook,
	ComboboxFormInputHook,
	ChecklistInputHook,
	FieldArrayInputHook,
	ArrayInputHook,
	NumberFormInputHook,
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

/**
 * Memoized hook generator function. Take a createHook
 * function and use it to return a new FormInputHook function.
 * 
 * @param createHook 
 * @returns 
 */
function inputHook(createHook: CreateHook): (_name: string, _opts: HookOpts) => FormInputHook {	
	return (name: string, opts?: HookOpts): FormInputHook => {
		// for dynamically generating attributes like 'setName'
		const Name = useMemo(() => capitalizeFirst(name), [name]);
		const selector = useMemo(() => selectorByKey(name), [name]);
		const valueSelector = opts?.valueSelector?? selector;

		if (opts) {
			opts.initialValue = useMemo(() => {
				if (opts.source == undefined) {
					return opts.defaultValue;
				} else {
					return valueSelector(opts.source) ?? opts.defaultValue;
				}
			}, [opts.source, opts.defaultValue, valueSelector]);
		}

		const hook = createHook({ name, Name }, opts ?? {});
		return Object.assign(hook, { name, Name });
	};
}

/**
 * Simplest form hook type in town.
 */
function value<T>(name: string, initialValue: T) {
	return {
		_default: initialValue,
		name,
		Name: "", 
		value: initialValue,
		hasChanged: () => true,
	};
}

export const useTextInput = inputHook(text) as (_name: string, _opts?: HookOpts<string>) => TextFormInputHook;
export const useNumberInput = inputHook(text) as (_name: string, _opts?: HookOpts<number>) => NumberFormInputHook;
export const useFileInput = inputHook(file) as (_name: string, _opts?: HookOpts<File>) => FileFormInputHook;
export const useBoolInput = inputHook(bool) as (_name: string, _opts?: HookOpts<boolean>) => BoolFormInputHook;
export const useRadioInput = inputHook(radio) as (_name: string, _opts?: HookOpts<string>) => RadioFormInputHook;
export const useComboBoxInput = inputHook(combobox) as (_name: string, _opts?: HookOpts<string>) => ComboboxFormInputHook;
export const useCheckListInput = inputHook(checklist) as (_name: string, _opts?: HookOpts<boolean>) => ChecklistInputHook;
export const useArrayInput = inputHook(array) as (_name: string, _opts?: HookOpts<string[]>) => ArrayInputHook;
export const useFieldArrayInput = inputHook(fieldarray) as (_name: string, _opts?: HookOpts<string>) => FieldArrayInputHook;
export const useValue = value as <T>(_name: string, _initialValue: T) => FormInputHook<T>;
