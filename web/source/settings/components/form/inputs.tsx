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
import { FormInputHook, FormInputHookWithOptions } from "../../lib/form/types";

function TextInput({ label, field, ...inputProps }) {
	const { onChange, value, ref } = field;

	return (
		<div className={`form-field text${field.valid ? "" : " invalid"}`}>
			<label>
				{label}
				<input
					type="text"
					{...{ onChange, value, ref }}
					{...inputProps}
				/>
			</label>
		</div>
	);
}

function TextArea({ label, field, ...inputProps }) {
	const { onChange, value, ref } = field;

	return (
		<div className="form-field textarea">
			<label>
				{label}
				<textarea
					type="text"
					{...{ onChange, value, ref }}
					{...inputProps}
				/>
			</label>
		</div>
	);
}

function FileInput({ label, field, ...inputProps }) {
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
					{...{ onChange, ref }}
					{...inputProps}
				/>
			</label>
		</div>
	);
}

function Checkbox({ label, field, ...inputProps }) {
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

export interface SelectProps {
	field: FormInputHook<any>;
	label?: string;
	options,
	inputProps: Object;
	children,
}

function Select({ label, field, options, children, ...inputProps }: SelectProps) {
	const { onChange, value, ref } = field;

	return (
		<div className="form-field select">
			<label>
				{label} {children}
				<select
					{...{ onChange, value, ref }}
					{...inputProps}
				>
					{options}
				</select>
			</label>
		</div>
	);
}

export interface RadioGroupProps {
	field: FormInputHookWithOptions<string>;
	label?: string;
	inputProps: Object;
}

function RadioGroup({ field, label, ...inputProps }: RadioGroupProps) {
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
						{...inputProps}
					/>
					{radioLabel}
				</label>
			))}
			{label}
		</div>
	);
}

export default {
	TextInput,
	TextArea,
	FileInput,
	Checkbox,
	Select,
	RadioGroup
};