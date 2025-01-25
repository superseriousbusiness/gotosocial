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

import { ComboboxState } from "ariakit";
import React from "react";

import {
	ChangeEventHandler,
	Dispatch,
	RefObject,
	SetStateAction,
	SyntheticEvent,
} from "react";

export interface CreateHookNames {
	name: string;
	Name: string;
}

export interface HookOpts<T = any> {
	initialValue?: T,
	defaultValue?: T,
	
	/**
	 * If true, don't submit this field as
	 * part of a mutation query's body.
	 * 
	 * Useful for 'internal' form fields.
	 */
	nosubmit?: boolean,
	dontReset?: boolean,
	validator?,
	showValidation?: boolean,
	initValidation?: string,
	length?: number;
	options?: { [_: string]: string },
	withPreview?: boolean,
	maxSize?,
	initialInfo?: string;
	valueSelector?: Function,
	source?,

	// checklist input types
	entries?: any[];
	uniqueKey?: string;
}

export type CreateHook = (
	name: CreateHookNames,
	opts: HookOpts,
) => FormInputHook;

export interface FormInputHook<T = any> {
	/**
	 * Name of this FormInputHook, as provided
	 * in the UseFormInputHook options.
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
	 * Return true if the values of this hook is considered 
	 * to have been changed from the default / initial value.
	 */
	hasChanged: () => boolean;

	/**
	 * If true, don't submit this field as
	 * part of a mutation query's body.
	 * 
	 * Useful for 'internal' form fields.
	 */
	nosubmit?: boolean;
}

interface _withReset {
	reset: () => void;
}

interface _withOnChange {
	onChange: ChangeEventHandler;
}

interface _withSetter<T> {
	setter: Dispatch<SetStateAction<T>>;
}

interface _withValidate {
	valid: boolean;
	validate: () => void;
}

interface _withRef {
	ref: RefObject<HTMLElement>;
}

interface _withFile {
	previewValue?: string;
	infoComponent: React.JSX.Element;
}

interface _withComboboxState {
	state: ComboboxState;
}

interface _withNew {
	isNew: boolean;
	setIsNew: Dispatch<SetStateAction<boolean>>;
}

interface _withSelectedValues {
	selectedValues: () => string[];
}

interface _withSelectedFieldValues {
	selectedValues: () => {
		[_: string]: any;
	}[]
}

interface _withCtx {
	ctx
}

interface _withMaxLength {
	maxLength: number;
}

interface _withOptions {
	options: { [_: string]: string };
}

interface _withToggleAll {
	toggleAll: _withRef & _withOnChange
}

interface _withSomeSelected {
	someSelected: boolean;
}

interface _withUpdateMultiple {
	updateMultiple: (entries: [key: string, value: Partial<Checkable>][]) => void;
}

export interface TextFormInputHook extends FormInputHook<string>,
	_withSetter<string>,
	_withOnChange,
	_withReset,
	_withValidate,
	_withRef {}

export interface NumberFormInputHook extends FormInputHook<number>,
	_withSetter<number>,
	_withOnChange,
	_withReset,
	_withValidate,
	_withRef {}

export interface RadioFormInputHook extends FormInputHook<string>,
	_withSetter<string>,
	_withOnChange,
	_withOptions,
	_withReset {}

export interface FileFormInputHook extends FormInputHook<File | undefined>,
	_withOnChange,
	_withReset,
	Partial<_withRef>,
	_withFile {}

export interface BoolFormInputHook extends FormInputHook<boolean>,
	_withSetter<boolean>,
	_withOnChange,
	_withReset {}

export interface ComboboxFormInputHook extends FormInputHook<string>,
	_withSetter<string>,
	_withComboboxState,
	_withNew,
	_withReset {}

export interface ArrayInputHook extends FormInputHook<HookedForm[]>,
	_withSelectedValues,
	_withMaxLength,
	_withCtx {}

export interface FieldArrayInputHook extends FormInputHook<HookedForm[]>,
	_withSelectedFieldValues,
	_withMaxLength,
	_withCtx {}

export interface Checkable {
	key: string;
	checked?: boolean;
}

export interface ChecklistInputHook<T = Checkable> extends FormInputHook<{[k: string]: T}>,
	_withReset,
	_withToggleAll,
	_withSelectedFieldValues,
	_withSomeSelected,
	_withUpdateMultiple {
		// Uses its own funky onChange handler.
		onChange: (key: any, value: any) => void
	}

export type AnyFormInputHook = 
	FormInputHook |
	TextFormInputHook |
	RadioFormInputHook |
	FileFormInputHook |
	BoolFormInputHook |
	ComboboxFormInputHook |
	FieldArrayInputHook |
	ChecklistInputHook;

export interface HookedForm {
	[_: string]: AnyFormInputHook
}

/**
 * Parameters for FormSubmitFunction.
 */
export type FormSubmitEvent = (string | SyntheticEvent<HTMLFormElement, Partial<SubmitEvent>> | undefined | void)


/**
 * Shadows "trigger" function for useMutation, but can also
 * be passed to onSubmit property of forms as a handler.
 * 
 * See: https://redux-toolkit.js.org/rtk-query/usage/mutations#mutation-hook-behavior
 */
export type FormSubmitFunction = ((_e: FormSubmitEvent) => void)

/**
 * Shadows redux mutation hook return values.
 * 
 * See: https://redux-toolkit.js.org/rtk-query/usage/mutations#frequently-used-mutation-hook-return-values
 */
export interface FormSubmitResult {
	/**
	 * Action used to submit the form, if any.
	 */
	action: FormSubmitEvent;
	data: any;
	error: any;
	isLoading: boolean;
	isSuccess: boolean;
	isError: boolean;
	reset: () => void;
}

/**
 * Shadows redux query hook return values.
 * 
 * See: https://redux-toolkit.js.org/rtk-query/usage/queries#frequently-used-query-hook-return-values
 */
export type FormWithDataQuery = (_queryArg: any) => {
	data?: any;
	isLoading: boolean;
	isError: boolean;
	error?: any;
}
