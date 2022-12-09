/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
const Redux = require("react-redux");

module.exports = function FakeProfile({}) {
	const account = Redux.useSelector(state => state.user.profile);

	return ( // Keep in sync with web/template/profile.tmpl
		<div className="profile">
			<div className="headerimage">
				<img className="headerpreview" src={account.header} alt={account.header ? `header image for ${account.username}` : "None set"} />
			</div>
			<div className="basic">
				<div id="profile-basic-filler2"></div>
				<span className="avatar"><img className="avatarpreview" src={account.avatar} alt={account.avatar ? `avatar image for ${account.username}` : "None set"} /></span>
				<div className="displayname">{account.display_name.trim().length > 0 ? account.display_name : account.username}</div>
				<div className="usernamecontainer">
					<div className="username"><span>@{account.username}</span></div>
					{(account.role && account.role != "user") && 
						<div className={`role ${account.role}`}>{account.role}</div>
					}
				</div>
			</div>
		</div>
	);
};