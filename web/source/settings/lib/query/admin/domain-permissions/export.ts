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

import fileDownload from "js-file-download";
import { unparse as csvUnparse } from "papaparse";
import { try as bbTry } from "bluebird";

import { unwrapRes } from "../../lib";
import { gtsApi } from "../../gts-api";
import { RootState } from "../../../../redux/store";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({		
		exportDomainList: build.mutation({
			queryFn: (formData, api, _extraOpts, baseQuery) => {
				let process;

				if (formData.exportType == "json") {
					process = {
						transformEntry: (entry) => ({
							domain: entry.domain,
							public_comment: entry.public_comment,
							obfuscate: entry.obfuscate
						}),
						stringify: (list) => JSON.stringify(list),
						extension: ".json",
						mime: "application/json"
					};
				} else if (formData.exportType == "csv") {
					process = {
						transformEntry: (entry) => [
							entry.domain,
							"suspend", // severity
							false, // reject_media
							false, // reject_reports
							entry.public_comment,
							entry.obfuscate ?? false
						],
						stringify: (list) => csvUnparse({
							fields: "#domain,#severity,#reject_media,#reject_reports,#public_comment,#obfuscate".split(","),
							data: list
						}),
						extension: ".csv",
						mime: "text/csv"
					};
				} else {
					process = {
						transformEntry: (entry) => entry.domain,
						stringify: (list) => list.join("\n"),
						extension: ".txt",
						mime: "text/plain"
					};
				}

				return bbTry(() => {
					return baseQuery({
						url: `/api/v1/admin/domain_blocks`
					});
				}).then(unwrapRes).then((blockedInstances) => {
					return blockedInstances.map(process.transformEntry);
				}).then((exportList) => {
					return process.stringify(exportList);
				}).then((exportAsString) => {
					if (formData.action == "export") {
						return {
							data: exportAsString
						};
					} else if (formData.action == "export-file") {
						const state = api.getState() as RootState;
						const instanceUrl = state.oauth.instanceUrl?? "unknown";

						let domain = new URL(instanceUrl).host;
						let date = new Date();

						let filename = [
							domain,
							"blocklist",
							date.getFullYear(),
							(date.getMonth() + 1).toString().padStart(2, "0"),
							date.getDate().toString().padStart(2, "0"),
						].join("-");

						fileDownload(
							exportAsString,
							filename + process.extension,
							process.mime
						);
					}
					return { data: null };
				}).catch((e) => {
					return { error: e };
				});
			}
		}),
	})
});

export const {
	useExportDomainListMutation,
} = extended;
