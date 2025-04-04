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

import { gtsApi } from "../gts-api";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		twoFactorQRCodeURI: build.mutation<string, void>({
			query: () => ({
				url: `/api/v1/user/2fa/qruri`,
				acceptContentType: "text/plain",
			})
		}),

		twoFactorQRCodePng: build.mutation<string, void>({
			async queryFn(_arg, _api, _extraOpts, fetchWithBQ) {
				const blobRes = await fetchWithBQ({
					url: `/api/v1/user/2fa/qr.png`,
					acceptContentType: "image/png",
				});
				if (blobRes.error) {
					return { error: blobRes.error as FetchBaseQueryError };
				}

				if (blobRes.meta?.response?.status !== 200) {
					return { error: blobRes.data };
				}

				const blob = blobRes.data as Blob;
				const url = URL.createObjectURL(blob);

				return { data: url };
			},
		}),

		twoFactorEnable: build.mutation<string[], { password: string }>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/user/2fa/enable`,
				asForm: true,
				body: formData,
				discardEmpty: true
			})
		}),

		twoFactorDisable: build.mutation<void, { password: string }>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/user/2fa/disable`,
				asForm: true,
				body: formData,
				discardEmpty: true,
				acceptContentType: "*/*",
			}),
			invalidatesTags: ["User"]
		}),
	})
});

export const {
	useTwoFactorQRCodeURIMutation,
	useTwoFactorQRCodePngMutation,
	useTwoFactorEnableMutation,
	useTwoFactorDisableMutation,
} = extended;
