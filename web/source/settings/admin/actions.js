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

const query = require("../lib/query");

const { useTextInput } = require("../lib/form");
const { TextInput } = require("../components/form/inputs");

const MutationButton = require("../components/form/mutation-button");

module.exports = function AdminActionPanel() {
	const daysField = useTextInput("days", { defaultValue: 30 });

	const [mediaCleanup, mediaCleanupResult] = query.useMediaCleanupMutation();

	function submitMediaCleanup(e) {
		e.preventDefault();
		mediaCleanup(daysField.value);
	}

	return (
		<>
			<h1>Admin Actions</h1>
			<form onSubmit={submitMediaCleanup}>
				<h2>Media cleanup</h2>
				<p>
					Clean up remote media older than the specified number of days.
					If the remote instance is still online they will be refetched when needed.
					Also cleans up unused headers and avatars from the media cache.
				</p>
				<TextInput
					field={daysField}
					label="Days"
					type="number"
					min="0"
					placeholder="30"
				/>
				<MutationButton label="Remove old media" result={mediaCleanupResult} />
			</form>
		</>
	);
};