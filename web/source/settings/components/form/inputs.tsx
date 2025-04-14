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

import React, { useRef } from "react";

import type {
	ReactNode,
	RefObject,
} from "react";

import type {
	FileFormInputHook,
	NumberFormInputHook,
	RadioFormInputHook,
	TextFormInputHook,
} from "../../lib/form/types";
import { nanoid } from "nanoid";

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

export interface NumberInputProps extends React.DetailedHTMLProps<
	React.InputHTMLAttributes<HTMLInputElement>,
	HTMLInputElement
> {
	label?: ReactNode;
	field: NumberFormInputHook;
}

export function NumberInput({label, field, ...props}: NumberInputProps) {
	const { onChange, value, ref } = field;

	return (
		<div className={`form-field number${field.valid ? "" : " invalid"}`}>
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
	const ref = useRef<HTMLInputElement>(null);
	const { onChange, infoComponent } = field;
	const id = nanoid();
	const onClick = (e) => {
		e.preventDefault();
		ref.current?.click();
	};

	return (
		<div className="form-field file">
			<label
				className="label-wrapper"
				htmlFor={id}
				tabIndex={0}
				onClick={onClick}
				onKeyDown={(e) => {
					if (e.key === "Enter") {
						e.preventDefault();
						onClick(e);
					}
				}}
				role="button"
			>
				<div className="label-label">
					{label}
				</div>
				<div className="label-button">
					<div className="file-input button">Browse</div>
				</div>
			</label>
			<input
				id={id}
				type="file"
				className="hidden"
				onChange={onChange}
				ref={ref}
				{...props}
			/>
			{infoComponent}
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

	/**
	 * Optional callback function that is
	 * triggered along with the select's onChange.
	 * 
	 * _selectValue is the current value of
	 * the select after onChange is triggered.
	 * 
	 * @param _selectValue 
	 * @returns 
	 */
	onChangeCallback?: (_selectValue: string | undefined) => void;
}

export function Select({
	label,
	field,
	children,
	options,
	onChangeCallback,
	...props
}: SelectProps) {
	const { onChange, value, ref } = field;

	return (
		<div className="form-field select">
			<label>
				{label}
				{children}
				<div className="select-wrapper">
					<select
						onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
							onChange(e);
							if (onChangeCallback !== undefined) {
								onChangeCallback(e.target.value);
							}
						}}
						value={value}
						ref={ref as RefObject<HTMLSelectElement>}
						{...props}
					>
						{options}
					</select>
				</div>
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
