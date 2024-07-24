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

export default function FakeProfile({ avatar, header, display_name, bot, username, role }) {
	return ( // Keep in sync with web/template/profile.tmpl
		<div className="profile">
			<div className="profile-header">
				<div className="header-image-wrapper">
					<img src={header} alt={header ? `header image for ${username}` : "None set"} />
				</div>
				<div className="basic-info" aria-hidden="true">
					<div className="avatar-image-wrapper">
						<a href={avatar}>
							<img className="avatar" src={avatar} alt={avatar ? `avatar image for ${username}` : "None set"} />
						</a>
					</div>
					<dl className="namerole">
						<dt className="sr-only">Display name</dt>
						<dd className="displayname text-cutoff">{display_name.trim().length > 0 ? display_name : username}</dd>
						<div className="bot-username-wrapper">
							{ bot && <>
								<dt className="sr-only">Bot account</dt>
								<dd>
									<span className="sr-only">true</span>
									<div
										className="bot-legend-wrapper"
										aria-hidden="true"
										title="This is a bot account."
									>
										<i className="bot-icon fa fa-microchip"></i>
										<span className="bot-legend">bot</span>
									</div>
								</dd>
							</>}
							<dt className="sr-only">Username</dt>
							<dd className="username text-cutoff">@{username}</dd>
						</div>
						<dt className="sr-only">Role</dt>
						{
							(role && role.name != "user") ?
								<>
									<dd className="sr-only">Role</dd>
									<dt className={`role ${role.name}`}>{role.name}</dt>
								</>
								: null
						}
					</dl>
				</div>
			</div>
		</div>
	);
}
