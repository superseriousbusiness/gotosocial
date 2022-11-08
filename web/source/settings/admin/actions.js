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

const Submit = require("../components/submit");

const api = require("../lib/api");
const submit = require("../lib/submit");

module.exports = function AdminActionPanel() {
	const dispatch = Redux.useDispatch();

	const [days, setDays] = React.useState(30);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const removeMedia = submit(
		() => dispatch(api.admin.mediaCleanup(days)),
		{setStatus, setError}
	);

	return (
		<>
			<h1>Admin Actions</h1>
			<div>
				<h2>Media cleanup</h2>
				<p>
					Clean up remote media older than the specified number of days.
					If the remote instance is still online they will be refetched when needed.
					Also cleans up unused headers and avatars from the media cache.
				</p>
				<div>
					<label htmlFor="days">Days: </label>
					<input id="days" type="number" value={days} onChange={(e) => setDays(e.target.value)}/>
				</div>
				<Submit onClick={removeMedia} label="Remove media" errorMsg={errorMsg} statusMsg={statusMsg} />
			</div>
		</>
	);
};