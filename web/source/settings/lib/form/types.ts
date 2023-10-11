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

/* eslint-disable no-unused-vars */

import React, { SyntheticEvent } from "react";

export interface HookOpts<T = any> {
	initialValue: T,
	dontReset: boolean,
	validator?: (_input: T) => string,
	showValidation: boolean,
	initValidation: string,
	length: number;
	options: { [_: string]: string },
	withPreview?: boolean,
	maxSize,
	initialInfo?: string;
}

export interface HookNames {
	name: string;
	Name: string;
}

export type CreateHook = (
	name: HookNames,
	opts: Object,
) => FormInputHook;

export type UseFormInputHook<T = any> = (
	name: string,
	opts: {
		valueSelector: (_arg: string) => any;
		initialValue;
		defaultValue;
		source: string;
		options?: Object;
	},
) => FormInputHook<T>;

export interface FormInputHook<T = any> {
	/**
	 * Name of this FormInputHook, as provided in the UseFormInputHook options.
	 */
	name: string;

	/**
	 * `name` with first letter capitalized.
	 */
	Name: string;
	
	/**
	 * Current value of this FormInputHook.
	 */
	value?: T;

	/**
	 * Default value of this FormInputHook.
	 */
	_default: T;

	/**
	 * Sets the `value` of the FormInputHook to the provided value.
	 */
	setter: (_new: T) => void;

	// TODO: move these to separate types.
	selectedValues?: () => any[];
	hasChanged: () => boolean;
	onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
	reset: () => void;
	ctx?,
	maxLength?,
}

export interface FormInputHookWithOptions<T = any> extends FormInputHook {
	options: { [_: string]: T };
}

export interface HookedForm {
	[_: string]: FormInputHook
}

/**
 * Parameters for FormSubmitFunction.
 */
export type FormSubmitEvent = (string | (SyntheticEvent<HTMLFormElement, SubmitEvent>) | undefined | void)

/**
 * Shadows "trigger" function for useMutation.
 * See: https://redux-toolkit.js.org/rtk-query/usage/mutations#mutation-hook-behavior
 */
export type FormSubmitFunction = (_e: FormSubmitEvent) => Promise<void>

/**
 * Shadows redux mutation hook return values.
 * See: https://redux-toolkit.js.org/rtk-query/usage/mutations#frequently-used-mutation-hook-return-values
 */
export interface FormSubmitResult {
	data: any,
	error: any,
	isLoading: boolean,
	isSuccess: boolean,
	isError: boolean,
}
