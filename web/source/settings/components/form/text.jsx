/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

module.exports = function useTextInput({name, Name}, {validator} = {}) {
	const [text, setText] = React.useState("");
	const textRef = React.useRef(null);

	function onChange(e) {
		let input = e.target.value;
		setText(input);

		if (validator) {
			validator(input);
		}
	}

	function reset() {
		setText("");
	}

	React.useEffect(() => {
		if (validator) {
			textRef.current.setCustomValidity(validator(text));
			textRef.current.reportValidity();
		}
	}, [text, textRef, validator]);

	return [
		onChange,
		reset,
		{
			[name]: text,
			[`${name}Ref`]: textRef,
			[`set${Name}`]: setText
		}
	];
};