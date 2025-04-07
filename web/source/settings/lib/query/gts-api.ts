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

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
	BaseQueryFn,
	FetchArgs,
	FetchBaseQueryError,
} from '@reduxjs/toolkit/query/react';
import { serialize as serializeForm } from "object-to-formdata";
import type { FetchBaseQueryMeta } from "@reduxjs/toolkit/dist/query/fetchBaseQuery";
import type { RootState } from '../../redux/store';
import { InstanceV1 } from '../types/instance';

/**
 * GTSFetchArgs extends standard FetchArgs used by
 * RTK Query with a couple helpers of our own.
 */
export interface GTSFetchArgs extends FetchArgs {
	/**
	 * If provided, will be used as base URL. Else,
	 * will fall back to authorized instance as baseUrl.
	 */
	baseUrl?: string;
	/**
	 * If true, and no args.body is set, or args.body is empty,
	 * then a null response will be returned from the API call.  
	 */
	discardEmpty?: boolean;
	/**
	 * If true, then args.body will be serialized
	 * as FormData before submission. 
	 */
	asForm?: boolean;
	/**
	 * If set, then Accept header will
	 * be set to the provided contentType.
	 */
	acceptContentType?: string;
}

/**
 * gtsBaseQuery wraps the redux toolkit fetchBaseQuery with some helper functionality.
 * 
 * For an explainer of what's happening in this function, see:
 * - https://redux-toolkit.js.org/rtk-query/usage/customizing-queries#customizing-queries-with-basequery
 * - https://redux-toolkit.js.org/rtk-query/usage/customizing-queries#constructing-a-dynamic-base-url-using-redux-state
 * 
 * @param args 
 * @param api 
 * @param extraOptions 
 * @returns 
 */
const gtsBaseQuery: BaseQueryFn<
	string | GTSFetchArgs,
	any,
	FetchBaseQueryError,
	{},
	FetchBaseQueryMeta
> = async (args, api, extraOptions) => {
	// Retrieve state at the moment
	// this function was called.
	const state = api.getState() as RootState;
	const { instanceUrl, token } = state.login;

	// Derive baseUrl dynamically.
	let baseUrl: string | undefined;

	// Assume Accept value of
	// "application/json" by default.
	let accept = "application/json";

	// Check if simple string baseUrl provided
	// as args, or if more complex args provided.
	if (typeof args === "string") {
		baseUrl = args;
	} else {
		if (args.baseUrl != undefined) {
			baseUrl = args.baseUrl;
		} else {
			baseUrl = instanceUrl;
		}

		if (args.discardEmpty) {
			if (args.body == undefined || Object.keys(args.body).length == 0) {
				return { data: null };
			}
		}

		if (args.asForm) {
			args.body = serializeForm(args.body, {
				// Array indices, for profile fields.
				indices: true,
			});
		}

		if (args.acceptContentType !== undefined) {
			accept = args.acceptContentType;
		}

		// Delete any of our extended arguments
		// to avoid confusing fetchBaseQuery.
		delete args.baseUrl;
		delete args.discardEmpty;
		delete args.asForm;
		delete args.acceptContentType;
	}

	if (!baseUrl) {
		return {
			error: {
				status: 400,
				statusText: 'Bad Request',
				data: {"error":"No baseUrl set for request"},
			},
		};
	}

	return fetchBaseQuery({
		baseUrl: baseUrl,
		prepareHeaders: (headers) => {
			if (token != undefined) {
				headers.set('Authorization', token);
			}
			
			headers.set("Accept", accept);
			return headers;
		},
		responseHandler: (response) => {
			switch (true) {
				case (accept === "application/json"):
					// return good old
					// fashioned JSON baby!
					return response.json();
				case (accept.startsWith("image/")):
					// It's an image,
					// return the blob.
					return response.blob();
				default:
					// God knows what it
					// is, just return text.
					return response.text();
			}
		},
	})(args, api, extraOptions);
};

export const gtsApi = createApi({
	reducerPath: "api",
	baseQuery: gtsBaseQuery,
	tagTypes: [
		"Application",
		"Auth",
		"Emoji",
		"Report",
		"Account",
		"InstanceRules",
		"HTTPHeaderAllows",
		"HTTPHeaderBlocks",
		"DefaultInteractionPolicies",
		"InteractionRequest",
		"DomainPermissionDraft",
		"DomainPermissionExclude",
		"DomainPermissionSubscription",
		"TokenInfo",
		"User",
	],
	endpoints: (build) => ({
		instanceV1: build.query<InstanceV1, void>({
			query: () => ({
				url: `/api/v1/instance`
			})
		})
	})
});

/**
 * Query /api/v1/instance to retrieve basic instance information.
 * This endpoint does not require authentication/authorization.
 * TODO: move this to ./instance.
 */
const useInstanceV1Query = gtsApi.useInstanceV1Query;

export { useInstanceV1Query };
