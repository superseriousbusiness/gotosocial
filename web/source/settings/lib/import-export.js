/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

function parseDomainList(list) {
	if (list[0] == "[") {
		return JSON.parse(list);
	} else {
		return list.split("\n").map((line) => {
			let trimmed = line.trim();
			return trimmed.length > 0
				? { domain: trimmed }
				: null;
		}).filter((a) => a); // not `null`
	}
}

function validateDomainList(list) {
	list.forEach((entry) => {
		entry.valid = isValidDomain(entry.domain, { wildcard: true, allowUnicode: true });
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

module.exports = function processDomainList(data) {
	return Promise.try(() => {
		return parseDomainList(data);
	}).then((parsed) => {
		return deduplicateDomainList(parsed);
	}).then((deduped) => {
		return validateDomainList(deduped);
	});
};