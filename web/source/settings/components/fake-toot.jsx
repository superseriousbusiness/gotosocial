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
const Redux = require("react-redux");

module.exports = function FakeToot({children}) {
	const account = Redux.useSelector((state) => state.user.profile);

	return (
		<div className="toot expanded">
			<div className="contentgrid">
				<span className="avatar">
					<img src={account.avatar} alt=""/>
				</span>
				<span className="displayname">{account.display_name.trim().length > 0 ? account.display_name : account.username}</span>
				<span className="username">@{account.username}</span>
				<div className="text">
					<div className="content">
						{children}
					</div>
				</div>
			</div>
		</div>
	);
};