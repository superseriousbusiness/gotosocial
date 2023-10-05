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

const syncpipe = require("syncpipe");

module.exports = function getFormMutations(form, { changedOnly }) {
	let updatedFields = [];
	return {
		updatedFields,
		mutationData: syncpipe(form, [
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
		])
	};
};