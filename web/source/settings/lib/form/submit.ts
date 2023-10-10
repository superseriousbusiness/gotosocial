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

import { try as bbTry } from "bluebird";
import getFormMutations from "./get-form-mutations";
import { HookedForm } from "./types";
import { FormEvent, SyntheticEvent, useRef } from "react";

export default function useFormSubmit(
	form: HookedForm,
	mutationQuery,
	{
		changedOnly = true,
		onFinish 
	}: {
		changedOnly: boolean;
		onFinish: ((_res: any) => void) | undefined;
	}
) {
	if (!Array.isArray(mutationQuery)) {
		throw "useFormSubmit: mutationQuery was not an Array. Is a valid useMutation RTK Query provided?";
	}
	const [runMutation, result] = mutationQuery;
	let action: any;
	let usedAction = useRef(action);
	
	const submitForm = (e: SyntheticEvent<HTMLFormElement, SubmitEvent>) => {
		if (e.preventDefault) {
			e.preventDefault();
			
			if (e.nativeEvent.submitter) {
				action = (e.nativeEvent.submitter as unknown as { name }).name;
			}
		} else {
			action = e;
		}

		if (action == "") {
			action = undefined;
		}
		usedAction = action;
		// transform the field definitions into an object with just their values 

		const { mutationData, updatedFields } = getFormMutations(form, { changedOnly });

		if (updatedFields.length == 0) {
			return;
		}

		mutationData.action = action;

		return bbTry(() => {
			return runMutation(mutationData);
		}).then((res) => {
			if (onFinish) {
				return onFinish(res);
			}
		});
	}
	
	return [ 
		submitForm,
		{
			...result,
			action: usedAction.current
		}
	];
};
