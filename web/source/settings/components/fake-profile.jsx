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

module.exports = function FakeProfile({ avatar, header, display_name, username, role }) {
	return ( // Keep in sync with web/template/profile.tmpl
		<div className="profile">
			<div className="headerimage">
				<img className="headerpreview" src={header} alt={header ? `header image for ${username}` : "None set"} />
			</div>
			<div className="basic">
				<div id="profile-basic-filler2"></div>
				<span className="avatar"><img className="avatarpreview" src={avatar} alt={avatar ? `avatar image for ${username}` : "None set"} /></span>
				<div className="displayname">{display_name.trim().length > 0 ? display_name : username}</div>
				<div className="usernamecontainer">
					<div className="username"><span>@{username}</span></div>
					{(role && role != "user") &&
						<div className={`role ${role}`}>{role}</div>
					}
				</div>
			</div>
		</div>
	);
};