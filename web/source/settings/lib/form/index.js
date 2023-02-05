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
const getByDot = require("get-by-dot").default;

function capitalizeFirst(str) {
	return str.slice(0, 1).toUpperCase + str.slice(1);
}

function selectorByKey(key) {
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

function makeHook(hookFunction) {
	return function (name, opts = {}) {
		// for dynamically generating attributes like 'setName'
		const Name = React.useMemo(() => capitalizeFirst(name), [name]);

		const selector = React.useMemo(() => selectorByKey(name), [name]);
		const valueSelector = opts.valueSelector ?? selector;

		opts.initialValue = React.useMemo(() => {
			if (opts.source == undefined) {
				return opts.defaultValue;
			} else {
				return valueSelector(opts.source) ?? opts.defaultValue;
			}
		}, [opts.source, opts.defaultValue, valueSelector]);

		const hook = hookFunction({ name, Name }, opts);

		return Object.assign(hook, {
			name, Name,
		});
	};
}

module.exports = {
	useTextInput: makeHook(require("./text")),
	useFileInput: makeHook(require("./file")),
	useBoolInput: makeHook(require("./bool")),
	useRadioInput: makeHook(require("./radio")),
	useComboBoxInput: makeHook(require("./combo-box")),
	useCheckListInput: makeHook(require("./check-list")),
	useValue: function (name, value) {
		return {
			name,
			value,
			hasChanged: () => true // always included
		};
	}
};