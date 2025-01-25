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

import React from "react";

export function DomainPermissionDraftHelpText() {
	return (
		<>
			Domain permission drafts are domain block or domain allow entries that are not yet in force.
			<br/>
			You can choose to accept or remove a draft.
		</>
	);
}

export function DomainPermissionDraftDocsLink() {
	return (
		<a
			href="https://docs.gotosocial.org/en/latest/admin/settings/#domain-permission-drafts"
			target="_blank"
			className="docslink"
			rel="noreferrer"
		>
			Learn more about domain permission drafts (opens in a new tab)
		</a>
	);
}
