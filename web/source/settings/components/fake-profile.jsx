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

module.exports = function FakeProfile({ avatar, header, display_name, username, role }) {
	return ( // Keep in sync with web/template/profile.tmpl
		<div className="profile">
			<div className="header">
				<div className="header-image">
					<img src={header} alt={header ? `header image for ${username}` : "None set"} />
				</div>
				<div className="basic-info" aria-hidden="true">
					<a className="avatar" href={avatar}>
						<img src={avatar} alt={avatar ? `avatar image for ${username}` : "None set"} />
					</a>
					<span className="displayname text-cutoff">
						{display_name.trim().length > 0 ? display_name : username}
						<span className="sr-only">.</span>
					</span>
					<span className="username text-cutoff">@{username}</span>
					{(role && role.name != "user") &&
						<div className={`role ${role.name}`}>
							<span className="sr-only">Role: </span>{role.name}
						</div>
					}
				</div>
			</div>
		</div>
	);
};