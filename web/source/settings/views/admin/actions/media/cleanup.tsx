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

import { useTextInput } from "../../../../lib/form";
import { TextInput } from "../../../../components/form/inputs";
import MutationButton from "../../../../components/form/mutation-button";
import { useMediaCleanupMutation } from "../../../../lib/query/admin/actions";

export default function Cleanup({}) {
	const daysField = useTextInput("days", { defaultValue: "7" });

	const [mediaCleanup, mediaCleanupResult] = useMediaCleanupMutation();

	function submitCleanup(e) {
		e.preventDefault();
		mediaCleanup(daysField.value);
	}
    
	return (
		<form onSubmit={submitCleanup}>
			<div className="form-section-docs">
				<h2>Cleanup</h2>
				<p>
					Clean up remote media older than the specified number of days.
					<br/>
					If the remote instance is still online they will be refetched when needed.
					<br/>
					Also cleans up unused headers and avatars from the media cache.
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/admin/media_caching/"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about media caching + cleanup (opens in a new tab)
				</a>
			</div>
			<TextInput
				field={daysField}
				label="Days"
				type="number"
				min="0"
				placeholder="30"
			/>
			<MutationButton
				disabled={!daysField.value}
				label="Remove old media"
				result={mediaCleanupResult}
			/>
		</form>
	);
}
