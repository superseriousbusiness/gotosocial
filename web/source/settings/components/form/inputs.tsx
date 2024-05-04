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

import React from "react";

import type {
	ReactNode,
	RefObject,
} from "react";

import type {
	FileFormInputHook,
	RadioFormInputHook,
	TextFormInputHook,
} from "../../lib/form/types";

export interface TextInputProps extends React.DetailedHTMLProps<
	React.InputHTMLAttributes<HTMLInputElement>,
	HTMLInputElement
> {
	label?: ReactNode;
	field: TextFormInputHook;
}

export function TextInput({label, field, ...props}: TextInputProps) {
	const { onChange, value, ref } = field;

	return (
		<div className={`form-field text${field.valid ? "" : " invalid"}`}>
			<label>
				{label}
				<input
					onChange={onChange}
					value={value}
					ref={ref as RefObject<HTMLInputElement>}
					{...props}
				/>
			</label>
		</div>
	);
}

export interface TextAreaProps extends React.DetailedHTMLProps<
	React.TextareaHTMLAttributes<HTMLTextAreaElement>,
	HTMLTextAreaElement
> {
	label?: ReactNode;
	field: TextFormInputHook;
}

export function TextArea({label, field, ...props}: TextAreaProps) {
	const { onChange, value, ref } = field;

	return (
		<div className="form-field textarea">
			<label>
				{label}
				<textarea
					onChange={onChange}
					value={value}
					ref={ref as RefObject<HTMLTextAreaElement>}
					{...props}
				/>
			</label>
		</div>
	);
}

export interface FileInputProps extends React.DetailedHTMLProps<
	React.InputHTMLAttributes<HTMLInputElement>,
	HTMLInputElement
> {
	label?: ReactNode;
	field: FileFormInputHook;
}

export function FileInput({ label, field, ...props }: FileInputProps) {
	const { onChange, ref, infoComponent } = field;

	return (
		<div className="form-field file">
			<label>
				<div className="label">{label}</div>
				<div className="file-input button">Browse</div>
				{infoComponent}
				{/* <a onClick={removeFile("header")}>remove</a> */}
				<input
					type="file"
					className="hidden"
					onChange={onChange}
					ref={ref ? ref as RefObject<HTMLInputElement> : undefined}
					{...props}
				/>
			</label>
		</div>
	);
}

export function Checkbox({ label, field, ...inputProps }) {
	const { onChange, value } = field;

	return (
		<div className="form-field checkbox">
			<label>
				<input
					type="checkbox"
					checked={value}
					onChange={onChange}
					{...inputProps}
				/> {label}
			</label>
		</div>
	);
}

export interface SelectProps extends React.DetailedHTMLProps<
	React.SelectHTMLAttributes<HTMLSelectElement>,
	HTMLSelectElement
> {
	label?: ReactNode;
	field: TextFormInputHook;
	children?: ReactNode;
	options: React.JSX.Element;
}

export function Select({ label, field, children, options, ...props }: SelectProps) {
	const { onChange, value, ref } = field;

	return (
		<div className="form-field select">
			<label>
				{label}
				{children}
				<select
					onChange={onChange}
					value={value}
					ref={ref as RefObject<HTMLSelectElement>}
					{...props}
				>
					{options}
				</select>
			</label>
		</div>
	);
}

export interface RadioGroupProps extends React.DetailedHTMLProps<
	React.InputHTMLAttributes<HTMLInputElement>,
	HTMLInputElement
> {
	label?: ReactNode;
	field: RadioFormInputHook;
}

export function RadioGroup({ label, field, ...props }: RadioGroupProps) {
	return (
		<div className="form-field radio">
			{Object.entries(field.options).map(([value, radioLabel]) => (
				<label key={value}>
					<input
						type="radio"
						name={field.name}
						value={value}
						checked={field.value == value}
						onChange={field.onChange}
						{...props}
					/>
					{radioLabel}
				</label>
			))}
			{label}
		</div>
	);
}
