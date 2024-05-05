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
import { useHasPermission } from "../../lib/navigation/util";

/*
	EXPORTED COMPONENTS
*/

/**
 * - /settings/admin/instance/settings
 * - /settings/admin/instance/rules
 * - /settings/admin/instance/rules/:ruleId
 * - /settings/admin/emojis
 * - /settings/admin/emojis/local
 * - /settings/admin/emojis/local/:emojiId
 * - /settings/admin/emojis/remote
 * - /settings/admin/actions
 * - /settings/admin/actions/media
 * - /settings/admin/actions/keys
 * - /settings/admin/http-header-permissions/blocks
 * - /settings/admin/http-header-permissions/blocks/:blockId\
 * - /settings/admin/http-header-permissions/allows
 * - /settings/admin/http-header-permissions/allows/:allowId
 */
export default function AdminMenu() {	
	const permissions = ["admin"];
	const admin = useHasPermission(permissions);
	if (!admin) {
		return null;
	}
	
	return (
		<MenuItem
			name="Administration"
			itemUrl="admin"
			defaultChild="actions"
			permissions={permissions}
		>
			<AdminInstanceMenu />
			<AdminEmojisMenu />
			<AdminActionsMenu />
			<AdminHTTPHeaderPermissionsMenu />
		</MenuItem>
	);
}

/*
	INTERNAL COMPONENTS
*/

function AdminInstanceMenu() {
	return (
		<MenuItem
			name="Instance"
			itemUrl="instance"
			defaultChild="settings"
			icon="fa-sitemap"
		>
			<MenuItem
				name="Settings"
				itemUrl="settings"
				icon="fa-sliders"
			/>
			<MenuItem
				name="Rules"
				itemUrl="rules"
				icon="fa-dot-circle-o"
			/>
		</MenuItem>
	);
}

function AdminActionsMenu() {
	return (
		<MenuItem
			name="Actions"
			itemUrl="actions"
			defaultChild="media"
			icon="fa-bolt"
		>
			<MenuItem
				name="Media"
				itemUrl="media"
				icon="fa-photo"
			/>
			<MenuItem
				name="Keys"
				itemUrl="keys"
				icon="fa-key-modern"
			/>
		</MenuItem>
	);
}

function AdminEmojisMenu() {
	return (
		<MenuItem
			name="Custom Emoji"
			itemUrl="emojis"
			defaultChild="local"
			icon="fa-smile-o"
		>
			<MenuItem
				name="Local"
				itemUrl="local"
				icon="fa-home"
			/>
			<MenuItem
				name="Remote"
				itemUrl="remote"
				icon="fa-cloud"
			/>
		</MenuItem>
	);
}

function AdminHTTPHeaderPermissionsMenu() {
	return (
		<MenuItem
			name="HTTP Header Permissions"
			itemUrl="http-header-permissions"
			defaultChild="blocks"
			icon="fa-hubzilla"
		>
			<MenuItem
				name="Blocks"
				itemUrl="blocks"
				icon="fa-close"
			/>
			<MenuItem
				name="Allows"
				itemUrl="allows"
				icon="fa-check"
			/>
		</MenuItem>
	);
}
