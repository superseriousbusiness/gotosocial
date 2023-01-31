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

const isValidDomain = require("is-valid-domain");

module.exports = function isValidDomainBlock(domain) {
	return isValidDomain(domain, {
		/* 
			Wildcard prefix *. can be stripped since it's equivalent to not having it,
			but wildcard anywhere else in the domain is not handled by the backend so it's invalid.
		*/
		wildcard: false,
		allowUnicode: true
	});
};