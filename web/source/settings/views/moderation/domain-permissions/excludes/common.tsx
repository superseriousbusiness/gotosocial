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

export function DomainPermissionExcludeHelpText() {
	return (
		<>
			Domain permission excludes prevent permissions for a domain (and all
			subdomains) from being automatically managed by domain permission subscriptions.
			<br/>
			For example, if you create an exclude entry for <code>example.org</code>, then
			a blocklist or allowlist subscription will <em>exclude</em> entries for <code>example.org</code>
			and any of its subdomains (<code>sub.example.org</code>, <code>another.sub.example.org</code> etc.)
			when creating domain permission drafts and domain blocks/allows.
			<br/>
			This functionality allows you to manually manage permissions for excluded domains,
			in cases where you know you definitely do or don't want to federate with a given domain,
			no matter what entries are contained in a domain permission subscription.
			<br/>
			Note that by itself, creation of an exclude entry for a given domain does not affect
			federation with that domain at all, it is only useful in combination with permission subscriptions.
		</>
	);
}

export function DomainPermissionExcludeDocsLink() {
	return (
		<a
			href="https://docs.gotosocial.org/en/latest/admin/settings/#domain-permission-excludes"
			target="_blank"
			className="docslink"
			rel="noreferrer"
		>
			Learn more about domain permission excludes (opens in a new tab)
		</a>
	);
}
