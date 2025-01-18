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

import React, {
	useState,
	useRef,
	useTransition,
	useEffect,
} from "react";

import type {
	CreateHookNames,
	HookOpts,
	NumberFormInputHook,
} from "./types";

const _default = 0;

export default function useNumberInput(
	{ name, Name }: CreateHookNames,
	{
		initialValue = _default,
		dontReset = false,
		validator,
		showValidation = true,
		initValidation,
		nosubmit = false,
	}: HookOpts<number>
): NumberFormInputHook {
	const [number, setNumber] = useState(initialValue);
	const numberRef = useRef<HTMLInputElement>(null);

	const [validation, setValidation] = useState(initValidation ?? "");
	const [_isValidating, startValidation] = useTransition();
	const valid = validation == "";

	function onChange(e: React.ChangeEvent<HTMLInputElement>) {
		const input = e.target.valueAsNumber;
		setNumber(input);

		if (validator) {
			startValidation(() => {
				setValidation(validator(input));
			});
		}
	}

	function reset() {
		if (!dontReset) {
			setNumber(initialValue);
		}
	}

	useEffect(() => {
		if (validator && numberRef.current) {
			if (showValidation) {
				numberRef.current.setCustomValidity(validation);
			} else {
				numberRef.current.setCustomValidity("");
			}
		}
	}, [validation, validator, showValidation]);

	// Array / Object hybrid, for easier access in different contexts
	return Object.assign([
		onChange,
		reset,
		{
			[name]: number,
			[`${name}Ref`]: numberRef,
			[`set${Name}`]: setNumber,
			[`${name}Valid`]: valid,
		}
	], {
		onChange,
		reset,
		name,
		Name: "", // Will be set by inputHook function.
		nosubmit,
		value: number,
		ref: numberRef,
		setter: setNumber,
		valid,
		validate: () => setValidation(validator ? validator(number): ""),
		hasChanged: () => number != initialValue,
		_default
	});
}
