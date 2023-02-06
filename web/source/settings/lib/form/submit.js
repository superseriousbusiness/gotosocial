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
const syncpipe = require("syncpipe");

module.exports = function useFormSubmit(form, mutationQuery, { changedOnly = true } = {}) {
	if (!Array.isArray(mutationQuery)) {
		throw new ("useFormSubmit: mutationQuery was not an Array. Is a valid useMutation RTK Query provided?");
	}
	const [runMutation, result] = mutationQuery;
	const usedAction = React.useRef(null);
	return [
		function submitForm(e) {
			let action;
			if (e?.preventDefault) {
				e.preventDefault();
				action = e.nativeEvent.submitter.name;
			} else {
				action = e;
			}

			if (action == "") {
				action = undefined;
			}
			usedAction.current = action;
			// transform the field definitions into an object with just their values 
			let updatedFields = [];
			const mutationData = syncpipe(form, [
				(_) => Object.values(_),
				(_) => _.map((field) => {
					if (field.selectedValues != undefined) {
						let selected = field.selectedValues();
						if (!changedOnly || selected.length > 0) {
							updatedFields.push(field);
							return [field.name, selected];
						}
					} else if (!changedOnly || field.hasChanged()) {
						updatedFields.push(field);
						return [field.name, field.value];
					}
					return null;
				}),
				(_) => _.filter((value) => value != null),
				(_) => Object.fromEntries(_)
			]);

			mutationData.action = action;

			return runMutation(mutationData);
		},
		{
			...result,
			action: usedAction.current
		}
	];
};