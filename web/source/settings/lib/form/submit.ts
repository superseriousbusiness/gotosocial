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

interface UseFormSubmitOptions {
	/**
	 * Include only changed fields when submitting the form.
	 * If no fields have been changed, submit will be a noop.
	 */
	changedOnly: boolean;
	/**
	 * Optional function to run when the form has been sent
	 * and a response has been returned from the server.
	 */
	onFinish?: ((_res: any) => void);
	/**
	 * Can be optionally used to modify the final mutation argument from the
	 * gathered mutation data before it's passed into the trigger function.
	 * 
	 * Useful if the mutation trigger function takes not just a simple key/value
	 * object but a more complicated object.
	 */
	customizeMutationArgs?: (_mutationData: { [k: string]: any }) => any;
}

/**
 * Parse changed values from the hooked form into a request
 * body, and submit it using the given mutation trigger.
 * 
 * This function basically wraps RTK Query's submit methods to
 * work with our hooked form interface.
 * 
 * An `onFinish` callback function can be provided, which will
 * be executed on a **successful** run of the given MutationTrigger,
 * with the mutation result passed into it.
 * 
 * If `changedOnly` is false, then **all** fields of the given HookedForm
 * will be submitted to the mutation endpoint, not just changed ones.
 * 
 * The returned function and result can be triggered and read
 * from just like an RTK Query mutation hook result would be.
 * 
 * See: https://redux-toolkit.js.org/rtk-query/usage/mutations#mutation-hook-behavior
 */
export default function useFormSubmit(
	form: HookedForm,
	mutationQuery: readonly [MutationTrigger<any>, UseMutationStateResult<any, any>],
	opts: UseFormSubmitOptions = { changedOnly: true }
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
			if (e !== "") {
				// String action name was provided.
				action = e;
			} else {
				// Empty string action name was provided.
				action = undefined;
			}
		} else if (e) {
			// Submit event action was provided.
			e.preventDefault();
			if (e.nativeEvent.submitter) {
				// We want the name of the element that was invoked to submit this form,
				// which will be something that extends HTMLElement, though we don't know
				// what at this point. If it's an empty string, fall back to undefined.
				// 
				// See: https://developer.mozilla.org/en-US/docs/Web/API/SubmitEvent/submitter
				action = (e.nativeEvent.submitter as Object as { name: string }).name || undefined;
			} else {
				// No submitter defined. Fall back
				// to just use the FormSubmitEvent.
				action = e;
			}
		} else {
			// Void or null or something
			// else was provided.
			action = undefined;
		}

		usedAction.current = action;

		// Transform the hooked form into an object.
		let {
			mutationData,
			updatedFields,
		} = getFormMutations(form, { changedOnly });
		
		// If there were no updated fields according to
		// the form parsing then there's nothing for us
		// to do, since remote and desired state match.
		if (updatedFields.length == 0) {
			return;
		}

		// Final tweaks on the mutation
		// argument before triggering it.
		mutationData.action = action;
		if (opts.customizeMutationArgs) {
			mutationData = opts.customizeMutationArgs(mutationData);
		}

		try {
			const res = await runMutation(mutationData);
			if (onFinish) {
				onFinish(res);
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
