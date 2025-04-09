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
import {
	useExportFollowingMutation,
	useExportFollowersMutation,
	useExportListsMutation,
	useExportBlocksMutation,
	useExportMutesMutation,
} from "../../../lib/query/user/export-import";
import MutationButton from "../../../components/form/mutation-button";
import useFormSubmit from "../../../lib/form/submit";
import { useValue } from "../../../lib/form";
import { AccountExportStats } from "../../../lib/types/account";

export default function Export({ exportStats }: { exportStats: AccountExportStats }) {
	const [exportFollowing, exportFollowingResult] = useFormSubmit(
		// Use a dummy value.
		{ type: useValue("exportFollowing", "exportFollowing") },
		// Mutation we're wrapping.
		useExportFollowingMutation(),
		// Form never changes but
		// we want to always trigger.
		{ changedOnly: false },
	);

	const [exportFollowers, exportFollowersResult] = useFormSubmit(
		// Use a dummy value.
		{ type: useValue("exportFollowers", "exportFollowers") },
		// Mutation we're wrapping.
		useExportFollowersMutation(),
		// Form never changes but
		// we want to always trigger.
		{ changedOnly: false },
	);

	const [exportLists, exportListsResult] = useFormSubmit(
		// Use a dummy value.
		{ type: useValue("exportLists", "exportLists") },
		// Mutation we're wrapping.
		useExportListsMutation(),
		// Form never changes but
		// we want to always trigger.
		{ changedOnly: false },
	);


	const [exportBlocks, exportBlocksResult] = useFormSubmit(
		// Use a dummy value.
		{ type: useValue("exportBlocks", "exportBlocks") },
		// Mutation we're wrapping.
		useExportBlocksMutation(),
		// Form never changes but
		// we want to always trigger.
		{ changedOnly: false },
	);

	const [exportMutes, exportMutesResult] = useFormSubmit(
		// Use a dummy value.
		{ type: useValue("exportMutes", "exportMutes") },
		// Mutation we're wrapping.
		useExportMutesMutation(),
		// Form never changes but
		// we want to always trigger.
		{ changedOnly: false },
	);
	
	return (
		<form className="export-data">
			<div className="form-section-docs">
				<h3>Export Data</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#export"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
				Learn more about this section (opens in a new tab)
				</a>
			</div>
			
			<div className="export-buttons-wrapper">
				<div className="stats-and-button">
					<span className="text-cutoff">
						Following {exportStats.following_count} account{ exportStats.following_count !== 1 && "s" }
					</span>
					<MutationButton
						label="Download following.csv"
						type="button"
						onClick={() => exportFollowing()}
						result={exportFollowingResult}
						showError={true}
						disabled={exportStats.following_count === 0}
					/>
				</div>
				<div className="stats-and-button">
					<span className="text-cutoff">
						Followed by {exportStats.followers_count} account{ exportStats.followers_count !== 1 && "s" }
					</span>
					<MutationButton
						label="Download followers.csv"
						type="button"
						onClick={() => exportFollowers()}
						result={exportFollowersResult}
						showError={true}
						disabled={exportStats.followers_count === 0}
					/>
				</div>
				<div className="stats-and-button">
					<span className="text-cutoff">
						Created {exportStats.lists_count} list{ exportStats.lists_count !== 1 && "s" }
					</span>
					<MutationButton
						label="Download lists.csv"
						type="button"
						onClick={() => exportLists()}
						result={exportListsResult}
						showError={true}
						disabled={exportStats.lists_count === 0}
					/>
				</div>
				<div className="stats-and-button">
					<span className="text-cutoff">
						Blocking {exportStats.blocks_count} account{ exportStats.blocks_count !== 1 && "s" }
					</span>
					<MutationButton
						label="Download blocks.csv"
						type="button"
						onClick={() => exportBlocks()}
						result={exportBlocksResult}
						showError={true}
						disabled={exportStats.blocks_count === 0}
					/>
				</div>
				<div className="stats-and-button">
					<span className="text-cutoff">
						Muting {exportStats.mutes_count} account{ exportStats.mutes_count !== 1 && "s" }
					</span>
					<MutationButton
						label="Download mutes.csv"
						type="button"
						onClick={() => exportMutes()}
						result={exportMutesResult}
						showError={true}
						disabled={exportStats.mutes_count === 0}
					/>
				</div>
			</div>
		</form>
	);
}
