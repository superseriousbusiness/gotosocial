/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

"use strict";

const Promise = require("bluebird");
const isValidDomain = require("is-valid-domain");
const fileDownload = require("js-file-download");

const {
	replaceCacheOnMutation,
	domainListToObject,
	unwrapRes
} = require("../lib");

function parseDomainList(list) {
	if (list[0] == "[") {
		return JSON.parse(list);
	} else {
		return list.split("\n").map((line) => {
			let domain = line.trim();
			let valid = true;
			if (domain.startsWith("http")) {
				try {
					domain = new URL(domain).hostname;
				} catch (e) {
					valid = false;
				}
			}
			return domain.length > 0
				? { domain, valid }
				: null;
		}).filter((a) => a); // not `null`
	}
}

function validateDomainList(list) {
	list.forEach((entry) => {
		entry.valid = (entry.valid !== false) && isValidDomain(entry.domain, { wildcard: true, allowUnicode: true });
		entry.checked = entry.valid;
	});

	return list;
}

function deduplicateDomainList(list) {
	let domains = new Set();
	return list.filter((entry) => {
		if (domains.has(entry.domain)) {
			return false;
		} else {
			domains.add(entry.domain);
			return true;
		}
	});
}

module.exports = (build) => ({
	processDomainList: build.mutation({
		queryFn: (formData) => {
			return Promise.try(() => {
				if (formData.domains == undefined || formData.domains.length == 0) {
					throw "No domains entered";
				}
				return parseDomainList(formData.domains);
			}).then((parsed) => {
				return deduplicateDomainList(parsed);
			}).then((deduped) => {
				return validateDomainList(deduped);
			}).then((data) => {
				return { data };
			}).catch((e) => {
				return { error: e.toString() };
			});
		}
	}),
	exportDomainList: build.mutation({
		queryFn: (formData, api, _extraOpts, baseQuery) => {
			return Promise.try(() => {
				return baseQuery({
					url: `/api/v1/admin/domain_blocks`
				});
			}).then(unwrapRes).then((blockedInstances) => {
				return blockedInstances.map((entry) => {
					if (formData.exportType == "json") {
						return {
							domain: entry.domain,
							public_comment: entry.public_comment
						};
					} else {
						return entry.domain;
					}
				});
			}).then((exportList) => {
				if (formData.exportType == "json") {
					return JSON.stringify(exportList);
				} else {
					return exportList.join("\n");
				}
			}).then((exportAsString) => {
				if (formData.action == "export") {
					return {
						data: exportAsString
					};
				} else if (formData.action == "export-file") {
					let domain = new URL(api.getState().oauth.instance).host;
					let date = new Date();
					let mime;

					let filename = [
						domain,
						"blocklist",
						date.getFullYear(),
						(date.getMonth() + 1).toString().padStart(2, "0"),
						date.getDate().toString().padStart(2, "0"),
					].join("-");

					if (formData.exportType == "json") {
						filename += ".json";
						mime = "application/json";
					} else {
						filename += ".txt";
						mime = "text/plain";
					}

					fileDownload(exportAsString, filename, mime);
				}
				return { data: null };
			}).catch((e) => {
				return { error: e };
			});
		}
	}),
	importDomainList: build.mutation({
		query: (formData) => {
			const { domains } = formData;

			// add/replace comments, obfuscation data
			let process = entryProcessor(formData);
			domains.forEach((entry) => {
				process(entry);
			});

			return {
				method: "POST",
				url: `/api/v1/admin/domain_blocks?import=true`,
				asForm: true,
				discardEmpty: true,
				body: {
					domains: new Blob([JSON.stringify(domains)], { type: "application/json" })
				}
			};
		},
		transformResponse: domainListToObject,
		...replaceCacheOnMutation("instanceBlocks")
	})
});

function entryProcessor(formData) {
	let funcs = [];

	["private_comment", "public_comment"].forEach((type) => {
		let text = formData[type].trim();

		if (text.length > 0) {
			let behavior = formData[`${type}_behavior`];

			if (behavior == "append") {
				funcs.push(function appendComment(entry) {
					if (entry[type] == undefined) {
						entry[type] = text;
					} else {
						entry[type] = [entry[type], text].join("\n");
					}
				});
			} else if (behavior == "replace") {
				funcs.push(function replaceComment(entry) {
					entry[type] = text;
				});
			}
		}
	});

	return function process(entry) {
		funcs.forEach((func) => {
			func(entry);
		});

		entry.obfuscate = formData.obfuscate;

		Object.entries(entry).forEach(([key, val]) => {
			if (val == undefined) {
				delete entry[key];
			}
		});
	};
}