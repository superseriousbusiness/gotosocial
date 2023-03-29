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

"use strict";

const React = require("react");
const RoleContext = React.createContext([]);
const BaseUrlContext = React.createContext(null);

function urlSafe(str) {
	return str.toLowerCase().replace(/[\s/]+/g, "-");
}

function useHasPermission(permissions) {
	const roles = React.useContext(RoleContext);
	return checkPermission(permissions, roles);
}

function checkPermission(requiredPermissisons, user) {
	// requiredPermissions can be 'false', in which case there are no restrictions
	if (requiredPermissisons === false) {
		return true;
	}

	// or an array of roles, check if one of the user's roles is sufficient
	return user.some((role) => requiredPermissisons.includes(role));
}

function useBaseUrl() {
	return React.useContext(BaseUrlContext);
}

module.exports = {
	urlSafe, RoleContext, useHasPermission, checkPermission, BaseUrlContext, useBaseUrl
};