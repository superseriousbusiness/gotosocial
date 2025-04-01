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

import { MenuItem } from "../../lib/navigation/menu";
import React from "react";

/**
 * - /settings/user/profile
 * - /settings/user/posts
 * - /settings/user/emailpassword
 * - /settings/user/migration
 */
export default function UserMenu() {	
	return (
		<MenuItem
			name="User"
			itemUrl="user"
			defaultChild="profile"
		>
			<MenuItem
				name="Profile"
				itemUrl="profile"
				icon="fa-user"
			/>
			<MenuItem
				name="Account"
				itemUrl="account"
				icon="fa-user-secret"
			/>
			<MenuItem
				name="Posts"
				itemUrl="posts"
				icon="fa-paper-plane"
			/>
			<MenuItem
				name="Interaction Requests"
				itemUrl="interaction_requests"
				icon="fa-commenting-o"
			/>
			<MenuItem
				name="Migration"
				itemUrl="migration"
				icon="fa-exchange"
			/>
			<MenuItem
				name="Export & Import"
				itemUrl="export-import"
				icon="fa-floppy-o"
			/>
			<MenuItem
				name="Access Tokens"
				itemUrl="tokens"
				icon="fa-certificate"
			/>
			<MenuItem
				name="Applications"
				itemUrl="applications"
				defaultChild="search"
				icon="fa-plug"
			>
				<MenuItem
					name="Search"
					itemUrl="search"
					icon="fa-list"
				/>
				<MenuItem
					name="New Application"
					itemUrl="new"
					icon="fa-plus"
				/>
			</MenuItem>
		</MenuItem>
	);
}
