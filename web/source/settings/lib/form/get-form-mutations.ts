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

import { FormInputHook, HookedForm } from "./types";

export default function getFormMutations(
	form: HookedForm,
	{ changedOnly }: { changedOnly: boolean },
): {
	updatedFields: FormInputHook<any>[];
	mutationData: {
		[k: string]: any;
	};
} {
	const updatedFields: FormInputHook[] = [];
	const mutationData: Array<[string, any]> = [];
	
	Object.values(form).forEach((field) => {
		if (field.nosubmit) {
			// Completely ignore
			// this field.
			return;
		}
		
		if ("selectedValues" in field) {
			// (Field)ArrayInputHook.
			const selected = field.selectedValues();
			if (!changedOnly || selected.length > 0) {
				updatedFields.push(field);
				mutationData.push([field.name, selected]);
			}
		} else if (!changedOnly || field.hasChanged()) {
			updatedFields.push(field);
			mutationData.push([field.name, field.value]);
		}
	});

	return {
		updatedFields,
		mutationData: Object.fromEntries(mutationData),
	};
}
