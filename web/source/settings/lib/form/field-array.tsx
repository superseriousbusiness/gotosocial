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

import { useRef, useMemo } from "react";

import getFormMutations from "./get-form-mutations";

import type {
	HookName,
	HookOpts,
} from "./types";

function parseFields(entries, length) {
	const fields: any[] = [];

	for (let i = 0; i < length; i++) {
		if (entries[i] != undefined) {
			fields[i] = Object.assign({}, entries[i]);
		} else {
			fields[i] = {};
		}
	}

	return fields;
}

export default function useArrayInput(
	{ name }: HookName,
	{
		initialValue,
		length = 0,
	}: HookOpts<any[]>,
) {
	const fields = useRef({});

	const value = useMemo(() => parseFields(initialValue, length), [initialValue, length]);

	return {
		name,
		value,
		ctx: fields.current,
		maxLength: length,
		selectedValues() {
			// if any form field changed, we need to re-send everything
			const hasUpdate = Object.values(fields.current).some((fieldSet) => {
				const { updatedFields } = getFormMutations(fieldSet, { changedOnly: true });
				return updatedFields.length > 0;
			});
			if (hasUpdate) {
				return Object.values(fields.current).map((fieldSet) => {
					return getFormMutations(fieldSet, { changedOnly: false }).mutationData;
				});
			} else {
				return [];
			}
		}
	};
}