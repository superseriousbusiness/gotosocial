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

import { gtsApi } from "../../gts-api";
import { RootState } from "../../../../redux/store";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { DomainPerm, ExportDomainPermsParams } from "../../../types/domain-permission";

interface _exportProcess {
	transformEntry: (_entry: DomainPerm) => any;
	stringify: (_list: any[]) => string;
	extension: string;
	mime: string;
}

/**
 * Derive process functions and metadata
 * from provided export request form.
 * 
 * @param formData 
 * @returns 
 */
function exportProcess(formData: ExportDomainPermsParams): _exportProcess {
	if (formData.exportType == "json") {
		return {
			transformEntry: (entry) => ({
				domain: entry.domain,
				public_comment: entry.public_comment,
				obfuscate: entry.obfuscate
			}),
			stringify: (list) => JSON.stringify(list),
			extension: ".json",
			mime: "application/json"
		};
	}
	
	if (formData.exportType == "csv") {
		return {
			transformEntry: (entry) => [
				entry.domain,               // domain
				"suspend",                  // severity
				false,                      // reject_media
				false,                      // reject_reports
				entry.public_comment ?? "", // public_comment
				entry.obfuscate ?? false    // obfuscate
			],
			stringify: (list) => csvUnparse({
				fields: [
					"#domain",
					"#severity",
					"#reject_media",
					"#reject_reports",
					"#public_comment",
					"#obfuscate",
				],
				data: list
			}),
			extension: ".csv",
			mime: "text/csv"
		};
	}

	// Fall back to plain text export.
	return {
		transformEntry: (entry) => entry.domain,
		stringify: (list) => list.join("\n"),
		extension: ".txt",
		mime: "text/plain"
	};
}

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({		
		exportDomainList: build.mutation<string | null, ExportDomainPermsParams>({
			async queryFn(formData, api, _extraOpts, fetchWithBQ) {
				// Fetch domain perms from relevant endpoint.
				// We could have used 'useDomainBlocksQuery'
				// or 'useDomainAllowsQuery' for this, but
				// we want the untransformed array version.
				const permsRes = await fetchWithBQ({ url: `/api/v1/admin/domain_${formData.permType}s` });
				if (permsRes.error) {
					return { error: permsRes.error as FetchBaseQueryError };
				}

				// Process perms into desired export format.
				const process = exportProcess(formData); 
				const transformed = (permsRes.data as DomainPerm[]).map(process.transformEntry);
				const exportAsString = process.stringify(transformed);

				if (formData.action == "export") {
					// Data will just be exported
					// to the domains text field.
					return { data: exportAsString };
				}

				// File export has been requested.
				// Parse filename to something like:
				// `example.org-blocklist-2023-10-09.json`.
				const state = api.getState() as RootState;
				const instanceUrl = state.login.instanceUrl?? "unknown";
				const domain = new URL(instanceUrl).host;
				const date = new Date();
				const filename = [
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

				// js-file-download handles the
				// nitty gritty for us, so we can
				// just return null data.
				return { data: null };
			}
		}),
	})
});

/**
 * Makes a GET to `/api/v1/admin/domain_{perm_type}s`
 * and exports the result in the requested format.
 * 
 * Return type will be string if `action` is "export",
 * else it will be null, since the file downloader handles
 * the rest of the request then.
 */
const useExportDomainListMutation = extended.useExportDomainListMutation;

export { useExportDomainListMutation };
