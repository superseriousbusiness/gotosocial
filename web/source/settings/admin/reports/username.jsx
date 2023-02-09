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

const React = require("react");

module.exports = function Username({ user, link = true }) {
	let className = "user";
	let isLocal = user.domain == null;

	if (user.suspended) {
		className += " suspended";
	}

	if (isLocal) {
		className += " local";
	}

	let icon = isLocal
		? { fa: "fa-home", info: "Local user" }
		: { fa: "fa-external-link-square", info: "Remote user" };

	let Element = "div";
	let href = null;

	if (link) {
		Element = "a";
		href = user.account.url;
	}

	return (
		<Element className={className} href={href} target="_blank" rel="noreferrer" >
			<span className="acct">@{user.account.acct}</span>
			<i className={`fa fa-fw ${icon.fa}`} aria-hidden="true" title={icon.info} />
			<span className="sr-only">{icon.info}</span>
		</Element>
	);
};