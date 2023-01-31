/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

"use strict";

const React = require("react");

module.exports = function useTextInput({ name, Name }, {
	defaultValue = "",
	dontReset = false,
	validator,
	initValidation: _initValidation
} = {}) {

	const [text, setText] = React.useState(defaultValue);
	const textRef = React.useRef(null);

	const initValidation = React.useRef(_initValidation); // memoize forever
	const [validation, setValidation] = React.useState(_initValidation ?? "");
	const [_isValidating, startValidation] = React.useTransition();
	const valid = validation == "";

	const isFirstUpdate = React.useRef(true);

	function onChange(e) {
		let input = e.target.value;
		setText(input);

		if (validator) {
			startValidation(() => {
				let validatorMsg = (isFirstUpdate.current && initValidation.current)
					? initValidation.current
					: validator(input);

				if (isFirstUpdate.current) {
					isFirstUpdate.current = false;
					return; // No need to update anything
				}

				setValidation(validatorMsg);
			});
		}
	}

	function reset() {
		if (!dontReset) {
			setText(defaultValue);
		}
	}

	React.useEffect(() => {
		if (validator && textRef.current) {
			textRef.current.setCustomValidity(validation);
		}
	}, [validation, validator]);

	// Array / Object hybrid, for easier access in different contexts
	return Object.assign([
		onChange,
		reset,
		{
			[name]: text,
			[`${name}Ref`]: textRef,
			[`set${Name}`]: setText,
			[`${name}Valid`]: valid,
		}
	], {
		onChange,
		reset,
		name,
		value: text,
		ref: textRef,
		setter: setText,
		valid,
		validate: () => setValidation(validator(text)),
		hasChanged: () => text != defaultValue
	});
};