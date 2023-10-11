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

import getFormMutations from "./get-form-mutations";

import { useRef } from "react";

import type {
	MutationTrigger,
	UseMutationStateResult,
} from "@reduxjs/toolkit/dist/query/react/buildHooks";

import type {
	FormSubmitEvent,
	FormSubmitFunction,
	FormSubmitResult,
	HookedForm,
} from "./types";

declare interface UseFormSubmitOptions {
	changedOnly: boolean;
	onFinish?: ((_res: any) => void);
}

export default function useFormSubmit(
	form: HookedForm,
	mutationQuery: readonly [MutationTrigger<any>, UseMutationStateResult<any, any>],
	opts: UseFormSubmitOptions = {
		changedOnly: true,
		onFinish: undefined,
	}
): [ FormSubmitFunction, FormSubmitResult ] {
	if (!Array.isArray(mutationQuery)) {
		throw "useFormSubmit: mutationQuery was not an Array. Is a valid useMutation RTK Query provided?";
	}

	const { changedOnly, onFinish } = opts;
	const [runMutation, mutationResult] = mutationQuery;
	const usedAction = useRef<FormSubmitEvent>(undefined);
	
	const submitForm = async(e: FormSubmitEvent) => {
		let action: FormSubmitEvent;
		
		if (typeof e === "string") {
			action = e !== "" ? e : undefined;
		} else if (e !== undefined) {
			e.preventDefault();
			if (e.nativeEvent.submitter) {
				action = (e.nativeEvent.submitter as Object)["name"];
			}
		}

		if (action !== undefined) {
			usedAction.current = action;
		}

		// Transform the field definitions into an object with just their values.
		const { mutationData, updatedFields } = getFormMutations(form, { changedOnly });

		if (updatedFields.length == 0) {
			// No updated fields.
			// Nothing to do.
			return;
		}

		mutationData.action = action;

		try {
			const res = await runMutation(mutationData);
			if (onFinish) {
				return onFinish(res);
			}
		} catch (e) {
			// eslint-disable-next-line no-console
			console.error(`caught error running mutation: ${e}`);
		}
	};
	
	return [
		submitForm,
		{
			...mutationResult,
			action: usedAction.current
		}
	];
}
