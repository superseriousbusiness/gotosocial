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
			<div class="header">
				<div class="header-image">
					<img src={header} alt={header ? `header image for ${username}` : "None set"} />
				</div>
				<div class="basic-info" aria-hidden="true">
					<a class="avatar" href="{{.account.Avatar}}">
						<img src={avatar} alt={avatar ? `avatar image for ${username}` : "None set"} />
					</a>
					<span class="displayname text-cutoff">
						{display_name.trim().length > 0 ? display_name : username}
						<span class="sr-only">.</span>
					</span>
					<span class="username text-cutoff">@{username}</span>
					{(role && role != "user") &&
						<div className={`role ${role}`}>
							<span class="sr-only">Role: </span>{role}
						</div>
					}
				</div>
			</div>
		</div>
	);
};