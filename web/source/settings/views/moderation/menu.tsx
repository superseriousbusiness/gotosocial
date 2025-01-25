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
 * - /settings/moderation/reports/overview
 * - /settings/moderation/reports/:reportId
 * - /settings/moderation/accounts/search
 * - /settings/moderation/accounts/pending
 * - /settings/moderation/accounts/:accountID
 * - /settings/moderation/domain-permissions/:permType
 * - /settings/moderation/domain-permissions/:permType/:domain
 * - /settings/moderation/domain-permissions/import-export
 * - /settings/moderation/domain-permissions/process
 */
export default function ModerationMenu() {
	const permissions = ["moderator"];
	const moderator = useHasPermission(permissions);
	if (!moderator) {
		return null;
	}
	
	return (
		<MenuItem
			name="Moderation"
			itemUrl="moderation"
			defaultChild="reports"
			permissions={permissions}
		>
			<ModerationReportsMenu />
			<ModerationAccountsMenu />
			<ModerationDomainPermsMenu />
		</MenuItem>
	);
}

/*
	INTERNAL COMPONENTS
*/

function ModerationReportsMenu() {
	return (
		<MenuItem
			name="Reports"
			itemUrl="reports"
			icon="fa-flag"
		/>
	);
}

function ModerationAccountsMenu() {
	return (
		<MenuItem
			name="Accounts"
			itemUrl="accounts"
			defaultChild="search"
			icon="fa-users"
		>
			<MenuItem
				name="Search"
				itemUrl="search"
				icon="fa-list"
			/>
			<MenuItem
				name="Pending"
				itemUrl="pending"
				icon="fa-question"
			/>
		</MenuItem>
	);
}

function ModerationDomainPermsMenu() {
	return (
		<MenuItem
			name="Domain Permissions"
			itemUrl="domain-permissions"
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
			<MenuItem
				name="Import/Export"
				itemUrl="import-export"
				icon="fa-floppy-o"
			/>
			<MenuItem
				name="Drafts"
				itemUrl="drafts"
				defaultChild="search"
				icon="fa-pencil"
			>
				<MenuItem
					name="Search"
					itemUrl="search"
					icon="fa-list"
				/>
				<MenuItem
					name="New draft"
					itemUrl="new"
					icon="fa-plus"
				/>
			</MenuItem>
			<MenuItem
				name="Excludes"
				itemUrl="excludes"
				defaultChild="search"
				icon="fa-minus-square"
			>
				<MenuItem
					name="Search"
					itemUrl="search"
					icon="fa-list"
				/>
				<MenuItem
					name="New exclude"
					itemUrl="new"
					icon="fa-plus"
				/>
			</MenuItem>
			<MenuItem
				name="Subscriptions"
				itemUrl="subscriptions"
				defaultChild="search"
				icon="fa-cloud-download"
			>
				<MenuItem
					name="Search"
					itemUrl="search"
					icon="fa-list"
				/>
				<MenuItem
					name="New subscription"
					itemUrl="new"
					icon="fa-plus"
				/>
				<MenuItem
					name="Preview"
					itemUrl="preview"
					icon="fa-eye"
				/>
			</MenuItem>
		</MenuItem>
	);
}
