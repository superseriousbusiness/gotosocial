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

import { createContext, useContext } from "react";
const RoleContext = createContext<string[]>([]);
const BaseUrlContext = createContext<string>("");
const MenuLevelContext = createContext<number>(0);

function urlSafe(str: string) {
	return str.toLowerCase().replace(/[\s/]+/g, "-");
}

function useHasPermission(permissions: string[] | undefined) {
	const roles = useContext<string[]>(RoleContext);
	return checkPermission(permissions, roles);
}

// checkPermission returns true if the user's roles
// include requiredPermissions, or false otherwise.
function checkPermission(requiredPermissions: string[] | undefined, userRoles: string[]): boolean {
	if (requiredPermissions === undefined) {
		// No perms defined, so user
		// implicitly has permission.
		return true;
	}

	if (requiredPermissions.length === 0) {
		// No perms defined, so user
		// implicitly has permission.
		return true;
	}

	// Check if one of the user's
	// roles is sufficient.
	return userRoles.some((role) => {
		if (role === "admin") {
			// Admins can
			// see everything.
			return true;
		}

		return requiredPermissions.includes(role);
	});
}

function useBaseUrl() {
	return useContext(BaseUrlContext);
}

function useMenuLevel() {
	return useContext(MenuLevelContext);
}

export {
	urlSafe,
	RoleContext,
	useHasPermission,
	checkPermission,
	BaseUrlContext,
	useBaseUrl,
	MenuLevelContext,
	useMenuLevel,
};
